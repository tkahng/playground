package shared

import (
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/tools/mapper"
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
	_                 struct{}          `db:"task_projects" json:"-"`
	ID                uuid.UUID         `db:"id" json:"id"`
	CreatedByMemberID *uuid.UUID        `db:"created_by_member_id" json:"created_by_member_id" nullable:"true"`
	TeamID            uuid.UUID         `db:"team_id" json:"team_id"`
	Name              string            `db:"name" json:"name"`
	Description       *string           `db:"description" json:"description"`
	Status            TaskProjectStatus `db:"status" json:"status" enum:"todo,in_progress,done"`
	StartAt           *time.Time        `db:"start_at" json:"start_at" nullable:"true"`
	EndAt             *time.Time        `db:"end_at" json:"end_at" nullable:"true"`
	AssigneeID        *uuid.UUID        `db:"assignee_id" json:"assignee_id" nullable:"true"`
	ReporterID        *uuid.UUID        `db:"reporter_id" json:"reporter_id" nullable:"true"`
	Rank              float64           `db:"rank" json:"rank"`
	CreatedAt         time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time         `db:"updated_at" json:"updated_at"`
	CreatedByMember   *TeamMember       `db:"created_by_member" src:"created_by_member_id" dest:"id" table:"team_members" json:"created_by_member,omitempty"`
	Team              *Team             `db:"team" src:"team_id" dest:"id" table:"teams" json:"team,omitempty"`
	Tasks             []*Task           `db:"tasks" src:"id" dest:"project_id" table:"tasks" json:"tasks,omitempty"`
}

func FromModelProject(task *models.TaskProject) *TaskProject {
	if task == nil {
		return nil
	}
	return &TaskProject{
		ID:                task.ID,
		CreatedByMemberID: task.CreatedByMemberID,
		TeamID:            task.TeamID,
		Name:              task.Name,
		Description:       task.Description,
		Status:            TaskProjectStatus(task.Status),
		StartAt:           task.StartAt,
		EndAt:             task.EndAt,
		AssigneeID:        task.AssigneeID,
		ReporterID:        task.ReporterID,
		Rank:              task.Rank,
		CreatedAt:         task.CreatedAt,
		UpdatedAt:         task.UpdatedAt,
		CreatedByMember:   FromTeamMemberModel(task.CreatedByMember),
		Team:              FromTeamModel(task.Team),
		Tasks:             mapper.Map(task.Tasks, FromModelTask),
	}
}
