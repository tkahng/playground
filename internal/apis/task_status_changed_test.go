//go:build integration

package apis_test

import (
	"context"
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

func TestApi_TaskUpdate_UsesWorkflowStatusCompletion(t *testing.T) {
	tests := []struct {
		name                  string
		statusName            string
		statusSlug            string
		category              string
		isCompleted           bool
		expectedTaskStatus    models.TaskStatus
		expectedChangedJobs   int64
		expectedCompletedJobs int64
	}{
		{
			name:                  "success: completed workflow status enqueues task_completed",
			statusName:            "Ready for release",
			statusSlug:            "ready-for-release",
			category:              string(models.TaskStatusInProgress),
			isCompleted:           true,
			expectedTaskStatus:    models.TaskStatusInProgress,
			expectedChangedJobs:   0,
			expectedCompletedJobs: 1,
		},
		{
			name:                  "success: non-completed done category enqueues status_changed",
			statusName:            "Done pending review",
			statusSlug:            "done-pending-review",
			category:              string(models.TaskStatusDone),
			isCompleted:           false,
			expectedTaskStatus:    models.TaskStatusDone,
			expectedChangedJobs:   1,
			expectedCompletedJobs: 0,
		},
	}
	for _, tt := range tests {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			testApi := apis.SetupApi(t, ctx, db)
			owner := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerifiedNow())
			team := core.CreateTeamAndMemberWithOptions(t, testApi.App, &owner.User)
			workflowID, _ := createAssignableTaskWorkflow(t, testApi.App, team.Team.ID, team.Member.ID)
			status, err := testApi.App.Adapter().Task().CreateWorkflowStatus(ctx, workflowID, &stores.CreateWorkflowStatusDTO{
				Name:        tt.statusName,
				Slug:        &tt.statusSlug,
				Category:    tt.category,
				IsCompleted: &tt.isCompleted,
			})
			require.NoError(t, err)
			project, err := testApi.App.Adapter().Task().CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
				TeamID:     team.Team.ID,
				MemberID:   team.Member.ID,
				Name:       "Workflow completion project",
				Status:     models.TaskProjectStatusTodo,
				WorkflowID: &workflowID,
			})
			require.NoError(t, err)
			task, err := testApi.App.Adapter().Task().CreateTask(ctx, &models.Task{
				TeamID:            team.Team.ID,
				ProjectID:         project.ID,
				CreatedByMemberID: &team.Member.ID,
				Name:              "Workflow completion task",
				Status:            models.TaskStatusTodo,
				Rank:              1000,
			})
			require.NoError(t, err)

			scenario := apis.ApiScenario{
				Name:           tt.name,
				Method:         http.MethodPut,
				URL:            fmt.Sprintf("/tasks/%s", task.ID),
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
					scenario.Headers = []string{header}
					scenario.Body = apis.JsonToReader(t, stores.UpdateTaskDto{
						Name:             "updated workflow completion task",
						WorkflowStatusID: &status.ID,
					})
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, _ *httptest.ResponseRecorder) {
					assert.Equal(t, tt.expectedChangedJobs, countJobsByKind(t, ctx, db, "task_status_changed"))
					assert.Equal(t, tt.expectedCompletedJobs, countJobsByKind(t, ctx, db, "task_completed"))
					updated, err := app.Adapter().Task().FindTaskByID(ctx, task.ID)
					require.NoError(t, err)
					require.NotNil(t, updated)
					assert.Equal(t, tt.expectedTaskStatus, updated.Status)
					require.NotNil(t, updated.WorkflowStatusID)
					assert.Equal(t, status.ID, *updated.WorkflowStatusID)
				},
			}
			scenario.Test(t)
		})
	}
}
