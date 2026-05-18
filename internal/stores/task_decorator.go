package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
)

type TaskDecorator struct {
	Delegate                        *DbTaskStore
	CalculateTaskRankStatusFunc     func(ctx context.Context, taskId uuid.UUID, taskProjectId uuid.UUID, status models.TaskStatus, currentRank float64, position int64) (float64, error)
	CountItemsFunc                  func(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (int64, error)
	CountTaskProjectsFunc           func(ctx context.Context, filter *TaskProjectsFilter) (int64, error)
	CountTasksFunc                  func(ctx context.Context, filter *TaskFilter) (int64, error)
	CreateTaskFunc                  func(ctx context.Context, task *models.Task) (*models.Task, error)
	CreateTaskFromInputFunc         func(ctx context.Context, teamID uuid.UUID, projectID uuid.UUID, memberID uuid.UUID, input *CreateTaskProjectTaskDTO) (*models.Task, error)
	CreateTaskProjectFunc           func(ctx context.Context, input *CreateTaskProjectDTO) (*models.TaskProject, error)
	CreateTaskProjectWithTasksFunc  func(ctx context.Context, input *CreateTaskProjectWithTasksDTO) (*models.TaskProject, error)
	DeleteTaskFunc                  func(ctx context.Context, taskID uuid.UUID) error
	DeleteTaskProjectFunc           func(ctx context.Context, taskProjectID uuid.UUID) error
	FindLastTaskRankFunc            func(ctx context.Context, taskProjectID uuid.UUID) (float64, error)
	FindTaskFunc                    func(ctx context.Context, task *TaskFilter) (*models.Task, error)
	FindTaskByIDFunc                func(ctx context.Context, id uuid.UUID) (*models.Task, error)
	FindTaskProjectByIDFunc         func(ctx context.Context, id uuid.UUID) (*models.TaskProject, error)
	FindWorkflowByIDFunc             func(ctx context.Context, id uuid.UUID) (*models.Workflow, error)
	FindWorkflowStatusByIDFunc       func(ctx context.Context, id uuid.UUID) (*models.WorkflowStatus, error)
	FindWorkflowStatusesByIDsFunc    func(ctx context.Context, ids ...uuid.UUID) ([]*models.WorkflowStatus, error)
	GetTaskFirstPositionFunc        func(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (float64, error)
	GetTaskLastPositionFunc         func(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (float64, error)
	GetTaskPositionsFunc            func(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID, offset int64) ([]float64, error)
	ListWorkflowsFunc               func(ctx context.Context, filter *WorkflowFilter) ([]*models.Workflow, error)
	ListTaskProjectsFunc            func(ctx context.Context, input *TaskProjectsFilter) ([]*models.TaskProject, error)
	ListTasksFunc                   func(ctx context.Context, input *TaskFilter) ([]*models.Task, error)
	LoadWorkflowStatusesFunc        func(ctx context.Context, workflowIds ...uuid.UUID) ([][]*models.WorkflowStatus, error)
	CreateWorkflowStatusFunc        func(ctx context.Context, workflowID uuid.UUID, input *CreateWorkflowStatusDTO) (*models.WorkflowStatus, error)
	UpdateWorkflowStatusFunc        func(ctx context.Context, workflowStatusID uuid.UUID, input *UpdateWorkflowStatusDTO) (*models.WorkflowStatus, error)
	ReorderWorkflowStatusesFunc     func(ctx context.Context, workflowID uuid.UUID, statusIDs []uuid.UUID) ([]*models.WorkflowStatus, error)
	DeleteWorkflowStatusFunc        func(ctx context.Context, workflowStatusID uuid.UUID) error
	LoadTaskProjectsTasksFunc       func(ctx context.Context, projectIds ...uuid.UUID) ([][]*models.Task, error)
	TaskWhereFunc                   func(task *TaskFilter) *map[string]any
	UpdateTaskFunc                  func(ctx context.Context, task *models.Task) error
	UpdateTaskProjectFunc           func(ctx context.Context, taskProjectID uuid.UUID, input *UpdateTaskProjectBaseDTO) error
	UpdateTaskProjectUpdateDateFunc func(ctx context.Context, taskProjectID uuid.UUID) error
	UpdateTaskRankStatusFunc        func(ctx context.Context, taskID uuid.UUID, position int64, status models.TaskStatus, workflowStatusID *uuid.UUID) error
	WithTxFunc                      func(dbx database.Dbx) *DbTaskStore
	GetTeamTaskStatsFunc            func(ctx context.Context, teamId uuid.UUID) (*models.TaskStats, error)
	FindAndUpdateTaskFunc           func(ctx context.Context, taskID uuid.UUID, input *UpdateTaskDto) error
	FindTasksDueTodayFunc           func(ctx context.Context) ([]*models.Task, error)
	FindTasksOverdueFunc            func(ctx context.Context) ([]*models.Task, error)
	ArchiveWorkflowFunc             func(ctx context.Context, workflowID uuid.UUID) (*models.Workflow, error)
	CreateWorkflowFunc              func(ctx context.Context, input *CreateWorkflowDTO) (*models.Workflow, error)
	UpdateWorkflowFunc              func(ctx context.Context, workflowID uuid.UUID, input *UpdateWorkflowDTO) (*models.Workflow, error)
	DeleteWorkflowFunc              func(ctx context.Context, workflowID uuid.UUID) error
	SetDefaultWorkflowFunc          func(ctx context.Context, workflowID uuid.UUID) (*models.Workflow, error)
}

// FindAndUpdateTask implements DbTaskStoreInterface.
func (t *TaskDecorator) FindAndUpdateTask(ctx context.Context, taskID uuid.UUID, input *UpdateTaskDto) error {
	if t.FindAndUpdateTaskFunc != nil {
		return t.FindAndUpdateTaskFunc(ctx, taskID, input)
	}
	if t.Delegate == nil {
		return ErrDelegateNil
	}
	return t.Delegate.FindAndUpdateTask(ctx, taskID, input)
}

// GetTeamTaskStats implements DbTaskStoreInterface.
func (t *TaskDecorator) GetTeamTaskStats(ctx context.Context, teamId uuid.UUID) (*models.TaskStats, error) {
	if t.GetTeamTaskStatsFunc != nil {
		return t.GetTeamTaskStatsFunc(ctx, teamId)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.GetTeamTaskStats(ctx, teamId)
}

// CreateTaskProject implements DbTaskStoreInterface.
func (t *TaskDecorator) CreateTaskProject(ctx context.Context, input *CreateTaskProjectDTO) (*models.TaskProject, error) {
	if t.CreateTaskProjectFunc != nil {
		return t.CreateTaskProjectFunc(ctx, input)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.CreateTaskProject(ctx, input)
}

// CreateTaskProjectWithTasks implements DbTaskStoreInterface.
func (t *TaskDecorator) CreateTaskProjectWithTasks(ctx context.Context, input *CreateTaskProjectWithTasksDTO) (*models.TaskProject, error) {
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

// FindTaskByIDForUpdate implements DbTaskStoreInterface.
func (t *TaskDecorator) FindTaskByIDForUpdate(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindTaskByIDForUpdate(ctx, id)
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

func (t *TaskDecorator) FindWorkflowByID(ctx context.Context, id uuid.UUID) (*models.Workflow, error) {
	if t.FindWorkflowByIDFunc != nil {
		return t.FindWorkflowByIDFunc(ctx, id)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindWorkflowByID(ctx, id)
}

func (t *TaskDecorator) FindWorkflowStatusByID(ctx context.Context, id uuid.UUID) (*models.WorkflowStatus, error) {
	if t.FindWorkflowStatusByIDFunc != nil {
		return t.FindWorkflowStatusByIDFunc(ctx, id)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindWorkflowStatusByID(ctx, id)
}

func (t *TaskDecorator) FindWorkflowStatusesByIDs(ctx context.Context, ids ...uuid.UUID) ([]*models.WorkflowStatus, error) {
	if t.FindWorkflowStatusesByIDsFunc != nil {
		return t.FindWorkflowStatusesByIDsFunc(ctx, ids...)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindWorkflowStatusesByIDs(ctx, ids...)
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

func (t *TaskDecorator) ListWorkflows(ctx context.Context, filter *WorkflowFilter) ([]*models.Workflow, error) {
	if t.ListWorkflowsFunc != nil {
		return t.ListWorkflowsFunc(ctx, filter)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.ListWorkflows(ctx, filter)
}

func (t *TaskDecorator) ArchiveWorkflow(ctx context.Context, workflowID uuid.UUID) (*models.Workflow, error) {
	if t.ArchiveWorkflowFunc != nil {
		return t.ArchiveWorkflowFunc(ctx, workflowID)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.ArchiveWorkflow(ctx, workflowID)
}

func (t *TaskDecorator) CreateWorkflow(ctx context.Context, input *CreateWorkflowDTO) (*models.Workflow, error) {
	if t.CreateWorkflowFunc != nil {
		return t.CreateWorkflowFunc(ctx, input)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.CreateWorkflow(ctx, input)
}

func (t *TaskDecorator) UpdateWorkflow(ctx context.Context, workflowID uuid.UUID, input *UpdateWorkflowDTO) (*models.Workflow, error) {
	if t.UpdateWorkflowFunc != nil {
		return t.UpdateWorkflowFunc(ctx, workflowID, input)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.UpdateWorkflow(ctx, workflowID, input)
}

func (t *TaskDecorator) DeleteWorkflow(ctx context.Context, workflowID uuid.UUID) error {
	if t.DeleteWorkflowFunc != nil {
		return t.DeleteWorkflowFunc(ctx, workflowID)
	}
	if t.Delegate == nil {
		return ErrDelegateNil
	}
	return t.Delegate.DeleteWorkflow(ctx, workflowID)
}

func (t *TaskDecorator) SetDefaultWorkflow(ctx context.Context, workflowID uuid.UUID) (*models.Workflow, error) {
	if t.SetDefaultWorkflowFunc != nil {
		return t.SetDefaultWorkflowFunc(ctx, workflowID)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.SetDefaultWorkflow(ctx, workflowID)
}

// ListTaskProjects implements DbTaskStoreInterface.
func (t *TaskDecorator) ListTaskProjects(ctx context.Context, input *TaskProjectsFilter) ([]*models.TaskProject, error) {
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

func (t *TaskDecorator) LoadWorkflowStatuses(ctx context.Context, workflowIds ...uuid.UUID) ([][]*models.WorkflowStatus, error) {
	if t.LoadWorkflowStatusesFunc != nil {
		return t.LoadWorkflowStatusesFunc(ctx, workflowIds...)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.LoadWorkflowStatuses(ctx, workflowIds...)
}

func (t *TaskDecorator) CreateWorkflowStatus(ctx context.Context, workflowID uuid.UUID, input *CreateWorkflowStatusDTO) (*models.WorkflowStatus, error) {
	if t.CreateWorkflowStatusFunc != nil {
		return t.CreateWorkflowStatusFunc(ctx, workflowID, input)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.CreateWorkflowStatus(ctx, workflowID, input)
}

func (t *TaskDecorator) UpdateWorkflowStatus(ctx context.Context, workflowStatusID uuid.UUID, input *UpdateWorkflowStatusDTO) (*models.WorkflowStatus, error) {
	if t.UpdateWorkflowStatusFunc != nil {
		return t.UpdateWorkflowStatusFunc(ctx, workflowStatusID, input)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.UpdateWorkflowStatus(ctx, workflowStatusID, input)
}

func (t *TaskDecorator) ReorderWorkflowStatuses(ctx context.Context, workflowID uuid.UUID, statusIDs []uuid.UUID) ([]*models.WorkflowStatus, error) {
	if t.ReorderWorkflowStatusesFunc != nil {
		return t.ReorderWorkflowStatusesFunc(ctx, workflowID, statusIDs)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.ReorderWorkflowStatuses(ctx, workflowID, statusIDs)
}

func (t *TaskDecorator) DeleteWorkflowStatus(ctx context.Context, workflowStatusID uuid.UUID) error {
	if t.DeleteWorkflowStatusFunc != nil {
		return t.DeleteWorkflowStatusFunc(ctx, workflowStatusID)
	}
	if t.Delegate == nil {
		return ErrDelegateNil
	}
	return t.Delegate.DeleteWorkflowStatus(ctx, workflowStatusID)
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

// taskWhere implements DbTaskStoreInterface.
func (t *TaskDecorator) taskWhere(task *TaskFilter) *map[string]any {
	if t.TaskWhereFunc != nil {
		return t.TaskWhereFunc(task)
	}
	if t.Delegate == nil {
		return nil
	}
	return t.Delegate.taskWhere(task)
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
func (t *TaskDecorator) UpdateTaskProject(ctx context.Context, taskProjectID uuid.UUID, input *UpdateTaskProjectBaseDTO) error {
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
func (t *TaskDecorator) UpdateTaskRankStatus(ctx context.Context, taskID uuid.UUID, position int64, status models.TaskStatus, workflowStatusID *uuid.UUID) error {
	if t.UpdateTaskRankStatusFunc != nil {
		return t.UpdateTaskRankStatusFunc(ctx, taskID, position, status, workflowStatusID)
	}
	if t.Delegate == nil {
		return ErrDelegateNil
	}
	return t.Delegate.UpdateTaskRankStatus(ctx, taskID, position, status, workflowStatusID)
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
func (t *TaskDecorator) CountTaskProjects(ctx context.Context, filter *TaskProjectsFilter) (int64, error) {
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
func (t *TaskDecorator) CreateTaskFromInput(ctx context.Context, teamID uuid.UUID, projectID uuid.UUID, memberID uuid.UUID, input *CreateTaskProjectTaskDTO) (*models.Task, error) {
	if t.CreateTaskFromInputFunc != nil {
		return t.CreateTaskFromInputFunc(ctx, teamID, projectID, memberID, input)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.CreateTaskFromInput(ctx, teamID, projectID, memberID, input)
}

// FindTasksDueToday implements DbTaskStoreInterface.
func (t *TaskDecorator) FindTasksDueToday(ctx context.Context) ([]*models.Task, error) {
	if t.FindTasksDueTodayFunc != nil {
		return t.FindTasksDueTodayFunc(ctx)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindTasksDueToday(ctx)
}

// FindTasksOverdue implements DbTaskStoreInterface.
func (t *TaskDecorator) FindTasksOverdue(ctx context.Context) ([]*models.Task, error) {
	if t.FindTasksOverdueFunc != nil {
		return t.FindTasksOverdueFunc(ctx)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindTasksOverdue(ctx)
}

var _ DbTaskStoreInterface = (*TaskDecorator)(nil)
