// package models contains the structs representing tables in the database.
//
//	type AuthUser struct {
//		_               struct{}        `db:"auth.users" json:"-"`
//		ID              uuid.UUID       `db:"id,pk" json:"id"`
//		Email           string          `db:"email" json:"email"`
//		EmailVerifiedAt *time.Time      `db:"email_verified_at" json:"email_verified_at"`
//		Name            *string         `db:"name" json:"name"`
//		Image           *string         `db:"image" json:"image"`
//		CreatedAt       time.Time       `db:"created_at" json:"created_at"`
//		UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`
//		Accounts        []*UserAccount  `db:"accounts,relation" src:"id" dest:"user_id" table:"auth.user_accounts" json:"accounts,omitempty"`
//		Roles           []*Role         `db:"roles,relation" src:"id" dest:"id" table:"auth.roles" through:"auth.user_roles" through_src:"user_id" through_dest:"role_id" json:"roles,omitempty"`
//		Permissions     []*Permission   `db:"permissions" src:"id" dest:"id" table:"auth.permissions" through:"auth.user_permissions" through_src:"user_id" through_dest:"permission_id" json:"permissions,omitempty"`
//		AiUsages        []*AiUsage      `db:"ai_usages" src:"id" dest:"user_id" table:"ai_usages" json:"ai_usages,omitempty"`
//		StripeCustomer  *StripeCustomer `db:"stripe_customer" src:"id" dest:"user_id" table:"stripe_customers" json:"stripe_customer,omitempty"`
//		TeamMembers     []*TeamMember   `db:"team_members" src:"id" dest:"user_id" table:"team_members" json:"team_members,omitempty"`
//	}
//
// each struct must have a info field `_ struct{}`
package models

import (
	"time"

	"github.com/google/uuid"
)

type AuthUser struct {
	_               struct{}       `db:"auth.users" json:"-"` // TODO: add quote option for table names
	ID              uuid.UUID      `db:"id,pk" json:"id"`
	Email           string         `db:"email" json:"email"`
	EmailVerifiedAt *time.Time     `db:"email_verified_at" json:"email_verified_at"`
	Name            *string        `db:"name" json:"name"`
	Image           *string        `db:"image" json:"image"`
	CreatedAt       time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at" json:"updated_at"`
	Accounts        []*UserAccount `db:"accounts" src:"id" dest:"user_id" table:"auth.user_accounts" json:"accounts,omitempty"`
	Roles           []*Role        `db:"roles" src:"id" dest:"id" table:"auth.roles" through_table:"auth.user_roles" through_src:"user_id" through_dest:"role_id" json:"roles,omitempty"`
	Permissions     []*Permission  `db:"permissions" src:"id" dest:"id" table:"auth.permissions" through_table:"auth.user_permissions" through_src:"user_id" through_dest:"permission_id" json:"permissions,omitempty"`
	// StripeCustomer  *StripeCustomer `db:"stripe_customer" src:"id" dest:"user_id" table:"stripe_customers" json:"stripe_customer,omitempty"`
	// TeamMembers     []*TeamMember   `db:"team_members" src:"id" dest:"user_id" table:"team_members" json:"team_members,omitempty"`
}
