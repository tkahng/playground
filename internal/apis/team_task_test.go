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
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/types"
)

func TestApi_TeamTaskList(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		team1 := CreateTeamAndOwner(t, testApi.App)
		project1 := core.CreateProjectAndTasks(
			t,
			testApi.App,
			&team1.Member,
			core.WithTaskByCountAndStatus(3, models.TaskStatusTodo),
			core.WithTaskByCountAndStatus(3, models.TaskStatusInProgress),
			core.WithTaskByCountAndStatus(3, models.TaskStatusDone),
		)
		// task
		tests := []ApiScenario{
			{
				Name:           "success: all tasks",
				Method:         http.MethodGet,
				URL:            "/task-projects/{task-project-id}/tasks",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/task-projects/%s/tasks", project1.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.Task]](t, res.Body.Bytes())
					assert.Equal(t, int64(9), result.Meta.Total)
				},
			},
			{
				Name:           "success: all tasks with done status filter",
				Method:         http.MethodGet,
				URL:            "/task-projects/{task-project-id}/tasks?status=done",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/task-projects/%s/tasks?status=done", project1.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.Task]](t, res.Body.Bytes())
					assert.Equal(t, int64(3), result.Meta.Total)
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

func TestApi_TeamTaskUpdate(t *testing.T) {
	tests := []ApiScenario{
		{
			Name:           "success: update task",
			Method:         http.MethodPut,
			URL:            "/tasks/{task-id}",
			ExpectedStatus: http.StatusNoContent,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team1 := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				project := core.CreateProjectAndTasks(t, app, &team1.Member)
				task := repository.MustCreateOneCtx(t, t.Context(), repository.Task, app.Db(), &models.Task{
					TeamID:    team1.Team.ID,
					ProjectID: project.ID,
					Status:    models.TaskStatusTodo,
					Rank:      0,
				})
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/tasks/%s", task.ID.String())
				input := stores.UpdateTaskDto{
					Name:        "updated",
					Description: types.Pointer("UpdateTaskDto"),
					Status:      models.TaskStatusInProgress,
				}
				scenario.Store.Set("input", input)
				scenario.Body = JsonToReader(t, input)
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
				result := repository.MustFindOneCtx(t, t.Context(), repository.Task, app.Db(), nil)
				input, ok := scenario.Store.Get("input").(stores.UpdateTaskDto)
				if !ok {
					t.Fatal("no input")
				}
				assert.Equal(t, input.Name, result.Name)
				assert.Equal(t, input.Description, result.Description)
				assert.Equal(t, input.Status, result.Status)
			},
		},
	}
	for _, tt := range tests {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			testApi := SetupApi(t, ctx, db)
			tt.TestAppFactory = func(t testing.TB) *TestApi {
				return testApi
			}
			tt.Test(t)
		})
	}
}
