package models

import (
	"time"

	"github.com/google/uuid"
)

type NotificationPreference struct {
	_            struct{}  `db:"team_notification_preferences" schema:"messaging" json:"-"`
	ID           uuid.UUID `db:"id,pk" json:"id"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
	TeamMemberID uuid.UUID `db:"team_member_id" json:"team_member_id"`
	Type         string    `db:"type" json:"type"`
	Enabled      bool      `db:"enabled" json:"enabled"`
}
