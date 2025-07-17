package models

import (
	"time"

	"github.com/google/uuid"
)

type UserReaction struct {
	_         struct{}   `db:"user_reactions" json:"-"`
	ID        uuid.UUID  `db:"id" json:"id"`
	UserID    *uuid.UUID `db:"user_id" json:"user_id"`
	Type      string     `db:"type" json:"type"`
	Reaction  *string    `db:"otp" json:"otp"`
	IpAddress *string    `db:"ip_address" json:"ip_address"`
	Country   *string    `db:"country" json:"country"`
	City      *string    `db:"city" json:"city"`
	Metadata  []byte     `db:"metadata" json:"metadata"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	User      *User      `db:"users" src:"user_id" dest:"id" table:"users" json:"user,omitempty"`
}
