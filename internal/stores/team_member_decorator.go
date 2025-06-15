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
	FindTeamMemberFunc                func(ctx context.Context, filter *TeamMemberFilter) (*models.TeamMember, error)
	FindTeamMemberByTeamAndUserIdFunc func(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) (*models.TeamMember, error)
	FindTeamMembersByUserIDFunc       func(ctx context.Context, userId uuid.UUID, paginate *shared.TeamMemberListInput) ([]*models.TeamMember, error)
	UpdateTeamMemberFunc              func(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error)
	UpdateTeamMemberSelectedAtFunc    func(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error
	FindTeamMembersFunc               func(ctx context.Context, filter *TeamMemberFilter) ([]*models.TeamMember, error)

	// Add any additional methods or fields for the decorator here
}

// LoadTeamMembersByUserAndTeamIds implements DbTeamMemberStoreInterface.
func (t *TeamMemberStoreDecorator) LoadTeamMembersByUserAndTeamIds(ctx context.Context, userId uuid.UUID, teamIds ...uuid.UUID) ([]*models.TeamMember, error) {
	panic("unimplemented")
}

// FindTeamMembers implements DbTeamMemberStoreInterface.
func (t *TeamMemberStoreDecorator) FindTeamMembers(ctx context.Context, filter *TeamMemberFilter) ([]*models.TeamMember, error) {
	if t.FindTeamMembersFunc != nil {
		return t.FindTeamMembersFunc(ctx, filter)
	}
	if t.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return t.Delegate.FindTeamMembers(ctx, filter)
}

func (t *TeamMemberStoreDecorator) Cleanup() {
	t.CountOwnerTeamMembersFunc = nil
	t.CountTeamMembersFunc = nil
	t.CountTeamMembersByUserIDFunc = nil
	t.CreateTeamFromUserFunc = nil
	t.CreateTeamMemberFunc = nil
	t.DeleteTeamMemberFunc = nil
	t.FindLatestTeamMemberByUserIDFunc = nil
	t.FindTeamMemberFunc = nil
	t.FindTeamMemberByTeamAndUserIdFunc = nil
	t.FindTeamMembersByUserIDFunc = nil
	t.UpdateTeamMemberFunc = nil
	t.UpdateTeamMemberSelectedAtFunc = nil

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
