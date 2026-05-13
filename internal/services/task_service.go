package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/apierrors"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/types"
)

type TaskFields struct {
	CreatedByMemberID *uuid.UUID        `db:"created_by_member_id" json:"created_by_member_id" nullable:"true"`
	TeamID            uuid.UUID         `db:"team_id" json:"team_id"`
	ProjectID         uuid.UUID         `db:"project_id" json:"project_id"`
	Name              string            `json:"name" required:"true"`
	Description       *string           `json:"description,omitempty" required:"false"`
	Status            models.TaskStatus `json:"status" required:"false" enum:"todo,in_progress,done" default:"todo"`
	WorkflowStatusID  *uuid.UUID        `db:"workflow_status_id" json:"workflow_status_id" nullable:"true" format:"uuid"`
	StartAt           *time.Time        `db:"start_at" json:"start_at"  nullable:"true"`
	EndAt             *time.Time        `db:"end_at" json:"end_at" nullable:"true"`
	AssigneeID        *uuid.UUID        `db:"assignee_id" json:"assignee_id" nullable:"true"`
	ReporterID        *uuid.UUID        `db:"reporter_id" json:"reporter_id" nullable:"true"`
	Rank              float64           `json:"rank,omitempty" required:"false"`
	Position          *int64            `json:"position,omitempty" required:"false"`
	ParentID          *uuid.UUID        `db:"parent_id" json:"parent_id" nullable:"true"`
}
type TaskService interface {
	CreateTask(ctx context.Context, teamID uuid.UUID, projectID uuid.UUID, createdByMemberID uuid.UUID, input *TaskFields) (*models.Task, error)

	// CreateTaskWithChildren(ctx context.Context, teamID uuid.UUID, projectID uuid.UUID, memberID uuid.UUID, input *shared.CreateTaskWithChildrenDTO) (*models.Task, error)
	UpdateTaskRankStatus(ctx context.Context, taskID uuid.UUID, position int64, status models.TaskStatus, workflowStatusID *uuid.UUID) error
	CalculateNewPosition(ctx context.Context, groupID uuid.UUID, status models.TaskStatus, targetIndex int64, excludeID uuid.UUID) (float64, error)
	ValidateTaskReferences(ctx context.Context, teamID uuid.UUID, projectID uuid.UUID, currentTaskID *uuid.UUID, assigneeID *uuid.UUID, reporterID *uuid.UUID, parentID *uuid.UUID) error
}
type taskService struct {
	// store   TaskStore
	adapter stores.StorageAdapterInterface

	jobService JobService
}

// CreateTask implements TaskService.
func (s *taskService) CreateTask(ctx context.Context, teamID uuid.UUID, projectID uuid.UUID, createdByMemberID uuid.UUID, input *TaskFields) (*models.Task, error) {
	if err := s.ValidateTaskReferences(ctx, teamID, projectID, nil, input.AssigneeID, input.ReporterID, input.ParentID); err != nil {
		return nil, err
	}

	setter := models.Task{
		ProjectID:         projectID,
		CreatedByMemberID: &createdByMemberID,
		TeamID:            teamID,
		Name:              input.Name,
		Description:       input.Description,
		Status:            models.TaskStatus(input.Status),
		WorkflowStatusID:  input.WorkflowStatusID,
		Rank:              input.Rank,
		AssigneeID:        input.AssigneeID,
		ReporterID:        input.ReporterID,
		StartAt:           input.StartAt,
		EndAt:             input.EndAt,
		ParentID:          input.ParentID,
	}
	task, err := s.adapter.Task().CreateTask(ctx, &setter)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *taskService) ValidateTaskReferences(ctx context.Context, teamID uuid.UUID, projectID uuid.UUID, currentTaskID *uuid.UUID, assigneeID *uuid.UUID, reporterID *uuid.UUID, parentID *uuid.UUID) error {
	if err := s.validateTaskMember(ctx, teamID, assigneeID, "assignee"); err != nil {
		return err
	}
	if err := s.validateTaskMember(ctx, teamID, reporterID, "reporter"); err != nil {
		return err
	}
	return s.validateParentTask(ctx, teamID, projectID, currentTaskID, parentID)
}

func (s *taskService) validateTaskMember(ctx context.Context, teamID uuid.UUID, memberID *uuid.UUID, field string) error {
	if memberID == nil {
		return nil
	}
	member, err := s.adapter.TeamMember().FindTeamMember(ctx, &stores.TeamMemberFilter{
		Ids: []uuid.UUID{*memberID},
		Active: types.OptionalParam[bool]{
			Value: true,
			IsSet: true,
		},
	})
	if err != nil {
		return err
	}
	if member == nil || member.TeamID != teamID {
		return apierrors.BadRequest(fmt.Sprintf("%s must be an active member of the task team", field))
	}
	return nil
}

func (s *taskService) validateParentTask(ctx context.Context, teamID uuid.UUID, projectID uuid.UUID, currentTaskID *uuid.UUID, parentID *uuid.UUID) error {
	if parentID == nil {
		return nil
	}
	if currentTaskID != nil && *parentID == *currentTaskID {
		return apierrors.BadRequest("parent task cannot be the task itself")
	}
	parent, err := s.adapter.Task().FindTaskByID(ctx, *parentID)
	if err != nil {
		return err
	}
	if parent == nil {
		return apierrors.BadRequest("parent task not found")
	}
	if parent.TeamID != teamID || parent.ProjectID != projectID {
		return apierrors.BadRequest("parent task must belong to the same task project")
	}
	return nil
}

func NewTaskService(adapter stores.StorageAdapterInterface, jobService JobService) TaskService {
	return &taskService{
		adapter:    adapter,
		jobService: jobService,
	}
}

var _ TaskService = (*taskService)(nil)

// FindAndUpdateTask implements TaskService.
type UpdateTaskDto struct {
	Name        string            `db:"name" json:"name"`
	Description *string           `db:"description" json:"description"`
	Status      models.TaskStatus `db:"status" json:"status" enum:"todo,in_progress,done"`
	StartAt     *time.Time        `db:"start_at" json:"start_at" nullable:"true"`
	EndAt       *time.Time        `db:"end_at" json:"end_at" nullable:"true"`
	AssigneeID  *uuid.UUID        `db:"assignee_id" json:"assignee_id" nullable:"true"`
	ReporterID  *uuid.UUID        `db:"reporter_id" json:"reporter_id" nullable:"true"`
	ParentID    *uuid.UUID        `db:"parent_id" json:"parent_id" nullable:"true"`
}

func (s *taskService) UpdateTaskRankStatus(ctx context.Context, taskID uuid.UUID, position int64, status models.TaskStatus, workflowStatusID *uuid.UUID) error {
	return s.adapter.Task().UpdateTaskRankStatus(ctx, taskID, position, status, workflowStatusID)
}

func (s *taskService) CalculateNewPosition(ctx context.Context, groupID uuid.UUID, status models.TaskStatus, targetIndex int64, excludeID uuid.UUID) (float64, error) {
	count, err := s.adapter.Task().CountItems(ctx, groupID, status, excludeID)
	if err != nil {
		return 0, fmt.Errorf("failed to count items: %w", err)
	}

	if count == 0 {
		return 1000.0, nil
	}

	if targetIndex <= 0 {
		// Insert at beginning
		firstPos, err := s.adapter.Task().GetTaskFirstPosition(ctx, groupID, status, excludeID)
		if err != nil {
			return 0, fmt.Errorf("failed to get first rank: %w", err)
		}
		return firstPos - 1000.0, nil
	}

	if targetIndex >= count {
		// Insert at end
		lastPos, err := s.adapter.Task().GetTaskLastPosition(ctx, groupID, status, excludeID)
		if err != nil {
			return 0, fmt.Errorf("failed to get last rank: %w", err)
		}
		return lastPos + 1000.0, nil
	}

	// Insert between two ranks
	ranks, err := s.adapter.Task().GetTaskPositions(ctx, groupID, status, excludeID, targetIndex-1)
	if err != nil {
		return 0, fmt.Errorf("failed to get ranks: %w", err)
	}

	if len(ranks) < 2 {
		return 0, fmt.Errorf("insufficient ranks returned")
	}

	return (ranks[0] + ranks[1]) / 2.0, nil
}

// // CreateTaskWithChildren implements TaskService.
// func (t *taskService) CreateTaskWithChildren(ctx context.Context, teamId uuid.UUID, projectID uuid.UUID, memberID uuid.UUID, input *shared.CreateTaskWithChildrenDTO) (*models.Task, error) {
// 	task, err := t.adapter.Task().CreateTaskFromInput(ctx, teamId, projectID, memberID, &input.CreateTaskProjectTaskDTO)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// for _, child := range input.Children {
// 	// 	childTask, err := CreateTask(ctx, userID, projectID, &child)
// 	// 	if err != nil {
// 	// 		return nil, err
// 	// 	}
// 	// }
// 	return task, nil
// }

func (t *taskService) Adapter() stores.StorageAdapterInterface {
	return t.adapter
}
