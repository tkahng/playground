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
