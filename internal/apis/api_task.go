package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

type TaskListResponse struct {
	Body *shared.PaginatedResponse[*shared.TaskWithSubtask]
}

func (api *Api) TaskList(ctx context.Context, input *shared.TaskListParams) (*TaskListResponse, error) {
	db := api.app.Db()
	tasks, err := queries.ListTasks(ctx, db, input)
	if err != nil {
		return nil, huma.Error500InternalServerError("error listing tasks", err)
	}
	total, err := queries.CountTasks(ctx, db, &input.TaskListFilter)
	if err != nil {
		return nil, huma.Error500InternalServerError("error counting tasks", err)
	}
	return &TaskListResponse{
		Body: &shared.PaginatedResponse[*shared.TaskWithSubtask]{
			Data: mapper.Map(tasks, func(task *models.Task) *shared.TaskWithSubtask {
				return &shared.TaskWithSubtask{
					Task: shared.CrudModelToTask(task),
					Children: mapper.Map(task.Children, func(child *models.Task) *shared.Task {
						return shared.CrudModelToTask(child)
					}),
				}
			}),
			Meta: shared.GenerateMeta(input.PaginatedInput, total),
		},
	}, nil
}

type TaskResposne struct {
	Body *shared.TaskWithSubtask
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
	id, err := uuid.Parse(input.TaskID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task ID")
	}
	err = queries.UpdateTask(ctx, db, id, &input.Body)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) UpdateTaskPositionOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "update-task-position",
		Method:      http.MethodPut,
		Path:        path,
		Summary:     "Update task position",
		Description: "Update task position",
		Tags:        []string{"Task"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

// func (api *Api) UpdateTaskPosition(ctx context.Context, input *shared.TaskPositionInput) (*struct{}, error) {
// 	if input == nil {
// 		return nil, huma.Error400BadRequest("Invalid input")
// 	}
// 	db := api.app.Db()
// 	id, err := uuid.Parse(input.TaskID)
// 	if err != nil {
// 		return nil, huma.Error400BadRequest("Invalid task ID")
// 	}
// 	err = queries.UpdateTaskPosition(ctx, db, id, input.Body.Position)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return nil, nil
// }

func (api *Api) UpdateTaskPositionStatusOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "update-task-position-status",
		Method:      http.MethodPut,
		Path:        path,
		Summary:     "Update task position and status",
		Description: "Update task position and status",
		Tags:        []string{"Task"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) UpdateTaskPositionStatus(ctx context.Context, input *shared.TaskPositionStatusInput) (*struct{}, error) {
	if input == nil {
		return nil, huma.Error400BadRequest("Invalid input")
	}
	db := api.app.Db()
	id, err := uuid.Parse(input.TaskID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task ID")
	}
	err = queries.UpdateTaskPositionStatus(ctx, db, id, input.Body.Position, models.TaskStatus(input.Body.Status))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) TaskDeleteOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "task-delete",
		Method:      http.MethodDelete,
		Path:        path,
		Summary:     "Task delete",
		Description: "Delete a task",
		Tags:        []string{"Task"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) TaskDelete(ctx context.Context, input *struct {
	TaskID string `path:"task-id"`
}) (*struct{}, error) {
	db := api.app.Db()
	userInfo := core.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	id, err := uuid.Parse(input.TaskID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task ID")
	}
	err = queries.DeleteTask(ctx, db, id)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) TaskGetOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "task-get",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Task get",
		Description: "Get a task",
		Tags:        []string{"Task"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) TaskGet(ctx context.Context, input *struct {
	TaskID string `path:"task-id"`
}) (*TaskResposne, error) {
	db := api.app.Db()
	userInfo := core.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	id, err := uuid.Parse(input.TaskID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task ID")
	}
	task, err := queries.FindTaskByID(ctx, db, id)
	if err != nil {
		return nil, err
	}
	return &TaskResposne{
		Body: &shared.TaskWithSubtask{
			Task: shared.CrudModelToTask(task),
			Children: mapper.Map(task.Children, func(child *models.Task) *shared.Task {
				return shared.CrudModelToTask(child)
			}),
		},
	}, nil
}
