package stores_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
)

func TestFindTasksDueToday(t *testing.T) {
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)
		user := stores.CreateUser(adapter, ctx, "due-today@example.com")
		team := stores.CreateTeam(adapter, ctx, "due-today-team")
		member := stores.CreateTeamMember(adapter, ctx, team, user, models.TeamMemberRoleOwner, true)
		project := stores.CreateTeamProject(adapter, ctx, member, "proj", "proj")

		now := time.Now().UTC()
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		midDay := startOfDay.Add(12 * time.Hour)
		yesterday := startOfDay.Add(-1 * time.Hour)
		tomorrow := startOfDay.Add(25 * time.Hour)

		dueToday := stores.CreateTask(adapter, ctx, &models.Task{
			Name:      "due today",
			Status:    models.TaskStatusTodo,
			TeamID:    team.ID,
			ProjectID: project.ID,
			EndAt:     &midDay,
		})
		dueYesterday := stores.CreateTask(adapter, ctx, &models.Task{
			Name:      "due yesterday",
			Status:    models.TaskStatusTodo,
			TeamID:    team.ID,
			ProjectID: project.ID,
			EndAt:     &yesterday,
		})
		dueTomorrow := stores.CreateTask(adapter, ctx, &models.Task{
			Name:      "due tomorrow",
			Status:    models.TaskStatusTodo,
			TeamID:    team.ID,
			ProjectID: project.ID,
			EndAt:     &tomorrow,
		})
		doneDueToday := stores.CreateTask(adapter, ctx, &models.Task{
			Name:      "done and due today",
			Status:    models.TaskStatusDone,
			TeamID:    team.ID,
			ProjectID: project.ID,
			EndAt:     &midDay,
		})
		_ = stores.CreateTask(adapter, ctx, &models.Task{
			Name:      "no due date",
			Status:    models.TaskStatusTodo,
			TeamID:    team.ID,
			ProjectID: project.ID,
		})

		tasks, err := adapter.Task().FindTasksDueToday(ctx)
		require.NoError(t, err)

		ids := make(map[string]bool)
		for _, tk := range tasks {
			ids[tk.ID.String()] = true
		}

		assert.True(t, ids[dueToday.ID.String()], "due-today task should be included")
		assert.False(t, ids[dueYesterday.ID.String()], "yesterday task should not be included")
		assert.False(t, ids[dueTomorrow.ID.String()], "tomorrow task should not be included")
		assert.False(t, ids[doneDueToday.ID.String()], "done task should not be included even if due today")
	})
}

func TestFindTasksOverdue(t *testing.T) {
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)
		user := stores.CreateUser(adapter, ctx, "overdue@example.com")
		team := stores.CreateTeam(adapter, ctx, "overdue-team")
		member := stores.CreateTeamMember(adapter, ctx, team, user, models.TeamMemberRoleOwner, true)
		project := stores.CreateTeamProject(adapter, ctx, member, "proj", "proj")

		now := time.Now().UTC()
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		yesterday := startOfDay.Add(-2 * time.Hour)
		midDay := startOfDay.Add(12 * time.Hour)

		overdue := stores.CreateTask(adapter, ctx, &models.Task{
			Name:      "overdue",
			Status:    models.TaskStatusInProgress,
			TeamID:    team.ID,
			ProjectID: project.ID,
			EndAt:     &yesterday,
		})
		doneOverdue := stores.CreateTask(adapter, ctx, &models.Task{
			Name:      "done but overdue",
			Status:    models.TaskStatusDone,
			TeamID:    team.ID,
			ProjectID: project.ID,
			EndAt:     &yesterday,
		})
		dueToday := stores.CreateTask(adapter, ctx, &models.Task{
			Name:      "due today not overdue",
			Status:    models.TaskStatusTodo,
			TeamID:    team.ID,
			ProjectID: project.ID,
			EndAt:     &midDay,
		})
		_ = stores.CreateTask(adapter, ctx, &models.Task{
			Name:      "no due date",
			Status:    models.TaskStatusTodo,
			TeamID:    team.ID,
			ProjectID: project.ID,
		})

		tasks, err := adapter.Task().FindTasksOverdue(ctx)
		require.NoError(t, err)

		ids := make(map[string]bool)
		for _, tk := range tasks {
			ids[tk.ID.String()] = true
		}

		assert.True(t, ids[overdue.ID.String()], "overdue task should be included")
		assert.False(t, ids[doneOverdue.ID.String()], "done task should not be included even if past due")
		assert.False(t, ids[dueToday.ID.String()], "due-today task should not be in overdue list")
	})
}
