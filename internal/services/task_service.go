package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type TaskStore interface {
	CountTaskProjects(ctx context.Context, filter *shared.TaskProjectsListFilter) (int64, error)
	CountTasks(ctx context.Context, filter *shared.TaskListFilter) (int64, error)
	CreateTask(ctx context.Context, projectID uuid.UUID, input *shared.CreateTaskBaseDTO) (*models.Task, error)
	CreateTaskProject(ctx context.Context, input *shared.CreateTaskProjectDTO) (*models.TaskProject, error)
	CreateTaskProjectWithTasks(ctx context.Context, input *shared.CreateTaskProjectWithTasksDTO) (*models.TaskProject, error)
	CreateTaskWithChildren(ctx context.Context, projectID uuid.UUID, input *shared.CreateTaskWithChildrenDTO) (*models.Task, error)
	DefineTaskOrderNumberByStatus(ctx context.Context, taskId uuid.UUID, taskProjectId uuid.UUID, status models.TaskStatus, currentOrder float64, position int64) (float64, error)
	DeleteTask(ctx context.Context, taskID uuid.UUID) error
	DeleteTaskProject(ctx context.Context, taskProjectID uuid.UUID) error
	FindLastTaskOrder(ctx context.Context, taskProjectID uuid.UUID) (float64, error)
	FindTaskByID(ctx context.Context, id uuid.UUID) (*models.Task, error)
	FindTaskProjectByID(ctx context.Context, id uuid.UUID) (*models.TaskProject, error)
	ListTaskProjects(ctx context.Context, input *shared.TaskProjectsListParams) ([]*models.TaskProject, error)
	ListTasks(ctx context.Context, input *shared.TaskListParams) ([]*models.Task, error)
	LoadTaskProjectsTasks(ctx context.Context, projectIds ...uuid.UUID) ([][]*models.Task, error)
	UpdateTask(ctx context.Context, taskID uuid.UUID, input *shared.UpdateTaskBaseDTO) error
	UpdateTaskPositionStatus(ctx context.Context, taskID uuid.UUID, position int64, status models.TaskStatus) error
	UpdateTaskProject(ctx context.Context, taskProjectID uuid.UUID, input *shared.UpdateTaskProjectBaseDTO) error
	UpdateTaskProjectUpdateDate(ctx context.Context, taskProjectID uuid.UUID) error
}

type TaskService interface {
	Store() TaskStore
}
type taskService struct {
	store TaskStore
}

// Store implements TaskService.
func (t *taskService) Store() TaskStore {
	return t.store
}

func NewTaskService(store TaskStore) TaskService {
	return &taskService{
		store: store,
	}
}
