package apis

import (
	"context"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

func (api *Api) TaskListOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "task-list",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Task list",
		Description: "List of tasks",
		Tags:        []string{"Task"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type TaskListResponse struct {
	Body *shared.PaginatedResponse[*shared.TaskWithSubtask]
}

func (api *Api) TaskList(ctx context.Context, input *shared.TaskListParams) (*TaskListResponse, error) {
	db := api.app.Db()
	tasks, err := repository.ListTasks(ctx, db, input)
	if err != nil {
		return nil, huma.Error500InternalServerError("error listing tasks", err)
	}
	total, err := repository.CountTasks(ctx, db, &input.TaskListFilter)
	if err != nil {
		return nil, huma.Error500InternalServerError("error counting tasks", err)
	}
	return &TaskListResponse{
		Body: &shared.PaginatedResponse[*shared.TaskWithSubtask]{
			Data: mapper.Map(tasks, func(task *models.Task) *shared.TaskWithSubtask {
				return &shared.TaskWithSubtask{
					Task: shared.ModelToTask(task),
					Children: mapper.Map(task.R.ReverseParents, func(child *models.Task) *shared.Task {
						return shared.ModelToTask(child)
					}),
				}
			}),
			Meta: shared.Meta{Total: int(total)},
		},
	}, nil
}

func (api *Api) TaskCreateOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "task-create",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Create task",
		Description: "Create a new task",
		Tags:        []string{"Task"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type TaskResposne struct {
	Body *shared.Task
}

func (api *Api) TaskCreate(ctx context.Context, input *shared.CreateTaskInput) (*TaskResposne, error) {
	db := api.app.Db()
	userInfo := core.GetContextUserClaims(ctx)
	if userInfo == nil || userInfo.User == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	task, err := repository.CreateTaskWithChildren(ctx, db, userInfo.User.ID, input.TaskProjectID, &input.CreateTaskWithChildrenDTO)
	if err != nil {
		return nil, err
	}
	return &TaskResposne{
		Body: shared.ModelToTask(task),
	}, nil
}

func (api *Api) TaskProjectListOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "task-project-list",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Task project list",
		Description: "List of task projects",
		Tags:        []string{"Task"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type TaskProjectListResponse struct {
	Body *shared.PaginatedResponse[*shared.TaskProjectWithTasks]
}

func (api *Api) TaskProjectList(ctx context.Context, input *shared.TaskProjectsListParams) (*TaskProjectListResponse, error) {
	db := api.app.Db()
	taskProject, err := repository.ListTaskProjects(ctx, db, input)
	if err != nil {
		return nil, err
	}
	total, err := repository.CountTaskProjects(ctx, db, &input.TaskProjectsListFilter)
	if err != nil {
		return nil, err
	}
	if input.Expand != nil && slices.Contains(input.Expand, "tasks") {
		if slices.Contains(input.Expand, "subtasks") {
			err = taskProject.LoadTaskProjectProjectTasks(ctx, db, models.ThenLoadTaskReverseParents())
			if err != nil {
				return nil, err
			}
		} else {
			err = taskProject.LoadTaskProjectProjectTasks(ctx, db)
			if err != nil {
				return nil, err
			}
		}
	}
	return &TaskProjectListResponse{
		Body: &shared.PaginatedResponse[*shared.TaskProjectWithTasks]{
			Data: mapper.Map(taskProject, func(taskProject *models.TaskProject) *shared.TaskProjectWithTasks {
				return &shared.TaskProjectWithTasks{
					TaskProject: shared.ModelToProject(taskProject),
					Tasks: mapper.Map(taskProject.R.ProjectTasks, func(task *models.Task) *shared.TaskWithSubtask {
						return &shared.TaskWithSubtask{
							Task: shared.ModelToTask(task),
							Children: mapper.Map(task.R.ReverseParents, func(child *models.Task) *shared.Task {
								return shared.ModelToTask(child)
							}),
						}
					}),
				}
			}),
			Meta: shared.Meta{Total: int(total)},
		},
	}, nil
}

func (api *Api) TaskProjectCreateOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "task-project-create",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Task project create",
		Description: "Create a new task project",
		Tags:        []string{"Task"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) TaskProjectCreate(ctx context.Context, input *struct {
	Body *shared.CreateTaskProjectWithTasksDTO
}) (*struct {
	Body *shared.TaskProject
}, error) {
	userInfo := core.GetContextUserClaims(ctx)
	if userInfo == nil || userInfo.User == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	db := api.app.Db()
	// db, err := pool.Begin(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	// defer tx.Rollback(ctx)
	// db := db.NewDBTx(tx)
	taskProject, err := repository.CreateTaskProjectWithTasks(ctx, db, userInfo.User.ID, input.Body)
	if err != nil {
		return nil, err
	}
	// err = tx.Commit(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	// err = taskProject.LoadTaskProjectProjectTasks(ctx, db)
	// if err != nil {
	// 	return nil, err
	// }
	return &struct {
		Body *shared.TaskProject
	}{
		Body: shared.ModelToProject(taskProject),
	}, nil
}

func (api *Api) TaskUpdateOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "task-update",
		Method:      http.MethodPut,
		Path:        path,
		Summary:     "Task update",
		Description: "Update a task",
		Tags:        []string{"Task"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) TaskUpdate(ctx context.Context, input *shared.UpdateTaskDTO) (*struct{}, error) {
	db := api.app.Db()
	err := repository.UpdateTask(ctx, db, input)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
