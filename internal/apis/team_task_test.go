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
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
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
