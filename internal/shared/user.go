package shared

import (
	"time"

	"github.com/google/uuid"
	crudModels "github.com/tkahng/authgo/internal/crud/crudModels"
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

type UserMutationInput struct {
	Email           string     `json:"email" required:"true" format:"email" maxLength:"100"`
	Name            *string    `json:"name,omitempty" required:"false" maxLength:"100"`
	Image           *string    `json:"image,omitempty" required:"false" format:"uri" maxLength:"200"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty" required:"false" format:"date-time"`
}

type UserCreateInput struct {
	*UserMutationInput
	Password string `json:"password" required:"true" minLength:"8" maxLength:"100"`
}

type UserWithAccounts struct {
	*User
	Accounts []*UserAccountOutput `json:"accounts"`
}

func FromCrudUser(user *crudModels.User) *User {
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

type EmailVerifiedStatus string

const (
	Verified   EmailVerifiedStatus = "verified"
	UnVerified EmailVerifiedStatus = "unverified"
)

type UserListFilter struct {
	Providers     []Providers         `query:"providers,omitempty" required:"false" uniqueItems:"true" minimum:"1" maximum:"100" enum:"google,apple,facebook,github,credentials"`
	Q             string              `query:"q,omitempty" required:"false"`
	Ids           []string            `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Emails        []string            `query:"emails,omitempty" required:"false" minimum:"1" maximum:"100" format:"email"`
	RoleIds       []string            `query:"role_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	EmailVerified EmailVerifiedStatus `query:"email_verified,omitempty" required:"false" enum:"verified,unverified"`
	// PermissionIds []string           `query:"permission_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
}
type UserListParams struct {
	PaginatedInput
	UserListFilter
	SortParams
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" uniqueItems:"true" enum:"roles,permissions,accounts,subscriptions"`
}

type UpdateMeInput struct {
	Name  *string `json:"name"`
	Image *string `json:"image"`
}
