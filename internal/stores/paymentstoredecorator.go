package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/shared"
)

type PaymentStoreDecorator struct {
	*PostgresStripeStore
	*PostgresRBACStore
	*PostgresTeamStore
}

type PaymentTeamStoreDecorator struct {
	Delegate                       *PostgresTeamStore
	CheckTeamSlugFunc              func(ctx context.Context, slug string) (bool, error)
	CountOwnerTeamMembersFunc      func(ctx context.Context, teamId uuid.UUID) (int64, error)
	CountTeamMembersFunc           func(ctx context.Context, teamId uuid.UUID) (int64, error)
	CountTeamMembersByUserIDFunc   func(ctx context.Context, userId uuid.UUID) (int64, error)
	CreateTeamFunc                 func(ctx context.Context, name string, slug string) (*models.Team, error)
	CreateTeamMemberFunc           func(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error)
	DeleteTeamMemberFunc           func(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error
	GetTeamFunc                    func(ctx context.Context, teamId uuid.UUID) (*models.Team, error)
	GetTeamBySlugFunc              func(ctx context.Context, slug string) (*models.Team, error)
	GetTeamMembersFunc             func(ctx context.Context, teamId uuid.UUID) ([]*models.TeamMember, error)
	GetTeamMembersByUserIDFunc     func(ctx context.Context, userId uuid.UUID) ([]*models.TeamMember, error)
	UpdateTeamFunc                 func(ctx context.Context, teamId uuid.UUID, team *models.Team) error
	CreateTeamWithOwnerMemberFunc  func(ctx context.Context, name string, slug string, userId uuid.UUID) (*shared.TeamInfo, error)
	DeleteTeamFunc                 func(ctx context.Context, teamId uuid.UUID) error
	FindTeamByIDFunc               func(ctx context.Context, teamId uuid.UUID) (*models.Team, error)
	FindTeamByStripeCustomerIdFunc func(ctx context.Context, stripeCustomerId string) (*models.Team, error)
}

// CheckTeamSlug implements services.TeamStore.
func (p *PaymentTeamStoreDecorator) CheckTeamSlug(ctx context.Context, slug string) (bool, error) {
	if p.CheckTeamSlugFunc != nil {
		return p.CheckTeamSlugFunc(ctx, slug)
	}
	return p.Delegate.CheckTeamSlug(ctx, slug)
}

// CountOwnerTeamMembers implements services.TeamStore.
func (p *PaymentTeamStoreDecorator) CountOwnerTeamMembers(ctx context.Context, teamId uuid.UUID) (int64, error) {
	if p.CountOwnerTeamMembersFunc != nil {
		return p.CountOwnerTeamMembersFunc(ctx, teamId)
	}
	return p.Delegate.CountOwnerTeamMembers(ctx, teamId)
}

// CountTeamMembers implements services.TeamStore.
func (p *PaymentTeamStoreDecorator) CountTeamMembers(ctx context.Context, teamId uuid.UUID) (int64, error) {
	if p.CountTeamMembersFunc != nil {
		return p.CountTeamMembersFunc(ctx, teamId)
	}
	return p.Delegate.CountTeamMembers(ctx, teamId)
}

// CountTeamMembersByUserID implements services.TeamStore.
func (p *PaymentTeamStoreDecorator) CountTeamMembersByUserID(ctx context.Context, userId uuid.UUID) (int64, error) {
	if p.CountTeamMembersByUserIDFunc != nil {
		return p.CountTeamMembersByUserIDFunc(ctx, userId)
	}
	return p.Delegate.CountTeamMembersByUserID(ctx, userId)
}

// CreateTeam implements services.TeamStore.
func (p *PaymentTeamStoreDecorator) CreateTeam(ctx context.Context, name string, slug string) (*models.Team, error) {
	if p.CreateTeamFunc != nil {
		return p.CreateTeamFunc(ctx, name, slug)
	}
	return p.Delegate.CreateTeam(ctx, name, slug)
}

// CreateTeamMember implements services.TeamStore.
func (p *PaymentTeamStoreDecorator) CreateTeamMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error) {
	if p.CreateTeamMemberFunc != nil {
		return p.CreateTeamMemberFunc(ctx, teamId, userId, role, hasBillingAccess)
	}
	return p.Delegate.CreateTeamMember(ctx, teamId, userId, role, hasBillingAccess)
}

// CreateTeamWithOwnerMember implements services.TeamStore.
func (p *PaymentTeamStoreDecorator) CreateTeamWithOwnerMember(ctx context.Context, name string, slug string, userId uuid.UUID) (*shared.TeamInfo, error) {
	if p.CreateTeamWithOwnerMemberFunc != nil {
		return p.CreateTeamWithOwnerMemberFunc(ctx, name, slug, userId)
	}
	return p.Delegate.CreateTeamWithOwnerMember(ctx, name, slug, userId)
}

// DeleteTeam implements services.TeamStore.
func (p *PaymentTeamStoreDecorator) DeleteTeam(ctx context.Context, teamId uuid.UUID) error {
	if p.DeleteTeamFunc != nil {
		return p.DeleteTeamFunc(ctx, teamId)
	}
	return p.Delegate.DeleteTeam(ctx, teamId)
}

// DeleteTeamMember implements services.TeamStore.
func (p *PaymentTeamStoreDecorator) DeleteTeamMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error {
	if p.DeleteTeamMemberFunc != nil {
		return p.DeleteTeamMemberFunc(ctx, teamId, userId)
	}
	return p.Delegate.DeleteTeamMember(ctx, teamId, userId)
}

// FindLatestTeamMemberByUserID implements services.TeamStore.
func (p *PaymentTeamStoreDecorator) FindLatestTeamMemberByUserID(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error) {
	return p.Delegate.FindLatestTeamMemberByUserID(ctx, userId)
}

// FindTeamByID implements services.TeamStore.
func (p *PaymentTeamStoreDecorator) FindTeamByID(ctx context.Context, teamId uuid.UUID) (*models.Team, error) {
	if p.FindTeamByIDFunc != nil {
		return p.GetTeamFunc(ctx, teamId)
	}
	return p.Delegate.FindTeamByID(ctx, teamId)
}

// FindTeamByStripeCustomerId implements services.TeamStore.
func (p *PaymentTeamStoreDecorator) FindTeamByStripeCustomerId(ctx context.Context, stripeCustomerId string) (*models.Team, error) {
	if p.FindTeamByStripeCustomerIdFunc != nil {
		return p.FindTeamByStripeCustomerIdFunc(ctx, stripeCustomerId)
	}
	return p.Delegate.FindTeamByStripeCustomerId(ctx, stripeCustomerId)
}

// FindTeamMemberByTeamAndUserId implements services.TeamStore.
func (p *PaymentTeamStoreDecorator) FindTeamMemberByTeamAndUserId(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) (*models.TeamMember, error) {
	return p.Delegate.FindTeamMemberByTeamAndUserId(ctx, teamId, userId)
}

// FindTeamMembersByUserID implements services.TeamStore.
func (p *PaymentTeamStoreDecorator) FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID, paginate *shared.PaginatedInput) ([]*models.TeamMember, error) {
	return p.Delegate.FindTeamMembersByUserID(ctx, userId, paginate)
}

// LoadTeamsByIds implements services.TeamStore.
func (p *PaymentTeamStoreDecorator) LoadTeamsByIds(ctx context.Context, teamIds ...uuid.UUID) ([]*models.Team, error) {
	return p.Delegate.LoadTeamsByIds(ctx, teamIds...)
}

// UpdateTeam implements services.TeamStore.
func (p *PaymentTeamStoreDecorator) UpdateTeam(ctx context.Context, teamId uuid.UUID, name string) (*models.Team, error) {
	return p.Delegate.UpdateTeam(ctx, teamId, name)
}

// UpdateTeamMember implements services.TeamStore.
func (p *PaymentTeamStoreDecorator) UpdateTeamMember(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error) {
	return p.Delegate.UpdateTeamMember(ctx, member)
}

// UpdateTeamMemberSelectedAt implements services.TeamStore.
func (p *PaymentTeamStoreDecorator) UpdateTeamMemberSelectedAt(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error {
	return p.Delegate.UpdateTeamMemberSelectedAt(ctx, teamId, userId)
}

var _ services.TeamStore = (*PaymentTeamStoreDecorator)(nil)
