package apis

import (
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/mapper"
)

type ApiUser struct {
	ID              uuid.UUID            `db:"id,pk" json:"id"`
	Email           string               `db:"email" json:"email"`
	EmailVerifiedAt *time.Time           `db:"email_verified_at" json:"email_verified_at"`
	Name            *string              `db:"name" json:"name"`
	Image           *string              `db:"image" json:"image"`
	CreatedAt       time.Time            `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time            `db:"updated_at" json:"updated_at"`
	Accounts        []*UserAccountOutput `json:"accounts,omitempty" required:"false"`
	Roles           []*Role              `json:"roles,omitempty" required:"false"`
	Permissions     []*Permission        `json:"permissions,omitempty" required:"false"`
}

func FromUserModel(user *models.User) *ApiUser {
	if user == nil {
		return nil
	}
	return &ApiUser{
		ID:              user.ID,
		Email:           user.Email,
		EmailVerifiedAt: user.EmailVerifiedAt,
		Name:            user.Name,
		Image:           user.Image,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
		Accounts:        mapper.Map(user.Accounts, FromModelUserAccountOutput),
		Roles:           mapper.Map(user.Roles, FromModelRole),
		Permissions:     mapper.Map(user.Permissions, FromModelPermission),
	}
}
