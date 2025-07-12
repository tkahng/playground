package notification

import "github.com/google/uuid"

type AssignedToTaskNotificationData struct {
	AssignedByMemeberID uuid.UUID `json:"assigned_by_member_id" required:"true"`
	AssigneeMemberID    uuid.UUID `json:"assignee_member_id" required:"true"`
	TaskID              uuid.UUID `json:"task_id" required:"true"`
}

func (n AssignedToTaskNotificationData) Kind() string {
	return "assigned_to_task"
}
