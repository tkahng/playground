package apis

import (
	"context"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

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

func (api *Api) TaskProjectUpdateOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "task-project-update",
		Method:      http.MethodPut,
		Path:        path,
		Summary:     "Task project update",
		Description: "Update a task project",
		Tags:        []string{"Task"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type TaskProjectResponse struct {
	Body *shared.TaskProject
}

func (api *Api) TaskProjectUpdate(ctx context.Context, input *shared.UpdateTaskProjectDTO) (*struct{}, error) {
	userInfo := core.GetContextUserClaims(ctx)
	if userInfo == nil || userInfo.User == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	// db := api.app.Db()
	// // err := repository.task(ctx, db, input.Body)
	// if err != nil {
	// 	return nil, err
	// }
	return nil, nil
}

func (api *Api) TaskProjectDeleteOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "task-project-delete",
		Method:      http.MethodDelete,
		Path:        path,
		Summary:     "Task project delete",
		Description: "Delete a task project",
		Tags:        []string{"Task"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) TaskProjectDelete(ctx context.Context, input *struct {
	TaskProjectID string `path:"task-project-id"`
}) (*struct{}, error) {
	userInfo := core.GetContextUserClaims(ctx)
	if userInfo == nil || userInfo.User == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	db := api.app.Db()
	id, err := uuid.Parse(input.TaskProjectID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task project id")
	}
	err = repository.DeleteTaskProject(ctx, db, id)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) TaskProjectGetOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "task-project-get",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Task project get",
		Description: "Get a task project",
		Tags:        []string{"Task"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) TaskProjectGet(ctx context.Context, input *struct {
	TaskProjectID string `path:"task-project-id"`
}) (*TaskProjectResponse, error) {
	userInfo := core.GetContextUserClaims(ctx)
	if userInfo == nil || userInfo.User == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	db := api.app.Db()
	id, err := uuid.Parse(input.TaskProjectID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task project id")
	}
	taskProject, err := repository.FindTaskProjectByID(ctx, db, id)
	if err != nil {
		return nil, err
	}
	return &TaskProjectResponse{
		Body: shared.ModelToProject(taskProject),
	}, nil
}

func (api *Api) TaskProjectTasksCreateOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "task-project-tasks-create",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Task project tasks create",
		Description: "Create a new task project task",
		Tags:        []string{"Task"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) TaskProjectTasksCreate(ctx context.Context, input *shared.CreateTaskWithProjectIdInput) (*TaskResposne, error) {
	userInfo := core.GetContextUserClaims(ctx)
	if userInfo == nil || userInfo.User == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	db := api.app.Db()
	id, err := uuid.Parse(input.TaskProjectID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task project id")
	}
	payload := input.Body
	order, err := repository.FindLastTaskOrder(ctx, db, id)
	if err != nil {
		return nil, err
	}
	payload.Order = order
	task, err := repository.CreateTaskWithChildren(ctx, db, userInfo.User.ID, id, &payload)
	if err != nil {
		return nil, err
	}
	return &TaskResposne{
		Body: shared.ModelToTask(task),
	}, nil
}
