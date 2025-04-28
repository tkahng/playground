package shared

import (
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db/models"
)

const (
	PermissionNameAdmin    string = "superuser"
	PermissionNameBasic    string = "basic"
	PermissionNamePro      string = "pro"
	PermissionNameAdvanced string = "advanced"
)

type Permission struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

func ToPermission(permission *models.Permission) *Permission {
	return &Permission{
		ID:          permission.ID,
		Name:        permission.Name,
		Description: permission.Description.Ptr(),
		CreatedAt:   permission.CreatedAt,
		UpdatedAt:   permission.UpdatedAt,
	}
}

type PermissionsListFilter struct {
	Q           string   `query:"q,omitempty" required:"false"`
	Ids         []string `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Names       []string `query:"names,omitempty" required:"false" minimum:"1" maximum:"100"`
	RoleId      string   `query:"role_id,omitempty" required:"false" format:"uuid"`
	RoleReverse bool     `query:"role_reverse,omitempty" required:"false" doc:"When role_id is provided, if this is true, it will return the permissions that the role does not have"`
}
type PermissionsListParams struct {
	PaginatedInput
	PermissionsListFilter
	SortParams
}
