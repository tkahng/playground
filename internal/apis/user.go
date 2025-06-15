package apis

import (
	"time"

	"github.com/google/uuid"
)

type ApiUser struct {
	ID              uuid.UUID  `db:"id,pk" json:"id"`
	Email           string     `db:"email" json:"email"`
	EmailVerifiedAt *time.Time `db:"email_verified_at" json:"email_verified_at"`
	Name            *string    `db:"name" json:"name"`
	Image           *string    `db:"image" json:"image"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}
