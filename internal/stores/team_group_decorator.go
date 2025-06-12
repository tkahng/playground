package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
)

type TeamGroupStoreDecorator struct {
	Delegate                       *DbTeamGroupStore
	CheckTeamSlugFunc              func(ctx context.Context, slug string) (bool, error)
	CountTeamsFunc                 func(ctx context.Context, params *TeamFilter) (int64, error)
	CreateTeamFunc                 func(ctx context.Context, name string, slug string) (*models.Team, error)
	DeleteTeamFunc                 func(ctx context.Context, teamId uuid.UUID) error
	FindTeamFunc                   func(ctx context.Context, team *TeamFilter) (*models.Team, error)
	FindTeamByIDFunc               func(ctx context.Context, teamId uuid.UUID) (*models.Team, error)
	FindTeamBySlugFunc             func(ctx context.Context, slug string) (*models.Team, error)
	FindTeamByStripeCustomerIdFunc func(ctx context.Context, stripeCustomerId string) (*models.Team, error)
	ListTeamsFunc                  func(ctx context.Context, params *TeamFilter) ([]*models.Team, error)
	LoadTeamsByIdsFunc             func(ctx context.Context, teamIds ...uuid.UUID) ([]*models.Team, error)
	UpdateTeamFunc                 func(ctx context.Context, teamId uuid.UUID, name string) (*models.Team, error)
}

func (t *TeamGroupStoreDecorator) Cleanup() {
	t.CheckTeamSlugFunc = nil
	t.CountTeamsFunc = nil
	t.CreateTeamFunc = nil
	t.DeleteTeamFunc = nil
	t.FindTeamFunc = nil
	t.FindTeamByIDFunc = nil
	t.FindTeamBySlugFunc = nil
	t.FindTeamByStripeCustomerIdFunc = nil
	t.ListTeamsFunc = nil
	t.LoadTeamsByIdsFunc = nil
	t.UpdateTeamFunc = nil
}

// CheckTeamSlug implements DbTeamGroupStoreInterface.
func (t *TeamGroupStoreDecorator) CheckTeamSlug(ctx context.Context, slug string) (bool, error) {
	if t.CheckTeamSlugFunc != nil {
		return t.CheckTeamSlugFunc(ctx, slug)
	}
	if t.Delegate == nil {
		return false, ErrDelegateNil
	}
	return t.Delegate.CheckTeamSlug(ctx, slug)
}

// CountTeams implements DbTeamGroupStoreInterface.
func (t *TeamGroupStoreDecorator) CountTeams(ctx context.Context, params *TeamFilter) (int64, error) {
	if t.CountTeamsFunc != nil {
		return t.CountTeamsFunc(ctx, params)
	}
	if t.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return t.Delegate.CountTeams(ctx, params)
}

// CreateTeam implements DbTeamGroupStoreInterface.
func (t *TeamGroupStoreDecorator) CreateTeam(ctx context.Context, name string, slug string) (*models.Team, error) {
	if t.CreateTeamFunc != nil {
		return t.CreateTeamFunc(ctx, name, slug)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.CreateTeam(ctx, name, slug)
}

// DeleteTeam implements DbTeamGroupStoreInterface.
func (t *TeamGroupStoreDecorator) DeleteTeam(ctx context.Context, teamId uuid.UUID) error {
	if t.DeleteTeamFunc != nil {
		return t.DeleteTeamFunc(ctx, teamId)
	}
	if t.Delegate == nil {
		return ErrDelegateNil
	}
	return t.Delegate.DeleteTeam(ctx, teamId)
}

// FindTeam implements DbTeamGroupStoreInterface.
func (t *TeamGroupStoreDecorator) FindTeam(ctx context.Context, team *TeamFilter) (*models.Team, error) {
	if t.FindTeamFunc != nil {
		return t.FindTeamFunc(ctx, team)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindTeam(ctx, team)
}

// FindTeamByID implements DbTeamGroupStoreInterface.
func (t *TeamGroupStoreDecorator) FindTeamByID(ctx context.Context, teamId uuid.UUID) (*models.Team, error) {
	if t.FindTeamByIDFunc != nil {
		return t.FindTeamByIDFunc(ctx, teamId)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindTeamByID(ctx, teamId)
}

// FindTeamBySlug implements DbTeamGroupStoreInterface.
func (t *TeamGroupStoreDecorator) FindTeamBySlug(ctx context.Context, slug string) (*models.Team, error) {
	if t.FindTeamBySlugFunc != nil {
		return t.FindTeamBySlugFunc(ctx, slug)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindTeamBySlug(ctx, slug)
}

// FindTeamByStripeCustomerId implements DbTeamGroupStoreInterface.
func (t *TeamGroupStoreDecorator) FindTeamByStripeCustomerId(ctx context.Context, stripeCustomerId string) (*models.Team, error) {
	if t.FindTeamByStripeCustomerIdFunc != nil {
		return t.FindTeamByStripeCustomerIdFunc(ctx, stripeCustomerId)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindTeamByStripeCustomerId(ctx, stripeCustomerId)
}

// ListTeams implements DbTeamGroupStoreInterface.
func (t *TeamGroupStoreDecorator) ListTeams(ctx context.Context, params *TeamFilter) ([]*models.Team, error) {
	if t.ListTeamsFunc != nil {
		return t.ListTeamsFunc(ctx, params)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.ListTeams(ctx, params)
}

// LoadTeamsByIds implements DbTeamGroupStoreInterface.
func (t *TeamGroupStoreDecorator) LoadTeamsByIds(ctx context.Context, teamIds ...uuid.UUID) ([]*models.Team, error) {
	if t.LoadTeamsByIdsFunc != nil {
		return t.LoadTeamsByIdsFunc(ctx, teamIds...)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.LoadTeamsByIds(ctx, teamIds...)
}

// UpdateTeam implements DbTeamGroupStoreInterface.
func (t *TeamGroupStoreDecorator) UpdateTeam(ctx context.Context, teamId uuid.UUID, name string) (*models.Team, error) {
	if t.UpdateTeamFunc != nil {
		return t.UpdateTeamFunc(ctx, teamId, name)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.UpdateTeam(ctx, teamId, name)
}

var _ DbTeamGroupStoreInterface = (*TeamGroupStoreDecorator)(nil)
