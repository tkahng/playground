package models

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	_           struct{}      `db:"roles" json:"-"`
	ID          uuid.UUID     `db:"id" json:"id"`
	Name        string        `db:"name" json:"name"`
	Description *string       `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time     `db:"updated_at" json:"updated_at"`
	Permissions []*Permission `db:"permissions" src:"id" dest:"role_id" table:"permissions" through:"role_permissions,permission_id,id" json:"permissions,omitempty"`
	Users       []*User       `db:"users" src:"id" dest:"role_id" table:"users" through:"user_roles,user_id,id" json:"users,omitempty"`
}

type UserRole struct {
	_      struct{}  `db:"user_roles" json:"-"`
	UserID uuid.UUID `db:"user_id" json:"user_id"`
	RoleID uuid.UUID `db:"role_id" json:"role_id"`
}
type ProductRole struct {
	_         struct{}  `db:"product_roles" json:"-"`
	ProductID string    `db:"product_id" json:"product_id"`
	RoleID    uuid.UUID `db:"role_id" json:"role_id"`
}
