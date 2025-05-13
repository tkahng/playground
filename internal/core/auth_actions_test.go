package core

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tkahng/authgo/internal/shared"
)

type mockAuthStorage struct {
	mock.Mock
}

// AssignUserRoles implements AuthStorage.
func (m *mockAuthStorage) AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error {
	args := m.Called(ctx, userId, roleNames)
	return args.Error(0)
}

// CreateUser implements AuthStorage.
func (m *mockAuthStorage) CreateUser(ctx context.Context, user *shared.User) (*shared.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(*shared.User), args.Error(1)
}

// DeleteToken implements AuthStorage.
func (m *mockAuthStorage) DeleteToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// DeleteUser implements AuthStorage.
func (m *mockAuthStorage) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// FindUserAccountByUserIdAndProvider implements AuthStorage.
func (m *mockAuthStorage) FindUserAccountByUserIdAndProvider(ctx context.Context, userId uuid.UUID, provider shared.Providers) (*shared.UserAccount, error) {
	args := m.Called(ctx, userId, provider)
	return args.Get(0).(*shared.UserAccount), args.Error(1)
}

// FindUserByEmail implements AuthStorage.
func (m *mockAuthStorage) FindUserByEmail(ctx context.Context, email string) (*shared.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*shared.User), args.Error(1)
}

// GetToken implements AuthStorage.
func (m *mockAuthStorage) GetToken(ctx context.Context, token string) (*shared.Token, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*shared.Token), args.Error(1)
}

// LinkAccount implements AuthStorage.
func (m *mockAuthStorage) LinkAccount(ctx context.Context, account *shared.UserAccount) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

// SaveToken implements AuthStorage.
func (m *mockAuthStorage) SaveToken(ctx context.Context, token *shared.CreateTokenDTO) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// UnlinkAccount implements AuthStorage.
func (m *mockAuthStorage) UnlinkAccount(ctx context.Context, userId uuid.UUID, provider shared.Providers) error {
	args := m.Called(ctx, userId, provider)
	return args.Error(0)
}

// UpdateUser implements AuthStorage.
func (m *mockAuthStorage) UpdateUser(ctx context.Context, user *shared.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// UpdateUserAccount implements AuthStorage.
func (m *mockAuthStorage) UpdateUserAccount(ctx context.Context, account *shared.UserAccount) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

var _ AuthStorage = (*mockAuthStorage)(nil)

func (m *mockAuthStorage) VerifyTokenStorage(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *mockAuthStorage) GetUserInfo(ctx context.Context, email string) (*shared.UserInfo, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*shared.UserInfo), args.Error(1)
}

func TestHandleRefreshToken(t *testing.T) {
	ctx := context.Background()
	mockStorage := new(mockAuthStorage)

	app := &AuthActionsBase{
		storage: mockStorage,
		options: &AppOptions{
			Auth: AuthOptions{
				RefreshToken: TokenOption{
					Type:     shared.TokenTypesRefreshToken,
					Secret:   string(shared.TokenTypesRefreshToken),
					Duration: 604800, // 7days
				},
			},
		},
	}

	testCases := []struct {
		name          string
		token         string
		setupMocks    func()
		expectedError bool
	}{
		{
			name:  "valid refresh token",
			token: "valid.token.here",
			setupMocks: func() {
				mockStorage.On("VerifyTokenStorage", ctx, mock.Anything).Return(nil)
				mockStorage.On("GetUserInfo", ctx, mock.Anything).Return(&shared.UserInfo{
					User: shared.User{
						ID:    uuid.New(),
						Email: "test@example.com",
					},
				}, nil)
			},
			expectedError: false,
		},
		{
			name:  "invalid token format",
			token: "invalid-token",
			setupMocks: func() {
				// mockStorage.On("VerifyTokenStorage", ctx, mock.Anything).Return(jwt.ErrSignatureInvalid)
			},
			expectedError: true,
		},
		{
			name:  "token verification fails",
			token: "valid.token.here",
			setupMocks: func() {
				mockStorage.On("VerifyTokenStorage", ctx, mock.Anything).Return(jwt.ErrSignatureInvalid)
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStorage.ExpectedCalls = nil
			mockStorage.Calls = nil

			if tc.setupMocks != nil {
				tc.setupMocks()
			}

			result, err := app.HandleRefreshToken(ctx, tc.token)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}
