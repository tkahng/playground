package shared

import (
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db/models"
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

func ToUser(user *models.User) *User {
	return &User{
		ID:              user.ID,
		Email:           user.Email,
		EmailVerifiedAt: user.EmailVerifiedAt.Ptr(),
		Name:            user.Name.Ptr(),
		Image:           user.Image.Ptr(),
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}

type UserListFilter struct {
	Providers []models.Providers `query:"providers,omitempty" required:"false" uniqueItems:"true" minimum:"1" maximum:"100" enum:"google,apple,facebook,github,credentials"`
	Q         string             `query:"q,omitempty" required:"false"`
	Ids       []string           `query:"ids,omitempty,explode" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Emails    []string           `query:"emails,omitempty" required:"false" minimum:"1" maximum:"100" format:"email"`
	RoleIds   []string           `query:"role_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	// PermissionIds []string           `query:"permission_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
}
type UserListParams struct {
	PaginatedInput
	UserListFilter
	SortParams
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"roles,permissions,accounts,subscriptions"`
}
