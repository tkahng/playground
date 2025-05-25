package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

type TaskListResponse struct {
	Body *shared.PaginatedResponse[*shared.Task]
}

func (api *Api) TeamTaskList(ctx context.Context, input *shared.TeamTaskListParams) (*TaskListResponse, error) {

	teamInfo := contextstore.GetContextTeamInfo(ctx)
	if teamInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	newInput := shared.TaskListParams{}
	newInput.SortParams = input.SortParams
	newInput.PaginatedInput = input.PaginatedInput
	newInput.TaskListFilter = shared.TaskListFilter{
		ProjectID: input.ProjectID,
		Status:    input.TeamTaskListFilter.Status,
		Ids:       input.TeamTaskListFilter.Ids,
		Q:         input.TeamTaskListFilter.Q,
		TeamID:    teamInfo.Team.ID.String(),
		ParentID:  input.TeamTaskListFilter.ParentID,
	}
	tasks, err := api.app.Task().Store().ListTasks(ctx, &newInput)
	if err != nil {
		return nil, huma.Error500InternalServerError("error listing tasks", err)
	}
	total, err := api.app.Task().Store().CountTasks(ctx, &newInput.TaskListFilter)
	if err != nil {
		return nil, huma.Error500InternalServerError("error counting tasks", err)
	}
	return &TaskListResponse{
		Body: &shared.PaginatedResponse[*shared.Task]{
			Data: mapper.Map(tasks, func(task *models.Task) *shared.Task {
				return shared.FromModelTask(task)
			}),
			Meta: shared.GenerateMeta(&input.PaginatedInput, total),
		},
	}, nil
}

type TaskResposne struct {
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
	err = api.app.Task().Store().DeleteTask(ctx, id)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) TaskGet(ctx context.Context, input *struct {
	TaskID string `path:"task-id"`
}) (*TaskResposne, error) {

	userInfo := contextstore.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	id, err := uuid.Parse(input.TaskID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task ID")
	}
	task, err := api.app.Task().Store().FindTaskByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &TaskResposne{
		Body: shared.FromModelTask(task),
	}, nil
}
