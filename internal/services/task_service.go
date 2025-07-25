package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

type TaskFields struct {
	CreatedByMemberID *uuid.UUID        `db:"created_by_member_id" json:"created_by_member_id" nullable:"true"`
	TeamID            uuid.UUID         `db:"team_id" json:"team_id"`
	ProjectID         uuid.UUID         `db:"project_id" json:"project_id"`
	Name              string            `json:"name" required:"true"`
	Description       *string           `json:"description,omitempty" required:"false"`
	Status            models.TaskStatus `json:"status" required:"false" enum:"todo,in_progress,done" default:"todo"`
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
	UpdateTaskRankStatus(ctx context.Context, taskID uuid.UUID, position int64, status models.TaskStatus) error
	CalculateNewPosition(ctx context.Context, groupID uuid.UUID, status models.TaskStatus, targetIndex int64, excludeID uuid.UUID) (float64, error)
}
type taskService struct {
	// store   TaskStore
	adapter stores.StorageAdapterInterface

	jobService JobService
}

// CreateTask implements TaskService.
func (s *taskService) CreateTask(ctx context.Context, teamID uuid.UUID, projectID uuid.UUID, createdByMemberID uuid.UUID, input *TaskFields) (*models.Task, error) {
	setter := models.Task{
		ProjectID:         projectID,
		CreatedByMemberID: &createdByMemberID,
		TeamID:            teamID,
		Name:              input.Name,
		Description:       input.Description,
		Status:            models.TaskStatus(input.Status),
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

func (s *taskService) UpdateTaskRankStatus(ctx context.Context, taskID uuid.UUID, position int64, status models.TaskStatus) error {
	task, err := s.adapter.Task().FindTaskByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}
	rank, err := s.CalculateNewPosition(ctx, task.ProjectID, status, position, task.ID)
	if err != nil {
		return err
	}
	task.Rank = rank
	task.Status = status
	err = s.adapter.Task().UpdateTask(ctx, task)
	if err != nil {
		return err
	}
	err = s.adapter.Task().UpdateTaskProjectUpdateDate(ctx, task.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to update task project update date: %w", err)
	}
	return nil
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
