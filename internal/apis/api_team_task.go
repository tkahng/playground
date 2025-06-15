package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/utils"
)

type TaskListResponse struct {
	Body *ApiPaginatedResponse[*shared.Task]
}

func (api *Api) TeamTaskList(ctx context.Context, input *shared.TeamTaskListParams) (*TaskListResponse, error) {

	teamInfo := contextstore.GetContextTeamInfo(ctx)
	if teamInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	newInput := &stores.TaskFilter{}
	newInput.SortBy = input.SortBy
	newInput.SortOrder = input.SortOrder
	newInput.Page = input.Page
	newInput.PerPage = input.PerPage
	newInput.Ids = utils.ParseValidUUIDs(input.TeamTaskListFilter.Ids...)
	newInput.Q = input.TeamTaskListFilter.Q
	newInput.Statuses = mapper.Map(input.TeamTaskListFilter.Status, func(status shared.TaskStatus) models.TaskStatus {
		return models.TaskStatus(status)
	})
	newInput.TeamIds = []uuid.UUID{teamInfo.Team.ID}
	newInput.ProjectIds = utils.ParseValidUUIDs(input.ProjectID)
	parentID, err := uuid.Parse(input.TeamTaskListFilter.ParentID)
	if err != nil && input.TeamTaskListFilter.ParentID != "" {
		return nil, huma.Error400BadRequest("Invalid parent ID format", err)
	}
	newInput.ParentIds = []uuid.UUID{parentID}

	tasks, err := api.app.Adapter().Task().ListTasks(ctx, newInput)
	if err != nil {
		return nil, huma.Error500InternalServerError("error listing tasks", err)
	}
	total, err := api.app.Adapter().Task().CountTasks(ctx, newInput)
	if err != nil {
		return nil, huma.Error500InternalServerError("error counting tasks", err)
	}
	return &TaskListResponse{
		Body: &ApiPaginatedResponse[*shared.Task]{
			Data: mapper.Map(tasks, func(task *models.Task) *shared.Task {
				return shared.FromModelTask(task)
			}),
			Meta: GenerateMeta(&input.PaginatedInput, total),
		},
	}, nil
}

type TaskResponse struct {
	Body *shared.Task
}

func (api *Api) TaskUpdate(ctx context.Context, input *shared.UpdateTaskInput) (*struct{}, error) {

	id, err := uuid.Parse(input.TaskID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task ID")
	}
	err = api.app.Task().FindAndUpdateTask(ctx, id, &input.Body)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) UpdateTaskPositionStatus(ctx context.Context, input *shared.TaskPositionStatusInput) (*struct{}, error) {
	if input == nil {
		return nil, huma.Error400BadRequest("Invalid input")
	}

	id, err := uuid.Parse(input.TaskID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task ID")
	}
	err = api.app.Task().UpdateTaskRankStatus(ctx, id, input.Body.Position, models.TaskStatus(input.Body.Status))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) TaskDelete(ctx context.Context, input *struct {
	TaskID string `path:"task-id"`
}) (*struct{}, error) {

	userInfo := contextstore.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	id, err := uuid.Parse(input.TaskID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task ID")
	}
	err = api.app.Adapter().Task().DeleteTask(ctx, id)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) TaskGet(ctx context.Context, input *struct {
	TaskID string `path:"task-id"`
}) (*TaskResponse, error) {

	userInfo := contextstore.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	id, err := uuid.Parse(input.TaskID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task ID")
	}
	task, err := api.app.Adapter().Task().FindTaskByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &TaskResponse{
		Body: shared.FromModelTask(task),
	}, nil
}
