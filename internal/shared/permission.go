package shared

import (
	"time"

	"github.com/google/uuid"
	crudModels "github.com/tkahng/authgo/internal/models"
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

func FromModelPermission(permission *crudModels.Permission) *Permission {
	return &Permission{
		ID:          permission.ID,
		Name:        permission.Name,
		Description: permission.Description,
		CreatedAt:   permission.CreatedAt,
		UpdatedAt:   permission.UpdatedAt,
	}
}
