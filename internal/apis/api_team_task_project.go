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
	Body *shared.PaginatedResponse[*shared.TaskProject]
}

func (api *Api) TeamTaskProjectList(ctx context.Context, input *shared.TeamTaskProjectsListParams) (*TaskProjectListResponse, error) {

	teamInfo := contextstore.GetContextTeamInfo(ctx)
	if teamInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	newInput := shared.TaskProjectsListParams{}
	newInput.SortParams = input.SortParams
	newInput.PaginatedInput = input.PaginatedInput
	newInput.TaskProjectsListFilter = shared.TaskProjectsListFilter{
		TeamID: teamInfo.Team.ID.String(),
		Status: input.TeamTaskProjectsListFilter.Status,
		Ids:    input.TeamTaskProjectsListFilter.Ids,
		Q:      input.TeamTaskProjectsListFilter.Q,
	}
	taskProject, err := api.app.Task().Store().ListTaskProjects(ctx, &newInput)
	if err != nil {
		return nil, err
	}
	total, err := api.app.Task().Store().CountTaskProjects(ctx, &newInput.TaskProjectsListFilter)
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
		Body: &shared.PaginatedResponse[*shared.TaskProject]{
			Data: mapper.Map(taskProject, func(taskProject *models.TaskProject) *shared.TaskProject {
				return shared.FromModelProject(taskProject)
			}),
			Meta: shared.GenerateMeta(&input.PaginatedInput, total),
		},
	}, nil
}

func (api *Api) TeamTaskProjectCreate(
	ctx context.Context,
	input *shared.CreateTaskProjectWithTasksInput,
) (
	*struct {
		Body *shared.TaskProject
	},
	error,
) {
	if input == nil {
		return nil, huma.Error400BadRequest("Input cannot be nil")
	}
	parsedTeamID, err := uuid.Parse(input.TeamID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid team id")
	}

	teamInfo := contextstore.GetContextTeamInfo(ctx)
	if teamInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}

	taskProject, err := api.app.Task().Store().CreateTaskProjectWithTasks(ctx, &shared.CreateTaskProjectWithTasksDTO{
		CreateTaskProjectDTO: shared.CreateTaskProjectDTO{
			TeamID:      parsedTeamID,
			MemberID:    teamInfo.Member.ID,
			Name:        input.Body.Name,
			Description: input.Body.Description,
			Status:      input.Body.Status,
			Rank:        input.Body.Rank,
		},
		Tasks: input.Body.Tasks,
	})
	if err != nil {
		return nil, err
	}
	return &struct {
		Body *shared.TaskProject
	}{
		Body: shared.FromModelProject(taskProject),
	}, nil
}

type TaskProjectCreateWithAiDto struct {
	Input string `json:"input" example:"Help me plan a 6 day vacation to Paris"`
}
type TaskProjectCreateWithAiInput struct {
	TeamID string                     `json:"team_id" path:"team-id" required:"true" format:"uuid"`
	Body   TaskProjectCreateWithAiDto `json:"body"`
}

func (api *Api) TeamTaskProjectCreateWithAi(ctx context.Context, input *TaskProjectCreateWithAiInput) (*struct {
	Body *shared.TaskProject
}, error) {
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
		Tasks: mapper.Map(taskProjectPlan.Tasks, func(task googleai.Task) shared.CreateTaskProjectTaskDTO {
			return shared.CreateTaskProjectTaskDTO{
				Name:        task.Name,
				Description: &task.Description,
				Status:      shared.TaskStatusTodo,
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
		Body: shared.FromModelProject(taskProject),
	}, nil
}

type TaskProjectResponse struct {
	Body *shared.TaskProject
}

func (api *Api) TeamTaskProjectUpdate(ctx context.Context, input *shared.UpdateTaskProjectDTO) (*struct{}, error) {
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

func (api *Api) TeamTaskProjectDelete(ctx context.Context, input *struct {
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

func (api *Api) TeamTaskProjectGet(ctx context.Context, input *struct {
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
		Body: shared.FromModelProject(taskProject),
	}, nil
}

func (api *Api) TeamTaskProjectTasksCreate(ctx context.Context, input *shared.CreateTaskWithProjectIdInput) (*TaskResposne, error) {
	userInfo := contextstore.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	teamInfo := contextstore.GetContextTeamInfo(ctx)
	if teamInfo == nil {
		return nil, huma.Error401Unauthorized("no team info")
	}

	parsedProjectID, err := uuid.Parse(input.TaskProjectID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task project id")
	}
	payload := input.Body
	order, err := api.app.Task().Store().FindLastTaskRank(ctx, parsedProjectID)
	if err != nil {
		return nil, err
	}
	payload.Rank = order
	task, err := api.app.Task().CreateTaskWithChildren(ctx, teamInfo.Member.TeamID, parsedProjectID, teamInfo.Member.ID, &payload)
	if err != nil {
		return nil, err
	}
	err = api.app.Task().Store().UpdateTaskProjectUpdateDate(ctx, parsedProjectID)
	return &TaskResposne{
		Body: shared.FromModelTask(task),
	}, nil
}
