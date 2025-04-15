package shared

import (
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db/models"
)

type Task struct {
	ID          uuid.UUID         `db:"id,pk" json:"id"`
	UserID      uuid.UUID         `db:"user_id" json:"user_id"`
	ProjectID   uuid.UUID         `db:"project_id" json:"project_id"`
	Name        string            `db:"name" json:"name"`
	Description *string           `db:"description" json:"description"`
	Status      models.TaskStatus `db:"status" json:"status"`
	Order       float64           `db:"order" json:"order"`
	ParentID    *uuid.UUID        `db:"parent_id" json:"parent_id"`
	CreatedAt   time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `db:"updated_at" json:"updated_at"`
}

type TaskProject struct {
	ID          uuid.UUID                `db:"id,pk" json:"id"`
	UserID      uuid.UUID                `db:"user_id" json:"user_id"`
	Name        string                   `db:"name" json:"name"`
	Description *string                  `db:"description" json:"description"`
	Status      models.TaskProjectStatus `db:"status" json:"status"`
	Order       float64                  `db:"order" json:"order"`
	CreatedAt   time.Time                `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time                `db:"updated_at" json:"updated_at"`
}

type TaskProjectWithTasks struct {
	*TaskProject
	Tasks []*TaskWithSubtask `json:"tasks"`
}

type TaskWithSubtask struct {
	*Task
	Children []*Task `json:"children,omitempty" required:"false"`
}

func ModelToTask(task *models.Task) *Task {
	return &Task{
		ID:          task.ID,
		UserID:      task.UserID,
		ProjectID:   task.ProjectID,
		Name:        task.Name,
		Description: task.Description.Ptr(),
		Status:      task.Status,
		Order:       task.Order,
		ParentID:    task.ParentID.Ptr(),
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

func ModelToProject(task *models.TaskProject) *TaskProject {
	return &TaskProject{
		ID:          task.ID,
		UserID:      task.UserID,
		Name:        task.Name,
		Description: task.Description.Ptr(),
		Status:      task.Status,
		Order:       task.Order,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

type CreateTaskBaseDTO struct {
	Name        string            `json:"name" required:"true"`
	Description *string           `json:"description,omitempty" required:"false"`
	Status      models.TaskStatus `json:"status" required:"false" enum:"todo,in_progress,done" default:"todo"`
	Order       float64           `json:"order,omitempty" required:"false"`
}

type UpdateTaskBaseDTO struct {
	Name        string            `json:"name" required:"true"`
	Description *string           `json:"description,omitempty" required:"false"`
	Status      models.TaskStatus `json:"status" enum:"todo,in_progress,done"`
	Order       float64           `json:"order"`
	Position    *int64            `json:"position,omitempty" required:"false"`
	ParentID    *uuid.UUID        `json:"parent_id,omitempty" required:"false"`
}

type UpdateTaskDTO struct {
	UpdateTaskBaseDTO
	TaskID uuid.UUID `path:"task-id" json:"task_id" required:"true"`
}

type CreateTaskInput struct {
	TaskProjectID uuid.UUID `path:"task-project-id"`
	CreateTaskWithChildrenDTO
}

type CreateTaskWithChildrenDTO struct {
	CreateTaskBaseDTO
	Children []CreateTaskBaseDTO `json:"children,omitempty" required:"false"`
}

type CreateTaskProjectDTO struct {
	Name        string                   `json:"name" required:"true"`
	Description *string                  `json:"description,omitempty" required:"false"`
	Status      models.TaskProjectStatus `json:"status" required:"false" enum:"todo,in_progress,done" default:"todo"`
	Order       float64                  `json:"order,omitempty" required:"false"`
}

type CreateTaskProjectWithTasksDTO struct {
	CreateTaskProjectDTO
	Tasks []CreateTaskBaseDTO `json:"tasks,omitempty" required:"false"`
}

type TaskProjectsListFilter struct {
	Q      string                     `query:"q,omitempty" required:"false"`
	UserID string                     `query:"user_id,omitempty" required:"false" format:"uuid"`
	Status []models.TaskProjectStatus `query:"status,omitempty,explode" required:"false" minimum:"1" maximum:"100" enum:"todo,in_progress,done"`
	Ids    []string                   `query:"ids,omitempty,explode" required:"false" minimum:"1" maximum:"100" format:"uuid"`
}

type TaskProjectsListParams struct {
	PaginatedInput
	TaskProjectsListFilter
	SortParams
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"tasks,subtasks"`
}

type TaskListFilter struct {
	Q            string              `query:"q,omitempty" required:"false"`
	Status       []models.TaskStatus `query:"status,omitempty,explode" required:"false" enum:"todo,in_progress,done"`
	ProjectID    string              `query:"project_id,omitempty" required:"false" format:"uuid"`
	UserID       string              `query:"user_id,omitempty" required:"false" format:"uuid"`
	Ids          []string            `query:"ids,omitempty,explode" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	ParentID     string              `query:"parent_id,omitempty" required:"false" format:"uuid"`
	ParentStatus ParentStatus        `query:"parent_status,omitempty" required:"false" enum:"parent,child"`
}

type ParentStatus string

const (
	ParentStatusParent ParentStatus = "parent"
	ParentStatusChild  ParentStatus = "child"
)

type TaskListParams struct {
	PaginatedInput
	TaskListFilter
	SortParams
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"subtasks"`
}
