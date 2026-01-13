package notification

import (
	"github.com/google/uuid"
)

type NewTeamMemberNotificationData struct {
	TeamMemberID uuid.UUID `json:"team_member_id" required:"true"`
	TeamID       uuid.UUID `json:"team_id" required:"true"`
	Email        string    `json:"email" required:"true"`
}

func (n NewTeamMemberNotificationData) Kind() string {
	return "new_team_member"
}
