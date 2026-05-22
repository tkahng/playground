//go:build integration

package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
)

func TestTaskNotificationScheduler_EnqueuesDueTodayJobs(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)

		user, err := adapter.User().CreateUser(ctx, &models.User{Email: "sched-due@example.com"})
		require.NoError(t, err)
		member, err := adapter.TeamMember().CreateTeamMemberFromUserAndSlug(ctx, user, "sched-team", models.TeamMemberRoleOwner)
		require.NoError(t, err)
		project, err := adapter.Task().CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name:     "proj",
			Status:   models.TaskProjectStatusTodo,
			TeamID:   member.TeamID,
			MemberID: member.ID,
		})
		require.NoError(t, err)

		now := time.Now().UTC()
		midDay := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.UTC)
		tomorrow := midDay.Add(25 * time.Hour)

		_, err = adapter.Task().CreateTask(ctx, &models.Task{
			Name:      "due today task",
			Status:    models.TaskStatusTodo,
			TeamID:    member.TeamID,
			ProjectID: project.ID,
			EndAt:     &midDay,
		})
		require.NoError(t, err)

		_, err = adapter.Task().CreateTask(ctx, &models.Task{
			Name:      "due tomorrow, should be skipped",
			Status:    models.TaskStatusTodo,
			TeamID:    member.TeamID,
			ProjectID: project.ID,
			EndAt:     &tomorrow,
		})
		require.NoError(t, err)

		_, err = adapter.Task().CreateTask(ctx, &models.Task{
			Name:      "done due today, should be skipped",
			Status:    models.TaskStatusDone,
			TeamID:    member.TeamID,
			ProjectID: project.ID,
			EndAt:     &midDay,
		})
		require.NoError(t, err)

		jobSvc := services.NewJobService(jobs.NewDbJobManager(db))
		scheduler := services.NewTaskNotificationScheduler(adapter.Task(), jobSvc)
		scheduler.RunOnce(ctx)

		count := repository.MustCountAllCtx(t, ctx, repository.Job, db, &map[string]any{
			"kind": map[string]any{"_eq": "task_due_today"},
		})
		assert.Equal(t, int64(1), count, "only the non-done due-today task should enqueue a job")
	})
}

func TestTaskNotificationScheduler_EnqueuesOverdueJobs(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)

		user, err := adapter.User().CreateUser(ctx, &models.User{Email: "sched-overdue@example.com"})
		require.NoError(t, err)
		member, err := adapter.TeamMember().CreateTeamMemberFromUserAndSlug(ctx, user, "sched-overdue-team", models.TeamMemberRoleOwner)
		require.NoError(t, err)
		project, err := adapter.Task().CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name:     "proj",
			Status:   models.TaskProjectStatusTodo,
			TeamID:   member.TeamID,
			MemberID: member.ID,
		})
		require.NoError(t, err)

		yesterday := time.Now().UTC().Truncate(24*time.Hour).Add(-2 * time.Hour)
		future := time.Now().UTC().Add(48 * time.Hour)

		_, err = adapter.Task().CreateTask(ctx, &models.Task{
			Name:      "overdue task",
			Status:    models.TaskStatusInProgress,
			TeamID:    member.TeamID,
			ProjectID: project.ID,
			EndAt:     &yesterday,
		})
		require.NoError(t, err)

		_, err = adapter.Task().CreateTask(ctx, &models.Task{
			Name:      "done overdue, should be skipped",
			Status:    models.TaskStatusDone,
			TeamID:    member.TeamID,
			ProjectID: project.ID,
			EndAt:     &yesterday,
		})
		require.NoError(t, err)

		_, err = adapter.Task().CreateTask(ctx, &models.Task{
			Name:      "future task, should be skipped",
			Status:    models.TaskStatusTodo,
			TeamID:    member.TeamID,
			ProjectID: project.ID,
			EndAt:     &future,
		})
		require.NoError(t, err)

		jobSvc := services.NewJobService(jobs.NewDbJobManager(db))
		scheduler := services.NewTaskNotificationScheduler(adapter.Task(), jobSvc)
		scheduler.RunOnce(ctx)

		count := repository.MustCountAllCtx(t, ctx, repository.Job, db, &map[string]any{
			"kind": map[string]any{"_eq": "task_overdue"},
		})
		assert.Equal(t, int64(1), count, "only the non-done overdue task should enqueue a job")
	})
}

func TestTaskNotificationScheduler_DeduplicatesJobs(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)

		user, err := adapter.User().CreateUser(ctx, &models.User{Email: "sched-dedup@example.com"})
		require.NoError(t, err)
		member, err := adapter.TeamMember().CreateTeamMemberFromUserAndSlug(ctx, user, "sched-dedup-team", models.TeamMemberRoleOwner)
		require.NoError(t, err)
		project, err := adapter.Task().CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name:     "proj",
			Status:   models.TaskProjectStatusTodo,
			TeamID:   member.TeamID,
			MemberID: member.ID,
		})
		require.NoError(t, err)

		yesterday := time.Now().UTC().Truncate(24*time.Hour).Add(-2 * time.Hour)

		_, err = adapter.Task().CreateTask(ctx, &models.Task{
			Name:      "overdue task",
			Status:    models.TaskStatusTodo,
			TeamID:    member.TeamID,
			ProjectID: project.ID,
			EndAt:     &yesterday,
		})
		require.NoError(t, err)

		jobSvc := services.NewJobService(jobs.NewDbJobManager(db))
		scheduler := services.NewTaskNotificationScheduler(adapter.Task(), jobSvc)

		// Run twice — unique key should prevent duplicates
		scheduler.RunOnce(ctx)
		scheduler.RunOnce(ctx)

		count := repository.MustCountAllCtx(t, ctx, repository.Job, db, &map[string]any{
			"kind": map[string]any{"_eq": "task_overdue"},
		})
		assert.Equal(t, int64(1), count, "second run should not create duplicate jobs")
	})
}
