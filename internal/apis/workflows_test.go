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
