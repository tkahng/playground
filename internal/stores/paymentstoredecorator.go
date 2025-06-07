package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/shared"
)

type PaymentStoreDecorator struct {
	*DbStripeStore
	*DbRBACStore
	*TeamStoreDecorator
}

type TeamStoreDecorator struct {
	Delegate                         *DbTeamStore
	CheckTeamSlugFunc                func(ctx context.Context, slug string) (bool, error)
	CountOwnerTeamMembersFunc        func(ctx context.Context, teamId uuid.UUID) (int64, error)
	CountTeamMembersFunc             func(ctx context.Context, teamId uuid.UUID) (int64, error)
	CountTeamMembersByUserIDFunc     func(ctx context.Context, userId uuid.UUID) (int64, error)
	CreateTeamFunc                   func(ctx context.Context, name string, slug string) (*models.Team, error)
	CreateTeamMemberFunc             func(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error)
	DeleteTeamMemberFunc             func(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error
	GetTeamFunc                      func(ctx context.Context, teamId uuid.UUID) (*models.Team, error)
	GetTeamBySlugFunc                func(ctx context.Context, slug string) (*models.Team, error)
	GetTeamMembersFunc               func(ctx context.Context, teamId uuid.UUID) ([]*models.TeamMember, error)
	GetTeamMembersByUserIDFunc       func(ctx context.Context, userId uuid.UUID) ([]*models.TeamMember, error)
	UpdateTeamFunc                   func(ctx context.Context, teamId uuid.UUID, name string) (*models.Team, error)
	CreateTeamWithOwnerMemberFunc    func(ctx context.Context, name string, slug string, userId uuid.UUID) (*shared.TeamInfoModel, error)
	DeleteTeamFunc                   func(ctx context.Context, teamId uuid.UUID) error
	FindTeamByIDFunc                 func(ctx context.Context, teamId uuid.UUID) (*models.Team, error)
	FindTeamByStripeCustomerIdFunc   func(ctx context.Context, stripeCustomerId string) (*models.Team, error)
	FindTeamFunc                     func(ctx context.Context, team *models.Team) (*models.Team, error)
	FindTeamMemberFunc               func(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error)
	FindUserByIDFunc                 func(ctx context.Context, userId uuid.UUID) (*models.User, error)
	CountTeamsFunc                   func(ctx context.Context, params *shared.ListTeamsParams) (int64, error)
	ListTeamsFunc                    func(ctx context.Context, params *shared.ListTeamsParams) ([]*models.Team, error)
	FindLatestTeamMemberByUserIDFunc func(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error)
	LoadTeamsByIdsFunc               func(ctx context.Context, teamIds ...uuid.UUID) ([]*models.Team, error)
	UpdateTeamMemberFunc             func(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error)
	UpdateTeamMemberSelectedAtFunc   func(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error
	FindTeamBySlugFunc               func(ctx context.Context, slug string) (*models.Team, error)
}

// FindTeamBySlug implements services.TeamStore.
func (p *TeamStoreDecorator) FindTeamBySlug(ctx context.Context, slug string) (*models.Team, error) {
	if p.FindTeamBySlugFunc != nil {
		return p.FindTeamBySlugFunc(ctx, slug)
	}
	return p.Delegate.FindTeamBySlug(ctx, slug)
}

// CountTeams implements services.TeamStore.
func (p *TeamStoreDecorator) CountTeams(ctx context.Context, params *shared.ListTeamsParams) (int64, error) {
	if p.CountTeamsFunc != nil {
		return p.CountTeamsFunc(ctx, params)
	}
	return p.Delegate.CountTeams(ctx, params)
}

// ListTeams implements services.TeamStore.
func (p *TeamStoreDecorator) ListTeams(ctx context.Context, params *shared.ListTeamsParams) ([]*models.Team, error) {
	if p.ListTeamsFunc != nil {
		return p.ListTeamsFunc(ctx, params)
	}
	return p.Delegate.ListTeams(ctx, params)
}

// FindUserByID implements services.TeamStore.
func (p *TeamStoreDecorator) FindUserByID(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	if p.FindUserByIDFunc != nil {
		return p.FindUserByIDFunc(ctx, userId)
	}
	return p.Delegate.FindUserByID(ctx, userId)
}

// FindTeam implements services.TeamStore.
func (p *TeamStoreDecorator) FindTeam(ctx context.Context, team *models.Team) (*models.Team, error) {
	if p.FindTeamFunc != nil {
		return p.FindTeamFunc(ctx, team)
	}
	return p.Delegate.FindTeam(ctx, team)
}

// FindTeamMember implements services.TeamStore.
func (p *TeamStoreDecorator) FindTeamMember(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error) {
	if p.FindTeamMemberFunc != nil {
		return p.FindTeamMemberFunc(ctx, member)
	}
	return p.Delegate.FindTeamMember(ctx, member)
}

func NewTeamStoreDecorator(delegate *DbTeamStore) *TeamStoreDecorator {
	return &TeamStoreDecorator{
		Delegate: delegate,
	}
}

// CheckTeamSlug implements services.TeamStore.
func (p *TeamStoreDecorator) CheckTeamSlug(ctx context.Context, slug string) (bool, error) {
	if p.CheckTeamSlugFunc != nil {
		return p.CheckTeamSlugFunc(ctx, slug)
	}
	return p.Delegate.CheckTeamSlug(ctx, slug)
}

// CountOwnerTeamMembers implements services.TeamStore.
func (p *TeamStoreDecorator) CountOwnerTeamMembers(ctx context.Context, teamId uuid.UUID) (int64, error) {
	if p.CountOwnerTeamMembersFunc != nil {
		return p.CountOwnerTeamMembersFunc(ctx, teamId)
	}
	return p.Delegate.CountOwnerTeamMembers(ctx, teamId)
}

// CountTeamMembers implements services.TeamStore.
func (p *TeamStoreDecorator) CountTeamMembers(ctx context.Context, teamId uuid.UUID) (int64, error) {
	if p.CountTeamMembersFunc != nil {
		return p.CountTeamMembersFunc(ctx, teamId)
	}
	return p.Delegate.CountTeamMembers(ctx, teamId)
}

// CountTeamMembersByUserID implements services.TeamStore.
func (p *TeamStoreDecorator) CountTeamMembersByUserID(ctx context.Context, userId uuid.UUID) (int64, error) {
	if p.CountTeamMembersByUserIDFunc != nil {
		return p.CountTeamMembersByUserIDFunc(ctx, userId)
	}
	return p.Delegate.CountTeamMembersByUserID(ctx, userId)
}

// CreateTeam implements services.TeamStore.
func (p *TeamStoreDecorator) CreateTeam(ctx context.Context, name string, slug string) (*models.Team, error) {
	if p.CreateTeamFunc != nil {
		return p.CreateTeamFunc(ctx, name, slug)
	}
	return p.Delegate.CreateTeam(ctx, name, slug)
}

// CreateTeamMember implements services.TeamStore.
func (p *TeamStoreDecorator) CreateTeamMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error) {
	if p.CreateTeamMemberFunc != nil {
		return p.CreateTeamMemberFunc(ctx, teamId, userId, role, hasBillingAccess)
	}
	return p.Delegate.CreateTeamMember(ctx, teamId, userId, role, hasBillingAccess)
}

// CreateTeamWithOwnerMember implements services.TeamStore.
func (p *TeamStoreDecorator) CreateTeamWithOwnerMember(ctx context.Context, name string, slug string, userId uuid.UUID) (*shared.TeamInfoModel, error) {
	if p.CreateTeamWithOwnerMemberFunc != nil {
		return p.CreateTeamWithOwnerMemberFunc(ctx, name, slug, userId)
	}
	return p.Delegate.CreateTeamWithOwnerMember(ctx, name, slug, userId)
}

// DeleteTeam implements services.TeamStore.
func (p *TeamStoreDecorator) DeleteTeam(ctx context.Context, teamId uuid.UUID) error {
	if p.DeleteTeamFunc != nil {
		return p.DeleteTeamFunc(ctx, teamId)
	}
	return p.Delegate.DeleteTeam(ctx, teamId)
}

// DeleteTeamMember implements services.TeamStore.
func (p *TeamStoreDecorator) DeleteTeamMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error {
	if p.DeleteTeamMemberFunc != nil {
		return p.DeleteTeamMemberFunc(ctx, teamId, userId)
	}
	return p.Delegate.DeleteTeamMember(ctx, teamId, userId)
}

// FindLatestTeamMemberByUserID implements services.TeamStore.
func (p *TeamStoreDecorator) FindLatestTeamMemberByUserID(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error) {
	if p.FindLatestTeamMemberByUserIDFunc != nil {
		return p.FindLatestTeamMemberByUserIDFunc(ctx, userId)
	}
	return p.Delegate.FindLatestTeamMemberByUserID(ctx, userId)
}

// FindTeamByID implements services.TeamStore.
func (p *TeamStoreDecorator) FindTeamByID(ctx context.Context, teamId uuid.UUID) (*models.Team, error) {
	if p.FindTeamByIDFunc != nil {
		return p.GetTeamFunc(ctx, teamId)
	}
	return p.Delegate.FindTeamByID(ctx, teamId)
}

// FindTeamByStripeCustomerId implements services.TeamStore.
func (p *TeamStoreDecorator) FindTeamByStripeCustomerId(ctx context.Context, stripeCustomerId string) (*models.Team, error) {
	if p.FindTeamByStripeCustomerIdFunc != nil {
		return p.FindTeamByStripeCustomerIdFunc(ctx, stripeCustomerId)
	}
	return p.Delegate.FindTeamByStripeCustomerId(ctx, stripeCustomerId)
}

// FindTeamMemberByTeamAndUserId implements services.TeamStore.
func (p *TeamStoreDecorator) FindTeamMemberByTeamAndUserId(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) (*models.TeamMember, error) {
	return p.Delegate.FindTeamMemberByTeamAndUserId(ctx, teamId, userId)
}

// FindTeamMembersByUserID implements services.TeamStore.
func (p *TeamStoreDecorator) FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID, paginate *shared.TeamMemberListInput) ([]*models.TeamMember, error) {
	if p.GetTeamMembersByUserIDFunc != nil {
		return p.GetTeamMembersByUserIDFunc(ctx, userId)
	}
	return p.Delegate.FindTeamMembersByUserID(ctx, userId, paginate)
}

// LoadTeamsByIds implements services.TeamStore.
func (p *TeamStoreDecorator) LoadTeamsByIds(ctx context.Context, teamIds ...uuid.UUID) ([]*models.Team, error) {
	if p.LoadTeamsByIdsFunc != nil {
		return p.LoadTeamsByIdsFunc(ctx, teamIds...)
	}
	return p.Delegate.LoadTeamsByIds(ctx, teamIds...)
}

// UpdateTeam implements services.TeamStore.
func (p *TeamStoreDecorator) UpdateTeam(ctx context.Context, teamId uuid.UUID, name string) (*models.Team, error) {
	if p.UpdateTeamFunc != nil {
		return p.UpdateTeamFunc(ctx, teamId, name)
	}
	return p.Delegate.UpdateTeam(ctx, teamId, name)
}

// UpdateTeamMember implements services.TeamStore.
func (p *TeamStoreDecorator) UpdateTeamMember(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error) {
	if p.UpdateTeamMemberFunc != nil {
		return p.UpdateTeamMemberFunc(ctx, member)
	}
	return p.Delegate.UpdateTeamMember(ctx, member)
}

// UpdateTeamMemberSelectedAt implements services.TeamStore.
func (p *TeamStoreDecorator) UpdateTeamMemberSelectedAt(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error {
	if p.UpdateTeamMemberSelectedAtFunc != nil {
		return p.UpdateTeamMemberSelectedAtFunc(ctx, teamId, userId)
	}
	return p.Delegate.UpdateTeamMemberSelectedAt(ctx, teamId, userId)
}

var _ services.TeamStore = (*TeamStoreDecorator)(nil)
