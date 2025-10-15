package models

import (
	"time"

	"github.com/google/uuid"
)

type AuthRole struct {
	_           struct{}          `db:"roles" json:"-"`
	ID          uuid.UUID         `db:"id" json:"id"`
	Name        string            `db:"name" json:"name"`
	Description *string           `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `db:"updated_at" json:"updated_at"`
	Permissions []*AuthPermission `db:"permissions" src:"id" dest:"id" table:"public.permissions" through:"public.role_permissions" through_src:"role_id" through_dest:"permission_id" json:"permissions,omitempty"`
	Users       []*AuthUser       `db:"users" src:"id" dest:"id" table:"public.users" through:"public.user_roles" through_src:"role_id" through_dest:"user_id" json:"users,omitempty"`
}

type AuthPermission struct {
	_           struct{}         `db:"permissions" json:"-"`
	ID          uuid.UUID        `db:"id" json:"id"`
	Name        string           `db:"name" json:"name"`
	Description *string          `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time        `db:"updated_at" json:"updated_at"`
	Roles       []*Role          `db:"roles" src:"id" dest:"id" table:"public.roles" through:"public.role_permissions" through_src:"permission_id" through_dest:"role_id" json:"roles,omitempty"`
	Users       []*User          `db:"users" src:"id" dest:"id" table:"public.users" through:"public.user_permissions" through_src:"permission_id" through_dest:"user_id" json:"users,omitempty"`
	Products    []*StripeProduct `db:"products" src:"id" dest:"id" table:"public.stripe_products" through:"public.product_permissions" through_src:"permission_id" through_dest:"product_id" json:"products,omitempty"`
}

type AuthPermissionSource struct {
	ID          uuid.UUID   `db:"id,pk" json:"id"`
	Name        string      `db:"name" json:"name"`
	Description *string     `db:"description" json:"description"`
	CreatedAt   time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time   `db:"updated_at" json:"updated_at"`
	RoleIDs     []uuid.UUID `db:"role_ids" json:"role_ids"`
	ProductIDs  []string    `db:"product_ids" json:"product_ids"`
	IsDirectly  bool        `db:"is_directly_assigned" json:"is_directly_assigned"`
}

type AuthUserRole struct {
	_      struct{}  `db:"user_roles" json:"-"`
	UserID uuid.UUID `db:"user_id" json:"user_id"`
	RoleID uuid.UUID `db:"role_id" json:"role_id"`
}

type AuthRolePermission struct {
	_            struct{}  `db:"role_permissions" json:"-"`
	RoleID       uuid.UUID `db:"role_id" json:"role_id"`
	PermissionID uuid.UUID `db:"permission_id" json:"permission_id"`
}
