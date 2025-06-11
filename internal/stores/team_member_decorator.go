package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type TeamMemberStoreDecorator struct {
	Delegate                          *DbTeamMemberStore
	CountOwnerTeamMembersFunc         func(ctx context.Context, teamId uuid.UUID) (int64, error)
	CountTeamMembersFunc              func(ctx context.Context, filter *TeamMemberFilter) (int64, error)
	CountTeamMembersByUserIDFunc      func(ctx context.Context, userId uuid.UUID) (int64, error)
	CreateTeamFromUserFunc            func(ctx context.Context, user *models.User) (*models.TeamMember, error)
	CreateTeamMemberFunc              func(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error)
	DeleteTeamMemberFunc              func(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error
	FindLatestTeamMemberByUserIDFunc  func(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error)
	FindTeamMemberFunc                func(ctx context.Context, member *TeamMemberFilter) (*models.TeamMember, error)
	FindTeamMemberByTeamAndUserIdFunc func(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) (*models.TeamMember, error)
	FindTeamMembersByUserIDFunc       func(ctx context.Context, userId uuid.UUID, paginate *shared.TeamMemberListInput) ([]*models.TeamMember, error)
	UpdateTeamMemberFunc              func(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error)
	UpdateTeamMemberSelectedAtFunc    func(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error

	// Add any additional methods or fields for the decorator here
}

// CountOwnerTeamMembers implements DbTeamMemberStoreInterface.
func (t *TeamMemberStoreDecorator) CountOwnerTeamMembers(ctx context.Context, teamId uuid.UUID) (int64, error) {
	if t.CountOwnerTeamMembersFunc != nil {
		return t.CountOwnerTeamMembersFunc(ctx, teamId)
	}
	if t.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return t.Delegate.CountOwnerTeamMembers(ctx, teamId)
}

// CountTeamMembers implements DbTeamMemberStoreInterface.
func (t *TeamMemberStoreDecorator) CountTeamMembers(ctx context.Context, filter *TeamMemberFilter) (int64, error) {
	if t.CountTeamMembersFunc != nil {
		return t.CountTeamMembersFunc(ctx, filter)
	}
	if t.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return t.Delegate.CountTeamMembers(ctx, filter)
}

// CountTeamMembersByUserID implements DbTeamMemberStoreInterface.
func (t *TeamMemberStoreDecorator) CountTeamMembersByUserID(ctx context.Context, userId uuid.UUID) (int64, error) {
	if t.CountTeamMembersByUserIDFunc != nil {
		return t.CountTeamMembersByUserIDFunc(ctx, userId)
	}
	if t.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return t.Delegate.CountTeamMembersByUserID(ctx, userId)
}

// CreateTeamFromUser implements DbTeamMemberStoreInterface.
func (t *TeamMemberStoreDecorator) CreateTeamFromUser(ctx context.Context, user *models.User) (*models.TeamMember, error) {
	if t.CreateTeamFromUserFunc != nil {
		return t.CreateTeamFromUserFunc(ctx, user)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.CreateTeamFromUser(ctx, user)
}

// CreateTeamMember implements DbTeamMemberStoreInterface.
func (t *TeamMemberStoreDecorator) CreateTeamMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error) {
	if t.CreateTeamMemberFunc != nil {
		return t.CreateTeamMemberFunc(ctx, teamId, userId, role, hasBillingAccess)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.CreateTeamMember(ctx, teamId, userId, role, hasBillingAccess)
}

// DeleteTeamMember implements DbTeamMemberStoreInterface.
func (t *TeamMemberStoreDecorator) DeleteTeamMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error {
	if t.DeleteTeamMemberFunc != nil {
		return t.DeleteTeamMemberFunc(ctx, teamId, userId)
	}
	if t.Delegate == nil {
		return ErrDelegateNil
	}
	return t.Delegate.DeleteTeamMember(ctx, teamId, userId)
}

// FindLatestTeamMemberByUserID implements DbTeamMemberStoreInterface.
func (t *TeamMemberStoreDecorator) FindLatestTeamMemberByUserID(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error) {
	if t.FindLatestTeamMemberByUserIDFunc != nil {
		return t.FindLatestTeamMemberByUserIDFunc(ctx, userId)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindLatestTeamMemberByUserID(ctx, userId)
}

// FindTeamMember implements DbTeamMemberStoreInterface.
func (t *TeamMemberStoreDecorator) FindTeamMember(ctx context.Context, member *TeamMemberFilter) (*models.TeamMember, error) {
	if t.FindTeamMemberFunc != nil {
		return t.FindTeamMemberFunc(ctx, member)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindTeamMember(ctx, member)
}

// FindTeamMemberByTeamAndUserId implements DbTeamMemberStoreInterface.
func (t *TeamMemberStoreDecorator) FindTeamMemberByTeamAndUserId(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) (*models.TeamMember, error) {
	if t.FindTeamMemberByTeamAndUserIdFunc != nil {
		return t.FindTeamMemberByTeamAndUserIdFunc(ctx, teamId, userId)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindTeamMemberByTeamAndUserId(ctx, teamId, userId)
}

// FindTeamMembersByUserID implements DbTeamMemberStoreInterface.
func (t *TeamMemberStoreDecorator) FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID, paginate *shared.TeamMemberListInput) ([]*models.TeamMember, error) {
	if t.FindTeamMembersByUserIDFunc != nil {
		return t.FindTeamMembersByUserIDFunc(ctx, userId, paginate)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindTeamMembersByUserID(ctx, userId, paginate)
}

// UpdateTeamMember implements DbTeamMemberStoreInterface.
func (t *TeamMemberStoreDecorator) UpdateTeamMember(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error) {
	if t.UpdateTeamMemberFunc != nil {
		return t.UpdateTeamMemberFunc(ctx, member)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.UpdateTeamMember(ctx, member)
}

// UpdateTeamMemberSelectedAt implements DbTeamMemberStoreInterface.
func (t *TeamMemberStoreDecorator) UpdateTeamMemberSelectedAt(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error {
	if t.UpdateTeamMemberSelectedAtFunc != nil {
		return t.UpdateTeamMemberSelectedAtFunc(ctx, teamId, userId)
	}
	if t.Delegate == nil {
		return ErrDelegateNil
	}
	return t.Delegate.UpdateTeamMemberSelectedAt(ctx, teamId, userId)
}

var _ DbTeamMemberStoreInterface = (*TeamMemberStoreDecorator)(nil)
