package models

import (
	"time"

	"github.com/google/uuid"
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
	StartAt           *time.Time   `db:"start_at" json:"start_at"  nullable:"true"`
	EndAt             *time.Time   `db:"end_at" json:"end_at" nullable:"true"`
	AssigneeID        *uuid.UUID   `db:"assignee_id" json:"assignee_id" nullable:"true"`
	ReporterID        *uuid.UUID   `db:"reporter_id" json:"reporter_id" nullable:"true"`
	Rank              float64      `db:"rank" json:"rank"`
	ParentID          *uuid.UUID   `db:"parent_id" json:"parent_id" nullable:"true"`
	CreatedAt         time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time    `db:"updated_at" json:"updated_at"`
	Children          []*Task      `db:"children" src:"id" dest:"parent_id" table:"tasks" json:"children,omitempty"`
	CreatedByMember   *TeamMember  `db:"created_by_member" src:"created_by_member_id" dest:"id" table:"team_members" json:"created_by_member,omitempty"`
	Assignee          *TeamMember  `db:"assignee" src:"assignee_id" dest:"id" table:"team_members" json:"assignee,omitempty"`
	Reporter          *TeamMember  `db:"reporter" src:"reporter_id" dest:"id" table:"team_members" json:"reporter,omitempty"`
	Team              *Team        `db:"team" src:"team_id" dest:"id" table:"teams" json:"team,omitempty"`
	Project           *TaskProject `db:"project" src:"project_id" dest:"id" table:"task_projects" json:"project,omitempty"`
}

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

type TaskStats struct {
	TotalProjects     int64 `db:"total_projects" json:"total_projects"`
	CompletedProjects int64 `db:"completed_projects" json:"completed_projects"`
	TotalTasks        int64 `db:"total_tasks" json:"total_tasks"`
	CompletedTasks    int64 `db:"completed_tasks" json:"completed_tasks"`
}
