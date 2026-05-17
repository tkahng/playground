package apis_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/populator"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/types"
)

func TestApi_TeamTaskList(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
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
		tests := []apis.ApiScenario{
			{
				Name:           "success: all tasks",
				Method:         http.MethodGet,
				URL:            "/task-projects/{task-project-id}/tasks",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/task-projects/%s/tasks", project1.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.Task]](t, res.Body.Bytes())
					assert.Equal(t, int64(9), result.Meta.Total)
				},
			},
			{
				Name:           "success: all tasks with done status filter",
				Method:         http.MethodGet,
				URL:            "/task-projects/{task-project-id}/tasks?status=done",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/task-projects/%s/tasks?status=done", project1.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.Task]](t, res.Body.Bytes())
					assert.Equal(t, int64(3), result.Meta.Total)
				},
			},
			{
				Name:           "success: tasks with workflow status filter",
				Method:         http.MethodGet,
				URL:            "/task-projects/{task-project-id}/tasks?workflow_status_ids={workflow-status-id}",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					doneStatusID := findWorkflowStatusID(t, app, team1.Team.ID, "task", "done")
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf(
						"/task-projects/%s/tasks?workflow_status_ids=%s",
						project1.ID.String(),
						doneStatusID.String(),
					)
					scenario.Store.Set("workflow_status_id", doneStatusID)
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.Task]](t, res.Body.Bytes())
					doneStatusID := scenario.Store.Get("workflow_status_id").(uuid.UUID)
					assert.Equal(t, int64(3), result.Meta.Total)
					for _, task := range result.Data {
						require.NotNil(t, task.WorkflowStatusID)
						assert.Equal(t, doneStatusID, *task.WorkflowStatusID)
						require.NotNil(t, task.WorkflowStatus)
						assert.Equal(t, doneStatusID, task.WorkflowStatus.ID)
						assert.Equal(t, string(models.TaskStatusDone), task.WorkflowStatus.Category)
					}
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

func TestApi_TeamTaskUpdate(t *testing.T) {
	tests := []apis.ApiScenario{
		{
			Name:           "success: update task",
			Method:         http.MethodPut,
			URL:            "/tasks/{task-id}",
			ExpectedStatus: http.StatusNoContent,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
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
					StartAt:     types.Pointer(time.Now().UTC()),
					EndAt:       types.Pointer(time.Now().UTC()),
					AssigneeID:  &team1.Member.ID,
					ReporterID:  &team1.Member.ID,
				}
				scenario.Store.Set("input", input)
				scenario.Body = apis.JsonToReader(t, input)
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := repository.MustFindOneCtx(t, t.Context(), repository.Task, app.Db(), nil)
				input, ok := scenario.Store.Get("input").(stores.UpdateTaskDto)
				if !ok {
					t.Fatal("no input")
				}
				assert.Equal(t, input.Name, result.Name)
				assert.Equal(t, input.Description, result.Description)
				assert.Equal(t, input.Status, result.Status)
				assert.Equal(
					t,
					input.StartAt.UTC().Truncate(time.Microsecond),
					result.StartAt.UTC().Truncate(time.Microsecond))
				assert.Equal(
					t,
					input.EndAt.UTC().Truncate(time.Microsecond),
					result.EndAt.UTC().Truncate(time.Microsecond),
				)
				assert.Equal(t, input.AssigneeID, result.AssigneeID)
				assert.Equal(t, input.ReporterID, result.ReporterID)
			},
		},
		{
			Name:           "success: update task with workflow status",
			Method:         http.MethodPut,
			URL:            "/tasks/{task-id}",
			ExpectedStatus: http.StatusNoContent,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team1 := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				project := core.CreateProjectAndTasks(t, app, &team1.Member)
				task := repository.MustCreateOneCtx(t, t.Context(), repository.Task, app.Db(), &models.Task{
					TeamID:    team1.Team.ID,
					ProjectID: project.ID,
					Status:    models.TaskStatusTodo,
					Rank:      0,
				})
				doneStatusID := findWorkflowStatusID(t, app, team1.Team.ID, "task", "done")
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/tasks/%s", task.ID.String())
				input := stores.UpdateTaskDto{
					Name:             "updated with workflow status",
					Status:           models.TaskStatusTodo,
					WorkflowStatusID: &doneStatusID,
				}
				scenario.Store.Set("workflow_status_id", doneStatusID)
				scenario.Body = apis.JsonToReader(t, input)
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := repository.MustFindOneCtx(t, t.Context(), repository.Task, app.Db(), nil)
				doneStatusID := scenario.Store.Get("workflow_status_id").(uuid.UUID)
				assert.Equal(t, models.TaskStatusDone, result.Status)
				assert.Equal(t, doneStatusID, *result.WorkflowStatusID)
			},
		},
		{
			Name:           "success: update task without status preserves workflow status",
			Method:         http.MethodPut,
			URL:            "/tasks/{task-id}",
			ExpectedStatus: http.StatusNoContent,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team1 := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				project := core.CreateProjectAndTasks(t, app, &team1.Member)
				task, err := app.Adapter().Task().CreateTask(t.Context(), &models.Task{
					TeamID:    team1.Team.ID,
					ProjectID: project.ID,
					Status:    models.TaskStatusInProgress,
					Rank:      1000,
				})
				require.NoError(t, err)
				require.NotNil(t, task.WorkflowStatusID)

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/tasks/%s", task.ID.String())
				scenario.Store.Set("task_id", task.ID)
				scenario.Store.Set("workflow_status_id", *task.WorkflowStatusID)
				scenario.Body = apis.JsonToReader(t, stores.UpdateTaskDto{
					Name: "renamed without status",
				})
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				taskID := scenario.Store.Get("task_id").(uuid.UUID)
				workflowStatusID := scenario.Store.Get("workflow_status_id").(uuid.UUID)
				result, err := app.Adapter().Task().FindTaskByID(t.Context(), taskID)
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, "renamed without status", result.Name)
				assert.Equal(t, models.TaskStatusInProgress, result.Status)
				require.NotNil(t, result.WorkflowStatusID)
				assert.Equal(t, workflowStatusID, *result.WorkflowStatusID)
			},
		},
		{
			Name:            "fail: update task rejects assignee from another team",
			Method:          http.MethodPut,
			URL:             "/tasks/{task-id}",
			ExpectedStatus:  http.StatusBadRequest,
			ExpectedContent: []string{"assignee must be an active member of the task team"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team1 := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				otherOwner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				team2 := core.CreateTeamAndMemberWithOptions(t, app, &otherOwner.User)
				project := core.CreateProjectAndTasks(t, app, &team1.Member)
				task := repository.MustCreateOneCtx(t, t.Context(), repository.Task, app.Db(), &models.Task{
					TeamID:            team1.Team.ID,
					ProjectID:         project.ID,
					Status:            models.TaskStatusTodo,
					Rank:              0,
					CreatedByMemberID: &team1.Member.ID,
				})
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/tasks/%s", task.ID.String())
				scenario.Body = apis.JsonToReader(t, stores.UpdateTaskDto{
					Name:       "updated",
					Status:     models.TaskStatusInProgress,
					AssigneeID: &team2.Member.ID,
				})
			},
		},
		{
			Name:            "fail: update task rejects parent from another project",
			Method:          http.MethodPut,
			URL:             "/tasks/{task-id}",
			ExpectedStatus:  http.StatusBadRequest,
			ExpectedContent: []string{"parent task must belong to the same task project"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team1 := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				project := core.CreateProjectAndTasks(t, app, &team1.Member)
				otherProject := core.CreateProjectAndTasks(t, app, &team1.Member)
				parent := repository.MustCreateOneCtx(t, t.Context(), repository.Task, app.Db(), &models.Task{
					TeamID:            team1.Team.ID,
					ProjectID:         otherProject.ID,
					Status:            models.TaskStatusTodo,
					Rank:              0,
					CreatedByMemberID: &team1.Member.ID,
				})
				task := repository.MustCreateOneCtx(t, t.Context(), repository.Task, app.Db(), &models.Task{
					TeamID:            team1.Team.ID,
					ProjectID:         project.ID,
					Status:            models.TaskStatusTodo,
					Rank:              0,
					CreatedByMemberID: &team1.Member.ID,
				})
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/tasks/%s", task.ID.String())
				scenario.Body = apis.JsonToReader(t, stores.UpdateTaskDto{
					Name:     "updated",
					Status:   models.TaskStatusInProgress,
					ParentID: &parent.ID,
				})
			},
		},
		{
			Name:            "fail: update task rejects workflow status from another team",
			Method:          http.MethodPut,
			URL:             "/tasks/{task-id}",
			ExpectedStatus:  http.StatusBadRequest,
			ExpectedContent: []string{"workflow status"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team1 := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				otherOwner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow(), core.UserWithEmail(randomEmail()))
				team2 := core.CreateTeamAndMemberWithOptions(t, app, &otherOwner.User)
				// create a project for team2 to trigger default workflow creation
				core.CreateProjectAndTasks(t, app, &team2.Member)
				project := core.CreateProjectAndTasks(t, app, &team1.Member)
				task := repository.MustCreateOneCtx(t, t.Context(), repository.Task, app.Db(), &models.Task{
					TeamID:            team1.Team.ID,
					ProjectID:         project.ID,
					Status:            models.TaskStatusTodo,
					Rank:              0,
					CreatedByMemberID: &team1.Member.ID,
				})
				// get a workflow status from team2's workflow
				otherTeamStatusID := findWorkflowStatusID(t, app, team2.Team.ID, "task", "done")
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/tasks/%s", task.ID.String())
				scenario.Body = apis.JsonToReader(t, stores.UpdateTaskDto{
					Name:             "updated",
					Status:           models.TaskStatusDone,
					WorkflowStatusID: &otherTeamStatusID,
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

func TestApi_TeamTaskPositionStatus(t *testing.T) {
	tests := []apis.ApiScenario{
		{
			Name:           "success: move task with workflow status",
			Method:         http.MethodPut,
			URL:            "/tasks/{task-id}/position-status",
			ExpectedStatus: http.StatusNoContent,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				workflowID, statusesByCategory := createAssignableTaskWorkflow(t, app, team.Team.ID, team.Member.ID)
				project, err := app.Adapter().Task().CreateTaskProject(t.Context(), &stores.CreateTaskProjectDTO{
					TeamID:     team.Team.ID,
					MemberID:   team.Member.ID,
					Name:       "Custom workflow project",
					Status:     models.TaskProjectStatusTodo,
					WorkflowID: &workflowID,
				})
				assert.NoError(t, err)
				task, err := app.Adapter().Task().CreateTask(t.Context(), &models.Task{
					TeamID:            team.Team.ID,
					ProjectID:         project.ID,
					CreatedByMemberID: &team.Member.ID,
					Name:              "Move me",
					Status:            models.TaskStatusTodo,
					Rank:              1000,
				})
				assert.NoError(t, err)
				reviewStatusID := statusesByCategory[string(models.TaskStatusInProgress)]

				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/tasks/%s/position-status", task.ID)
				scenario.Store.Set("task_id", task.ID)
				scenario.Store.Set("workflow_status_id", reviewStatusID)
				scenario.Body = apis.JsonToReader(t, apis.TaskPositionStatusDTO{
					Position:         0,
					WorkflowStatusID: &reviewStatusID,
				})
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				taskID := scenario.Store.Get("task_id").(uuid.UUID)
				workflowStatusID := scenario.Store.Get("workflow_status_id").(uuid.UUID)
				task, err := app.Adapter().Task().FindTaskByID(t.Context(), taskID)
				assert.NoError(t, err)
				require.NotNil(t, task)
				assert.Equal(t, models.TaskStatusInProgress, task.Status)
				require.NotNil(t, task.WorkflowStatusID)
				assert.Equal(t, workflowStatusID, *task.WorkflowStatusID)
			},
		},
		{
			Name:            "fail: position-status with no status or workflow_status_id returns 422",
			Method:          http.MethodPut,
			URL:             "/tasks/{task-id}/position-status",
			ExpectedStatus:  http.StatusUnprocessableEntity,
			ExpectedContent: []string{"status or workflow_status_id is required"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				project := core.CreateProjectAndTasks(t, app, &team.Member)
				task, err := app.Adapter().Task().CreateTask(t.Context(), &models.Task{
					TeamID:    team.Team.ID,
					ProjectID: project.ID,
					Status:    models.TaskStatusTodo,
					Rank:      1000,
				})
				require.NoError(t, err)
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/tasks/%s/position-status", task.ID)
				scenario.Body = apis.JsonToReader(t, apis.TaskPositionStatusDTO{Position: 0})
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

func TestApi_TeamTaskGet(t *testing.T) {
	tests := []apis.ApiScenario{
		{
			Name:           "success: get task",
			Method:         http.MethodGet,
			URL:            "/tasks/{task-id}",
			ExpectedStatus: http.StatusOK,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				owner := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team1 := core.CreateTeamAndMemberWithOptions(t, app, &owner.User)
				project := core.CreateProjectAndTasks(t, app, &team1.Member)
				task2 := repository.MustCreateOneCtx(t, t.Context(), repository.Task, app.Db(), &models.Task{
					TeamID:            team1.Team.ID,
					ProjectID:         project.ID,
					Status:            models.TaskStatusTodo,
					Rank:              0,
					CreatedByMemberID: &team1.Member.ID,
				})
				task := repository.MustCreateOneCtx(t, t.Context(), repository.Task, app.Db(), &models.Task{
					TeamID:            team1.Team.ID,
					ProjectID:         project.ID,
					Status:            models.TaskStatusTodo,
					Rank:              0,
					CreatedByMemberID: &team1.Member.ID,
					ParentID:          &task2.ID,
				})
				pop := populator.New(app.Adapter())
				err := populator.PopulateTask(t.Context(), pop, task)
				if err != nil {
					t.Fatal("failed to populate task", err)
				}
				scenario.Store.Set("task", task)
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.URL = fmt.Sprintf("/tasks/%s", task.ID.String())
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.Task](t, res.Body.Bytes())
				input, ok := scenario.Store.Get("task").(*models.Task)
				if !ok {
					t.Fatal("no input")
				}
				assert.Equal(t, input.Name, result.Name)
				assert.Equal(t, input.Description, result.Description)
				assert.Equal(t, input.Status, result.Status)
				assert.Equal(t, input.AssigneeID, result.AssigneeID)
				assert.Equal(t, input.ReporterID, result.ReporterID)
				assert.Equal(t, input.TeamID, result.TeamID)
				assert.Equal(t, input.ProjectID, result.ProjectID)
				assert.Equal(t, input.Rank, result.Rank)
				assert.Equal(t, input.CreatedByMemberID, result.CreatedByMemberID)
				assert.NotNil(t, result.CreatedByMember)
				assert.NotNil(t, result.Team)
				assert.NotNil(t, result.Project)
				assert.NotNil(t, result.Parent)
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
