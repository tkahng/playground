package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
)

type TeamInvitationStoreDecorator struct {
	Delegate                  *DbTeamInvitationStore
	CreateInvitationFunc      func(ctx context.Context, invitation *models.TeamInvitation) error
	FindInvitationByIDFunc    func(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error)
	FindInvitationByTokenFunc func(ctx context.Context, token string) (*models.TeamInvitation, error)
	FindPendingInvitationFunc func(ctx context.Context, teamId uuid.UUID, email string) (*models.TeamInvitation, error)
	FindTeamInvitationsFunc   func(ctx context.Context, params *TeamInvitationFilter) ([]*models.TeamInvitation, error)
	GetInvitationByIDFunc     func(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error)
	UpdateInvitationFunc      func(ctx context.Context, invitation *models.TeamInvitation) error
	CountTeamInvitationsFunc  func(ctx context.Context, filter *TeamInvitationFilter) (int64, error)
}

// CountTeamInvitations implements DbTeamInvitationStoreInterface.
func (t *TeamInvitationStoreDecorator) CountTeamInvitations(ctx context.Context, params *TeamInvitationFilter) (int64, error) {
	if t.CountTeamInvitationsFunc != nil {
		return t.CountTeamInvitationsFunc(ctx, params)
	}
	if t.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return t.Delegate.CountTeamInvitations(ctx, params)
}

func (t *TeamInvitationStoreDecorator) Cleanup() {
	t.CreateInvitationFunc = nil
	t.FindInvitationByIDFunc = nil
	t.FindInvitationByTokenFunc = nil
	t.FindPendingInvitationFunc = nil
	t.FindTeamInvitationsFunc = nil
	t.GetInvitationByIDFunc = nil
	t.UpdateInvitationFunc = nil
}

// CreateInvitation implements DbTeamInvitationStoreInterface.
func (t *TeamInvitationStoreDecorator) CreateInvitation(ctx context.Context, invitation *models.TeamInvitation) error {
	if t.CreateInvitationFunc != nil {
		return t.CreateInvitationFunc(ctx, invitation)
	}
	if t.Delegate == nil {
		return ErrDelegateNil
	}
	return t.Delegate.CreateInvitation(ctx, invitation)
}

// FindInvitationByID implements DbTeamInvitationStoreInterface.
func (t *TeamInvitationStoreDecorator) FindInvitationByID(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error) {
	if t.FindInvitationByIDFunc != nil {
		return t.FindInvitationByIDFunc(ctx, invitationId)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindInvitationByID(ctx, invitationId)
}

// FindInvitationByToken implements DbTeamInvitationStoreInterface.
func (t *TeamInvitationStoreDecorator) FindInvitationByToken(ctx context.Context, token string) (*models.TeamInvitation, error) {
	if t.FindInvitationByTokenFunc != nil {
		return t.FindInvitationByTokenFunc(ctx, token)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindInvitationByToken(ctx, token)
}

// FindPendingInvitation implements DbTeamInvitationStoreInterface.
func (t *TeamInvitationStoreDecorator) FindPendingInvitation(ctx context.Context, teamId uuid.UUID, email string) (*models.TeamInvitation, error) {
	if t.FindPendingInvitationFunc != nil {
		return t.FindPendingInvitationFunc(ctx, teamId, email)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindPendingInvitation(ctx, teamId, email)
}

// FindTeamInvitations implements DbTeamInvitationStoreInterface.
func (t *TeamInvitationStoreDecorator) FindTeamInvitations(ctx context.Context, params *TeamInvitationFilter) ([]*models.TeamInvitation, error) {
	if t.FindTeamInvitationsFunc != nil {
		return t.FindTeamInvitationsFunc(ctx, params)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindTeamInvitations(ctx, params)
}

// GetInvitationByID implements DbTeamInvitationStoreInterface.
func (t *TeamInvitationStoreDecorator) GetInvitationByID(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error) {
	if t.GetInvitationByIDFunc != nil {
		return t.GetInvitationByIDFunc(ctx, invitationId)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.GetInvitationByID(ctx, invitationId)
}

// UpdateInvitation implements DbTeamInvitationStoreInterface.
func (t *TeamInvitationStoreDecorator) UpdateInvitation(ctx context.Context, invitation *models.TeamInvitation) error {
	if t.UpdateInvitationFunc != nil {
		return t.UpdateInvitationFunc(ctx, invitation)
	}
	if t.Delegate == nil {
		return ErrDelegateNil
	}
	return t.Delegate.UpdateInvitation(ctx, invitation)
}

var _ DbTeamInvitationStoreInterface = (*TeamInvitationStoreDecorator)(nil)
