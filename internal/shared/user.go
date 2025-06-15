package shared

import (
	"time"

	"github.com/google/uuid"
	crudModels "github.com/tkahng/authgo/internal/models"
)

const (
	SuperUserEmail string = "admin@k2dv.io"
)

type User struct {
	ID              uuid.UUID  `db:"id,pk" json:"id"`
	Email           string     `db:"email" json:"email"`
	EmailVerifiedAt *time.Time `db:"email_verified_at" json:"email_verified_at"`
	Name            *string    `db:"name" json:"name"`
	Image           *string    `db:"image" json:"image"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

func FromUserModel(user *crudModels.User) *User {
	if user == nil {
		return nil
	}
	return &User{
		ID:              user.ID,
		Email:           user.Email,
		EmailVerifiedAt: user.EmailVerifiedAt,
		Name:            user.Name,
		Image:           user.Image,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}

type UpdateMeInput struct {
	Name  *string `json:"name"`
	Image *string `json:"image"`
}
