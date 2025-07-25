package apis

import (
	"context"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/utils"
	"github.com/tkahng/playground/internal/workers"
)

type Task struct {
	_                 struct{}          `db:"tasks" json:"-"`
	ID                uuid.UUID         `db:"id" json:"id"`
	CreatedByMemberID *uuid.UUID        `db:"created_by_member_id" json:"created_by_member_id" nullable:"true"`
	TeamID            uuid.UUID         `db:"team_id" json:"team_id"`
	ProjectID         uuid.UUID         `db:"project_id" json:"project_id"`
	Name              string            `db:"name" json:"name"`
	Description       *string           `db:"description" json:"description"`
	Status            models.TaskStatus `db:"status" json:"status" enum:"todo,in_progress,done"`
	StartAt           *time.Time        `db:"start_at" json:"start_at" nullable:"true"`
	EndAt             *time.Time        `db:"end_at" json:"end_at" nullable:"true"`
	AssigneeID        *uuid.UUID        `db:"assignee_id" json:"assignee_id" nullable:"true"`
	ReporterID        *uuid.UUID        `db:"reporter_id" json:"reporter_id" nullable:"true"`
	Rank              float64           `db:"rank" json:"rank"`
	ParentID          *uuid.UUID        `db:"parent_id" json:"parent_id" nullable:"true"`
	CreatedAt         time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time         `db:"updated_at" json:"updated_at"`
	Children          []*Task           `db:"children" src:"id" dest:"parent_id" table:"tasks" json:"children,omitempty"`
	CreatedByMember   *TeamMember       `db:"created_by_member" src:"created_by_member_id" dest:"id" table:"team_members" json:"created_by_member,omitempty"`
	Assignee          *TeamMember       `db:"assignee" src:"assignee_id" dest:"id" table:"team_members" json:"assignee,omitempty"`
	Reporter          *TeamMember       `db:"reporter" src:"reporter_id" dest:"id" table:"team_members" json:"reporter,omitempty"`
	Team              *Team             `db:"team" src:"team_id" dest:"id" table:"teams" json:"team,omitempty"`
	Project           *TaskProject      `db:"project" src:"project_id" dest:"id" table:"task_projects" json:"project,omitempty"`
}

func FromModelTask(task *models.Task) *Task {
	if task == nil {
		return nil
	}
	return &Task{
		ID:                task.ID,
		CreatedByMemberID: task.CreatedByMemberID,
		TeamID:            task.TeamID,
		ProjectID:         task.ProjectID,
		Name:              task.Name,
		Description:       task.Description,
		Status:            task.Status,
		StartAt:           task.StartAt,
		EndAt:             task.EndAt,
		AssigneeID:        task.AssigneeID,
		ReporterID:        task.ReporterID,
		Rank:              task.Rank,
		ParentID:          task.ParentID,
		CreatedAt:         task.CreatedAt,
		UpdatedAt:         task.UpdatedAt,
		Children:          mapper.Map(task.Children, FromModelTask),
		CreatedByMember:   FromTeamMemberModel(task.CreatedByMember),
		Team:              FromTeamModel(task.Team),
		Project:           FromModelProject(task.Project),
	}
}

type CreateTaskProjectTaskDTO struct {
	Name        string            `json:"name" required:"true"`
	Description *string           `json:"description,omitempty" required:"false"`
	Status      models.TaskStatus `json:"status" required:"false" enum:"todo,in_progress,done" default:"todo"`
	Rank        float64           `json:"rank,omitempty" required:"false"`
}

type UpdateTaskInput struct {
	Body   stores.UpdateTaskDto
	TaskID string `path:"task-id" json:"task_id" required:"true" format:"uuid"`
}

type TaskPositionStatusDTO struct {
	Position int64             `json:"position" required:"true"`
	Status   models.TaskStatus `json:"status" required:"true" enum:"todo,in_progress,done"`
}

type TaskPositionStatusInput struct {
	TaskID string `path:"task-id" json:"task_id" required:"true" format:"uuid"`
	Body   TaskPositionStatusDTO
}

type TaskListResponse struct {
	Body *ApiPaginatedResponse[*Task]
}
type TeamTaskListParams struct {
	ProjectID string `path:"task-project-id" json:"project_id" required:"true" format:"uuid"`
	PaginatedInput
	Q                 string              `query:"q,omitempty" required:"false"`
	Status            []models.TaskStatus `query:"status,omitempty" required:"false" enum:"todo,in_progress,done"`
	CreatedByMemberID string              `query:"created_by,omitempty" required:"false" format:"uuid"`
	Ids               []string            `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	ParentID          string              `query:"parent_id,omitempty" required:"false" format:"uuid"`
	SortParams
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"subtasks"`
}

func (api *Api) TeamTaskList(ctx context.Context, input *TeamTaskListParams) (*TaskListResponse, error) {

	teamInfo := contextstore.GetContextTeamInfo(ctx)
	if teamInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	newInput := &stores.TaskFilter{}
	newInput.SortBy = input.SortBy
	newInput.SortOrder = input.SortOrder
	newInput.Page = input.Page
	newInput.PerPage = input.PerPage
	newInput.Ids = utils.ParseValidUUIDs(input.Ids...)
	newInput.Q = input.Q
	newInput.Statuses = input.Status
	newInput.TeamIds = []uuid.UUID{teamInfo.Team.ID}
	newInput.ProjectIds = utils.ParseValidUUIDs(input.ProjectID)
	if input.ParentID != "" {
		parentID, err := uuid.Parse(input.ParentID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid parent ID format", err)
		}
		newInput.ParentIds = []uuid.UUID{parentID}
	}

	tasks, err := api.App().Adapter().Task().ListTasks(ctx, newInput)
	if err != nil {
		return nil, huma.Error500InternalServerError("error listing tasks", err)
	}
	total, err := api.App().Adapter().Task().CountTasks(ctx, newInput)
	if err != nil {
		return nil, huma.Error500InternalServerError("error counting tasks", err)
	}
	return &TaskListResponse{
		Body: &ApiPaginatedResponse[*Task]{
			Data: mapper.Map(tasks, FromModelTask),
			Meta: ApiGenerateMeta(&input.PaginatedInput, total),
		},
	}, nil
}

type TaskResponse struct {
	Body *Task
}

func (api *Api) TaskUpdate(ctx context.Context, input *UpdateTaskInput) (*struct{}, error) {
	teamInfo := contextstore.GetContextTeamInfo(ctx)
	if teamInfo == nil {
		return nil, huma.Error401Unauthorized("Team not found")
	}
	id, err := uuid.Parse(input.TaskID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task ID")
	}
	task, err := api.App().Adapter().Task().FindTaskByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, huma.Error404NotFound("Task not found")
	}
	if task.AssigneeID == nil && input.Body.AssigneeID != nil {
		err = api.App().JobService().EnqueAssignedToTaskJob(ctx, &workers.AssignedToTasJobArgs{
			TaskID:              task.ID,
			AssignedByMemeberID: teamInfo.Member.ID,
			AssigneeMemberID:    *input.Body.AssigneeID,
		})
		if err != nil {
			return nil, err
		}
	}
	task.Name = input.Body.Name
	task.Description = input.Body.Description
	task.Status = models.TaskStatus(input.Body.Status)
	task.StartAt = input.Body.StartAt
	task.EndAt = input.Body.EndAt
	if task.EndAt != nil {
		err = api.App().JobService().EnqueTaskDueJob(ctx, &workers.TaskDueTodayJobArgs{
			TaskID:  task.ID,
			DueDate: *task.EndAt,
		})
		if err != nil {
			return nil, err
		}
	}
	if task.Status == models.TaskStatusDone {
		err = api.App().JobService().EnqueueTaskCompletedJob(ctx, &workers.TaskCompletedJobArgs{
			TaskID:              id,
			CompletedByMemberID: teamInfo.Member.ID,
			CompletedAt:         time.Now(),
		})
		if err != nil {
			return nil, err
		}
	}
	task.AssigneeID = input.Body.AssigneeID
	task.ReporterID = input.Body.ReporterID
	task.ParentID = input.Body.ParentID

	err = api.App().Adapter().Task().UpdateTask(ctx, task)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) UpdateTaskPositionStatus(ctx context.Context, input *TaskPositionStatusInput) (*struct{}, error) {

	if input == nil {
		return nil, huma.Error400BadRequest("Invalid input")
	}

	id, err := uuid.Parse(input.TaskID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task ID")
	}
	teamInfo := contextstore.GetContextTeamInfo(ctx)
	if teamInfo == nil {
		return nil, huma.Error401Unauthorized("team info not found")
	}
	err = api.App().Task().UpdateTaskRankStatus(ctx, id, input.Body.Position, models.TaskStatus(input.Body.Status))
	if err != nil {
		return nil, err
	}
	if models.TaskStatus(input.Body.Status) == models.TaskStatusDone {
		err = api.App().JobService().EnqueueTaskCompletedJob(ctx, &workers.TaskCompletedJobArgs{
			TaskID:              id,
			CompletedByMemberID: teamInfo.Member.ID,
			CompletedAt:         time.Now(),
		})
		if err != nil {
			return nil, err
		}
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
	err = api.App().Adapter().Task().DeleteTask(ctx, id)
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
	task, err := api.App().Adapter().Task().FindTaskByID(ctx, id)
	if err != nil {
		return nil, err
	}
	outputTask := FromModelTask(task)
	if outputTask != nil {
		if outputTask.AssigneeID != nil {
			teamMemberId := *outputTask.AssigneeID
			taskTeamInfo, err := api.app.Team().FindTeamInfoByMemberID(ctx, teamMemberId)
			if err != nil {
				return nil, err
			}
			outputTask.Assignee = FromTeamMemberModel(&taskTeamInfo.Member)
			outputTask.Assignee.User = FromUserModel(&taskTeamInfo.User)
			outputTask.Assignee.Team = FromTeamModel(&taskTeamInfo.Team)
		}
	}
	return &TaskResponse{
		Body: outputTask,
	}, nil
}
