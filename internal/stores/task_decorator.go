package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type TaskDecorator struct {
	Delegate                        *DbTaskStore
	CalculateTaskRankStatusFunc     func(ctx context.Context, taskId uuid.UUID, taskProjectId uuid.UUID, status models.TaskStatus, currentRank float64, position int64) (float64, error)
	CountItemsFunc                  func(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (int64, error)
	CountTaskProjectsFunc           func(ctx context.Context, filter *shared.TaskProjectsListFilter) (int64, error)
	CountTasksFunc                  func(ctx context.Context, filter *TaskFilter) (int64, error)
	CreateTaskFunc                  func(ctx context.Context, task *models.Task) (*models.Task, error)
	CreateTaskFromInputFunc         func(ctx context.Context, teamID uuid.UUID, projectID uuid.UUID, memberID uuid.UUID, input *shared.CreateTaskProjectTaskDTO) (*models.Task, error)
	CreateTaskProjectFunc           func(ctx context.Context, input *shared.CreateTaskProjectDTO) (*models.TaskProject, error)
	CreateTaskProjectWithTasksFunc  func(ctx context.Context, input *shared.CreateTaskProjectWithTasksDTO) (*models.TaskProject, error)
	DeleteTaskFunc                  func(ctx context.Context, taskID uuid.UUID) error
	DeleteTaskProjectFunc           func(ctx context.Context, taskProjectID uuid.UUID) error
	FindLastTaskRankFunc            func(ctx context.Context, taskProjectID uuid.UUID) (float64, error)
	FindTaskFunc                    func(ctx context.Context, task *TaskFilter) (*models.Task, error)
	FindTaskByIDFunc                func(ctx context.Context, id uuid.UUID) (*models.Task, error)
	FindTaskProjectByIDFunc         func(ctx context.Context, id uuid.UUID) (*models.TaskProject, error)
	GetTaskFirstPositionFunc        func(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (float64, error)
	GetTaskLastPositionFunc         func(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (float64, error)
	GetTaskPositionsFunc            func(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID, offset int64) ([]float64, error)
	ListTaskProjectsFunc            func(ctx context.Context, input *shared.TaskProjectsListParams) ([]*models.TaskProject, error)
	ListTasksFunc                   func(ctx context.Context, input *TaskFilter) ([]*models.Task, error)
	LoadTaskProjectsTasksFunc       func(ctx context.Context, projectIds ...uuid.UUID) ([][]*models.Task, error)
	TaskWhereFunc                   func(task *TaskFilter) *map[string]any
	UpdateTaskFunc                  func(ctx context.Context, task *models.Task) error
	UpdateTaskProjectFunc           func(ctx context.Context, taskProjectID uuid.UUID, input *shared.UpdateTaskProjectBaseDTO) error
	UpdateTaskProjectUpdateDateFunc func(ctx context.Context, taskProjectID uuid.UUID) error
	UpdateTaskRankStatusFunc        func(ctx context.Context, taskID uuid.UUID, position int64, status models.TaskStatus) error
	WithTxFunc                      func(dbx database.Dbx) *DbTaskStore
}

// CreateTaskProject implements DbTaskStoreInterface.
func (t *TaskDecorator) CreateTaskProject(ctx context.Context, input *shared.CreateTaskProjectDTO) (*models.TaskProject, error) {
	if t.CreateTaskProjectFunc != nil {
		return t.CreateTaskProjectFunc(ctx, input)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.CreateTaskProject(ctx, input)
}

// CreateTaskProjectWithTasks implements DbTaskStoreInterface.
func (t *TaskDecorator) CreateTaskProjectWithTasks(ctx context.Context, input *shared.CreateTaskProjectWithTasksDTO) (*models.TaskProject, error) {
	if t.CreateTaskProjectWithTasksFunc != nil {
		return t.CreateTaskProjectWithTasksFunc(ctx, input)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.CreateTaskProjectWithTasks(ctx, input)
}

// DeleteTask implements DbTaskStoreInterface.
func (t *TaskDecorator) DeleteTask(ctx context.Context, taskID uuid.UUID) error {
	if t.DeleteTaskFunc != nil {
		return t.DeleteTaskFunc(ctx, taskID)
	}
	if t.Delegate == nil {
		return ErrDelegateNil
	}
	return t.Delegate.DeleteTask(ctx, taskID)
}

// DeleteTaskProject implements DbTaskStoreInterface.
func (t *TaskDecorator) DeleteTaskProject(ctx context.Context, taskProjectID uuid.UUID) error {
	if t.DeleteTaskProjectFunc != nil {
		return t.DeleteTaskProjectFunc(ctx, taskProjectID)
	}
	if t.Delegate == nil {
		return ErrDelegateNil
	}
	return t.Delegate.DeleteTaskProject(ctx, taskProjectID)
}

// FindLastTaskRank implements DbTaskStoreInterface.
func (t *TaskDecorator) FindLastTaskRank(ctx context.Context, taskProjectID uuid.UUID) (float64, error) {
	if t.FindLastTaskRankFunc != nil {
		return t.FindLastTaskRankFunc(ctx, taskProjectID)
	}
	if t.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return t.Delegate.FindLastTaskRank(ctx, taskProjectID)
}

// FindTask implements DbTaskStoreInterface.
func (t *TaskDecorator) FindTask(ctx context.Context, task *TaskFilter) (*models.Task, error) {
	if t.FindTaskFunc != nil {
		return t.FindTaskFunc(ctx, task)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindTask(ctx, task)
}

// FindTaskByID implements DbTaskStoreInterface.
func (t *TaskDecorator) FindTaskByID(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	if t.FindTaskByIDFunc != nil {
		return t.FindTaskByIDFunc(ctx, id)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindTaskByID(ctx, id)
}

// FindTaskProjectByID implements DbTaskStoreInterface.
func (t *TaskDecorator) FindTaskProjectByID(ctx context.Context, id uuid.UUID) (*models.TaskProject, error) {
	if t.FindTaskProjectByIDFunc != nil {
		return t.FindTaskProjectByIDFunc(ctx, id)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindTaskProjectByID(ctx, id)
}

// GetTaskFirstPosition implements DbTaskStoreInterface.
func (t *TaskDecorator) GetTaskFirstPosition(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (float64, error) {
	if t.GetTaskFirstPositionFunc != nil {
		return t.GetTaskFirstPositionFunc(ctx, projectID, status, excludeID)
	}
	if t.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return t.Delegate.GetTaskFirstPosition(ctx, projectID, status, excludeID)
}

// GetTaskLastPosition implements DbTaskStoreInterface.
func (t *TaskDecorator) GetTaskLastPosition(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (float64, error) {
	if t.GetTaskLastPositionFunc != nil {
		return t.GetTaskLastPositionFunc(ctx, projectID, status, excludeID)
	}
	if t.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return t.Delegate.GetTaskLastPosition(ctx, projectID, status, excludeID)
}

// GetTaskPositions implements DbTaskStoreInterface.
func (t *TaskDecorator) GetTaskPositions(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID, offset int64) ([]float64, error) {
	if t.GetTaskPositionsFunc != nil {
		return t.GetTaskPositionsFunc(ctx, projectID, status, excludeID, offset)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.GetTaskPositions(ctx, projectID, status, excludeID, offset)
}

// ListTaskProjects implements DbTaskStoreInterface.
func (t *TaskDecorator) ListTaskProjects(ctx context.Context, input *shared.TaskProjectsListParams) ([]*models.TaskProject, error) {
	if t.ListTaskProjectsFunc != nil {
		return t.ListTaskProjectsFunc(ctx, input)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.ListTaskProjects(ctx, input)
}

// ListTasks implements DbTaskStoreInterface.
func (t *TaskDecorator) ListTasks(ctx context.Context, input *TaskFilter) ([]*models.Task, error) {
	if t.ListTasksFunc != nil {
		return t.ListTasksFunc(ctx, input)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.ListTasks(ctx, input)
}

// LoadTaskProjectsTasks implements DbTaskStoreInterface.
func (t *TaskDecorator) LoadTaskProjectsTasks(ctx context.Context, projectIds ...uuid.UUID) ([][]*models.Task, error) {
	if t.LoadTaskProjectsTasksFunc != nil {
		return t.LoadTaskProjectsTasksFunc(ctx, projectIds...)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.LoadTaskProjectsTasks(ctx, projectIds...)
}

// TaskWhere implements DbTaskStoreInterface.
func (t *TaskDecorator) TaskWhere(task *TaskFilter) *map[string]any {
	if t.TaskWhereFunc != nil {
		return t.TaskWhereFunc(task)
	}
	if t.Delegate == nil {
		return nil
	}
	return t.Delegate.TaskWhere(task)
}

// UpdateTask implements DbTaskStoreInterface.
func (t *TaskDecorator) UpdateTask(ctx context.Context, task *models.Task) error {
	if t.UpdateTaskFunc != nil {
		return t.UpdateTaskFunc(ctx, task)
	}
	if t.Delegate == nil {
		return ErrDelegateNil
	}
	return t.Delegate.UpdateTask(ctx, task)
}

// UpdateTaskProject implements DbTaskStoreInterface.
func (t *TaskDecorator) UpdateTaskProject(ctx context.Context, taskProjectID uuid.UUID, input *shared.UpdateTaskProjectBaseDTO) error {
	if t.UpdateTaskProjectFunc != nil {
		return t.UpdateTaskProjectFunc(ctx, taskProjectID, input)
	}
	if t.Delegate == nil {
		return ErrDelegateNil
	}
	return t.Delegate.UpdateTaskProject(ctx, taskProjectID, input)
}

// UpdateTaskProjectUpdateDate implements DbTaskStoreInterface.
func (t *TaskDecorator) UpdateTaskProjectUpdateDate(ctx context.Context, taskProjectID uuid.UUID) error {
	if t.UpdateTaskProjectUpdateDateFunc != nil {
		return t.UpdateTaskProjectUpdateDateFunc(ctx, taskProjectID)
	}
	if t.Delegate == nil {
		return ErrDelegateNil
	}
	return t.Delegate.UpdateTaskProjectUpdateDate(ctx, taskProjectID)
}

// UpdateTaskRankStatus implements DbTaskStoreInterface.
func (t *TaskDecorator) UpdateTaskRankStatus(ctx context.Context, taskID uuid.UUID, position int64, status models.TaskStatus) error {
	if t.UpdateTaskRankStatusFunc != nil {
		return t.UpdateTaskRankStatusFunc(ctx, taskID, position, status)
	}
	if t.Delegate == nil {
		return ErrDelegateNil
	}
	return t.Delegate.UpdateTaskRankStatus(ctx, taskID, position, status)
}

// WithTx implements DbTaskStoreInterface.
func (t *TaskDecorator) WithTx(dbx database.Dbx) *DbTaskStore {
	if t.WithTxFunc != nil {
		return t.WithTxFunc(dbx)
	}
	if t.Delegate == nil {
		return nil
	}
	return t.Delegate.WithTx(dbx)
}

// CalculateTaskRankStatus implements DbTaskStoreInterface.
func (t *TaskDecorator) CalculateTaskRankStatus(ctx context.Context, taskId uuid.UUID, taskProjectId uuid.UUID, status models.TaskStatus, currentRank float64, position int64) (float64, error) {
	if t.CalculateTaskRankStatusFunc != nil {
		return t.CalculateTaskRankStatusFunc(ctx, taskId, taskProjectId, status, currentRank, position)
	}
	if t.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return t.Delegate.CalculateTaskRankStatus(ctx, taskId, taskProjectId, status, currentRank, position)
}

// CountItems implements DbTaskStoreInterface.
func (t *TaskDecorator) CountItems(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (int64, error) {
	if t.CountItemsFunc != nil {
		return t.CountItemsFunc(ctx, projectID, status, excludeID)
	}
	if t.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return t.Delegate.CountItems(ctx, projectID, status, excludeID)
}

// CountTaskProjects implements DbTaskStoreInterface.
func (t *TaskDecorator) CountTaskProjects(ctx context.Context, filter *shared.TaskProjectsListFilter) (int64, error) {
	if t.CountTaskProjectsFunc != nil {
		return t.CountTaskProjectsFunc(ctx, filter)
	}
	if t.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return t.Delegate.CountTaskProjects(ctx, filter)
}

// CountTasks implements DbTaskStoreInterface.
func (t *TaskDecorator) CountTasks(ctx context.Context, filter *TaskFilter) (int64, error) {
	if t.CountTasksFunc != nil {
		return t.CountTasksFunc(ctx, filter)
	}
	if t.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return t.Delegate.CountTasks(ctx, filter)
}

// CreateTask implements DbTaskStoreInterface.
func (t *TaskDecorator) CreateTask(ctx context.Context, task *models.Task) (*models.Task, error) {
	if t.CreateTaskFunc != nil {
		return t.CreateTaskFunc(ctx, task)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.CreateTask(ctx, task)
}

// CreateTaskFromInput implements DbTaskStoreInterface.
func (t *TaskDecorator) CreateTaskFromInput(ctx context.Context, teamID uuid.UUID, projectID uuid.UUID, memberID uuid.UUID, input *shared.CreateTaskProjectTaskDTO) (*models.Task, error) {
	if t.CreateTaskFromInputFunc != nil {
		return t.CreateTaskFromInputFunc(ctx, teamID, projectID, memberID, input)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.CreateTaskFromInput(ctx, teamID, projectID, memberID, input)
}

var _ DbTaskStoreInterface = (*TaskDecorator)(nil)
