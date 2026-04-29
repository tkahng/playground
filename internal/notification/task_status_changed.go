package notification

import "github.com/google/uuid"

type TaskStatusChangedNotificationData struct {
	TaskID            uuid.UUID `json:"task_id" required:"true"`
	OldStatus         string    `json:"old_status" required:"true"`
	NewStatus         string    `json:"new_status" required:"true"`
	ChangedByMemberID uuid.UUID `json:"changed_by_member_id" required:"true"`
}

func (n TaskStatusChangedNotificationData) Kind() string {
	return "task_status_changed"
}
