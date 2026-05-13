package apis

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/populator"
	"github.com/tkahng/playground/internal/shared"
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
	WorkflowStatusID  *uuid.UUID        `db:"workflow_status_id" json:"workflow_status_id" nullable:"true"`
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
	Parent            *Task             `db:"parent" src:"parent_id" dest:"id" table:"tasks" json:"parent,omitempty"`
}

func fromModelTask(task *models.Task) *Task {
	if task == nil {
		return nil
	}
	return &Task{
		ID:                task.ID,
		CreatedByMemberID: task.CreatedByMemberID,
		TeamID:            task.TeamID,
		ProjectID:         task.ProjectID,
		WorkflowStatusID:  task.WorkflowStatusID,
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
		Children:          mapper.Map(task.Children, fromModelTask),
		CreatedByMember:   fromTeamMemberModel(task.CreatedByMember),
		Team:              fromTeamModel(task.Team),
		Project:           FromModelProject(task.Project),
		Assignee:          fromTeamMemberModel(task.Assignee),
		Reporter:          fromTeamMemberModel(task.Reporter),
		Parent:            fromModelTask(task.Parent),
	}
}

type CreateTaskProjectTaskDTO struct {
	Name             string            `json:"name" required:"true"`
	Description      *string           `json:"description,omitempty" required:"false"`
	Status           models.TaskStatus `json:"status" required:"false" enum:"todo,in_progress,done" default:"todo"`
	WorkflowStatusID *uuid.UUID        `json:"workflow_status_id,omitempty" required:"false" format:"uuid"`
	Rank             float64           `json:"rank,omitempty" required:"false"`
}

type UpdateTaskInput struct {
	Body   stores.UpdateTaskDto
	TaskID string `path:"task-id" json:"task_id" required:"true" format:"uuid"`
}

type TaskPositionStatusDTO struct {
	Position         int64             `json:"position" required:"true"`
	Status           models.TaskStatus `json:"status,omitempty" required:"false" enum:"todo,in_progress,done"`
	WorkflowStatusID *uuid.UUID        `json:"workflow_status_id,omitempty" required:"false" format:"uuid"`
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

func (api *Api) TeamTaskListBind(humaApi huma.API) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "task-list",
			Method:      http.MethodGet,
			Path:        "/task-projects/{task-project-id}/tasks",
			Summary:     "Task list",
			Description: "List of tasks",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
			),
		},
		func(ctx context.Context, input *TeamTaskListParams) (*TaskListResponse, error) {
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
			pop := populator.New(api.App().Adapter())
			for _, task := range tasks {
				err := populator.PopulateTask(ctx, pop, task)
				if err != nil {
					return nil, err
				}
			}
			total, err := api.App().Adapter().Task().CountTasks(ctx, newInput)
			if err != nil {
				return nil, huma.Error500InternalServerError("error counting tasks", err)
			}
			return &TaskListResponse{
				Body: &ApiPaginatedResponse[*Task]{
					Data: mapper.Map(tasks, fromModelTask),
					Meta: ApiGenerateMeta(&input.PaginatedInput, total),
				},
			}, nil
		},
	)
}
func (api *Api) TeamTaskUpdateBind(humaApi huma.API) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "task-update",
			Method:      http.MethodPut,
			Path:        "/tasks/{task-id}",
			Summary:     "Task update",
			Description: "Update a task",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
				middleware.RequireTeamPermission(api.App(), shared.TeamPermissionTasksEdit),
			),
		},
		func(ctx context.Context, input *UpdateTaskInput) (*struct{}, error) {
			teamInfo := contextstore.GetContextTeamInfo(ctx)
			if teamInfo == nil {
				return nil, huma.Error401Unauthorized("Team not found")
			}
			id, err := uuid.Parse(input.TaskID)
			if err != nil {
				return nil, huma.Error400BadRequest("Invalid task ID")
			}

			// previousState is captured inside the transaction from the locked read,
			// ensuring the notification decisions below reflect the actual DB state at
			// the time of the update (not a stale snapshot from a concurrent request).
			var (
				previousStatus   models.TaskStatus
				previousDueDate  *time.Time
				previousAssignee *uuid.UUID
				newStatus        models.TaskStatus
			)

			txErr := api.App().Adapter().RunInTx(ctx, func(tx stores.StorageAdapterInterface) error {
				task, err := tx.Task().FindTaskByIDForUpdate(ctx, id)
				if err != nil {
					return err
				}
				if task == nil {
					return huma.Error404NotFound("Task not found")
				}

				previousStatus = task.Status
				previousDueDate = task.EndAt
				previousAssignee = task.AssigneeID

				if err := api.App().Task().ValidateTaskReferences(ctx, task.TeamID, task.ProjectID, &task.ID, input.Body.AssigneeID, input.Body.ReporterID, input.Body.ParentID); err != nil {
					return err
				}

				task.Name = input.Body.Name
				task.Description = input.Body.Description
				task.Status = models.TaskStatus(input.Body.Status)
				task.WorkflowStatusID = input.Body.WorkflowStatusID
				task.StartAt = input.Body.StartAt
				task.EndAt = input.Body.EndAt
				task.AssigneeID = input.Body.AssigneeID
				task.ReporterID = input.Body.ReporterID
				task.ParentID = input.Body.ParentID

				if err := tx.Task().UpdateTask(ctx, task); err != nil {
					return err
				}
				newStatus = task.Status
				return nil
			})
			if txErr != nil {
				return nil, txErr
			}

			newAssignee := previousAssignee == nil && input.Body.AssigneeID != nil
			differentAssignee := previousAssignee != nil && input.Body.AssigneeID != nil && *previousAssignee != *input.Body.AssigneeID
			if newAssignee || differentAssignee {
				err = api.App().JobService().EnqueueAssignedToTaskJob(ctx, &workers.AssignedToTaskJobArgs{
					TaskID:             id,
					AssignedByMemberID: teamInfo.Member.ID,
					AssigneeMemberID:   *input.Body.AssigneeID,
				})
				if err != nil {
					return nil, err
				}
			}

			newDueDate := previousDueDate == nil && input.Body.EndAt != nil
			differentDueDate := previousDueDate != nil && input.Body.EndAt != nil && *previousDueDate != *input.Body.EndAt
			if newDueDate || differentDueDate {
				dueDate := *input.Body.EndAt
				if dueDate.Before(time.Now()) {
					dueDate = time.Now().Add(10 * time.Second)
				}
				err = api.App().JobService().EnqueTaskDueJob(ctx, &workers.TaskDueTodayJobArgs{
					TaskID:  id,
					DueDate: dueDate,
				})
				if err != nil {
					return nil, err
				}
				err = api.App().JobService().EnqueueTaskOverdueJob(ctx, &workers.TaskOverdueJobArgs{
					TaskID:  id,
					DueDate: dueDate,
				})
				if err != nil {
					return nil, err
				}
			}

			newDoneStatus := previousStatus != newStatus && newStatus == models.TaskStatusDone
			if newDoneStatus {
				err = api.App().JobService().EnqueueTaskCompletedJob(ctx, &workers.TaskCompletedJobArgs{
					TaskID:              id,
					CompletedByMemberID: teamInfo.Member.ID,
					CompletedAt:         time.Now(),
				})
				if err != nil {
					return nil, err
				}
			}

			statusChanged := previousStatus != newStatus && newStatus != models.TaskStatusDone
			if statusChanged {
				err = api.App().JobService().EnqueueTaskStatusChangedJob(ctx, &workers.TaskStatusChangedJobArgs{
					TaskID:            id,
					OldStatus:         string(previousStatus),
					NewStatus:         string(newStatus),
					ChangedByMemberID: teamInfo.Member.ID,
				})
				if err != nil {
					return nil, err
				}
			}

			return nil, nil
		},
	)
}

type TaskResponse struct {
	Body *Task
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
	task, err := api.App().Adapter().Task().FindTaskByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, huma.Error404NotFound("Task not found")
	}
	err = api.App().Task().UpdateTaskRankStatus(ctx, id, input.Body.Position, models.TaskStatus(input.Body.Status), input.Body.WorkflowStatusID)
	if err != nil {
		return nil, err
	}

	updatedTask, err := api.App().Adapter().Task().FindTaskByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if updatedTask == nil {
		return nil, huma.Error404NotFound("Task not found")
	}
	newStatus := updatedTask.Status
	if newStatus == models.TaskStatusDone {
		if task.Status != models.TaskStatusDone {
			err = api.App().JobService().EnqueueTaskCompletedJob(ctx, &workers.TaskCompletedJobArgs{
				TaskID:              id,
				CompletedByMemberID: teamInfo.Member.ID,
				CompletedAt:         time.Now(),
			})
			if err != nil {
				return nil, err
			}
		}
	} else if task.Status != newStatus {
		err = api.App().JobService().EnqueueTaskStatusChangedJob(ctx, &workers.TaskStatusChangedJobArgs{
			TaskID:            id,
			OldStatus:         string(task.Status),
			NewStatus:         string(newStatus),
			ChangedByMemberID: teamInfo.Member.ID,
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
	pop := populator.New(api.App().Adapter())

	task, err := pop.GetTaskByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, huma.Error404NotFound("Task not found")
	}
	err = populator.PopulateTask(ctx, pop, task)
	if err != nil {
		return nil, err
	}
	outputTask := fromModelTask(task)
	return &TaskResponse{
		Body: outputTask,
	}, nil
}
