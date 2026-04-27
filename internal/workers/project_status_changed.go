package workers

import "github.com/google/uuid"

type ProjectStatusChangedJobArgs struct {
	ProjectID         uuid.UUID `json:"project_id" required:"true"`
	OldStatus         string    `json:"old_status" required:"true"`
	NewStatus         string    `json:"new_status" required:"true"`
	ChangedByMemberID uuid.UUID `json:"changed_by_member_id" required:"true"`
}

func (j ProjectStatusChangedJobArgs) Kind() string {
	return "project_status_changed"
}

