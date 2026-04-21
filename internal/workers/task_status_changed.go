package workers

import (
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/jobs"
)

type TaskStatusChangedJobArgs struct {
	TaskID            uuid.UUID `json:"task_id" required:"true"`
	OldStatus         string    `json:"old_status" required:"true"`
	NewStatus         string    `json:"new_status" required:"true"`
	ChangedByMemberID uuid.UUID `json:"changed_by_member_id" required:"true"`
}

func (j TaskStatusChangedJobArgs) Kind() string {
	return "task_status_changed"
}

type TaskStatusChangedWorker jobs.Worker[TaskStatusChangedJobArgs]
