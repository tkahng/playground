package apis

import (
	"context"
	"slices"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/ai/googleai"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/utils"
	"github.com/tkahng/authgo/internal/workers"
)

type TaskProject struct {
	_                 struct{}                 `db:"task_projects" json:"-"`
	ID                uuid.UUID                `db:"id" json:"id"`
	CreatedByMemberID *uuid.UUID               `db:"created_by_member_id" json:"created_by_member_id" nullable:"true"`
	TeamID            uuid.UUID                `db:"team_id" json:"team_id"`
	Name              string                   `db:"name" json:"name"`
	Description       *string                  `db:"description" json:"description"`
	Status            models.TaskProjectStatus `db:"status" json:"status" enum:"todo,in_progress,done"`
	StartAt           *time.Time               `db:"start_at" json:"start_at" nullable:"true"`
	EndAt             *time.Time               `db:"end_at" json:"end_at" nullable:"true"`
	AssigneeID        *uuid.UUID               `db:"assignee_id" json:"assignee_id" nullable:"true"`
	ReporterID        *uuid.UUID               `db:"reporter_id" json:"reporter_id" nullable:"true"`
	Rank              float64                  `db:"rank" json:"rank"`
	CreatedAt         time.Time                `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time                `db:"updated_at" json:"updated_at"`
	CreatedByMember   *TeamMember              `db:"created_by_member" src:"created_by_member_id" dest:"id" table:"team_members" json:"created_by_member,omitempty"`
	Team              *Team                    `db:"team" src:"team_id" dest:"id" table:"teams" json:"team,omitempty"`
	Tasks             []*Task                  `db:"tasks" src:"id" dest:"project_id" table:"tasks" json:"tasks,omitempty"`
}

func FromModelProject(task *models.TaskProject) *TaskProject {
	if task == nil {
		return nil
	}
	return &TaskProject{
		ID:                task.ID,
		CreatedByMemberID: task.CreatedByMemberID,
		TeamID:            task.TeamID,
		Name:              task.Name,
		Description:       task.Description,
		Status:            task.Status,
		StartAt:           task.StartAt,
		EndAt:             task.EndAt,
		AssigneeID:        task.AssigneeID,
		ReporterID:        task.ReporterID,
		Rank:              task.Rank,
		CreatedAt:         task.CreatedAt,
		UpdatedAt:         task.UpdatedAt,
		CreatedByMember:   FromTeamMemberModel(task.CreatedByMember),
		Team:              FromTeamModel(task.Team),
		Tasks:             mapper.Map(task.Tasks, FromModelTask),
	}
}

type CreateTaskProjectDTO struct {
	TeamID      uuid.UUID                `json:"team_id" required:"true" format:"uuid"`
	MemberID    uuid.UUID                `json:"member_id" required:"true" format:"uuid"`
	Name        string                   `json:"name" required:"true"`
	Description *string                  `json:"description,omitempty" required:"false"`
	Status      models.TaskProjectStatus `json:"status" required:"false" enum:"todo,in_progress,done" default:"todo"`
	Rank        float64                  `json:"rank,omitempty" required:"false"`
}

type CreateTaskProjectWithoutTeamDTO struct {
	Name        string                   `json:"name" required:"true"`
	Description *string                  `json:"description,omitempty" required:"false"`
	Status      models.TaskProjectStatus `json:"status" required:"false" enum:"todo,in_progress,done" default:"todo"`
	Rank        float64                  `json:"rank,omitempty" required:"false"`
}

type CreateTaskProjectWithoutTeamWithTasks struct {
	CreateTaskProjectWithoutTeamDTO
	Tasks []CreateTaskProjectTaskDTO `json:"tasks,omitempty" required:"false"`
}
type CreateTaskProjectWithTasksInput struct {
	TeamID string `json:"team_id" path:"team-id" required:"true" format:"uuid"`
	Body   CreateTaskProjectWithoutTeamWithTasks
}

type UpdateTaskProjectDTO struct {
	Body          stores.UpdateTaskProjectBaseDTO
	TaskProjectID string `path:"task-project-id" json:"task_project_id" required:"true" format:"uuid"`
}

type TaskProjectListResponse struct {
	Body *ApiPaginatedResponse[*TaskProject]
}

type TeamTaskProjectsListParams struct {
	TeamID string `path:"team-id" required:"true" format:"uuid"`
	PaginatedInput
	Q        string                     `query:"q,omitempty" required:"false"`
	Status   []models.TaskProjectStatus `query:"status,omitempty" required:"false" minimum:"1" maximum:"100" enum:"todo,in_progress,done"`
	Ids      []string                   `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Statuses []models.TaskProjectStatus `query:"task_status,omitempty" required:"false" minimum:"1" maximum:"100" enum:"todo,in_progress,done"`
	SortParams
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"tasks,subtasks"`
}

func (api *Api) TeamTaskProjectList(ctx context.Context, input *TeamTaskProjectsListParams) (*TaskProjectListResponse, error) {

	teamInfo := contextstore.GetContextTeamInfo(ctx)
	if teamInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	newInput := &stores.TaskProjectsFilter{}
	newInput.SortBy = input.SortBy
	newInput.SortOrder = input.SortOrder
	newInput.Page = input.Page
	newInput.PerPage = input.PerPage
	newInput.Ids = utils.ParseValidUUIDs(input.Ids...)
	newInput.Q = input.Q
	newInput.Statuses = input.Statuses
	newInput.TeamIds = []uuid.UUID{teamInfo.Team.ID}
	taskProject, err := api.App().Adapter().Task().ListTaskProjects(ctx, newInput)
	if err != nil {
		return nil, err
	}
	total, err := api.App().Adapter().Task().CountTaskProjects(ctx, newInput)
	if err != nil {
		return nil, err
	}
	taskProjectIds := mapper.Map(taskProject, func(taskProject *models.TaskProject) uuid.UUID {
		return taskProject.ID
	})

	if input.Expand != nil && slices.Contains(input.Expand, "tasks") {
		tasks, err := api.App().Adapter().Task().LoadTaskProjectsTasks(ctx, taskProjectIds...)
		if err != nil {
			return nil, err
		}
		for idx, taskProject := range taskProject {
			taskProject.Tasks = tasks[idx]
		}
	}
	return &TaskProjectListResponse{
		Body: &ApiPaginatedResponse[*TaskProject]{
			Data: mapper.Map(taskProject, func(taskProject *models.TaskProject) *TaskProject {
				return FromModelProject(taskProject)
			}),
			Meta: ApiGenerateMeta(&input.PaginatedInput, total),
		},
	}, nil
}

func (api *Api) TeamTaskProjectCreate(
	ctx context.Context,
	input *CreateTaskProjectWithTasksInput,
) (
	*struct {
		Body *TaskProject
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

	taskProject, err := api.App().Adapter().Task().CreateTaskProjectWithTasks(ctx, &stores.CreateTaskProjectWithTasksDTO{
		CreateTaskProjectDTO: stores.CreateTaskProjectDTO{
			TeamID:      parsedTeamID,
			MemberID:    teamInfo.Member.ID,
			Name:        input.Body.Name,
			Description: input.Body.Description,
			Status:      input.Body.Status,
			Rank:        input.Body.Rank,
		},
		Tasks: mapper.Map(input.Body.Tasks, func(task CreateTaskProjectTaskDTO) stores.CreateTaskProjectTaskDTO {
			return stores.CreateTaskProjectTaskDTO{
				Name:        task.Name,
				Description: task.Description,
				Status:      models.TaskStatus(task.Status),
				Rank:        task.Rank,
			}
		}),
	})
	if err != nil {
		return nil, err
	}
	return &struct {
		Body *TaskProject
	}{
		Body: FromModelProject(taskProject),
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
	Body *TaskProject
}, error) {
	teamInfo := contextstore.GetContextTeamInfo(ctx)
	if teamInfo == nil {
		return nil, huma.Error401Unauthorized("no team info")
	}

	aiService := googleai.NewAiService(ctx, api.App().Config().AiConfig)
	taskProjectPlan, err := aiService.GenerateProjectPlan(ctx, input.Body.Input)
	if err != nil {
		return nil, err
	}
	args := stores.CreateTaskProjectWithTasksDTO{
		CreateTaskProjectDTO: stores.CreateTaskProjectDTO{
			Name:        taskProjectPlan.Project.Name,
			Description: &taskProjectPlan.Project.Description,
			Status:      models.TaskProjectStatusTodo,
			TeamID:      teamInfo.Member.TeamID,
			MemberID:    teamInfo.Member.ID,
		},
		Tasks: mapper.Map(taskProjectPlan.Tasks, func(task googleai.Task) stores.CreateTaskProjectTaskDTO {
			return stores.CreateTaskProjectTaskDTO{
				Name:        task.Name,
				Description: &task.Description,
				Status:      models.TaskStatusTodo,
			}
		}),
	}
	taskProject, err := api.App().Adapter().Task().CreateTaskProjectWithTasks(ctx, &args)
	if err != nil {
		return nil, err
	}
	return &struct {
		Body *TaskProject
	}{
		Body: FromModelProject(taskProject),
	}, nil
}

type TaskProjectResponse struct {
	Body *TaskProject
}

func (api *Api) TeamTaskProjectUpdate(ctx context.Context, input *UpdateTaskProjectDTO) (*struct{}, error) {
	userInfo := contextstore.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}

	id, err := uuid.Parse(input.TaskProjectID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task project id")
	}
	payload := input.Body
	err = api.App().Adapter().Task().UpdateTaskProject(ctx, id, &payload)
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
	err = api.App().Adapter().Task().DeleteTaskProject(ctx, id)
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
	taskProject, err := api.App().Adapter().Task().FindTaskProjectByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Expand != nil && slices.Contains(input.Expand, "tasks") {
		tasks, err := api.App().Adapter().Task().LoadTaskProjectsTasks(ctx, taskProject.ID)
		if err != nil {
			return nil, err
		}
		if len(tasks) > 0 {
			taskProject.Tasks = tasks[0]
		}
	}
	return &TaskProjectResponse{
		Body: FromModelProject(taskProject),
	}, nil
}

type ApiCreateTaskWithProjectIdInput struct {
	TaskProjectID string `path:"task-project-id" json:"task_project_id" required:"true" format:"uuid"`
	Body          services.TaskFields
}

func (api *Api) TeamTaskProjectTasksCreate(ctx context.Context, input *ApiCreateTaskWithProjectIdInput) (*TaskResponse, error) {
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

	task, err := api.App().Task().CreateTask(ctx, teamInfo.Team.ID, parsedProjectID, teamInfo.Member.ID, &input.Body)
	if err != nil {
		return nil, err
	}
	if task.EndAt != nil {
		err = api.App().JobService().EnqueTaskDueJob(ctx, &workers.TaskDueTodayJobArgs{
			TaskID:  task.ID,
			DueDate: *task.EndAt,
		})
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to create task project update date job")
		}
	}
	err = api.App().Adapter().Task().UpdateTaskProjectUpdateDate(ctx, parsedProjectID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to update task project update date")
	}
	return &TaskResponse{
		Body: FromModelTask(task),
	}, nil
}
