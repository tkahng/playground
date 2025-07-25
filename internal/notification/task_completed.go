package notification

import (
	"time"

	"github.com/google/uuid"
)

type TaskCompletedNotificationData struct {
	TaskID              uuid.UUID `json:"task_id" required:"true"`
	CompletedByMemberID uuid.UUID `json:"completed_by_member_id" required:"true"`
	CompletedAt         time.Time `json:"completed_at" required:"true"`
}

func (n TaskCompletedNotificationData) Kind() string {
	return "task_completed"
}
