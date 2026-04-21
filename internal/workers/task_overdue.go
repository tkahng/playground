package workers

import (
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/jobs"
)

type TaskOverdueJobArgs struct {
	TaskID  uuid.UUID `json:"task_id" required:"true"`
	DueDate time.Time `json:"due_date" required:"true"`
}

func (j TaskOverdueJobArgs) Kind() string {
	return "task_overdue"
}

type TaskOverdueJobWorker jobs.Worker[TaskOverdueJobArgs]
