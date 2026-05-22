//go:build integration

package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

func TestNotifyTaskStatusChanged_NotifiesRelevantMembers(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)
		notifier := newTestNotifier(adapter)

		user1, err := adapter.User().CreateUser(ctx, &models.User{Email: "creator@example.com"})
		require.NoError(t, err)
		creator, err := adapter.TeamMember().CreateTeamMemberFromUserAndSlug(ctx, user1, "team", models.TeamMemberRoleOwner)
		require.NoError(t, err)

		user2, err := adapter.User().CreateUser(ctx, &models.User{Email: "assignee@example.com"})
		require.NoError(t, err)
		assignee, err := adapter.TeamMember().CreateTeamMember(ctx, &models.TeamMember{
			TeamID: creator.TeamID,
			UserID: &user2.ID,
			Role:   models.TeamMemberRoleMember,
			Active: true,
		})
		require.NoError(t, err)

		user3, err := adapter.User().CreateUser(ctx, &models.User{Email: "reporter@example.com"})
		require.NoError(t, err)
		reporter, err := adapter.TeamMember().CreateTeamMember(ctx, &models.TeamMember{
			TeamID: creator.TeamID,
			UserID: &user3.ID,
			Role:   models.TeamMemberRoleMember,
			Active: true,
		})
		require.NoError(t, err)

		project, err := adapter.Task().CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name:     "proj",
			Status:   models.TaskProjectStatusTodo,
			TeamID:   creator.TeamID,
			MemberID: creator.ID,
		})
		require.NoError(t, err)

		pastDue := time.Now().Add(24 * time.Hour)
		task, err := adapter.Task().CreateTask(ctx, &models.Task{
			Name:              "status change task",
			Status:            models.TaskStatusTodo,
			TeamID:            creator.TeamID,
			ProjectID:         project.ID,
			CreatedByMemberID: &creator.ID,
			AssigneeID:        &assignee.ID,
			ReporterID:        &reporter.ID,
			EndAt:             &pastDue,
		})
		require.NoError(t, err)

		err = notifier.NotifyTaskStatusChanged(ctx, task.ID, string(models.TaskStatusTodo), string(models.TaskStatusInProgress), creator.ID)
		require.NoError(t, err)

		for _, memberID := range []uuid.UUID{creator.ID, assignee.ID, reporter.ID} {
			count, err := adapter.Notification().CountNotification(ctx, &stores.NotificationFilter{
				TeamMemberIds: []uuid.UUID{memberID},
				Types:         []string{"task_status_changed"},
			})
			require.NoError(t, err)
			assert.Equal(t, int64(1), count, "member %s should have 1 status_changed notification", memberID)
		}
	})
}

func TestNotifyTaskStatusChanged_NoMembersNoError(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)
		notifier := newTestNotifier(adapter)

		user, err := adapter.User().CreateUser(ctx, &models.User{Email: "solo@example.com"})
		require.NoError(t, err)
		member, err := adapter.TeamMember().CreateTeamMemberFromUserAndSlug(ctx, user, "team", models.TeamMemberRoleOwner)
		require.NoError(t, err)

		project, err := adapter.Task().CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name:     "proj",
			Status:   models.TaskProjectStatusTodo,
			TeamID:   member.TeamID,
			MemberID: member.ID,
		})
		require.NoError(t, err)

		// task with no assignee, reporter, or creator — no notifications expected
		task, err := adapter.Task().CreateTask(ctx, &models.Task{
			Name:      "orphan task",
			Status:    models.TaskStatusTodo,
			TeamID:    member.TeamID,
			ProjectID: project.ID,
		})
		require.NoError(t, err)

		err = notifier.NotifyTaskStatusChanged(ctx, task.ID, string(models.TaskStatusTodo), string(models.TaskStatusInProgress), member.ID)
		require.NoError(t, err)

		count, err := adapter.Notification().CountNotification(ctx, &stores.NotificationFilter{
			TeamMemberIds: []uuid.UUID{member.ID},
			Types:         []string{"task_status_changed"},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}
