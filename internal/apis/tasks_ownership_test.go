package apis_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

func TestApi_TaskOwnership(t *testing.T) {
	tests := []apis.ApiScenario{
		{
			Name:            "non-owner PUT is rejected with 403",
			Method:          http.MethodPut,
			URL:             "/tasks/{task-id}",
			ExpectedStatus:  http.StatusForbidden,
			ExpectedContent: []string{"only the task creator"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, sc *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				other := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				core.CreateTeamMemberWithOptions(t, app, team.Team.ID, other.User.ID,
					core.TeamWithRole(models.TeamMemberRoleMember),
					core.TeamWithBilling(false),
				)
				project := core.CreateProjectAndTasks(t, app, &team.Member)
				task := repository.MustCreateOneCtx(t, t.Context(), repository.Task, app.Db(), &models.Task{
					TeamID:            team.Team.ID,
					ProjectID:         project.ID,
					Status:            models.TaskStatusTodo,
					CreatedByMemberID: &team.Member.ID,
				})
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, other.User.Email)
				sc.Headers = []string{header}
				sc.URL = fmt.Sprintf("/tasks/%s", task.ID)
				sc.Body = apis.JsonToReader(t, stores.UpdateTaskDto{Name: "hijacked", Status: models.TaskStatusInProgress})
			},
		},
		{
			Name:            "non-owner DELETE is rejected with 403",
			Method:          http.MethodDelete,
			URL:             "/tasks/{task-id}",
			ExpectedStatus:  http.StatusForbidden,
			ExpectedContent: []string{"only the task creator"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, sc *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				other := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				core.CreateTeamMemberWithOptions(t, app, team.Team.ID, other.User.ID,
					core.TeamWithRole(models.TeamMemberRoleMember),
					core.TeamWithBilling(false),
				)
				project := core.CreateProjectAndTasks(t, app, &team.Member)
				task := repository.MustCreateOneCtx(t, t.Context(), repository.Task, app.Db(), &models.Task{
					TeamID:            team.Team.ID,
					ProjectID:         project.ID,
					Status:            models.TaskStatusTodo,
					CreatedByMemberID: &team.Member.ID,
				})
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, other.User.Email)
				sc.Headers = []string{header}
				sc.URL = fmt.Sprintf("/tasks/%s", task.ID)
			},
		},
		{
			Name:           "owner can PUT their own task",
			Method:         http.MethodPut,
			URL:            "/tasks/{task-id}",
			ExpectedStatus: http.StatusNoContent,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, sc *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				project := core.CreateProjectAndTasks(t, app, &team.Member)
				task := repository.MustCreateOneCtx(t, t.Context(), repository.Task, app.Db(), &models.Task{
					TeamID:            team.Team.ID,
					ProjectID:         project.ID,
					Status:            models.TaskStatusTodo,
					CreatedByMemberID: &team.Member.ID,
				})
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, owner.User.Email)
				sc.Headers = []string{header}
				sc.URL = fmt.Sprintf("/tasks/%s", task.ID)
				sc.Body = apis.JsonToReader(t, stores.UpdateTaskDto{Name: "my update", Status: models.TaskStatusInProgress})
			},
		},
		{
			Name:           "owner can DELETE their own task",
			Method:         http.MethodDelete,
			URL:            "/tasks/{task-id}",
			ExpectedStatus: http.StatusNoContent,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, sc *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				project := core.CreateProjectAndTasks(t, app, &team.Member)
				task := repository.MustCreateOneCtx(t, t.Context(), repository.Task, app.Db(), &models.Task{
					TeamID:            team.Team.ID,
					ProjectID:         project.ID,
					Status:            models.TaskStatusTodo,
					CreatedByMemberID: &team.Member.ID,
				})
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, owner.User.Email)
				sc.Headers = []string{header}
				sc.URL = fmt.Sprintf("/tasks/%s", task.ID)
			},
		},
		{
			Name:            "non-owner GET is still allowed",
			Method:          http.MethodGet,
			URL:             "/tasks/{task-id}",
			ExpectedStatus:  http.StatusOK,
			ExpectedContent: []string{`"status"`},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, sc *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				other := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				core.CreateTeamMemberWithOptions(t, app, team.Team.ID, other.User.ID,
					core.TeamWithRole(models.TeamMemberRoleMember),
					core.TeamWithBilling(false),
				)
				project := core.CreateProjectAndTasks(t, app, &team.Member)
				task := repository.MustCreateOneCtx(t, t.Context(), repository.Task, app.Db(), &models.Task{
					TeamID:            team.Team.ID,
					ProjectID:         project.ID,
					Status:            models.TaskStatusTodo,
					CreatedByMemberID: &team.Member.ID,
				})
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, other.User.Email)
				sc.Headers = []string{header}
				sc.URL = fmt.Sprintf("/tasks/%s", task.ID)
			},
		},
		{
			Name:           "superuser can PUT any task",
			Method:         http.MethodPut,
			URL:            "/tasks/{task-id}",
			ExpectedStatus: http.StatusNoContent,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, sc *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				// superuser is a separate team member with the superuser permission
				super := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(),
					core.UserWithEmail(randomEmail()),
					core.UserWithPermission("superuser"),
				)
				core.CreateTeamMemberWithOptions(t, app, team.Team.ID, super.User.ID,
					core.TeamWithRole(models.TeamMemberRoleMember),
					core.TeamWithBilling(false),
				)
				project := core.CreateProjectAndTasks(t, app, &team.Member)
				task := repository.MustCreateOneCtx(t, t.Context(), repository.Task, app.Db(), &models.Task{
					TeamID:            team.Team.ID,
					ProjectID:         project.ID,
					Status:            models.TaskStatusTodo,
					CreatedByMemberID: &team.Member.ID, // owned by owner, not super
				})
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, super.User.Email)
				sc.Headers = []string{header}
				sc.URL = fmt.Sprintf("/tasks/%s", task.ID)
				sc.Body = apis.JsonToReader(t, stores.UpdateTaskDto{Name: "superuser override", Status: models.TaskStatusDone})
			},
		},
	}

	for _, tt := range tests {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			testApi := apis.SetupApi(t, ctx, db)
			tt.TestAppFactory = func(t testing.TB) *apis.TestApi { return testApi }
			tt.Test(t)
		})
	}
}
