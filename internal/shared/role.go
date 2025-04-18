package shared

import (
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

type Role struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

func ToRole(role *models.Role) *Role {
	return &Role{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description.Ptr(),
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

type RoleWithPermissions struct {
	*Role
	Permissions []*Permission `json:"permissions,omitempty" required:"false"`
}

func ToRoleWithPermissions(role *models.Role) *RoleWithPermissions {
	return &RoleWithPermissions{
		Permissions: mapper.Map(role.R.Permissions, ToPermission),
		Role:        ToRole(role),
	}
}

type RoleListFilter struct {
	Q           string   `query:"q,omitempty" required:"false"`
	Ids         []string `query:"ids,omitempty,explode" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Names       []string `query:"names,omitempty" required:"false" minimum:"1" maximum:"100"`
	UserId      string   `query:"user_id,omitempty" required:"false" format:"uuid"`
	UserReverse bool     `query:"user_reverse,omitempty" required:"false" doc:"When user_id is provided, if this is true, it will return the roles that the user does not have"`
}
type RolesListParams struct {
	PaginatedInput
	RoleListFilter
	SortParams
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"users,permissions"`
}
