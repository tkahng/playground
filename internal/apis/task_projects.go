package apis

import (
	"context"
	"net/http"
	"slices"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/gemini"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/utils"
	"github.com/tkahng/playground/internal/workers"
)

type TaskProject struct {
	_                 struct{}                 `db:"task_projects" json:"-"`
	ID                uuid.UUID                `db:"id" json:"id"`
	CreatedByMemberID *uuid.UUID               `db:"created_by_member_id" json:"created_by_member_id" nullable:"true"`
	TeamID            uuid.UUID                `db:"team_id" json:"team_id"`
	WorkflowID        *uuid.UUID               `db:"workflow_id" json:"workflow_id" nullable:"true"`
	WorkflowStatusID  *uuid.UUID               `db:"workflow_status_id" json:"workflow_status_id" nullable:"true"`
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
		WorkflowID:        task.WorkflowID,
		WorkflowStatusID:  task.WorkflowStatusID,
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
		CreatedByMember:   fromTeamMemberModel(task.CreatedByMember),
		Team:              fromTeamModel(task.Team),
		Tasks:             mapper.Map(task.Tasks, fromModelTask),
	}
}

type CreateTaskProjectDTO struct {
	TeamID           uuid.UUID                `json:"team_id" required:"true" format:"uuid"`
	MemberID         uuid.UUID                `json:"member_id" required:"true" format:"uuid"`
	Name             string                   `json:"name" required:"true"`
	Description      *string                  `json:"description,omitempty" required:"false"`
	Status           models.TaskProjectStatus `json:"status" required:"false" enum:"todo,in_progress,done" default:"todo"`
	WorkflowStatusID *uuid.UUID               `json:"workflow_status_id,omitempty" required:"false" format:"uuid"`
	WorkflowID       *uuid.UUID               `json:"workflow_id,omitempty" required:"false" format:"uuid"`
	Rank             float64                  `json:"rank,omitempty" required:"false"`
}

type CreateTaskProjectWithoutTeamDTO struct {
	Name             string                   `json:"name" required:"true"`
	Description      *string                  `json:"description,omitempty" required:"false"`
	Status           models.TaskProjectStatus `json:"status" required:"false" enum:"todo,in_progress,done" default:"todo"`
	WorkflowStatusID *uuid.UUID               `json:"workflow_status_id,omitempty" required:"false" format:"uuid"`
	WorkflowID       *uuid.UUID               `json:"workflow_id,omitempty" required:"false" format:"uuid"`
	Rank             float64                  `json:"rank,omitempty" required:"false"`
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
	Q                 string                     `query:"q,omitempty" required:"false"`
	Status            []models.TaskProjectStatus `query:"status,omitempty" required:"false" minimum:"1" maximum:"100" enum:"todo,in_progress,done"`
	WorkflowIds       []string                   `query:"workflow_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	WorkflowStatusIds []string                   `query:"workflow_status_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Ids               []string                   `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	SortParams
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"tasks,subtasks"`
}

func (api *Api) TeamTaskProjectListBind(humaApi huma.API) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "task-project-list",
			Method:      http.MethodGet,
			Path:        "/teams/{team-id}/task-projects",
			Summary:     "Task project list",
			Description: "List of task projects",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
			),
		},
		func(ctx context.Context, input *TeamTaskProjectsListParams) (*TaskProjectListResponse, error) {
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
			newInput.Status = input.Status
			newInput.WorkflowIds = utils.ParseValidUUIDs(input.WorkflowIds...)
			newInput.WorkflowStatusIds = utils.ParseValidUUIDs(input.WorkflowStatusIds...)
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
		},
	)
}

func (api *Api) TeamTaskProjectCreateBind(humaApi huma.API) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "task-project-create",
			Method:      http.MethodPost,
			Path:        "/teams/{team-id}/task-projects",
			Summary:     "Task project create",
			Description: "Create a new task project",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
				middleware.RequireTeamPermission(api.App(), shared.TeamPermissionProjectsCreate),
			),
		},
		func(ctx context.Context, input *CreateTaskProjectWithTasksInput) (*struct {
			Body *TaskProject
		}, error) {
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
					TeamID:           parsedTeamID,
					MemberID:         teamInfo.Member.ID,
					Name:             input.Body.Name,
					Description:      input.Body.Description,
					Status:           input.Body.Status,
					WorkflowStatusID: input.Body.WorkflowStatusID,
					WorkflowID:       input.Body.WorkflowID,
					Rank:             input.Body.Rank,
				},
				Tasks: mapper.Map(input.Body.Tasks, func(task CreateTaskProjectTaskDTO) stores.CreateTaskProjectTaskDTO {
					return stores.CreateTaskProjectTaskDTO{
						Name:             task.Name,
						Description:      task.Description,
						Status:           models.TaskStatus(task.Status),
						WorkflowStatusID: task.WorkflowStatusID,
						Rank:             task.Rank,
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
		},
	)
}
func (api *Api) TeamTaskProjectCreateWithAiBind(humaApi huma.API) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "task-project-create-with-ai",
			Method:      http.MethodPost,
			Path:        "/teams/{team-id}/task-projects/ai",
			Summary:     "Task project create with ai",
			Description: "Create a new task project with ai",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
			),
		},
		func(ctx context.Context, input *TaskProjectCreateWithAiInput) (*struct {
			Body *TaskProject
		}, error) {
			teamInfo := contextstore.GetContextTeamInfo(ctx)
			if teamInfo == nil {
				return nil, huma.Error401Unauthorized("no team info")
			}
			if teamInfo.Member.UserID == nil {
				return nil, huma.Error401Unauthorized("no user info")
			}

			if err := api.App().AiUsage().CheckQuota(ctx, teamInfo.Member.TeamID); err != nil {
				return nil, err
			}

			geminiClient := gemini.NewClient(api.App().Config().AiConfig)
			taskProjectPlan, err := geminiClient.GenerateProjectPlan(ctx, input.Body.Input)
			if err != nil {
				return nil, err
			}

			_ = api.App().AiUsage().RecordUsage(ctx, *teamInfo.Member.UserID, teamInfo.Member.ID, teamInfo.Member.TeamID, taskProjectPlan.Usage)

			args := stores.CreateTaskProjectWithTasksDTO{
				CreateTaskProjectDTO: stores.CreateTaskProjectDTO{
					Name:        taskProjectPlan.Project.Name,
					Description: &taskProjectPlan.Project.Description,
					Status:      models.TaskProjectStatusTodo,
					TeamID:      teamInfo.Member.TeamID,
					MemberID:    teamInfo.Member.ID,
				},
				Tasks: mapper.Map(taskProjectPlan.Tasks, func(task gemini.Task) stores.CreateTaskProjectTaskDTO {
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
		},
	)
}

type TaskProjectCreateWithAiDto struct {
	Input string `json:"input" example:"Help me plan a 6 day vacation to Paris"`
}
type TaskProjectCreateWithAiInput struct {
	TeamID string                     `json:"team_id" path:"team-id" required:"true" format:"uuid"`
	Body   TaskProjectCreateWithAiDto `json:"body"`
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

	existing, err := api.App().Adapter().Task().FindTaskProjectByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, huma.Error404NotFound("Task project not found")
	}
	oldStatus := string(existing.Status)

	payload := input.Body
	err = api.App().Adapter().Task().UpdateTaskProject(ctx, id, &payload)
	if err != nil {
		return nil, err
	}

	if oldStatus != string(payload.Status) {
		teamInfo := contextstore.GetContextTeamInfo(ctx)
		if teamInfo != nil {
			_ = api.App().JobService().EnqueueProjectStatusChangedJob(ctx, &workers.ProjectStatusChangedJobArgs{
				ProjectID:         id,
				OldStatus:         oldStatus,
				NewStatus:         string(payload.Status),
				ChangedByMemberID: teamInfo.Member.ID,
			})
		}
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
	TaskProjectID string              `path:"task-project-id" json:"task_project_id" required:"true" format:"uuid"`
	Body          services.TaskFields `json:"body" required:"true"`
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
		taskDue := *task.EndAt
		if taskDue.Before(time.Now()) {
			taskDue = time.Now().Add(10 * time.Second)
		}
		err = api.App().JobService().EnqueTaskDueJob(ctx, &workers.TaskDueTodayJobArgs{
			TaskID:  task.ID,
			DueDate: taskDue,
		})
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to create task project update date job")
		}
		err = api.App().JobService().EnqueueTaskOverdueJob(ctx, &workers.TaskOverdueJobArgs{
			TaskID:  task.ID,
			DueDate: taskDue,
		})
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to enqueue task overdue job")
		}
	}
	err = api.App().Adapter().Task().UpdateTaskProjectUpdateDate(ctx, parsedProjectID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to update task project update date")
	}
	return &TaskResponse{
		Body: fromModelTask(task),
	}, nil
}
