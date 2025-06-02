package models

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	_            struct{}       `db:"notifications" json:"-"`
	ID           uuid.UUID      `db:"id,pk" json:"id"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updated_at"`
	ReadAt       *time.Time     `db:"read_at" json:"read_at,omitempty"`
	Channel      string         `db:"channel" json:"channel"`
	Payload      []byte         `db:"payload" json:"payload"`
	UserID       *uuid.UUID     `db:"user_id" json:"user_id,omitempty"`
	TeamMemberID *uuid.UUID     `db:"team_member_id" json:"team_member_id,omitempty"`
	TeamID       *uuid.UUID     `db:"team_id" json:"team_id,omitempty"`
	Metadata     map[string]any `db:"metadata" json:"metadata"`
	Type         string         `db:"type" json:"type"`
	User         *User          `db:"user" src:"user_id" dest:"id" table:"users" json:"user,omitempty"`
	TeamMember   *TeamMember    `db:"team_member" src:"team_member_id" dest:"id" table:"team_members" json:"team_member,omitempty"`
	Team         *Team          `db:"team" src:"team_id" dest:"id" table:"teams" json:"team,omitempty"`
}
