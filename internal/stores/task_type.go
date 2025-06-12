package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type DbTaskStoreInterface interface { // size=16 (0x10)
	CalculateTaskRankStatus(ctx context.Context, taskId uuid.UUID, taskProjectId uuid.UUID, status models.TaskStatus, currentRank float64, position int64) (float64, error)
	CountItems(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (int64, error)
	CountTaskProjects(ctx context.Context, filter *shared.TaskProjectsListFilter) (int64, error)
	CountTasks(ctx context.Context, filter *TaskFilter) (int64, error)
	CreateTask(ctx context.Context, task *models.Task) (*models.Task, error)
	CreateTaskFromInput(ctx context.Context, teamID uuid.UUID, projectID uuid.UUID, memberID uuid.UUID, input *shared.CreateTaskProjectTaskDTO) (*models.Task, error)
	CreateTaskProject(ctx context.Context, input *shared.CreateTaskProjectDTO) (*models.TaskProject, error)
	CreateTaskProjectWithTasks(ctx context.Context, input *shared.CreateTaskProjectWithTasksDTO) (*models.TaskProject, error)
	DeleteTask(ctx context.Context, taskID uuid.UUID) error
	DeleteTaskProject(ctx context.Context, taskProjectID uuid.UUID) error
	FindLastTaskRank(ctx context.Context, taskProjectID uuid.UUID) (float64, error)
	FindTask(ctx context.Context, task *TaskFilter) (*models.Task, error)
	FindTaskByID(ctx context.Context, id uuid.UUID) (*models.Task, error)
	FindTaskProjectByID(ctx context.Context, id uuid.UUID) (*models.TaskProject, error)
	GetTaskFirstPosition(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (float64, error)
	GetTaskLastPosition(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (float64, error)
	GetTaskPositions(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID, offset int64) ([]float64, error)
	ListTaskProjects(ctx context.Context, input *shared.TaskProjectsListParams) ([]*models.TaskProject, error)
	ListTasks(ctx context.Context, input *TaskFilter) ([]*models.Task, error)
	LoadTaskProjectsTasks(ctx context.Context, projectIds ...uuid.UUID) ([][]*models.Task, error)
	TaskWhere(task *TaskFilter) *map[string]any
	UpdateTask(ctx context.Context, task *models.Task) error
	UpdateTaskProject(ctx context.Context, taskProjectID uuid.UUID, input *shared.UpdateTaskProjectBaseDTO) error
	UpdateTaskProjectUpdateDate(ctx context.Context, taskProjectID uuid.UUID) error
	UpdateTaskRankStatus(ctx context.Context, taskID uuid.UUID, position int64, status models.TaskStatus) error
	WithTx(dbx database.Dbx) *DbTaskStore
}
