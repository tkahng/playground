package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	_               struct{}        `db:"users" json:"-"`
	ID              uuid.UUID       `db:"id" json:"id"`
	Email           string          `db:"email" json:"email"`
	EmailVerifiedAt *time.Time      `db:"email_verified_at" json:"email_verified_at"`
	Name            *string         `db:"name" json:"name"`
	Image           *string         `db:"image" json:"image"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`
	Accounts        []*UserAccount  `db:"accounts" src:"id" dest:"user_id" table:"user_accounts" json:"accounts,omitempty"`
	Roles           []*Role         `db:"roles" src:"id" dest:"user_id" table:"roles" through:"user_roles,role_id,id" json:"roles,omitempty"`
	Permissions     []*Permission   `db:"permissions" src:"id" dest:"user_id" table:"permissions" through:"user_permissions,permission_id,id" json:"permissions,omitempty"`
	AiUsages        []*AiUsage      `db:"ai_usages" src:"id" dest:"user_id" table:"ai_usages" json:"ai_usages,omitempty"`
	StripeCustomer  *StripeCustomer `db:"stripe_customer" src:"id" dest:"user_id" table:"stripe_customers" json:"stripe_customer,omitempty"`
	TeamMembers     []*TeamMember   `db:"team_members" src:"id" dest:"user_id" table:"team_members" json:"team_members,omitempty"`
}

type userTable struct {
	Columns         []string
	ID              string
	Email           string
	EmailVerifiedAt string
	Name            string
	Image           string
	CreatedAt       string
	UpdatedAt       string
	Accounts        string
	Roles           string
	Permissions     string
	AiUsages        string
	StripeCustomer  string
	TeamMembers     string
}

var UserTable = userTable{
	Columns: []string{
		"id",
		"email",
		"email_verified_at",
		"name",
		"image",
		"created_at",
		"updated_at",
	},
	ID:              "id",
	Email:           "email",
	EmailVerifiedAt: "email_verified_at",
	Name:            "name",
	Image:           "image",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
	Accounts:        "accounts",
	Roles:           "roles",
	Permissions:     "permissions",
	AiUsages:        "ai_usages",
	StripeCustomer:  "stripe_customer",
	TeamMembers:     "team_members",
}
