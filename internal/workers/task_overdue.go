package workers

import (
	"time"

	"github.com/google/uuid"
)

type TaskOverdueJobArgs struct {
	TaskID  uuid.UUID `json:"task_id" required:"true"`
	DueDate time.Time `json:"due_date" required:"true"`
}

func (j TaskOverdueJobArgs) Kind() string {
	return "task_overdue"
}
