//go:build integration

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
	"github.com/tkahng/playground/internal/tools/types"
)

func TestApi_TaskUpdate_AssignedToTaskDeduplication(t *testing.T) {
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

		assigneeID := team.Member.ID

		tests := []apis.ApiScenario{
			{
				Name:           "success: assign member enqueues assigned_to_task",
				Method:         http.MethodPut,
				URL:            fmt.Sprintf("/tasks/%s", task.ID),
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
					scenario.Headers = []string{header}
					scenario.Body = apis.JsonToReader(t, stores.UpdateTaskDto{
						Name:       "task",
						Status:     models.TaskStatusTodo,
						AssigneeID: types.Pointer(assigneeID),
					})
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, _ *httptest.ResponseRecorder) {
					assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "assigned_to_task"))
				},
			},
			{
				Name:           "success: same assignee again deduplicates (count stays 1)",
				Method:         http.MethodPut,
				URL:            fmt.Sprintf("/tasks/%s", task.ID),
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
					scenario.Headers = []string{header}
					scenario.Body = apis.JsonToReader(t, stores.UpdateTaskDto{
						Name:       "task renamed",
						Status:     models.TaskStatusTodo,
						AssigneeID: types.Pointer(assigneeID),
					})
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, _ *httptest.ResponseRecorder) {
					// unique key upsert — still 1, not 2
					assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "assigned_to_task"))
				},
			},
			{
				Name:           "success: no assignee change enqueues no job",
				Method:         http.MethodPut,
				URL:            fmt.Sprintf("/tasks/%s", task.ID),
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
					scenario.Headers = []string{header}
					scenario.Body = apis.JsonToReader(t, stores.UpdateTaskDto{
						Name:       "task renamed again",
						Status:     models.TaskStatusTodo,
						AssigneeID: types.Pointer(assigneeID),
					})
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, _ *httptest.ResponseRecorder) {
					assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "assigned_to_task"))
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}
