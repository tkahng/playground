package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type MockAuthStore struct {
	mock.Mock
}

// WithTx implements AuthStore.
func (m *MockAuthStore) WithTx(dbx database.Dbx) AuthStore {
	args := m.Called(dbx)
	if args.Get(0) != nil {
		return args.Get(0).(AuthStore)
	}
	return nil
}

// RunInTransaction implements AuthStore.
func (m *MockAuthStore) RunInTransaction(ctx context.Context, fn func(store AuthStore) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

var _ AuthStore = (*MockAuthStore)(nil)

// AssignUserRoles implements AuthStorage.
func (m *MockAuthStore) AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error {
	args := m.Called(ctx, userId, roleNames)
	return args.Error(0)
}

// CreateUser implements AuthStorage.
func (m *MockAuthStore) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	var createdUser *models.User
	if args.Get(0) != nil {
		createdUser = args.Get(0).(*models.User)
	}
	return createdUser, args.Error(1)
}

// DeleteToken implements AuthStorage.
func (m *MockAuthStore) DeleteToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// DeleteUser implements AuthStorage.
func (m *MockAuthStore) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// FindUserAccountByUserIdAndProvider implements AuthStorage.
func (m *MockAuthStore) FindUserAccountByUserIdAndProvider(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error) {
	args := m.Called(ctx, userId, provider)
	var userAccount *models.UserAccount
	if args.Get(0) != nil {
		userAccount = args.Get(0).(*models.UserAccount)
	}

	return userAccount, args.Error(1)
}

// FindUserByEmail implements AuthStorage.
func (m *MockAuthStore) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	var user *models.User
	if args.Get(0) != nil {
		user = args.Get(0).(*models.User)
	}
	return user, args.Error(1)
}

// GetToken implements AuthStorage.
func (m *MockAuthStore) GetToken(ctx context.Context, token string) (*models.Token, error) {
	args := m.Called(ctx, token)
	var tokenModel *models.Token
	if args.Get(0) != nil {
		tokenModel = args.Get(0).(*models.Token)
	}
	return tokenModel, args.Error(1)

}

// LinkAccount implements AuthStorage.
func (m *MockAuthStore) LinkAccount(ctx context.Context, account *models.UserAccount) (*models.UserAccount, error) {
	args := m.Called(ctx, account)
	var linkedAccount *models.UserAccount
	if args.Get(0) != nil {
		linkedAccount = args.Get(0).(*models.UserAccount)
	}
	return linkedAccount, args.Error(1)
}

// SaveToken implements AuthStorage.
func (m *MockAuthStore) SaveToken(ctx context.Context, token *shared.CreateTokenDTO) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// UnlinkAccount implements AuthStorage.
func (m *MockAuthStore) UnlinkAccount(ctx context.Context, userId uuid.UUID, provider models.Providers) error {
	args := m.Called(ctx, userId, provider)
	return args.Error(0)
}

// UpdateUser implements AuthStorage.
func (m *MockAuthStore) UpdateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// UpdateUserAccount implements AuthStorage.
func (m *MockAuthStore) UpdateUserAccount(ctx context.Context, account *models.UserAccount) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *MockAuthStore) VerifyTokenStorage(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockAuthStore) GetUserInfo(ctx context.Context, email string) (*shared.UserInfo, error) {
	args := m.Called(ctx, email)
	var userInfo *shared.UserInfo
	if args.Get(0) != nil {
		userInfo = args.Get(0).(*shared.UserInfo)
	}
	return userInfo, args.Error(1)
}
