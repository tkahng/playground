package services_test

// Tests for the assigned_to_task notification pipeline:
//
//   EnqueueAssignedToTaskJob  →  job stored with correct JSON field names
//   NotifyAssignedToTask      →  notification persisted for assignee with full payload
//   AssignedToTaskWorker.Work →  routes job args to NotifyAssignedToTask correctly

import (
	"context"
	"encoding/json"
	"testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/notification"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/sse"
	"github.com/tkahng/playground/internal/workers"
)

// noopSSE satisfies sse.Manager without sending anything.
type noopSSE struct{}

func (noopSSE) Send(_ string, _ any) error                          { return nil }
func (noopSSE) SendAll(_ any) error                                 { return nil }
func (noopSSE) Clients() []sse.Client                               { return nil }
func (noopSSE) RegisterClient(_ context.Context, _ context.CancelFunc, _ sse.Client) {}
func (noopSSE) UnregisterClient(_ sse.Client)                       {}
func (noopSSE) Run(_ context.Context)                               {}

var _ sse.Manager = noopSSE{}

func setupAssignNotifyFixture(t *testing.T, ctx context.Context, db database.Dbx) (
	adapter stores.StorageAdapterInterface,
	assigner *models.TeamMember,
	assignee *models.TeamMember,
	task *models.Task,
) {
	t.Helper()
	adapter = stores.NewStorageAdapter(db)

	assignerUser, err := adapter.User().CreateUser(ctx, &models.User{Email: "assigner@example.com"})
	require.NoError(t, err)
	assigner, err = adapter.TeamMember().CreateTeamMemberFromUserAndSlug(ctx, assignerUser, "notify-team", models.TeamMemberRoleOwner)
	require.NoError(t, err)

	assigneeUser, err := adapter.User().CreateUser(ctx, &models.User{Email: "assignee@example.com"})
	require.NoError(t, err)
	// Add assignee to the same team as assigner.
	assigneeMember, err := adapter.TeamMember().CreateTeamMemberFromUserAndSlug(ctx, assigneeUser, "", models.TeamMemberRoleMember)
	require.NoError(t, err)
	assigneeMember.TeamID = assigner.TeamID
	assignee, err = adapter.TeamMember().UpdateTeamMember(ctx, assigneeMember)
	require.NoError(t, err)

	project, err := adapter.Task().CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
		Name:     "test-project",
		Status:   models.TaskProjectStatusTodo,
		TeamID:   assigner.TeamID,
		MemberID: assigner.ID,
	})
	require.NoError(t, err)

	task, err = adapter.Task().CreateTask(ctx, &models.Task{
		Name:      "test-task",
		Status:    models.TaskStatusTodo,
		TeamID:    assigner.TeamID,
		ProjectID: project.ID,
	})
	require.NoError(t, err)
	return
}

// TestEnqueueAssignedToTaskJob_FieldNamesRoundtrip verifies that job args are
// stored with the correct JSON keys after the typo rename
// (AssignedByMemeberID → AssignedByMemberID).
func TestEnqueueAssignedToTaskJob_FieldNamesRoundtrip(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		_, assigner, assignee, task := setupAssignNotifyFixture(t, ctx, db)

		jobSvc := services.NewJobService(jobs.NewDbJobManager(db))
		err := jobSvc.EnqueueAssignedToTaskJob(ctx, &workers.AssignedToTaskJobArgs{
			AssignedByMemberID: assigner.ID,
			AssigneeMemberID:   assignee.ID,
			TaskID:             task.ID,
		})
		require.NoError(t, err)

		// Decode the raw payload JSON from the DB to confirm field names.
		var rawArgs map[string]string
		row := db.QueryRow(ctx,
			"SELECT payload FROM app.jobs WHERE kind = $1 LIMIT 1",
			workers.AssignedToTaskJobArgs{}.Kind(),
		)
		require.NoError(t, row.Scan(&rawArgs))

		assert.Equal(t, assigner.ID.String(), rawArgs["assigned_by_member_id"],
			"assigned_by_member_id must match the assigner")
		assert.Equal(t, assignee.ID.String(), rawArgs["assignee_member_id"],
			"assignee_member_id must match the assignee")
		assert.Equal(t, task.ID.String(), rawArgs["task_id"],
			"task_id must match the task")
	})
}

// TestNotifyAssignedToTask_SendsNotificationToAssignee verifies that exactly
// one notification row is persisted for the assignee's member ID.
func TestNotifyAssignedToTask_SendsNotificationToAssignee(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter, assigner, assignee, task := setupAssignNotifyFixture(t, ctx, db)

		notifier := services.NewDbNotificationPublisher(noopSSE{}, services.NewTeamService(adapter), adapter)
		require.NoError(t, notifier.NotifyAssignedToTask(ctx, task.ID, assigner.ID, assignee.ID))

		notifications, err := adapter.Notification().FindNotifications(ctx, &stores.NotificationFilter{
			TeamMemberIds: []uuid.UUID{assignee.ID},
			Types:         []string{"assigned_to_task"},
		})
		require.NoError(t, err)
		assert.Len(t, notifications, 1, "exactly one notification should be created for the assignee")
	})
}

// TestNotifyAssignedToTask_PayloadContainsAllFields verifies that the stored
// payload has assigned_by_member_id, assignee_member_id, and task_id all set —
// catching any field dropped or misnamed during the typo rename.
func TestNotifyAssignedToTask_PayloadContainsAllFields(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter, assigner, assignee, task := setupAssignNotifyFixture(t, ctx, db)

		notifier := services.NewDbNotificationPublisher(noopSSE{}, services.NewTeamService(adapter), adapter)
		require.NoError(t, notifier.NotifyAssignedToTask(ctx, task.ID, assigner.ID, assignee.ID))

		notif, err := adapter.Notification().FindNotification(ctx, &stores.NotificationFilter{
			TeamMemberIds: []uuid.UUID{assignee.ID},
			Types:         []string{"assigned_to_task"},
		})
		require.NoError(t, err)
		require.NotNil(t, notif)

		var payload notification.NotificationPayload[notification.AssignedToTaskNotificationData]
		require.NoError(t, json.Unmarshal(notif.Payload, &payload))

		assert.Equal(t, assigner.ID, payload.Data.AssignedByMemberID,
			"payload.data.assigned_by_member_id must be the assigner")
		assert.Equal(t, assignee.ID, payload.Data.AssigneeMemberID,
			"payload.data.assignee_member_id must be the assignee")
		assert.Equal(t, task.ID, payload.Data.TaskID,
			"payload.data.task_id must be the task")
	})
}

// TestAssignedToTaskWorker_RoutesArgsToNotifier verifies that
// AssignedToTaskWorker.Work reads AssignedByMemberID, AssigneeMemberID, and
// TaskID from the job args and passes them to NotifyAssignedToTask in the right
// positions — catching any field-to-arg positional mismatch.
func TestAssignedToTaskWorker_RoutesArgsToNotifier(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter, assigner, assignee, task := setupAssignNotifyFixture(t, ctx, db)

		notifier := services.NewDbNotificationPublisher(noopSSE{}, services.NewTeamService(adapter), adapter)
		worker := services.NewAssignedToTaskWorker(notifier)

		err := worker.Work(ctx, &jobs.Job[workers.AssignedToTaskJobArgs]{
			Args: workers.AssignedToTaskJobArgs{
				AssignedByMemberID: assigner.ID,
				AssigneeMemberID:   assignee.ID,
				TaskID:             task.ID,
			},
		})
		require.NoError(t, err)

		// If the worker swapped AssignedByMemberID ↔ AssigneeMemberID the
		// notifier would try to look up a non-existent assigner and return an
		// error (caught above), or write a payload with swapped IDs (caught here).
		notif, err := adapter.Notification().FindNotification(ctx, &stores.NotificationFilter{
			TeamMemberIds: []uuid.UUID{assignee.ID},
			Types:         []string{"assigned_to_task"},
		})
		require.NoError(t, err)
		require.NotNil(t, notif, "notification must be created for the assignee")

		var payload notification.NotificationPayload[notification.AssignedToTaskNotificationData]
		require.NoError(t, json.Unmarshal(notif.Payload, &payload))
		assert.Equal(t, assigner.ID, payload.Data.AssignedByMemberID)
		assert.Equal(t, assignee.ID, payload.Data.AssigneeMemberID)
	})
}

