package models

import (
	"time"

	"github.com/google/uuid"
)

type Team struct {
	_              struct{}        `db:"teams" json:"-"`
	ID             uuid.UUID       `db:"id" json:"id"`
	Name           string          `db:"name" json:"name"`
	Slug           string          `db:"slug" json:"slug"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time       `db:"updated_at" json:"updated_at"`
	Members        []*TeamMember   `db:"members" src:"id" dest:"team_id" table:"team_members" json:"members,omitempty"`
	StripeCustomer *StripeCustomer `db:"stripe_customer" src:"id" dest:"team_id" table:"stripe_customers" json:"stripe_customer,omitempty" required:"false"`
}

type teamTable struct {
	ID             string
	Name           string
	Slug           string
	CreatedAt      string
	UpdatedAt      string
	Members        string
	StripeCustomer string
}

var TeamTable = teamTable{
	ID:             "id",
	Name:           "name",
	Slug:           "slug",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
	Members:        "members",
	StripeCustomer: "stripe_customer",
}

type TeamMemberRole string

const (
	TeamMemberRoleOwner  TeamMemberRole = "owner"
	TeamMemberRoleMember TeamMemberRole = "member"
	TeamMemberRoleGuest  TeamMemberRole = "guest"
)

type TeamInvitationStatus string

const (
	TeamInvitationStatusPending  TeamInvitationStatus = "pending"
	TeamInvitationStatusAccepted TeamInvitationStatus = "accepted"
	TeamInvitationStatusDeclined TeamInvitationStatus = "declined"
	TeamInvitationStatusCanceled TeamInvitationStatus = "canceled"
)

type TeamInvitation struct {
	_               struct{}             `db:"team_invitations" json:"-"`
	ID              uuid.UUID            `db:"id" json:"id"`
	TeamID          uuid.UUID            `db:"team_id" json:"team_id"`
	InviterMemberID uuid.UUID            `db:"inviter_member_id" json:"inviter_member_id"`
	Email           string               `db:"email" json:"email"`
	Role            TeamMemberRole       `db:"role" json:"role"`
	Token           string               `db:"token" json:"token"`
	Status          TeamInvitationStatus `db:"status" json:"status" enum:"pending,accepted,declined,canceled"`
	ExpiresAt       time.Time            `db:"expires_at" json:"expires_at"`
	CreatedAt       time.Time            `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time            `db:"updated_at" json:"updated_at"`
	Team            *Team                `db:"team" src:"team_id" dest:"id" table:"teams" json:"team,omitempty"`
	InviterMember   *TeamMember          `db:"inviter_member" src:"inviter_member_id" dest:"id" table:"member" json:"inviter_member,omitempty"`
}

type teamInvitationTable struct {
	ID              string
	TeamID          string
	InviterMemberID string
	Email           string
	Role            string
	Token           string
	Status          string
	ExpiresAt       string
	CreatedAt       string
	UpdatedAt       string
	Team            string
	InviterMember   string
}

var TeamInvitationTable = teamInvitationTable{
	ID:              "id",
	TeamID:          "team_id",
	InviterMemberID: "inviter_member_id",
	Email:           "email",
	Role:            "role",
	Token:           "token",
	Status:          "status",
	ExpiresAt:       "expires_at",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
	Team:            "team",
	InviterMember:   "inviter_member",
}

type TeamMember struct {
	_                struct{}       `db:"team_members" json:"-"`
	ID               uuid.UUID      `db:"id" json:"id"`
	TeamID           uuid.UUID      `db:"team_id" json:"team_id"`
	UserID           *uuid.UUID     `db:"user_id" json:"user_id"`
	Active           bool           `db:"active" json:"active"`
	Role             TeamMemberRole `db:"role" json:"role" enum:"owner,member,guest"`
	HasBillingAccess bool           `db:"has_billing_access" json:"has_billing_access"`
	LastSelectedAt   time.Time      `db:"last_selected_at" json:"last_selected_at"`
	CreatedAt        time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at" json:"updated_at"`
	Team             *Team          `db:"team" src:"team_id" dest:"id" table:"team" json:"team,omitempty"`
	User             *User          `db:"user" src:"user_id" dest:"id" table:"users" json:"user,omitempty"`
}

type teamMemberTable struct {
	ID               string
	TeamID           string
	UserID           string
	Active           string
	Role             string
	HasBillingAccess string
	LastSelectedAt   string
	CreatedAt        string
	UpdatedAt        string
	Team             string
	User             string
}

var TeamMemberTable = teamMemberTable{
	ID:               "id",
	TeamID:           "team_id",
	UserID:           "user_id",
	Active:           "active",
	Role:             "role",
	HasBillingAccess: "has_billing_access",
	LastSelectedAt:   "last_selected_at",
	CreatedAt:        "created_at",
	UpdatedAt:        "updated_at",
	Team:             "team",
	User:             "user",
}
