package apis_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
)

func TestApi_TeamTaskProjectList(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		owner := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerifiedNow())
		team1 := core.CreateTeamAndMemberWithOptions(t, testApi.App, &owner.User)
		project1 := core.CreateProjectAndTasks(
			t,
			testApi.App,
			&team1.Member,
			core.WithProjectName("Project 1"),
			core.WithProjectStatus(models.TaskProjectStatusTodo),
		)
		project2 := core.CreateProjectAndTasks(
			t,
			testApi.App,
			&team1.Member,
			core.WithProjectName("Project 2"),
			core.WithProjectStatus(models.TaskProjectStatusInProgress),
		)
		project3 := core.CreateProjectAndTasks(
			t,
			testApi.App,
			&team1.Member,
			core.WithProjectName("Project 3"),
			core.WithProjectStatus(models.TaskProjectStatusDone),
		)
		var allIds []uuid.UUID = []uuid.UUID{project1.ID, project2.ID, project3.ID}
		// task
		tests := []ApiScenario{
			{
				Name:           "success: all projects",
				Method:         http.MethodGet,
				URL:            "/teams/{team-id}/task-projects",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/task-projects", team1.Team.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.TaskProject]](t, res.Body.Bytes())
					assert.Equal(t, int64(3), result.Meta.Total)
					test.TestSliceEveryFunc(t, "", result.Data, func(item *apis.TaskProject) bool {
						return slices.Contains(allIds, item.ID)
					})
				},
			},
			{
				Name:           "success: projects with status todo",
				Method:         http.MethodGet,
				URL:            "/teams/{team-id}/task-projects?status=todo",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/task-projects?status=todo", team1.Team.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.TaskProject]](t, res.Body.Bytes())
					assert.Equal(t, int64(1), result.Meta.Total)
					assert.Equal(t, project1.ID.String(), result.Data[0].ID.String())
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}
