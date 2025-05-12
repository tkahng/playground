package models

import (
	"time"

	"github.com/google/uuid"
)

type UserPermission struct {
	_            struct{}  `db:"user_permissions" json:"-"`
	UserID       uuid.UUID `db:"user_id" json:"user_id"`
	PermissionID uuid.UUID `db:"permission_id" json:"permission_id"`
}
type RolePermission struct {
	_            struct{}  `db:"role_permissions" json:"-"`
	RoleID       uuid.UUID `db:"role_id" json:"role_id"`
	PermissionID uuid.UUID `db:"permission_id" json:"permission_id"`
}

type Permission struct {
	_           struct{}  `db:"permissions" json:"-"`
	ID          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type ProductPermission struct {
	_            struct{}  `db:"product_permissions" json:"-"`
	ProductID    string    `db:"product_id" json:"product_id"`
	PermissionID uuid.UUID `db:"permission_id" json:"permission_id"`
}
