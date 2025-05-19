package services

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

func TestHandleRefreshToken(t *testing.T) {
	ctx := context.Background()
	mockStorage := new(mockAuthStore)
	mockToken := new(mockJwtService)
	app := &authService{
		token:     mockToken,
		authStore: mockStorage,
		options: &conf.AppOptions{
			Auth: conf.AuthOptions{
				RefreshToken: conf.TokenOption{
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
				mockStorage.On("GetToken", ctx, mock.Anything).Return(&models.Token{
					Type:  models.TokenTypesRefreshToken,
					Token: "valid.token.here",
				}, nil)
				mockStorage.On("DeleteToken", ctx, mock.Anything).Return(nil)
				mockToken.On("ParseToken", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockToken.On("CreateJwtToken", mock.Anything, mock.Anything).Return("new.valid.token.here", nil)
				// mockStorage.On("VerifyTokenStorage", ctx, mock.Anything).Return(nil)
				mockStorage.On("GetUserInfo", ctx, mock.Anything).Return(&shared.UserInfo{
					User: shared.User{
						ID:    uuid.New(),
						Email: "test@example.com",
					},
				}, nil)
				mockStorage.On("SaveToken", ctx, mock.Anything).Return(nil)
			},
			expectedError: false,
		},
		{
			name:  "invalid token format",
			token: "invalid-token",
			setupMocks: func() {
				mockToken.On("ParseToken", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("invalid token"))
			},
			expectedError: true,
		},
		{
			name:  "token verification fails",
			token: "valid.token.here",
			setupMocks: func() {
				mockToken.On("ParseToken", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("invalid token"))
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockToken.ExpectedCalls = nil
			mockToken.Calls = nil
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

			mockToken.AssertExpectations(t)
			mockStorage.AssertExpectations(t)
		})
	}
}
func TestResetPassword(t *testing.T) {
	ctx := context.Background()
	mockStorage := new(mockAuthStore)
	passwordManager := NewPasswordService()
	app := &authService{
		authStore: mockStorage,
		password:  passwordManager,
	}

	testUserId := uuid.New()
	testOldPassword := "oldPassword123"
	testNewPassword := "newPassword123"
	testHashedPassword, err := passwordManager.HashPassword(testOldPassword)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	// Simulate the hashed password for the new password
	// In a real scenario, you would hash the new password as well
	// but here we are just checking the flow
	// so we can use a dummy value
	// for the new hashed password
	testNewHashedPassword, err := passwordManager.HashPassword(testNewPassword)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	fmt.Println("testHashedPassword", testHashedPassword)
	fmt.Println("testNewHashedPassword", testNewHashedPassword)

	testCases := []struct {
		name          string
		userId        uuid.UUID
		oldPassword   string
		newPassword   string
		setupMocks    func()
		expectedError bool
	}{
		{
			name:        "successful password reset",
			userId:      testUserId,
			oldPassword: testOldPassword,
			newPassword: testNewPassword,
			setupMocks: func() {
				mockStorage.On("FindUserAccountByUserIdAndProvider", ctx, testUserId, models.ProvidersCredentials).
					Return(&models.UserAccount{
						Password: &testHashedPassword,
						UserID:   testUserId,
						Provider: models.ProvidersCredentials,
						Type:     models.ProviderTypeCredentials,
					}, nil)
				mockStorage.On("UpdateUserAccount", ctx, mock.Anything).Return(nil)
			},
			expectedError: false,
		},
		{
			name:        "user account not found",
			userId:      testUserId,
			oldPassword: testOldPassword,
			newPassword: testNewPassword,
			setupMocks: func() {
				mockStorage.On("FindUserAccountByUserIdAndProvider", ctx, testUserId, models.ProvidersCredentials).
					Return(nil, nil)
			},
			expectedError: true,
		},
		{
			name:        "incorrect old password",
			userId:      testUserId,
			oldPassword: "wrongPassword",
			newPassword: testNewPassword,
			setupMocks: func() {
				mockStorage.On("FindUserAccountByUserIdAndProvider", ctx, testUserId, models.ProvidersCredentials).
					Return(&models.UserAccount{Password: &testHashedPassword}, nil)
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

			err := app.ResetPassword(ctx, tc.userId, tc.oldPassword, tc.newPassword)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}
