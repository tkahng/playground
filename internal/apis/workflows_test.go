package apis_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/types"
)

func TestApi_TeamWorkflowList(t *testing.T) {
	tests := []apis.ApiScenario{
		{
			Name:           "success: lists default workflows with statuses",
			Method:         http.MethodGet,
			URL:            "/teams/{team-id}/workflows",
			ExpectedStatus: http.StatusOK,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				core.CreateProjectAndTasks(t, app, &team.Member)

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows", team.Team.ID.String())
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[[]*apis.Workflow](t, res.Body.Bytes())
				assert.Len(t, result, 2)
				assertWorkflowStatuses(t, result, "project")
				assertWorkflowStatuses(t, result, "task")
			},
		},
		{
			Name:           "success: filters by workflow target",
			Method:         http.MethodGet,
			URL:            "/teams/{team-id}/workflows?applies_to=task",
			ExpectedStatus: http.StatusOK,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				core.CreateProjectAndTasks(t, app, &team.Member)

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows?applies_to=task", team.Team.ID.String())
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[[]*apis.Workflow](t, res.Body.Bytes())
				assert.Len(t, result, 1)
				assertWorkflowStatuses(t, result, "task")
			},
		},
		{
			Name:            "fail: guest cannot list workflows",
			Method:          http.MethodGet,
			URL:             "/teams/{team-id}/workflows",
			ExpectedStatus:  http.StatusForbidden,
			ExpectedContent: []string{"team info not found. you are not a member of the team related to this request"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				other := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, other.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows", team.Team.ID.String())
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

func assertWorkflowStatuses(t testing.TB, workflows []*apis.Workflow, appliesTo string) {
	t.Helper()
	var workflow *apis.Workflow
	for _, item := range workflows {
		if item.AppliesTo == appliesTo {
			workflow = item
			break
		}
	}
	if assert.NotNil(t, workflow) {
		assert.True(t, workflow.IsDefault)
		assert.Len(t, workflow.Statuses, 3)
		assert.Equal(t, models.TaskStatusTodo, models.TaskStatus(workflow.Statuses[0].Slug))
		assert.Equal(t, models.TaskStatusInProgress, models.TaskStatus(workflow.Statuses[1].Slug))
		assert.Equal(t, models.TaskStatusDone, models.TaskStatus(workflow.Statuses[2].Slug))
	}
}

func TestApi_TeamWorkflowCreate(t *testing.T) {
	tests := []apis.ApiScenario{
		{
			Name:           "success: owner can create workflow",
			Method:         http.MethodPost,
			URL:            "/teams/{team-id}/workflows",
			ExpectedStatus: http.StatusOK,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows", team.Team.ID)
				scenario.Body = apis.JsonToReader(t, apis.WorkflowCreateRequestDTO{
					AppliesTo:   "task",
					Name:        "Delivery board",
					Description: types.Pointer("Workflow for delivery tasks."),
				})
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.Workflow](t, res.Body.Bytes())
				assert.Equal(t, "task", result.AppliesTo)
				assert.Equal(t, "Delivery board", result.Name)
				assert.False(t, result.IsDefault)
				assert.NotNil(t, result.CreatedByMemberID)
			},
		},
		{
			Name:            "fail: member cannot create workflow",
			Method:          http.MethodPost,
			URL:             "/teams/{team-id}/workflows",
			ExpectedStatus:  http.StatusForbidden,
			ExpectedContent: []string{"You do not have the required team permission: workflow.manage"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				memberUser := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				core.CreateTeamMemberWithOptions(t, app, team.Team.ID, memberUser.User.ID, core.TeamWithRole(models.TeamMemberRoleMember))

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, memberUser.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows", team.Team.ID)
				scenario.Body = apis.JsonToReader(t, apis.WorkflowCreateRequestDTO{
					AppliesTo: "task",
					Name:      "Delivery board",
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

func TestApi_TeamWorkflowUpdate(t *testing.T) {
	tests := []apis.ApiScenario{
		{
			Name:           "success: owner can update workflow",
			Method:         http.MethodPut,
			URL:            "/teams/{team-id}/workflows/{workflow-id}",
			ExpectedStatus: http.StatusOK,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				core.CreateProjectAndTasks(t, app, &team.Member)
				workflowID := findWorkflowID(t, app, team.Team.ID, "task")

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows/%s", team.Team.ID, workflowID)
				scenario.Body = apis.JsonToReader(t, stores.UpdateWorkflowDTO{
					Name:        types.Pointer("Renamed task workflow"),
					Description: types.Pointer("Updated task workflow description."),
				})
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.Workflow](t, res.Body.Bytes())
				assert.Equal(t, "Renamed task workflow", result.Name)
				assert.Equal(t, "Updated task workflow description.", *result.Description)
			},
		},
		{
			Name:            "fail: cannot update another team workflow",
			Method:          http.MethodPut,
			URL:             "/teams/{team-id}/workflows/{workflow-id}",
			ExpectedStatus:  http.StatusNotFound,
			ExpectedContent: []string{"workflow not found"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				otherOwner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				otherTeam := core.CreateTeamAndMemberWithOptions(t, app, &otherOwner.User)
				core.CreateProjectAndTasks(t, app, &team.Member)
				core.CreateProjectAndTasks(t, app, &otherTeam.Member)
				otherWorkflowID := findWorkflowID(t, app, otherTeam.Team.ID, "task")

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows/%s", team.Team.ID, otherWorkflowID)
				scenario.Body = apis.JsonToReader(t, stores.UpdateWorkflowDTO{
					Name: types.Pointer("Renamed task workflow"),
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

func TestApi_TeamWorkflowDelete(t *testing.T) {
	tests := []apis.ApiScenario{
		{
			Name:           "success: owner can delete unused workflow",
			Method:         http.MethodDelete,
			URL:            "/teams/{team-id}/workflows/{workflow-id}",
			ExpectedStatus: http.StatusNoContent,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				workflow, err := app.Adapter().Task().CreateWorkflow(t.Context(), &stores.CreateWorkflowDTO{
					TeamID:            team.Team.ID,
					CreatedByMemberID: &team.Member.ID,
					AppliesTo:         "task",
					Name:              "Unused workflow",
				})
				assert.NoError(t, err)

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows/%s", team.Team.ID, workflow.ID)
				scenario.Store.Set("workflow_id", workflow.ID)
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				workflowID := scenario.Store.Get("workflow_id").(uuid.UUID)
				workflow, err := app.Adapter().Task().FindWorkflowByID(t.Context(), workflowID)
				assert.NoError(t, err)
				assert.Nil(t, workflow)
			},
		},
		{
			Name:            "fail: cannot delete default workflow",
			Method:          http.MethodDelete,
			URL:             "/teams/{team-id}/workflows/{workflow-id}",
			ExpectedStatus:  http.StatusConflict,
			ExpectedContent: []string{"default workflow cannot be deleted"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				core.CreateProjectAndTasks(t, app, &team.Member)
				workflowID := findWorkflowID(t, app, team.Team.ID, "task")

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows/%s", team.Team.ID, workflowID)
			},
		},
		{
			Name:            "fail: cannot delete workflow in use",
			Method:          http.MethodDelete,
			URL:             "/teams/{team-id}/workflows/{workflow-id}",
			ExpectedStatus:  http.StatusConflict,
			ExpectedContent: []string{"workflow is in use"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				workflowID, _ := createAssignableTaskWorkflow(t, app, team.Team.ID, team.Member.ID)
				_, err := app.Adapter().Task().CreateTaskProject(t.Context(), &stores.CreateTaskProjectDTO{
					TeamID:     team.Team.ID,
					MemberID:   team.Member.ID,
					Name:       "Assigned workflow project",
					Status:     models.TaskProjectStatusTodo,
					WorkflowID: &workflowID,
				})
				assert.NoError(t, err)

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows/%s", team.Team.ID, workflowID)
			},
		},
		{
			Name:            "fail: member cannot delete workflow",
			Method:          http.MethodDelete,
			URL:             "/teams/{team-id}/workflows/{workflow-id}",
			ExpectedStatus:  http.StatusForbidden,
			ExpectedContent: []string{"You do not have the required team permission: workflow.manage"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				memberUser := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				core.CreateTeamMemberWithOptions(t, app, team.Team.ID, memberUser.User.ID, core.TeamWithRole(models.TeamMemberRoleMember))
				workflow, err := app.Adapter().Task().CreateWorkflow(t.Context(), &stores.CreateWorkflowDTO{
					TeamID:            team.Team.ID,
					CreatedByMemberID: &team.Member.ID,
					AppliesTo:         "task",
					Name:              "Member delete target",
				})
				assert.NoError(t, err)

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, memberUser.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows/%s", team.Team.ID, workflow.ID)
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

func TestApi_TeamWorkflowDefault(t *testing.T) {
	tests := []apis.ApiScenario{
		{
			Name:           "success: owner can set default workflow",
			Method:         http.MethodPut,
			URL:            "/teams/{team-id}/workflows/{workflow-id}/default",
			ExpectedStatus: http.StatusOK,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				core.CreateProjectAndTasks(t, app, &team.Member)
				oldDefaultID := findWorkflowID(t, app, team.Team.ID, "task")
				workflowID, _ := createAssignableTaskWorkflow(t, app, team.Team.ID, team.Member.ID)

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows/%s/default", team.Team.ID, workflowID)
				scenario.Store.Set("workflow_id", workflowID)
				scenario.Store.Set("old_default_id", oldDefaultID)
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.Workflow](t, res.Body.Bytes())
				workflowID := scenario.Store.Get("workflow_id").(uuid.UUID)
				oldDefaultID := scenario.Store.Get("old_default_id").(uuid.UUID)

				assert.Equal(t, workflowID, result.ID)
				assert.True(t, result.IsDefault)

				oldDefault, err := app.Adapter().Task().FindWorkflowByID(t.Context(), oldDefaultID)
				assert.NoError(t, err)
				assert.False(t, oldDefault.IsDefault)
			},
		},
		{
			Name:            "fail: cannot set incomplete workflow as default",
			Method:          http.MethodPut,
			URL:             "/teams/{team-id}/workflows/{workflow-id}/default",
			ExpectedStatus:  http.StatusBadRequest,
			ExpectedContent: []string{"workflow must have statuses before assignment"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				workflow, err := app.Adapter().Task().CreateWorkflow(t.Context(), &stores.CreateWorkflowDTO{
					TeamID:            team.Team.ID,
					CreatedByMemberID: &team.Member.ID,
					AppliesTo:         "task",
					Name:              "Incomplete workflow",
				})
				assert.NoError(t, err)

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows/%s/default", team.Team.ID, workflow.ID)
			},
		},
		{
			Name:            "fail: member cannot set default workflow",
			Method:          http.MethodPut,
			URL:             "/teams/{team-id}/workflows/{workflow-id}/default",
			ExpectedStatus:  http.StatusForbidden,
			ExpectedContent: []string{"You do not have the required team permission: workflow.manage"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				memberUser := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				core.CreateTeamMemberWithOptions(t, app, team.Team.ID, memberUser.User.ID, core.TeamWithRole(models.TeamMemberRoleMember))
				workflowID, _ := createAssignableTaskWorkflow(t, app, team.Team.ID, team.Member.ID)

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, memberUser.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows/%s/default", team.Team.ID, workflowID)
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

func TestApi_TeamWorkflowStatusCreate(t *testing.T) {
	tests := []apis.ApiScenario{
		{
			Name:           "success: owner can create workflow status",
			Method:         http.MethodPost,
			URL:            "/teams/{team-id}/workflows/{workflow-id}/statuses",
			ExpectedStatus: http.StatusOK,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				core.CreateProjectAndTasks(t, app, &team.Member)
				workflowID := findWorkflowID(t, app, team.Team.ID, "task")

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows/%s/statuses", team.Team.ID, workflowID)
				scenario.Body = apis.JsonToReader(t, stores.CreateWorkflowStatusDTO{
					Name:        "Code review",
					Description: types.Pointer("Ready for peer review."),
					Category:    string(models.TaskStatusInProgress),
					Color:       types.Pointer("#7c3aed"),
				})
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.WorkflowStatus](t, res.Body.Bytes())
				assert.Equal(t, "Code review", result.Name)
				assert.Equal(t, "code-review", result.Slug)
				assert.Equal(t, string(models.TaskStatusInProgress), result.Category)
				assert.Equal(t, float64(4000), result.Rank)
				assert.False(t, result.IsCompleted)
			},
		},
		{
			Name:            "fail: member cannot create workflow status",
			Method:          http.MethodPost,
			URL:             "/teams/{team-id}/workflows/{workflow-id}/statuses",
			ExpectedStatus:  http.StatusForbidden,
			ExpectedContent: []string{"You do not have the required team permission: workflow.manage"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				memberUser := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				core.CreateTeamMemberWithOptions(t, app, team.Team.ID, memberUser.User.ID, core.TeamWithRole(models.TeamMemberRoleMember))
				core.CreateProjectAndTasks(t, app, &team.Member)
				workflowID := findWorkflowID(t, app, team.Team.ID, "task")

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, memberUser.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows/%s/statuses", team.Team.ID, workflowID)
				scenario.Body = apis.JsonToReader(t, stores.CreateWorkflowStatusDTO{
					Name:     "Blocked",
					Category: string(models.TaskStatusInProgress),
				})
			},
		},
		{
			Name:            "fail: cannot create workflow status for another team workflow",
			Method:          http.MethodPost,
			URL:             "/teams/{team-id}/workflows/{workflow-id}/statuses",
			ExpectedStatus:  http.StatusNotFound,
			ExpectedContent: []string{"workflow not found"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				otherOwner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				otherTeam := core.CreateTeamAndMemberWithOptions(t, app, &otherOwner.User)
				core.CreateProjectAndTasks(t, app, &team.Member)
				core.CreateProjectAndTasks(t, app, &otherTeam.Member)
				otherWorkflowID := findWorkflowID(t, app, otherTeam.Team.ID, "task")

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows/%s/statuses", team.Team.ID, otherWorkflowID)
				scenario.Body = apis.JsonToReader(t, stores.CreateWorkflowStatusDTO{
					Name:     "Blocked",
					Category: string(models.TaskStatusInProgress),
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

func TestApi_TeamWorkflowStatusUpdate(t *testing.T) {
	tests := []apis.ApiScenario{
		{
			Name:           "success: owner can update workflow status",
			Method:         http.MethodPut,
			URL:            "/teams/{team-id}/workflows/{workflow-id}/statuses/{workflow-status-id}",
			ExpectedStatus: http.StatusOK,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				core.CreateProjectAndTasks(t, app, &team.Member)
				workflowID := findWorkflowID(t, app, team.Team.ID, "task")
				statusID := findWorkflowStatusID(t, app, team.Team.ID, "task", "in_progress")
				rank := 1500.0

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows/%s/statuses/%s", team.Team.ID, workflowID, statusID)
				scenario.Body = apis.JsonToReader(t, stores.UpdateWorkflowStatusDTO{
					Name:     types.Pointer("Doing"),
					Slug:     types.Pointer("doing"),
					Category: types.Pointer(string(models.TaskStatusInProgress)),
					Rank:     &rank,
				})
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.WorkflowStatus](t, res.Body.Bytes())
				assert.Equal(t, "Doing", result.Name)
				assert.Equal(t, "doing", result.Slug)
				assert.Equal(t, float64(1500), result.Rank)
			},
		},
		{
			Name:            "fail: member cannot update workflow status",
			Method:          http.MethodPut,
			URL:             "/teams/{team-id}/workflows/{workflow-id}/statuses/{workflow-status-id}",
			ExpectedStatus:  http.StatusForbidden,
			ExpectedContent: []string{"You do not have the required team permission: workflow.manage"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				memberUser := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				core.CreateTeamMemberWithOptions(t, app, team.Team.ID, memberUser.User.ID, core.TeamWithRole(models.TeamMemberRoleMember))
				core.CreateProjectAndTasks(t, app, &team.Member)
				workflowID := findWorkflowID(t, app, team.Team.ID, "task")
				statusID := findWorkflowStatusID(t, app, team.Team.ID, "task", "in_progress")

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, memberUser.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows/%s/statuses/%s", team.Team.ID, workflowID, statusID)
				scenario.Body = apis.JsonToReader(t, stores.UpdateWorkflowStatusDTO{
					Name: types.Pointer("Doing"),
				})
			},
		},
		{
			Name:            "fail: cannot remove required category from default workflow",
			Method:          http.MethodPut,
			URL:             "/teams/{team-id}/workflows/{workflow-id}/statuses/{workflow-status-id}",
			ExpectedStatus:  http.StatusConflict,
			ExpectedContent: []string{"workflow must keep statuses for todo, in_progress, and done while in use"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				workflowID, statusesByCategory := createAssignableTaskWorkflow(t, app, team.Team.ID, team.Member.ID)
				_, err := app.Adapter().Task().SetDefaultWorkflow(t.Context(), workflowID)
				assert.NoError(t, err)

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf(
					"/teams/%s/workflows/%s/statuses/%s",
					team.Team.ID,
					workflowID,
					statusesByCategory[string(models.TaskStatusTodo)],
				)
				scenario.Body = apis.JsonToReader(t, stores.UpdateWorkflowStatusDTO{
					Category: types.Pointer(string(models.TaskStatusInProgress)),
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

func TestApi_TeamWorkflowStatusDelete(t *testing.T) {
	tests := []apis.ApiScenario{
		{
			Name:           "success: owner can delete unused workflow status",
			Method:         http.MethodDelete,
			URL:            "/teams/{team-id}/workflows/{workflow-id}/statuses/{workflow-status-id}",
			ExpectedStatus: http.StatusNoContent,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				core.CreateProjectAndTasks(t, app, &team.Member)
				workflowID := findWorkflowID(t, app, team.Team.ID, "task")
				status, err := app.Adapter().Task().CreateWorkflowStatus(t.Context(), workflowID, &stores.CreateWorkflowStatusDTO{
					Name:     "Blocked",
					Category: string(models.TaskStatusInProgress),
				})
				assert.NoError(t, err)

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows/%s/statuses/%s", team.Team.ID, workflowID, status.ID)
				scenario.Store.Set("workflow_status_id", status.ID)
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				statusID := scenario.Store.Get("workflow_status_id").(uuid.UUID)
				status, err := app.Adapter().Task().FindWorkflowStatusByID(t.Context(), statusID)
				assert.NoError(t, err)
				assert.Nil(t, status)
			},
		},
		{
			Name:            "fail: cannot delete workflow status in use",
			Method:          http.MethodDelete,
			URL:             "/teams/{team-id}/workflows/{workflow-id}/statuses/{workflow-status-id}",
			ExpectedStatus:  http.StatusConflict,
			ExpectedContent: []string{"workflow status is in use"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				core.CreateProjectAndTasks(t, app, &team.Member, core.WithTaskByCountAndStatus(1, models.TaskStatusTodo))
				workflowID := findWorkflowID(t, app, team.Team.ID, "task")
				statusID := findWorkflowStatusID(t, app, team.Team.ID, "task", "todo")

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows/%s/statuses/%s", team.Team.ID, workflowID, statusID)
			},
		},
		{
			Name:            "fail: cannot remove required category from default workflow",
			Method:          http.MethodDelete,
			URL:             "/teams/{team-id}/workflows/{workflow-id}/statuses/{workflow-status-id}",
			ExpectedStatus:  http.StatusConflict,
			ExpectedContent: []string{"workflow must keep statuses for todo, in_progress, and done while in use"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				workflowID, statusesByCategory := createAssignableTaskWorkflow(t, app, team.Team.ID, team.Member.ID)
				_, err := app.Adapter().Task().SetDefaultWorkflow(t.Context(), workflowID)
				assert.NoError(t, err)

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf(
					"/teams/%s/workflows/%s/statuses/%s",
					team.Team.ID,
					workflowID,
					statusesByCategory[string(models.TaskStatusDone)],
				)
			},
		},
		{
			Name:            "fail: member cannot delete workflow status",
			Method:          http.MethodDelete,
			URL:             "/teams/{team-id}/workflows/{workflow-id}/statuses/{workflow-status-id}",
			ExpectedStatus:  http.StatusForbidden,
			ExpectedContent: []string{"You do not have the required team permission: workflow.manage"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				memberUser := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				core.CreateTeamMemberWithOptions(t, app, team.Team.ID, memberUser.User.ID, core.TeamWithRole(models.TeamMemberRoleMember))
				core.CreateProjectAndTasks(t, app, &team.Member)
				workflowID := findWorkflowID(t, app, team.Team.ID, "task")
				statusID := findWorkflowStatusID(t, app, team.Team.ID, "task", "todo")

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, memberUser.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/teams/%s/workflows/%s/statuses/%s", team.Team.ID, workflowID, statusID)
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

func findWorkflowID(t testing.TB, app *core.BaseApp, teamID uuid.UUID, appliesTo string) uuid.UUID {
	t.Helper()
	workflows, err := app.Adapter().Task().ListWorkflows(t.Context(), &stores.WorkflowFilter{
		TeamIds:   []uuid.UUID{teamID},
		AppliesTo: []string{appliesTo},
	})
	if err != nil {
		t.Fatalf("list workflows: %v", err)
	}
	if len(workflows) == 0 {
		t.Fatalf("workflow not found: %s", appliesTo)
	}
	return workflows[0].ID
}
