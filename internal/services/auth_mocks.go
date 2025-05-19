package services

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"

	"github.com/stretchr/testify/mock"
)

type mockJwtService struct {
	mock.Mock
}

var _ JwtService = (*mockJwtService)(nil)

// CreateJwtToken implements TokenManager.
func (m *mockJwtService) CreateJwtToken(payload jwt.Claims, signingKey string) (string, error) {
	args := m.Called(payload, signingKey)
	return args.String(0), args.Error(1)
}

// ParseToken implements TokenManager.
func (m *mockJwtService) ParseToken(token string, config conf.TokenOption, data any) error {
	args := m.Called(token, config, data)
	return args.Error(0)
}

type mockAuthStore struct {
	mock.Mock
}

var _ AuthStore = (*mockAuthStore)(nil)

// AssignUserRoles implements AuthStorage.
func (m *mockAuthStore) AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error {
	args := m.Called(ctx, userId, roleNames)
	return args.Error(0)
}

// CreateUser implements AuthStorage.
func (m *mockAuthStore) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(*models.User), args.Error(1)
}

// DeleteToken implements AuthStorage.
func (m *mockAuthStore) DeleteToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// DeleteUser implements AuthStorage.
func (m *mockAuthStore) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// FindUserAccountByUserIdAndProvider implements AuthStorage.
func (m *mockAuthStore) FindUserAccountByUserIdAndProvider(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error) {
	args := m.Called(ctx, userId, provider)
	var userAccount *models.UserAccount
	if args.Get(0) != nil {
		userAccount = args.Get(0).(*models.UserAccount)
	}

	return userAccount, args.Error(1)
}

// FindUserByEmail implements AuthStorage.
func (m *mockAuthStore) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*models.User), args.Error(1)
}

// GetToken implements AuthStorage.
func (m *mockAuthStore) GetToken(ctx context.Context, token string) (*models.Token, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*models.Token), args.Error(1)
}

// LinkAccount implements AuthStorage.
func (m *mockAuthStore) LinkAccount(ctx context.Context, account *models.UserAccount) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

// SaveToken implements AuthStorage.
func (m *mockAuthStore) SaveToken(ctx context.Context, token *shared.CreateTokenDTO) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// UnlinkAccount implements AuthStorage.
func (m *mockAuthStore) UnlinkAccount(ctx context.Context, userId uuid.UUID, provider models.Providers) error {
	args := m.Called(ctx, userId, provider)
	return args.Error(0)
}

// UpdateUser implements AuthStorage.
func (m *mockAuthStore) UpdateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// UpdateUserAccount implements AuthStorage.
func (m *mockAuthStore) UpdateUserAccount(ctx context.Context, account *models.UserAccount) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *mockAuthStore) VerifyTokenStorage(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *mockAuthStore) GetUserInfo(ctx context.Context, email string) (*shared.UserInfo, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*shared.UserInfo), args.Error(1)
}

type mockPasswordService struct {
	mock.Mock
}

// HashPassword implements PasswordManager.
func (m *mockPasswordService) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

// VerifyPassword implements PasswordManager.
func (m *mockPasswordService) VerifyPassword(hashedPassword string, password string) (match bool, err error) {
	args := m.Called(hashedPassword, password)
	return args.Bool(0), args.Error(1)
}

var _ PasswordService = (*mockPasswordService)(nil)
