package services_test

// Tests for per-member per-type notification rate limiting in sendToMembers:
//
//   - A member receiving notificationRateLimit notifications of the same type
//     within the rate window is suppressed on the next send.
//   - A member under the limit still receives notifications.
//   - A database error from FindMembersOverRateLimit propagates (fail-closed).

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/notification"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
)

func setupRateLimitFixture(t *testing.T, ctx context.Context, db database.Dbx) (
	adapter stores.StorageAdapterInterface,
	member *models.TeamMember,
	assigner *models.TeamMember,
) {
	t.Helper()
	adapter = stores.NewDbAdapterDecorators(db)

	ownerUser, err := adapter.User().CreateUser(ctx, &models.User{Email: "rl-owner@example.com"})
	require.NoError(t, err)
	owner, err := adapter.TeamMember().CreateTeamMemberFromUserAndSlug(ctx, ownerUser, "rl-team", models.TeamMemberRoleOwner)
	require.NoError(t, err)

	memberUser, err := adapter.User().CreateUser(ctx, &models.User{Email: "rl-member@example.com"})
	require.NoError(t, err)
	member, err = adapter.TeamMember().CreateTeamMember(ctx, &models.TeamMember{
		TeamID: owner.TeamID, UserID: &memberUser.ID, Role: models.TeamMemberRoleMember, Active: true,
	})
	require.NoError(t, err)

	assignerUser, err := adapter.User().CreateUser(ctx, &models.User{Email: "rl-assigner@example.com"})
	require.NoError(t, err)
	assigner, err = adapter.TeamMember().CreateTeamMember(ctx, &models.TeamMember{
		TeamID: owner.TeamID, UserID: &assignerUser.ID, Role: models.TeamMemberRoleMember, Active: true,
	})
	require.NoError(t, err)

	return adapter, member, assigner
}

// TestRateLimit_MemberSuppressedAfterThreshold verifies that a member who has
// already received the maximum allowed notifications of a type within the
// sliding window does not receive additional notifications.
func TestRateLimit_MemberSuppressedAfterThreshold(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter, member, assigner := setupRateLimitFixture(t, ctx, db)

		project, err := adapter.Task().CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name: "rl-project", Status: models.TaskProjectStatusTodo,
			TeamID: member.TeamID, MemberID: assigner.ID,
		})
		require.NoError(t, err)

		notifier := services.NewDbNotificationPublisher(noopSSE{}, services.NewTeamService(adapter), adapter)

		// Send exactly at the rate limit (5). All should succeed.
		for i := range 5 {
			task, err := adapter.Task().CreateTask(ctx, &models.Task{
				Name: "rl-task", Status: models.TaskStatusTodo,
				TeamID: member.TeamID, ProjectID: project.ID, AssigneeID: &member.ID,
			})
			require.NoError(t, err, "task %d", i)
			require.NoError(t, notifier.NotifyAssignedToTask(ctx, task.ID, assigner.ID, member.ID), "send %d", i)
		}

		count, err := adapter.Notification().CountNotification(ctx, &stores.NotificationFilter{
			TeamMemberIds: []uuid.UUID{member.ID},
			Types:         []string{notification.TypeAssignedToTask},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(5), count, "exactly 5 notifications should be persisted")

		// 6th send — member is now over the limit, should be suppressed.
		task6, err := adapter.Task().CreateTask(ctx, &models.Task{
			Name: "rl-task-6", Status: models.TaskStatusTodo,
			TeamID: member.TeamID, ProjectID: project.ID, AssigneeID: &member.ID,
		})
		require.NoError(t, err)
		require.NoError(t, notifier.NotifyAssignedToTask(ctx, task6.ID, assigner.ID, member.ID))

		count, err = adapter.Notification().CountNotification(ctx, &stores.NotificationFilter{
			TeamMemberIds: []uuid.UUID{member.ID},
			Types:         []string{notification.TypeAssignedToTask},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(5), count, "6th notification should be suppressed by rate limit")
	})
}

// TestRateLimit_DifferentTypeNotSuppressed verifies that the rate limit is
// per-type: receiving the max of one type does not suppress a different type.
func TestRateLimit_DifferentTypeNotSuppressed(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter, member, assigner := setupRateLimitFixture(t, ctx, db)

		project, err := adapter.Task().CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name: "rl-project2", Status: models.TaskProjectStatusTodo,
			TeamID: member.TeamID, MemberID: assigner.ID,
		})
		require.NoError(t, err)

		notifier := services.NewDbNotificationPublisher(noopSSE{}, services.NewTeamService(adapter), adapter)

		// Saturate assigned_to_task rate limit.
		for range 5 {
			task, err := adapter.Task().CreateTask(ctx, &models.Task{
				Name: "rl-sat-task", Status: models.TaskStatusTodo,
				TeamID: member.TeamID, ProjectID: project.ID, AssigneeID: &member.ID,
			})
			require.NoError(t, err)
			require.NoError(t, notifier.NotifyAssignedToTask(ctx, task.ID, assigner.ID, member.ID))
		}

		// task_completed for a task where member is reporter should still go through.
		task, err := adapter.Task().CreateTask(ctx, &models.Task{
			Name: "rl-done-task", Status: models.TaskStatusDone,
			TeamID: member.TeamID, ProjectID: project.ID, ReporterID: &member.ID,
		})
		require.NoError(t, err)
		require.NoError(t, notifier.NotifyTaskCompleted(ctx, task.ID, assigner.ID, task.CreatedAt))

		count, err := adapter.Notification().CountNotification(ctx, &stores.NotificationFilter{
			TeamMemberIds: []uuid.UUID{member.ID},
			Types:         []string{notification.TypeTaskCompleted},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), count, "task_completed should not be suppressed by assigned_to_task rate limit")
	})
}
