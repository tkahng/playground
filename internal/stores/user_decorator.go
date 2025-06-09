package stores

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type UserStoreDecorator struct {
	Delegate               *DbUserStore
	WithTxFunc             func(dbx database.Dbx) *UserStoreDecorator
	AssignUserRolesFunc    func(ctx context.Context, userId uuid.UUID, roleNames ...string) error
	CreateUserFunc         func(ctx context.Context, user *models.User) (*models.User, error)
	DeleteUserFunc         func(ctx context.Context, userId uuid.UUID) error
	FindUserFunc           func(ctx context.Context, user *UserFilter) (*models.User, error)
	FindUserByIDFunc       func(ctx context.Context, userId uuid.UUID) (*models.User, error)
	GetUserInfoFunc        func(ctx context.Context, email string) (*shared.UserInfo, error)
	LoadUsersByUserIdsFunc func(ctx context.Context, userIds ...uuid.UUID) ([]*models.User, error)
	UpdateUserFunc         func(ctx context.Context, user *models.User) error
	UserWhereFunc          func(user *models.User) *map[string]any
	FindUsersFunc          func(ctx context.Context, filter *UserFilter) ([]*models.User, error)
	CountUsersFunc         func(ctx context.Context, filter *UserFilter) (int64, error)
}

// CountUsers implements DbUserStoreInterface.
func (u *UserStoreDecorator) CountUsers(ctx context.Context, filter *UserFilter) (int64, error) {
	if u.CountUsersFunc != nil {
		return u.CountUsersFunc(ctx, filter)
	}
	if u.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return u.Delegate.CountUsers(ctx, filter)
}

// FindUsers implements DbUserStoreInterface.
func (u *UserStoreDecorator) FindUsers(ctx context.Context, filter *UserFilter) ([]*models.User, error) {
	if u.FindUsersFunc != nil {
		return u.FindUsersFunc(ctx, filter)
	}
	if u.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return u.Delegate.FindUsers(ctx, filter)
}

func (u *UserStoreDecorator) Cleanup() {
	u.WithTxFunc = nil
	u.AssignUserRolesFunc = nil
	u.CreateUserFunc = nil
	u.DeleteUserFunc = nil
	u.FindUserFunc = nil
	u.FindUserByIDFunc = nil
	u.GetUserInfoFunc = nil
	u.LoadUsersByUserIdsFunc = nil
	u.UpdateUserFunc = nil
	u.UserWhereFunc = nil
}

var ErrDelegateNil = errors.New("delegate is nil, cannot call method on nil")

// AssignUserRoles implements DbUserStoreInterface.
func (u *UserStoreDecorator) AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error {
	if u.AssignUserRolesFunc != nil {
		return u.AssignUserRolesFunc(ctx, userId, roleNames...)
	}
	if u.Delegate == nil {
		return ErrDelegateNil
	}
	return u.Delegate.AssignUserRoles(ctx, userId, roleNames...)
}

// CreateUser implements DbUserStoreInterface.
func (u *UserStoreDecorator) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	if u.CreateUserFunc != nil {
		return u.CreateUserFunc(ctx, user)
	}
	if u.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return u.Delegate.CreateUser(ctx, user)
}

// DeleteUser implements DbUserStoreInterface.
func (u *UserStoreDecorator) DeleteUser(ctx context.Context, userId uuid.UUID) error {
	if u.DeleteUserFunc != nil {
		return u.DeleteUserFunc(ctx, userId)
	}
	if u.Delegate == nil {
		return ErrDelegateNil
	}
	return u.Delegate.DeleteUser(ctx, userId)
}

// FindUser implements DbUserStoreInterface.
func (u *UserStoreDecorator) FindUser(ctx context.Context, user *UserFilter) (*models.User, error) {
	if u.FindUserFunc != nil {
		return u.FindUserFunc(ctx, user)
	}
	if u.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return u.Delegate.FindUser(ctx, user)
}

// FindUserByID implements DbUserStoreInterface.
func (u *UserStoreDecorator) FindUserByID(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	if u.FindUserByIDFunc != nil {
		return u.FindUserByIDFunc(ctx, userId)
	}
	if u.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return u.Delegate.FindUserByID(ctx, userId)
}

// GetUserInfo implements DbUserStoreInterface.
func (u *UserStoreDecorator) GetUserInfo(ctx context.Context, email string) (*shared.UserInfo, error) {
	if u.GetUserInfoFunc != nil {
		return u.GetUserInfoFunc(ctx, email)
	}
	if u.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return u.Delegate.GetUserInfo(ctx, email)
}

// LoadUsersByUserIds implements DbUserStoreInterface.
func (u *UserStoreDecorator) LoadUsersByUserIds(ctx context.Context, userIds ...uuid.UUID) ([]*models.User, error) {
	if u.LoadUsersByUserIdsFunc != nil {
		return u.LoadUsersByUserIdsFunc(ctx, userIds...)
	}
	if u.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return u.Delegate.LoadUsersByUserIds(ctx, userIds...)
}

// UpdateUser implements DbUserStoreInterface.
func (u *UserStoreDecorator) UpdateUser(ctx context.Context, user *models.User) error {
	if u.UpdateUserFunc != nil {
		return u.UpdateUserFunc(ctx, user)
	}
	if u.Delegate == nil {
		return ErrDelegateNil
	}
	return u.Delegate.UpdateUser(ctx, user)
}

func (u *UserStoreDecorator) WithTx(dbx database.Dbx) *UserStoreDecorator {
	if u.WithTxFunc != nil {
		return u.WithTxFunc(dbx)
	}
	if u.Delegate == nil {
		panic(ErrDelegateNil)
	}
	return &UserStoreDecorator{
		Delegate: u.Delegate.WithTx(dbx),
	}
}

var _ DbUserStoreInterface = (*UserStoreDecorator)(nil)
