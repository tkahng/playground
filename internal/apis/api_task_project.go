package apis

import (
	"context"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/ai/googleai"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

type TaskProjectListResponse struct {
	Body *shared.PaginatedResponse[*shared.TaskProjectWithTasks]
}

func (api *Api) TaskProjectList(ctx context.Context, input *shared.TaskProjectsListParams) (*TaskProjectListResponse, error) {

	userInfo := contextstore.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	input.TaskProjectsListFilter.UserID = userInfo.User.ID.String()
	taskProject, err := api.app.Task().Store().ListTaskProjects(ctx, input)
	if err != nil {
		return nil, err
	}
	total, err := api.app.Task().Store().CountTaskProjects(ctx, &input.TaskProjectsListFilter)
	if err != nil {
		return nil, err
	}
	taskProjectIds := mapper.Map(taskProject, func(taskProject *models.TaskProject) uuid.UUID {
		return taskProject.ID
	})

	if input.Expand != nil && slices.Contains(input.Expand, "tasks") {
		tasks, err := api.app.Task().Store().LoadTaskProjectsTasks(ctx, taskProjectIds...)
		if err != nil {
			return nil, err
		}
		for idx, taskProject := range taskProject {
			taskProject.Tasks = tasks[idx]
		}
	}
	return &TaskProjectListResponse{
		Body: &shared.PaginatedResponse[*shared.TaskProjectWithTasks]{
			Data: mapper.Map(taskProject, func(taskProject *models.TaskProject) *shared.TaskProjectWithTasks {
				return &shared.TaskProjectWithTasks{
					TaskProject: shared.CrudToProject(taskProject),
					Tasks: mapper.Map(taskProject.Tasks, func(task *models.Task) *shared.TaskWithSubtask {
						return &shared.TaskWithSubtask{
							Task: shared.CrudModelToTask(task),
							Children: mapper.Map(task.Children, func(child *models.Task) *shared.Task {
								return shared.CrudModelToTask(child)
							}),
						}
					}),
				}
			}),
			Meta: shared.GenerateMeta(input.PaginatedInput, total),
		},
	}, nil
}

func (api *Api) TaskProjectCreate(ctx context.Context, input *struct {
	Body shared.CreateTaskProjectWithTasksDTO
}) (*struct {
	Body *shared.TaskProject
}, error) {
	userInfo := contextstore.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}

	taskProject, err := api.app.Task().Store().CreateTaskProjectWithTasks(ctx, &input.Body)
	if err != nil {
		return nil, err
	}
	return &struct {
		Body *shared.TaskProject
	}{
		Body: shared.CrudToProject(taskProject),
	}, nil
}

type TaskProjectCreateWithAiDto struct {
	Input string `json:"input" example:"Help me plan a 6 day vacation to Paris"`
}
type TaskProjectCreateWithAiInput struct {
	Body TaskProjectCreateWithAiDto `json:"body"`
}

func (api *Api) TaskProjectCreateWithAi(ctx context.Context, input *TaskProjectCreateWithAiInput) (*struct {
	Body *shared.TaskProject
}, error) {
	userInfo := contextstore.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	teamInfo := contextstore.GetContextTeamInfo(ctx)
	if teamInfo == nil {
		return nil, huma.Error401Unauthorized("no team info")
	}

	aiService := googleai.NewAiService(ctx, api.app.Cfg().AiConfig)
	taskProjectPlan, err := aiService.GenerateProjectPlan(ctx, input.Body.Input)
	if err != nil {
		return nil, err
	}
	args := shared.CreateTaskProjectWithTasksDTO{
		CreateTaskProjectDTO: shared.CreateTaskProjectDTO{
			Name:        taskProjectPlan.Project.Name,
			Description: &taskProjectPlan.Project.Description,
			Status:      shared.TaskProjectStatusTodo,
			TeamID:      teamInfo.Member.TeamID,
			MemberID:    teamInfo.Member.ID,
		},
		Tasks: mapper.Map(taskProjectPlan.Tasks, func(task googleai.Task) shared.CreateTaskBaseDTO {
			return shared.CreateTaskBaseDTO{
				Name:        task.Name,
				Description: &task.Description,
				Status:      shared.TaskStatusTodo,
				CreatedBy:   teamInfo.Member.ID,
				TeamID:      teamInfo.Member.TeamID,
			}
		}),
	}
	taskProject, err := api.app.Task().Store().CreateTaskProjectWithTasks(ctx, &args)
	if err != nil {
		return nil, err
	}
	return &struct {
		Body *shared.TaskProject
	}{
		Body: shared.CrudToProject(taskProject),
	}, nil
}

type TaskProjectResponse struct {
	Body *shared.TaskProjectWithTasks
}

func (api *Api) TaskProjectUpdate(ctx context.Context, input *shared.UpdateTaskProjectDTO) (*struct{}, error) {
	userInfo := contextstore.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}

	id, err := uuid.Parse(input.TaskProjectID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task project id")
	}
	payload := input.Body
	err = api.app.Task().Store().UpdateTaskProject(ctx, id, &payload)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) TaskProjectDelete(ctx context.Context, input *struct {
	TaskProjectID string `path:"task-project-id"`
}) (*struct{}, error) {
	userInfo := contextstore.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}

	id, err := uuid.Parse(input.TaskProjectID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task project id")
	}
	err = api.app.Task().Store().DeleteTaskProject(ctx, id)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) TaskProjectGet(ctx context.Context, input *struct {
	TaskProjectID string   `path:"task-project-id" json:"task_project_id" required:"true" format:"uuid"`
	Expand        []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"tasks"`
}) (*TaskProjectResponse, error) {
	userInfo := contextstore.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}

	id, err := uuid.Parse(input.TaskProjectID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task project id")
	}
	taskProject, err := api.app.Task().Store().FindTaskProjectByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Expand != nil && slices.Contains(input.Expand, "tasks") {
		tasks, err := api.app.Task().Store().LoadTaskProjectsTasks(ctx, taskProject.ID)
		if err != nil {
			return nil, err
		}
		if len(tasks) > 0 {
			taskProject.Tasks = tasks[0]
		}
	}
	return &TaskProjectResponse{
		Body: &shared.TaskProjectWithTasks{
			TaskProject: shared.CrudToProject(taskProject),
			Tasks: mapper.Map(taskProject.Tasks, func(task *models.Task) *shared.TaskWithSubtask {
				return &shared.TaskWithSubtask{
					Task: shared.CrudModelToTask(task),
					Children: mapper.Map(task.Children, func(child *models.Task) *shared.Task {
						return shared.CrudModelToTask(child)
					}),
				}
			}),
		},
	}, nil
}

func (api *Api) TaskProjectTasksCreate(ctx context.Context, input *shared.CreateTaskWithProjectIdInput) (*TaskResposne, error) {
	userInfo := contextstore.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	teamInfo := contextstore.GetContextTeamInfo(ctx)
	if teamInfo == nil {
		return nil, huma.Error401Unauthorized("no team info")
	}

	id, err := uuid.Parse(input.TaskProjectID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task project id")
	}
	payload := input.Body
	order, err := api.app.Task().Store().FindLastTaskOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	payload.Order = order
	payload.CreatedBy = teamInfo.Member.ID
	payload.TeamID = teamInfo.Member.TeamID
	task, err := api.app.Task().Store().CreateTaskWithChildren(ctx, id, &payload)
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
