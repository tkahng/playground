package apis_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/workers"
)

func TestApi_ProjectUpdate_EnquesStatusChangedJob(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		owner := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerifiedNow())
		team := core.CreateTeamAndMemberWithOptions(t, testApi.App, &owner.User)
		project := core.CreateProjectAndTasks(t, testApi.App, &team.Member)

		tests := []apis.ApiScenario{
			{
				Name:           "success: todo→in_progress enqueues project_status_changed",
				Method:         http.MethodPut,
				URL:            fmt.Sprintf("/task-projects/%s", project.ID),
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
					scenario.Headers = []string{header}
					scenario.Body = apis.JsonToReader(t, stores.UpdateTaskProjectBaseDTO{
						Name:   project.Name,
						Status: models.TaskProjectStatusInProgress,
					})
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, _ *httptest.ResponseRecorder) {
					assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "project_status_changed"))
				},
			},
			{
				Name:           "success: no status change enqueues no new job",
				Method:         http.MethodPut,
				URL:            fmt.Sprintf("/task-projects/%s", project.ID),
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
					scenario.Headers = []string{header}
					scenario.Body = apis.JsonToReader(t, stores.UpdateTaskProjectBaseDTO{
						Name:   "renamed but same status",
						Status: models.TaskProjectStatusInProgress,
					})
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, _ *httptest.ResponseRecorder) {
					// unique key upsert — still 1
					assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "project_status_changed"))
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

func TestApi_ProjectUpdate_UsesWorkflowStatusInStatusChangedJob(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		owner := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerifiedNow())
		team := core.CreateTeamAndMemberWithOptions(t, testApi.App, &owner.User)
		project := core.CreateProjectAndTasks(t, testApi.App, &team.Member)
		workflowID := findWorkflowID(t, testApi.App, team.Team.ID, "project")
		statusSlug := "waiting-on-budget"
		status, err := testApi.App.Adapter().Task().CreateWorkflowStatus(ctx, workflowID, &stores.CreateWorkflowStatusDTO{
			Name:     "Waiting on budget",
			Slug:     &statusSlug,
			Category: string(models.TaskStatusInProgress),
		})
		require.NoError(t, err)

		scenario := apis.ApiScenario{
			Name:           "success: workflow status slug is used in project_status_changed job",
			Method:         http.MethodPut,
			URL:            fmt.Sprintf("/task-projects/%s", project.ID),
			ExpectedStatus: http.StatusNoContent,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
				scenario.Headers = []string{header}
				scenario.Body = apis.JsonToReader(t, stores.UpdateTaskProjectBaseDTO{
					Name:             project.Name,
					WorkflowStatusID: &status.ID,
				})
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, _ *httptest.ResponseRecorder) {
				assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "project_status_changed"))
				job, err := repository.Job.GetOne(ctx, db, &map[string]any{
					"kind": map[string]any{"_eq": "project_status_changed"},
				})
				require.NoError(t, err)
				require.NotNil(t, job)
				var args workers.ProjectStatusChangedJobArgs
				require.NoError(t, json.Unmarshal(job.Payload, &args))
				assert.Equal(t, "todo", args.OldStatus)
				assert.Equal(t, statusSlug, args.NewStatus)
			},
		}
		scenario.Test(t)
	})
}
