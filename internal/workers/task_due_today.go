package workers

import (
	"time"

	"github.com/google/uuid"
)

type TaskDueTodayJobArgs struct {
	TaskID  uuid.UUID `json:"task_id" required:"true"`
	DueDate time.Time `json:"due_date" required:"true"`
}

func (j TaskDueTodayJobArgs) Kind() string {
	return "task_due_today"
}
