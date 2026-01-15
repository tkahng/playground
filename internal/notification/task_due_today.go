package notification

import (
	"time"

	"github.com/google/uuid"
)

type TaskDueTodayNotificationData struct {
	TaskID  uuid.UUID `json:"task_id" required:"true"`
	DueDate time.Time `json:"due_date" required:"true"`
}

func (n TaskDueTodayNotificationData) Kind() string {
	return "task_due_today"
}
