package models

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	_           struct{}      `db:"roles" schema:"auth" json:"-"`
	ID          uuid.UUID     `db:"id" json:"id"`
	Name        string        `db:"name" json:"name"`
	Description *string       `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time     `db:"updated_at" json:"updated_at"`
	Permissions []*Permission `db:"permissions" src:"id" dest:"id" table:"auth.permissions" through:"auth.role_permissions" through_src:"role_id" through_dest:"permission_id" json:"permissions,omitempty"`
	Users       []*User       `db:"users" src:"id" dest:"id" table:"auth.users" through:"auth.user_roles" through_src:"role_id" through_dest:"user_id" json:"users,omitempty"`
}

type roleTable struct {
	ID          string
	Name        string
	Description string
	CreatedAt   string
	UpdatedAt   string
	Permissions string
	Users       string
}

var RoleTable = roleTable{
	ID:          "id",
	Name:        "name",
	Description: "description",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
	Permissions: "permissions",
	Users:       "users",
}

type Permission struct {
	_           struct{}         `db:"permissions" schema:"auth" json:"-"`
	ID          uuid.UUID        `db:"id" json:"id"`
	Name        string           `db:"name" json:"name"`
	Description *string          `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time        `db:"updated_at" json:"updated_at"`
	Roles       []*Role          `db:"roles" src:"id" dest:"id" table:"auth.roles" through:"auth.role_permissions" through_src:"permission_id" through_dest:"role_id" json:"roles,omitempty"`
	Users       []*User          `db:"users" src:"id" dest:"id" table:"auth.users" through:"auth.user_permissions" through_src:"permission_id" through_dest:"user_id" json:"users,omitempty"`
	Products    []*StripeProduct `db:"products" src:"id" dest:"id" table:"billing.stripe_products" through:"billing.product_permissions" through_src:"permission_id" through_dest:"product_id" json:"products,omitempty"`
}

type permissionTable struct {
	Columns     []string
	ID          string
	Name        string
	Description string
	CreatedAt   string
	UpdatedAt   string
	Roles       string
	Users       string
	Products    string
}

var PermissionTable = permissionTable{
	Columns: []string{
		"id",
		"name",
		"description",
		"created_at",
		"updated_at",
	},
	ID:          "id",
	Name:        "name",
	Description: "description",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
	Roles:       "roles",
	Users:       "users",
	Products:    "products",
}

type PermissionSource struct {
	ID          uuid.UUID   `db:"id,pk" json:"id"`
	Name        string      `db:"name" json:"name"`
	Description *string     `db:"description" json:"description"`
	CreatedAt   time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time   `db:"updated_at" json:"updated_at"`
	RoleIDs     []uuid.UUID `db:"role_ids" json:"role_ids"`
	ProductIDs  []string    `db:"product_ids" json:"product_ids"`
	IsDirectly  bool        `db:"is_directly_assigned" json:"is_directly_assigned"`
}

type UserRole struct {
	_      struct{}  `db:"user_roles" schema:"auth" json:"-"`
	UserID uuid.UUID `db:"user_id" json:"user_id"`
	RoleID uuid.UUID `db:"role_id" json:"role_id"`
}

type UserPermission struct {
	_            struct{}  `db:"user_permissions" schema:"auth" json:"-"`
	UserID       uuid.UUID `db:"user_id" json:"user_id"`
	PermissionID uuid.UUID `db:"permission_id" json:"permission_id"`
}

type RolePermission struct {
	_            struct{}  `db:"role_permissions" schema:"auth" json:"-"`
	RoleID       uuid.UUID `db:"role_id" json:"role_id"`
	PermissionID uuid.UUID `db:"permission_id" json:"permission_id"`
}
