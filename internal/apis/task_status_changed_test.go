package apis_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

func TestApi_TaskUpdate_EnquesStatusChangedJob(t *testing.T) {
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

		tests := []apis.ApiScenario{
			{
				Name:           "success: todo→in_progress enqueues task_status_changed",
				Method:         http.MethodPut,
				URL:            fmt.Sprintf("/tasks/%s", task.ID),
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
					scenario.Headers = []string{header}
					scenario.Body = apis.JsonToReader(t, stores.UpdateTaskDto{
						Name:   "task",
						Status: models.TaskStatusInProgress,
					})
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, _ *httptest.ResponseRecorder) {
					assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "task_status_changed"))
					assert.Equal(t, int64(0), countJobsByKind(t, ctx, db, "task_completed"))
				},
			},
			{
				Name:           "success: in_progress→done enqueues task_completed, not task_status_changed",
				Method:         http.MethodPut,
				URL:            fmt.Sprintf("/tasks/%s", task.ID),
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
					scenario.Headers = []string{header}
					scenario.Body = apis.JsonToReader(t, stores.UpdateTaskDto{
						Name:   "task",
						Status: models.TaskStatusDone,
					})
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, _ *httptest.ResponseRecorder) {
					// still 1 from previous sub-test (upsert by unique key)
					assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "task_status_changed"))
					assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "task_completed"))
				},
			},
			{
				Name:           "success: no status change enqueues no new status_changed job",
				Method:         http.MethodPut,
				URL:            fmt.Sprintf("/tasks/%s", task.ID),
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
					scenario.Headers = []string{header}
					scenario.Body = apis.JsonToReader(t, stores.UpdateTaskDto{
						Name:   "renamed but still done",
						Status: models.TaskStatusDone,
					})
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, _ *httptest.ResponseRecorder) {
					assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "task_status_changed"))
					assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "task_completed"))
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}
