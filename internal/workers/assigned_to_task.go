package workers

import "github.com/google/uuid"

type AssignedToTaskJobArgs struct {
	AssignedByMemberID uuid.UUID `json:"assigned_by_member_id" required:"true"`
	AssigneeMemberID   uuid.UUID `json:"assignee_member_id" required:"true"`
	TaskID             uuid.UUID `json:"task_id" required:"true"`
}

func (a AssignedToTaskJobArgs) Kind() string {
	return "assigned_to_task"
}
