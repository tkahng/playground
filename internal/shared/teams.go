package shared

import (
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

type TeamInfo struct {
	User   models.User       `json:"user"`
	Team   models.Team       `json:"team"`
	Member models.TeamMember `json:"member"`
}

type TeamMemberRole string

const (
	TeamMemberRoleOwner  TeamMemberRole = "owner"
	TeamMemberRoleMember TeamMemberRole = "member"
	TeamMemberRoleGuest  TeamMemberRole = "guest"
)

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

type Team struct {
	_    struct{}  `db:"teams" json:"-"`
	ID   uuid.UUID `db:"id" json:"id"`
	Name string    `db:"name" json:"name"`
	Slug string    `db:"slug" json:"slug"`
	// StripeCustomerID *string       `db:"stripe_customer_id" json:"stripe_customer_id"`
	CreatedAt time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt time.Time     `db:"updated_at" json:"updated_at"`
	Members   []*TeamMember `db:"members" src:"id" dest:"team_id" table:"team_members" json:"members,omitempty"`
	// StripeCustomer *Stripecus `db:"stripe_customer" src:"id" dest:"team_id" table:"stripe_customers" json:"stripe_customer,omitempty" required:"false"`
}

func FromTeamModel(team *models.Team) *Team {
	if team == nil {
		return nil
	}
	return &Team{
		ID:        team.ID,
		Name:      team.Name,
		Slug:      team.Slug,
		CreatedAt: team.CreatedAt,
		UpdatedAt: team.UpdatedAt,
		Members:   mapper.Map(team.Members, FromTeamMemberModel),
	}
}
func FromTeamMemberModel(member *models.TeamMember) *TeamMember {
	if member == nil {
		return nil
	}
	return &TeamMember{
		ID:               member.ID,
		TeamID:           member.TeamID,
		UserID:           member.UserID,
		Active:           member.Active,
		Role:             TeamMemberRole(member.Role),
		HasBillingAccess: member.HasBillingAccess,
		LastSelectedAt:   member.LastSelectedAt,
		CreatedAt:        member.CreatedAt,
		UpdatedAt:        member.UpdatedAt,
		Team:             FromTeamModel(member.Team),
		User:             FromUserModel(member.User),
	}
}

type TeamMemberListInput struct {
	PaginatedInput
	SortParams
}
