package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
)

type TaskStore interface {
	CreateTask(ctx context.Context, task *models.Task) (*models.Task, error)
	FindTask(ctx context.Context, task *models.Task) (*models.Task, error)
	UpdateTask(ctx context.Context, task *models.Task) error
	// UpdateTask(ctx context.Context, task *models.Task) error
	// task methods
	CountTasks(ctx context.Context, filter *shared.TaskListFilter) (int64, error)
	CreateTaskFromInput(ctx context.Context, teamID uuid.UUID, projectID uuid.UUID, memberID uuid.UUID, input *shared.CreateTaskProjectTaskDTO) (*models.Task, error)
	DeleteTask(ctx context.Context, taskID uuid.UUID) error
	FindLastTaskRank(ctx context.Context, taskProjectID uuid.UUID) (float64, error)
	FindTaskByID(ctx context.Context, id uuid.UUID) (*models.Task, error)
	ListTasks(ctx context.Context, input *shared.TaskListParams) ([]*models.Task, error)

	CalculateTaskRankStatus(ctx context.Context, taskId uuid.UUID, taskProjectId uuid.UUID, status models.TaskStatus, currentRank float64, position int64) (float64, error)
	UpdateTaskRankStatus(ctx context.Context, taskID uuid.UUID, position int64, status models.TaskStatus) error

	// task project methods
	LoadTaskProjectsTasks(ctx context.Context, projectIds ...uuid.UUID) ([][]*models.Task, error)
	FindTaskProjectByID(ctx context.Context, id uuid.UUID) (*models.TaskProject, error)
	ListTaskProjects(ctx context.Context, input *shared.TaskProjectsListParams) ([]*models.TaskProject, error)
	UpdateTaskProject(ctx context.Context, taskProjectID uuid.UUID, input *shared.UpdateTaskProjectBaseDTO) error
	UpdateTaskProjectUpdateDate(ctx context.Context, taskProjectID uuid.UUID) error
	DeleteTaskProject(ctx context.Context, taskProjectID uuid.UUID) error
	CreateTaskProject(ctx context.Context, input *shared.CreateTaskProjectDTO) (*models.TaskProject, error)
	CreateTaskProjectWithTasks(ctx context.Context, input *shared.CreateTaskProjectWithTasksDTO) (*models.TaskProject, error)
	CountTaskProjects(ctx context.Context, filter *shared.TaskProjectsListFilter) (int64, error)
	CountItems(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (int64, error)
	GetTaskFirstPosition(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (float64, error)
	GetTaskLastPosition(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (float64, error)
	GetTaskPositions(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID, offset int64) ([]float64, error)
}

type TaskService interface {
	FindAndUpdateTask(ctx context.Context, taskID uuid.UUID, input *shared.UpdateTaskDto) error

	CreateTaskWithChildren(ctx context.Context, teamID uuid.UUID, projectID uuid.UUID, memberID uuid.UUID, input *shared.CreateTaskWithChildrenDTO) (*models.Task, error)
	UpdateTaskRankStatus(ctx context.Context, taskID uuid.UUID, position int64, status models.TaskStatus) error
	CalculateNewPosition(ctx context.Context, groupID uuid.UUID, status models.TaskStatus, targetIndex int64, excludeID uuid.UUID) (float64, error)
}
type taskService struct {
	// store   TaskStore
	adapter *stores.StorageAdapter
}

// FindAndUpdateTask implements TaskService.
func (s *taskService) FindAndUpdateTask(ctx context.Context, taskID uuid.UUID, input *shared.UpdateTaskDto) error {
	task, err := s.adapter.Task().FindTask(ctx, &stores.TaskFilter{Ids: []uuid.UUID{taskID}})
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	task.Name = input.Name
	task.Description = input.Description
	task.Status = models.TaskStatus(input.Status)
	task.StartAt = input.StartAt
	task.EndAt = input.EndAt
	task.AssigneeID = input.AssigneeID
	task.ReporterID = input.ReporterID
	task.ParentID = input.ParentID
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

// CreateTaskWithChildren implements TaskService.
func (t *taskService) CreateTaskWithChildren(ctx context.Context, teamId uuid.UUID, projectID uuid.UUID, memberID uuid.UUID, input *shared.CreateTaskWithChildrenDTO) (*models.Task, error) {
	task, err := t.adapter.Task().CreateTaskFromInput(ctx, teamId, projectID, memberID, &input.CreateTaskProjectTaskDTO)
	if err != nil {
		return nil, err
	}
	// for _, child := range input.Children {
	// 	childTask, err := CreateTask(ctx, userID, projectID, &child)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	return task, nil
}

func (t *taskService) Adapter() stores.StorageAdapterInterface {
	return t.adapter
}

func NewTaskService(adapter *stores.StorageAdapter) TaskService {
	return &taskService{
		adapter: adapter,
	}
}
