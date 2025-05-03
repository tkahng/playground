package shared

import (
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crud/models"
)

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

type TaskStatus string

// Enum values for TaskProjectStatus
const (
	TaskProjectStatusTodo       TaskProjectStatus = "todo"
	TaskProjectStatusInProgress TaskProjectStatus = "in_progress"
	TaskProjectStatusDone       TaskProjectStatus = "done"
)

type TaskProjectStatus string

type TaskProject struct {
	ID          uuid.UUID         `db:"id,pk" json:"id"`
	UserID      uuid.UUID         `db:"user_id" json:"user_id"`
	Name        string            `db:"name" json:"name"`
	Description *string           `db:"description" json:"description"`
	Status      TaskProjectStatus `db:"status" json:"status" enum:"todo,in_progress,done"`
	Order       float64           `db:"order" json:"order"`
	CreatedAt   time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `db:"updated_at" json:"updated_at"`
}
type TaskProjectWithTasks struct {
	*TaskProject
	Tasks []*TaskWithSubtask `json:"tasks,omitempty" required:"false"`
}

func CrudToProject(task *models.TaskProject) *TaskProject {
	if task == nil {
		return nil
	}
	return &TaskProject{
		ID:          task.ID,
		UserID:      task.UserID,
		Name:        task.Name,
		Description: task.Description,
		Status:      TaskProjectStatus(task.Status),
		Order:       task.Order,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

func ModelToProject(task *models.TaskProject) *TaskProject {
	if task == nil {
		return nil
	}
	return &TaskProject{
		ID:          task.ID,
		UserID:      task.UserID,
		Name:        task.Name,
		Description: task.Description,
		Status:      TaskProjectStatus(task.Status),
		Order:       task.Order,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

type CreateTaskProjectDTO struct {
	Name        string            `json:"name" required:"true"`
	Description *string           `json:"description,omitempty" required:"false"`
	Status      TaskProjectStatus `json:"status" required:"false" enum:"todo,in_progress,done" default:"todo"`
	Order       float64           `json:"order,omitempty" required:"false"`
}

type CreateTaskProjectWithTasksDTO struct {
	CreateTaskProjectDTO
	Tasks []CreateTaskBaseDTO `json:"tasks,omitempty" required:"false"`
}

type UpdateTaskProjectBaseDTO struct {
	Name        string            `json:"name" required:"true"`
	Description *string           `json:"description,omitempty" required:"false"`
	Status      TaskProjectStatus `json:"status" enum:"todo,in_progress,done"`
	Order       float64           `json:"order"`
	Position    *int64            `json:"position,omitempty" required:"false"`
}

type UpdateTaskProjectDTO struct {
	Body          UpdateTaskProjectBaseDTO
	TaskProjectID string `path:"task-project-id" json:"task_project_id" required:"true" format:"uuid"`
}

type TaskProjectsListFilter struct {
	Q      string              `query:"q,omitempty" required:"false"`
	UserID string              `query:"user_id,omitempty" required:"false" format:"uuid"`
	Status []TaskProjectStatus `query:"status,omitempty" required:"false" minimum:"1" maximum:"100" enum:"todo,in_progress,done"`
	Ids    []string            `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
}

type TaskProjectsListParams struct {
	PaginatedInput
	TaskProjectsListFilter
	SortParams
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"tasks,subtasks"`
}
