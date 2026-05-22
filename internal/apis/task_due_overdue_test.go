//go:build integration

package apis_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/types"
)

func countJobsByKind(t testing.TB, ctx context.Context, db database.Dbx, kind string) int64 {
	t.Helper()
	return repository.MustCountAllCtx(t, ctx, repository.Job, db, &map[string]any{
		"kind": map[string]any{"_eq": kind},
	})
}

// Task created with a due date → task_due_today and task_overdue jobs enqueued.
func TestApi_TaskCreate_EnquesDueDateJobs(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		owner := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerifiedNow())
		team := core.CreateTeamAndMemberWithOptions(t, testApi.App, &owner.User)
		project := core.CreateProjectAndTasks(t, testApi.App, &team.Member)

		dueDate := time.Now().Add(48 * time.Hour)

		tests := []apis.ApiScenario{
			{
				Name:           "success: create task with due date enqueues due+overdue jobs",
				Method:         http.MethodPost,
				URL:            fmt.Sprintf("/task-projects/%s", project.ID),
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
					scenario.Headers = []string{header}
					scenario.Body = apis.JsonToReader(t, services.TaskFields{
						Name:   "Task with due date",
						Status: models.TaskStatusTodo,
						EndAt:  types.Pointer(dueDate),
					})
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, _ *httptest.ResponseRecorder) {
					assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "task_due_today"))
					assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "task_overdue"))
				},
			},
			{
				Name:           "success: create task without due date enqueues no jobs",
				Method:         http.MethodPost,
				URL:            fmt.Sprintf("/task-projects/%s", project.ID),
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
					scenario.Headers = []string{header}
					scenario.Body = apis.JsonToReader(t, services.TaskFields{
						Name:   "Task without due date",
						Status: models.TaskStatusTodo,
					})
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, _ *httptest.ResponseRecorder) {
					// still 1 from previous sub-test (shared tx)
					assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "task_due_today"))
					assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "task_overdue"))
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

// Task updated with a new due date → both jobs enqueued.
// Task updated with no due date change → no new jobs.
func TestApi_TaskUpdate_EnquesDueDateJobs(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		owner := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerifiedNow())
		team := core.CreateTeamAndMemberWithOptions(t, testApi.App, &owner.User)
		project := core.CreateProjectAndTasks(t, testApi.App, &team.Member)

		task := repository.MustCreateOneCtx(t, ctx, repository.Task, db, &models.Task{
			TeamID:    team.Team.ID,
			ProjectID: project.ID,
			Status:    models.TaskStatusTodo,
			Rank:      0,
		})

		dueDate := time.Now().Add(48 * time.Hour)
		newDueDate := dueDate.Add(24 * time.Hour)

		tests := []apis.ApiScenario{
			{
				Name:           "success: set due date for first time enqueues both jobs",
				Method:         http.MethodPut,
				URL:            fmt.Sprintf("/tasks/%s", task.ID),
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
					scenario.Headers = []string{header}
					scenario.Body = apis.JsonToReader(t, stores.UpdateTaskDto{
						Name:   "task",
						Status: models.TaskStatusTodo,
						EndAt:  types.Pointer(dueDate),
					})
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, _ *httptest.ResponseRecorder) {
					assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "task_due_today"))
					assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "task_overdue"))
				},
			},
			{
				Name:           "success: change due date upserts jobs (count stays 1 per unique key)",
				Method:         http.MethodPut,
				URL:            fmt.Sprintf("/tasks/%s", task.ID),
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
					scenario.Headers = []string{header}
					scenario.Body = apis.JsonToReader(t, stores.UpdateTaskDto{
						Name:   "task",
						Status: models.TaskStatusTodo,
						EndAt:  types.Pointer(newDueDate),
					})
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, _ *httptest.ResponseRecorder) {
					// Unique key upserts — still 1 per kind, not 2
					assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "task_due_today"))
					assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "task_overdue"))
				},
			},
			{
				Name:           "success: no due date change enqueues no new jobs",
				Method:         http.MethodPut,
				URL:            fmt.Sprintf("/tasks/%s", task.ID),
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
					scenario.Headers = []string{header}
					// same due date as previous — no change
					scenario.Body = apis.JsonToReader(t, stores.UpdateTaskDto{
						Name:   "task renamed",
						Status: models.TaskStatusTodo,
						EndAt:  types.Pointer(newDueDate),
					})
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, _ *httptest.ResponseRecorder) {
					assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "task_due_today"))
					assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "task_overdue"))
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}
