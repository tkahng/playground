package notification

import (
	"time"

	"github.com/google/uuid"
)

type TaskOverdueNotificationData struct {
	TaskID  uuid.UUID `json:"task_id" required:"true"`
	DueDate time.Time `json:"due_date" required:"true"`
}

func (n TaskOverdueNotificationData) Kind() string {
	return "task_overdue"
}
