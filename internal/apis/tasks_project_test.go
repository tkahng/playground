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
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/types"
)

func TestApi_TeamTaskProjectList(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
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
		tests := []apis.ApiScenario{
			{
				Name:           "success: all projects",
				Method:         http.MethodGet,
				URL:            "/teams/{team-id}/task-projects",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/task-projects", team1.Team.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
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
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/task-projects?status=todo", team1.Team.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
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
func TestApi_TeamTaskProjectCreate(t *testing.T) {
	// task
	tests := []apis.ApiScenario{
		{
			Name:           "success: create project",
			Method:         http.MethodPost,
			URL:            "/teams/{team-id}/task-projects",
			ExpectedStatus: http.StatusOK,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team1 := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/task-projects", team1.Team.ID.String())
				input := apis.CreateTaskProjectWithoutTeamWithTasks{
					CreateTaskProjectWithoutTeamDTO: apis.CreateTaskProjectWithoutTeamDTO{
						Name:        "New Project",
						Description: types.Pointer("This is a new project."),
						Status:      models.TaskProjectStatusTodo,
					},
				}
				scenario.Store.Set("input", input)
				scenario.Body = apis.JsonToReader(t, input)
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.TaskProject](t, res.Body.Bytes())
				input, ok := scenario.Store.Get("input").(apis.CreateTaskProjectWithoutTeamWithTasks)
				if !ok {
					t.Fatal("no input")
				}
				assert.Equal(t, input.CreateTaskProjectWithoutTeamDTO.Name, result.Name)
				assert.Equal(t, input.CreateTaskProjectWithoutTeamDTO.Description, result.Description)
				assert.Equal(t, input.CreateTaskProjectWithoutTeamDTO.Status, result.Status)
				assert.NotNil(t, result.WorkflowID)
				assert.NotNil(t, result.WorkflowStatusID)
			},
		},
		{
			Name:           "success: create project with workflow status",
			Method:         http.MethodPost,
			URL:            "/teams/{team-id}/task-projects",
			ExpectedStatus: http.StatusOK,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team1 := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				core.CreateProjectAndTasks(t, app, &team1.Member)
				doneStatusID := findWorkflowStatusID(t, app, team1.Team.ID, "project", "done")

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/task-projects", team1.Team.ID.String())
				scenario.Store.Set("workflow_status_id", doneStatusID)
				scenario.Body = apis.JsonToReader(t, apis.CreateTaskProjectWithoutTeamWithTasks{
					CreateTaskProjectWithoutTeamDTO: apis.CreateTaskProjectWithoutTeamDTO{
						Name:             "New Project",
						Status:           models.TaskProjectStatusTodo,
						WorkflowStatusID: &doneStatusID,
					},
				})
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.TaskProject](t, res.Body.Bytes())
				doneStatusID := scenario.Store.Get("workflow_status_id").(uuid.UUID)
				assert.Equal(t, models.TaskProjectStatusDone, result.Status)
				assert.Equal(t, doneStatusID, *result.WorkflowStatusID)
			},
		},
		{
			Name:            "fail: guest cannot create project",
			Method:          http.MethodPost,
			URL:             "/teams/{team-id}/task-projects",
			ExpectedStatus:  http.StatusForbidden,
			ExpectedContent: []string{"You do not have the required team permission: projects.create"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team1 := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				guestUser := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				guest := core.CreateTeamMemberWithOptions(t, app, team1.Team.ID, guestUser.User.ID, core.TeamWithRole(models.TeamMemberRoleGuest))
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, guestUser.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/task-projects", guest.TeamID.String())
				input := apis.CreateTaskProjectWithoutTeamWithTasks{
					CreateTaskProjectWithoutTeamDTO: apis.CreateTaskProjectWithoutTeamDTO{
						Name:   "New Project",
						Status: models.TaskProjectStatusTodo,
					},
				}
				scenario.Body = apis.JsonToReader(t, input)
			},
		},
	}
	for _, tt := range tests {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			testApi := apis.SetupApi(t, ctx, db)
			tt.TestAppFactory = func(t testing.TB) *apis.TestApi {
				return testApi
			}
			tt.Test(t)
		})
	}
}

func TestApi_TeamTaskProjectPermissions(t *testing.T) {
	tests := []apis.ApiScenario{
		{
			Name:           "success: owner can update project with workflow status",
			Method:         http.MethodPut,
			URL:            "/task-projects/{task-project-id}",
			ExpectedStatus: http.StatusNoContent,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team1 := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				project := core.CreateProjectAndTasks(t, app, &team1.Member)
				doneStatusID := findWorkflowStatusID(t, app, team1.Team.ID, "project", "done")

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/task-projects/%s", project.ID.String())
				scenario.Store.Set("project_id", project.ID)
				scenario.Store.Set("workflow_status_id", doneStatusID)
				scenario.Body = apis.JsonToReader(t, stores.UpdateTaskProjectBaseDTO{
					Name:             "Updated Project",
					Status:           models.TaskProjectStatusTodo,
					WorkflowStatusID: &doneStatusID,
				})
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				projectID := scenario.Store.Get("project_id").(uuid.UUID)
				doneStatusID := scenario.Store.Get("workflow_status_id").(uuid.UUID)
				project, err := app.Adapter().Task().FindTaskProjectByID(t.Context(), projectID)
				assert.NoError(t, err)
				assert.Equal(t, models.TaskProjectStatusDone, project.Status)
				assert.Equal(t, doneStatusID, *project.WorkflowStatusID)
			},
		},
		{
			Name:            "fail: member cannot update project",
			Method:          http.MethodPut,
			URL:             "/task-projects/{task-project-id}",
			ExpectedStatus:  http.StatusForbidden,
			ExpectedContent: []string{"You do not have the required team permission: projects.manage"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team1 := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				memberUser := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				core.CreateTeamMemberWithOptions(t, app, team1.Team.ID, memberUser.User.ID, core.TeamWithRole(models.TeamMemberRoleMember))
				project := core.CreateProjectAndTasks(t, app, &team1.Member)

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, memberUser.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/task-projects/%s", project.ID.String())
				scenario.Body = apis.JsonToReader(t, stores.UpdateTaskProjectBaseDTO{
					Name:   "Updated Project",
					Status: models.TaskProjectStatusInProgress,
				})
			},
		},
		{
			Name:            "fail: member cannot delete project",
			Method:          http.MethodDelete,
			URL:             "/task-projects/{task-project-id}",
			ExpectedStatus:  http.StatusForbidden,
			ExpectedContent: []string{"You do not have the required team permission: projects.delete"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team1 := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				memberUser := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				core.CreateTeamMemberWithOptions(t, app, team1.Team.ID, memberUser.User.ID, core.TeamWithRole(models.TeamMemberRoleMember))
				project := core.CreateProjectAndTasks(t, app, &team1.Member)

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, memberUser.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/task-projects/%s", project.ID.String())
			},
		},
	}
	for _, tt := range tests {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			testApi := apis.SetupApi(t, ctx, db)
			tt.TestAppFactory = func(t testing.TB) *apis.TestApi {
				return testApi
			}
			tt.Test(t)
		})
	}
}

func TestApi_TeamTaskProjectTasksCreateReferences(t *testing.T) {
	tests := []apis.ApiScenario{
		{
			Name:           "success: create task returns workflow status",
			Method:         http.MethodPost,
			URL:            "/task-projects/{task-project-id}",
			ExpectedStatus: http.StatusOK,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team1 := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				project := core.CreateProjectAndTasks(t, app, &team1.Member)

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/task-projects/%s", project.ID.String())
				scenario.Body = apis.JsonToReader(t, services.TaskFields{
					Name:   "Task with workflow status",
					Status: models.TaskStatusTodo,
				})
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.Task](t, res.Body.Bytes())
				assert.NotNil(t, result.WorkflowStatusID)
			},
		},
		{
			Name:           "success: create task with workflow status",
			Method:         http.MethodPost,
			URL:            "/task-projects/{task-project-id}",
			ExpectedStatus: http.StatusOK,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team1 := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				project := core.CreateProjectAndTasks(t, app, &team1.Member)
				doneStatusID := findWorkflowStatusID(t, app, team1.Team.ID, "task", "done")

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/task-projects/%s", project.ID.String())
				scenario.Store.Set("workflow_status_id", doneStatusID)
				scenario.Body = apis.JsonToReader(t, services.TaskFields{
					Name:             "Task with done workflow status",
					Status:           models.TaskStatusTodo,
					WorkflowStatusID: &doneStatusID,
				})
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.Task](t, res.Body.Bytes())
				doneStatusID := scenario.Store.Get("workflow_status_id").(uuid.UUID)
				assert.Equal(t, models.TaskStatusDone, result.Status)
				assert.Equal(t, doneStatusID, *result.WorkflowStatusID)
			},
		},
		{
			Name:            "fail: guest cannot create task",
			Method:          http.MethodPost,
			URL:             "/task-projects/{task-project-id}",
			ExpectedStatus:  http.StatusForbidden,
			ExpectedContent: []string{"You do not have the required team permission: tasks.create"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team1 := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				guestUser := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				core.CreateTeamMemberWithOptions(t, app, team1.Team.ID, guestUser.User.ID, core.TeamWithRole(models.TeamMemberRoleGuest))
				project := core.CreateProjectAndTasks(t, app, &team1.Member)

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, guestUser.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/task-projects/%s", project.ID.String())
				scenario.Body = apis.JsonToReader(t, services.TaskFields{
					Name:   "Task by guest",
					Status: models.TaskStatusTodo,
				})
			},
		},
		{
			Name:            "fail: create task rejects assignee from another team",
			Method:          http.MethodPost,
			URL:             "/task-projects/{task-project-id}",
			ExpectedStatus:  http.StatusBadRequest,
			ExpectedContent: []string{"assignee must be an active member of the task team"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team1 := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				otherOwner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				team2 := core.CreateTeamAndMemberWithOptions(t, app, &otherOwner.User)
				project := core.CreateProjectAndTasks(t, app, &team1.Member)

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/task-projects/%s", project.ID.String())
				scenario.Body = apis.JsonToReader(t, services.TaskFields{
					Name:       "Task with invalid assignee",
					Status:     models.TaskStatusTodo,
					AssigneeID: &team2.Member.ID,
				})
			},
		},
	}
	for _, tt := range tests {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			testApi := apis.SetupApi(t, ctx, db)
			tt.TestAppFactory = func(t testing.TB) *apis.TestApi {
				return testApi
			}
			tt.Test(t)
		})
	}
}

func findWorkflowStatusID(t testing.TB, app *core.BaseApp, teamID uuid.UUID, appliesTo string, slug string) uuid.UUID {
	t.Helper()
	workflows, err := app.Adapter().Task().ListWorkflows(t.Context(), &stores.WorkflowFilter{
		TeamIds:   []uuid.UUID{teamID},
		AppliesTo: []string{appliesTo},
	})
	if err != nil {
		t.Fatalf("list workflows: %v", err)
	}
	workflowIds := make([]uuid.UUID, len(workflows))
	for idx, workflow := range workflows {
		workflowIds[idx] = workflow.ID
	}
	statuses, err := app.Adapter().Task().LoadWorkflowStatuses(t.Context(), workflowIds...)
	if err != nil {
		t.Fatalf("load workflow statuses: %v", err)
	}
	for _, group := range statuses {
		for _, status := range group {
			if status.Slug == slug {
				return status.ID
			}
		}
	}
	t.Fatalf("workflow status %s/%s not found", appliesTo, slug)
	return uuid.Nil
}
