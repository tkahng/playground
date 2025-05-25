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
	_               struct{}          `db:"task_projects" json:"-"`
	ID              uuid.UUID         `db:"id" json:"id"`
	CreatedBy       uuid.UUID         `db:"created_by" json:"created_by"`
	TeamID          uuid.UUID         `db:"team_id" json:"team_id"`
	Name            string            `db:"name" json:"name"`
	Description     *string           `db:"description" json:"description"`
	Status          TaskProjectStatus `db:"status" json:"status" enum:"todo,in_progress,done"`
	StartAt         *time.Time        `db:"start_at" json:"start_at,omitempty" required:"false"`
	EndAt           *time.Time        `db:"end_at" json:"end_at,omitempty" required:"false"`
	AssigneeID      *uuid.UUID        `db:"assignee_id" json:"assignee_id,omitempty"`
	AssignerID      *uuid.UUID        `db:"assigner_id" json:"assigner_id,omitempty"`
	Rank            float64           `db:"rank" json:"rank"`
	CreatedAt       time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time         `db:"updated_at" json:"updated_at"`
	CreatedByMember *TeamMember       `db:"created_by_member" src:"created_by" dest:"id" table:"team_members" json:"created_by_member,omitempty"`
	Team            *Team             `db:"team" src:"team_id" dest:"id" table:"teams" json:"team,omitempty"`
	Tasks           []*Task           `db:"tasks" src:"id" dest:"project_id" table:"tasks" json:"tasks,omitempty"`
}

func FromModelProject(task *models.TaskProject) *TaskProject {
	if task == nil {
		return nil
	}
	return &TaskProject{
		ID:              task.ID,
		CreatedBy:       task.CreatedBy,
		TeamID:          task.TeamID,
		Name:            task.Name,
		Description:     task.Description,
		Status:          TaskProjectStatus(task.Status),
		StartAt:         task.StartAt,
		EndAt:           task.EndAt,
		AssigneeID:      task.AssigneeID,
		AssignerID:      task.AssignerID,
		Rank:            task.Rank,
		CreatedAt:       task.CreatedAt,
		UpdatedAt:       task.UpdatedAt,
		CreatedByMember: FromTeamMemberModel(task.CreatedByMember),
		Team:            FromTeamModel(task.Team),
		Tasks:           mapper.Map(task.Tasks, FromModelTask),
	}
}

type CreateTaskProjectDTO struct {
	TeamID      uuid.UUID         `json:"team_id" required:"true" format:"uuid"`
	MemberID    uuid.UUID         `json:"member_id" required:"true" format:"uuid"`
	Name        string            `json:"name" required:"true"`
	Description *string           `json:"description,omitempty" required:"false"`
	Status      TaskProjectStatus `json:"status" required:"false" enum:"todo,in_progress,done" default:"todo"`
	Rank        float64           `json:"rank,omitempty" required:"false"`
}

type CreateTaskProjectWithTasksDTO struct {
	CreateTaskProjectDTO
	Tasks []CreateTaskBaseDTO `json:"tasks,omitempty" required:"false"`
}

type UpdateTaskProjectBaseDTO struct {
	Name        string            `json:"name" required:"true"`
	Description *string           `json:"description,omitempty" required:"false"`
	Status      TaskProjectStatus `json:"status" enum:"todo,in_progress,done"`
	Rank        float64           `json:"rank"`
	Position    *int64            `json:"position,omitempty" required:"false"`
}

type UpdateTaskProjectDTO struct {
	Body          UpdateTaskProjectBaseDTO
	TaskProjectID string `path:"task-project-id" json:"task_project_id" required:"true" format:"uuid"`
}

type TaskProjectsListFilter struct {
	Q      string              `query:"q,omitempty" required:"false"`
	TeamID string              `query:"team_id,omitempty" required:"false" format:"uuid"`
	Status []TaskProjectStatus `query:"status,omitempty" required:"false" minimum:"1" maximum:"100" enum:"todo,in_progress,done"`
	Ids    []string            `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
}

type TaskProjectsListParams struct {
	PaginatedInput
	TaskProjectsListFilter
	SortParams
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"tasks,subtasks"`
}
