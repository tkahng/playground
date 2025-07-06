package workers

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/stores"
)

type NewMemberNotificationJobArgs struct {
	TeamID uuid.UUID
}

func (j NewMemberNotificationJobArgs) Kind() string {
	return "new_member_notification_job"
}

type NewMemberNotificationService interface {
	FindTeamMember(ctx context.Context, member *stores.TeamMemberFilter) (*models.TeamMember, error)
}
