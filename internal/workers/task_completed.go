package workers

import (
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/jobs"
)

type TaskCompletedJobArgs struct {
	TaskID              uuid.UUID `json:"task_id" required:"true"`
	CompletedByMemberID uuid.UUID `json:"completed_by_member_id" required:"true"`
	CompletedAt         time.Time `json:"completed_at" required:"true"`
}

func (j TaskCompletedJobArgs) Kind() string {
	return "task_completed"
}

type TaskCompletedJobWorker jobs.Worker[TaskCompletedJobArgs]
