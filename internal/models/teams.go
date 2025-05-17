package models

import (
	"time"

	"github.com/google/uuid"
)

type Team struct {
	_                struct{}      `db:"teams" json:"-"`
	ID               uuid.UUID     `db:"id" json:"id"`
	Name             string        `db:"name" json:"name"`
	StripeCustomerID *string       `db:"stripe_customer_id" json:"stripe_customer_id"`
	CreatedAt        time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time     `db:"updated_at" json:"updated_at"`
	Members          []*TeamMember `db:"members" src:"id" dest:"team_id" table:"team_members" json:"members,omitempty"`
}

type TeamMemberRole string

const (
	TeamMemberRoleAdmin  TeamMemberRole = "admin"
	TeamMemberRoleMember TeamMemberRole = "member"
	TeamMemberRoleGuest  TeamMemberRole = "guest"
)

type TeamInvitationStatus string

const (
	TeamInvitationStatusPending  TeamInvitationStatus = "pending"
	TeamInvitationStatusAccepted TeamInvitationStatus = "accepted"
	TeamInvitationStatusDeclined TeamInvitationStatus = "declined"
)

type TeamInvitation struct {
	_             struct{}             `db:"team_invitations" json:"-"`
	ID            uuid.UUID            `db:"id" json:"id"`
	TeamID        uuid.UUID            `db:"team_id" json:"team_id"`
	InvitedBy     uuid.UUID            `db:"invited_by" json:"invited_by"`
	Email         string               `db:"email" json:"email"`
	Role          TeamMemberRole       `db:"role" json:"role"`
	Token         string               `db:"token" json:"token"`
	Status        TeamInvitationStatus `db:"status" json:"status"`
	CreatedAt     time.Time            `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time            `db:"updated_at" json:"updated_at"`
	Team          *Team                `db:"team" src:"team_id" dest:"id" table:"teams" json:"team,omitempty"`
	InvitedByUser *TeamMember          `db:"invited_by_member" src:"invited_by" dest:"id" table:"member" json:"invited_by_member,omitempty"`
}

type TeamMember struct {
	_         struct{}       `db:"team_members" json:"-"`
	ID        uuid.UUID      `db:"id" json:"id"`
	TeamID    uuid.UUID      `db:"team_id" json:"team_id"`
	UserID    *uuid.UUID     `db:"user_id" json:"user_id"`
	Role      TeamMemberRole `db:"role" json:"role"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	Team      *Team          `db:"team" src:"team_id" dest:"id" table:"team" json:"team,omitempty"`
	User      *User          `db:"user" src:"user_id" dest:"id" table:"users" json:"user,omitempty"`
}
