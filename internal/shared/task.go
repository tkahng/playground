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

func ModelToTaskWithSubtask(task *models.TaskProject) *TaskProject {
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
