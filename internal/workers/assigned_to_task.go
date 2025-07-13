package workers

import (
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/jobs"
)

type AssignedToTasJobArgs struct {
	AssignedByMemeberID uuid.UUID `json:"assigned_by_member_id" required:"true"`
	AssigneeMemberID    uuid.UUID `json:"assignee_member_id" required:"true"`
	TaskID              uuid.UUID `json:"task_id" required:"true"`
}

func (a AssignedToTasJobArgs) Kind() string {
	return "assigned_to_task"
}

type AssignedToTaskWorker jobs.Worker[AssignedToTasJobArgs]
