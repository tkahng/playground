package shared

import (
	"time"

	"github.com/google/uuid"
	crudModels "github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

type Task struct {
	_                 struct{}     `db:"tasks" json:"-"`
	ID                uuid.UUID    `db:"id" json:"id"`
	CreatedByMemberID *uuid.UUID   `db:"created_by_member_id" json:"created_by_member_id" nullable:"true"`
	TeamID            uuid.UUID    `db:"team_id" json:"team_id"`
	ProjectID         uuid.UUID    `db:"project_id" json:"project_id"`
	Name              string       `db:"name" json:"name"`
	Description       *string      `db:"description" json:"description"`
	Status            TaskStatus   `db:"status" json:"status" enum:"todo,in_progress,done"`
	StartAt           *time.Time   `db:"start_at" json:"start_at" nullable:"true"`
	EndAt             *time.Time   `db:"end_at" json:"end_at" nullable:"true"`
	AssigneeID        *uuid.UUID   `db:"assignee_id" json:"assignee_id" nullable:"true"`
	ReporterID        *uuid.UUID   `db:"reporter_id" json:"reporter_id" nullable:"true"`
	Rank              float64      `db:"rank" json:"rank"`
	ParentID          *uuid.UUID   `db:"parent_id" json:"parent_id" nullable:"true"`
	CreatedAt         time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time    `db:"updated_at" json:"updated_at"`
	Children          []*Task      `db:"children" src:"id" dest:"parent_id" table:"tasks" json:"children,omitempty"`
	CreatedByMember   *TeamMember  `db:"created_by_member" src:"created_by_member_id" dest:"id" table:"team_members" json:"created_by_member,omitempty"`
	Team              *Team        `db:"team" src:"team_id" dest:"id" table:"teams" json:"team,omitempty"`
	Project           *TaskProject `db:"project" src:"project_id" dest:"id" table:"task_projects" json:"project,omitempty"`
}

func FromModelTask(task *crudModels.Task) *Task {
	if task == nil {
		return nil
	}
	return &Task{
		ID:                task.ID,
		CreatedByMemberID: task.CreatedByMemberID,
		TeamID:            task.TeamID,
		ProjectID:         task.ProjectID,
		Name:              task.Name,
		Description:       task.Description,
		Status:            TaskStatus(task.Status),
		StartAt:           task.StartAt,
		EndAt:             task.EndAt,
		AssigneeID:        task.AssigneeID,
		ReporterID:        task.ReporterID,
		Rank:              task.Rank,
		ParentID:          task.ParentID,
		CreatedAt:         task.CreatedAt,
		UpdatedAt:         task.UpdatedAt,
		Children:          mapper.Map(task.Children, FromModelTask),
		CreatedByMember:   FromTeamMemberModel(task.CreatedByMember),
		Team:              FromTeamModel(task.Team),
		Project:           FromModelProject(task.Project),
	}
}

type CreateTaskProjectTaskDTO struct {
	Name        string     `json:"name" required:"true"`
	Description *string    `json:"description,omitempty" required:"false"`
	Status      TaskStatus `json:"status" required:"false" enum:"todo,in_progress,done" default:"todo"`
	Rank        float64    `json:"rank,omitempty" required:"false"`
}

type UpdateTaskDto struct {
	Name        string     `db:"name" json:"name"`
	Description *string    `db:"description" json:"description"`
	Status      TaskStatus `db:"status" json:"status" enum:"todo,in_progress,done"`
	StartAt     *time.Time `db:"start_at" json:"start_at" nullable:"true"`
	EndAt       *time.Time `db:"end_at" json:"end_at" nullable:"true"`
	AssigneeID  *uuid.UUID `db:"assignee_id" json:"assignee_id" nullable:"true"`
	ReporterID  *uuid.UUID `db:"reporter_id" json:"reporter_id" nullable:"true"`
	ParentID    *uuid.UUID `db:"parent_id" json:"parent_id" nullable:"true"`
}

type UpdateTaskInput struct {
	Body   UpdateTaskDto
	TaskID string `path:"task-id" json:"task_id" required:"true" format:"uuid"`
}

type TaskPositionStatusDTO struct {
	Position int64      `json:"position" required:"true"`
	Status   TaskStatus `json:"status" required:"true" enum:"todo,in_progress,done"`
}

type TaskPositionStatusInput struct {
	TaskID string `path:"task-id" json:"task_id" required:"true" format:"uuid"`
	Body   TaskPositionStatusDTO
}

type CreateTaskWithChildrenDTO struct {
	CreateTaskProjectTaskDTO
	Children []CreateTaskProjectTaskDTO `json:"children,omitempty" required:"false"`
}
