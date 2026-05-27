package services_test

// Tests for the task comment notification pipeline:
//
//   NotifyTaskCommentCreated  →  notifies assignee/reporter (not the author)
//   NotifyTaskCommentMention  →  notifies mentioned member (not the author)
//   TaskCommentCreatedWorker  →  routes args correctly to notifier
//   TaskCommentMentionWorker  →  routes args correctly to notifier

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
	"github.com/tkahng/playground/internal/workers"
)

func setupCommentNotifyFixture(t *testing.T, ctx context.Context, db database.Dbx) (
	adapter stores.StorageAdapterInterface,
	author *models.TeamMember,
	assignee *models.TeamMember,
	task *models.Task,
	comment *models.TaskComment,
) {
	t.Helper()
	adapter = stores.NewStorageAdapter(db)

	authorUser, err := adapter.User().CreateUser(ctx, &models.User{Email: "comment-author@example.com"})
	require.NoError(t, err)
	author, err = adapter.TeamMember().CreateTeamMemberFromUserAndSlug(ctx, authorUser, "comment-team", models.TeamMemberRoleOwner)
	require.NoError(t, err)

	assigneeUser, err := adapter.User().CreateUser(ctx, &models.User{Email: "comment-assignee@example.com"})
	require.NoError(t, err)
	assigneeMember, err := adapter.TeamMember().CreateTeamMemberFromUserAndSlug(ctx, assigneeUser, "", models.TeamMemberRoleMember)
	require.NoError(t, err)
	assigneeMember.TeamID = author.TeamID
	assignee, err = adapter.TeamMember().UpdateTeamMember(ctx, assigneeMember)
	require.NoError(t, err)

	project, err := adapter.Task().CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
		Name:     "comment-project",
		Status:   models.TaskProjectStatusTodo,
		TeamID:   author.TeamID,
		MemberID: author.ID,
	})
	require.NoError(t, err)

	task, err = adapter.Task().CreateTask(ctx, &models.Task{
		Name:       "comment-task",
		Status:     models.TaskStatusTodo,
		TeamID:     author.TeamID,
		ProjectID:  project.ID,
		AssigneeID: &assignee.ID,
	})
	require.NoError(t, err)

	comment, err = adapter.TaskComment().CreateTaskComment(ctx, &models.TaskComment{
		TaskID:            task.ID,
		CreatedByMemberID: author.ID,
		Content:           "this is a test comment",
	})
	require.NoError(t, err)
	return
}

// TestNotifyTaskCommentCreated_SkipsAuthor ensures the author does not receive
// a notification for their own comment.
func TestNotifyTaskCommentCreated_SkipsAuthor(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter, author, assignee, task, comment := setupCommentNotifyFixture(t, ctx, db)

		notifier := services.NewDbNotificationPublisher(noopSSE{}, services.NewTeamService(adapter), adapter)
		require.NoError(t, notifier.NotifyTaskCommentCreated(ctx, task.ID, comment.ID, author.ID, nil))

		authorNotifs, err := adapter.Notification().FindNotifications(ctx, &stores.NotificationFilter{
			TeamMemberIds: []uuid.UUID{author.ID},
			Types:         []string{"task_comment_created"},
		})
		require.NoError(t, err)
		assert.Empty(t, authorNotifs, "author should not receive a notification for their own comment")

		assigneeNotifs, err := adapter.Notification().FindNotifications(ctx, &stores.NotificationFilter{
			TeamMemberIds: []uuid.UUID{assignee.ID},
			Types:         []string{"task_comment_created"},
		})
		require.NoError(t, err)
		assert.Len(t, assigneeNotifs, 1, "assignee should receive exactly one notification")
	})
}

// TestNotifyTaskCommentCreated_PayloadFields verifies the stored payload has
// the expected task_id, comment_id, author_id, task_name, and excerpt fields.
func TestNotifyTaskCommentCreated_PayloadFields(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter, author, assignee, task, comment := setupCommentNotifyFixture(t, ctx, db)

		notifier := services.NewDbNotificationPublisher(noopSSE{}, services.NewTeamService(adapter), adapter)
		require.NoError(t, notifier.NotifyTaskCommentCreated(ctx, task.ID, comment.ID, author.ID, nil))

		notif, err := adapter.Notification().FindNotification(ctx, &stores.NotificationFilter{
			TeamMemberIds: []uuid.UUID{assignee.ID},
			Types:         []string{"task_comment_created"},
		})
		require.NoError(t, err)
		require.NotNil(t, notif)

		var payload notification.NotificationPayload[notification.TaskCommentCreatedNotificationData]
		require.NoError(t, json.Unmarshal(notif.Payload, &payload))

		assert.Equal(t, task.ID, payload.Data.TaskID)
		assert.Equal(t, comment.ID, payload.Data.CommentID)
		assert.Equal(t, author.ID, payload.Data.AuthorID)
		assert.Equal(t, "comment-task", payload.Data.TaskName)
		assert.Equal(t, "this is a test comment", payload.Data.Excerpt)
	})
}

// TestNotifyTaskCommentMention_SendsToMentioned verifies that the mentioned
// member receives a notification and the author does not.
func TestNotifyTaskCommentMention_SendsToMentioned(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter, author, mentioned, task, comment := setupCommentNotifyFixture(t, ctx, db)

		notifier := services.NewDbNotificationPublisher(noopSSE{}, services.NewTeamService(adapter), adapter)
		require.NoError(t, notifier.NotifyTaskCommentMention(ctx, task.ID, comment.ID, author.ID, mentioned.ID))

		authorNotifs, err := adapter.Notification().FindNotifications(ctx, &stores.NotificationFilter{
			TeamMemberIds: []uuid.UUID{author.ID},
			Types:         []string{"task_comment_mention"},
		})
		require.NoError(t, err)
		assert.Empty(t, authorNotifs)

		mentionedNotifs, err := adapter.Notification().FindNotifications(ctx, &stores.NotificationFilter{
			TeamMemberIds: []uuid.UUID{mentioned.ID},
			Types:         []string{"task_comment_mention"},
		})
		require.NoError(t, err)
		assert.Len(t, mentionedNotifs, 1)
	})
}

// TestNotifyTaskCommentMention_SelfMentionNoOp verifies that self-mentions
// produce no notification.
func TestNotifyTaskCommentMention_SelfMentionNoOp(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter, author, _, task, comment := setupCommentNotifyFixture(t, ctx, db)

		notifier := services.NewDbNotificationPublisher(noopSSE{}, services.NewTeamService(adapter), adapter)
		require.NoError(t, notifier.NotifyTaskCommentMention(ctx, task.ID, comment.ID, author.ID, author.ID))

		notifs, err := adapter.Notification().FindNotifications(ctx, &stores.NotificationFilter{
			TeamMemberIds: []uuid.UUID{author.ID},
			Types:         []string{"task_comment_mention"},
		})
		require.NoError(t, err)
		assert.Empty(t, notifs, "self-mention should produce no notification")
	})
}

// TestTaskCommentCreatedWorker_RoutesArgs verifies that the worker passes args
// to NotifyTaskCommentCreated in the correct positions.
func TestTaskCommentCreatedWorker_RoutesArgs(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter, author, assignee, task, comment := setupCommentNotifyFixture(t, ctx, db)

		notifier := services.NewDbNotificationPublisher(noopSSE{}, services.NewTeamService(adapter), adapter)
		worker := services.NewTaskCommentCreatedWorker(notifier)

		err := worker.Work(ctx, &jobs.Job[workers.TaskCommentCreatedJobArgs]{
			Args: workers.TaskCommentCreatedJobArgs{
				TaskID:    task.ID,
				CommentID: comment.ID,
				AuthorID:  author.ID,
			},
		})
		require.NoError(t, err)

		notif, err := adapter.Notification().FindNotification(ctx, &stores.NotificationFilter{
			TeamMemberIds: []uuid.UUID{assignee.ID},
			Types:         []string{"task_comment_created"},
		})
		require.NoError(t, err)
		require.NotNil(t, notif, "assignee should receive a notification via the worker")

		var payload notification.NotificationPayload[notification.TaskCommentCreatedNotificationData]
		require.NoError(t, json.Unmarshal(notif.Payload, &payload))
		assert.Equal(t, author.ID, payload.Data.AuthorID)
		assert.Equal(t, comment.ID, payload.Data.CommentID)
	})
}

// TestTaskCommentMentionWorker_RoutesArgs verifies that the worker passes args
// to NotifyTaskCommentMention in the correct positions.
func TestTaskCommentMentionWorker_RoutesArgs(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter, author, mentioned, task, comment := setupCommentNotifyFixture(t, ctx, db)

		notifier := services.NewDbNotificationPublisher(noopSSE{}, services.NewTeamService(adapter), adapter)
		worker := services.NewTaskCommentMentionWorker(notifier)

		err := worker.Work(ctx, &jobs.Job[workers.TaskCommentMentionJobArgs]{
			Args: workers.TaskCommentMentionJobArgs{
				TaskID:      task.ID,
				CommentID:   comment.ID,
				AuthorID:    author.ID,
				MentionedID: mentioned.ID,
			},
		})
		require.NoError(t, err)

		notif, err := adapter.Notification().FindNotification(ctx, &stores.NotificationFilter{
			TeamMemberIds: []uuid.UUID{mentioned.ID},
			Types:         []string{"task_comment_mention"},
		})
		require.NoError(t, err)
		require.NotNil(t, notif, "mentioned member should receive a notification via the worker")

		var payload notification.NotificationPayload[notification.TaskCommentMentionNotificationData]
		require.NoError(t, json.Unmarshal(notif.Payload, &payload))
		assert.Equal(t, mentioned.ID, payload.Data.MentionedID)
		assert.Equal(t, author.ID, payload.Data.AuthorID)
	})
}
