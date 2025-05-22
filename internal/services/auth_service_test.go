package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mailer"
)

func TestHandleRefreshToken(t *testing.T) {
	ctx := context.Background()
	mockStorage := new(MockAuthStore)
	mockToken := new(MockJwtService)
	app := &BaseAuthService{
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
	mockStorage := new(MockAuthStore)
	passwordManager := NewPasswordService()
	app := &BaseAuthService{
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

func TestAuthenticate(t *testing.T) {
	ctx := context.Background()
	mockStorage := new(MockAuthStore)
	mockToken := NewJwtService()
	mockPassword := new(MockPasswordService)
	mockMailService := &MockMailService{
		delegate: NewMailService(&mailer.LogMailer{}),
	}
	settings := (&conf.EnvConfig{
		AppConfig: conf.AppConfig{
			AppUrl:        "http://localhost:8080",
			AppName:       "TestApp",
			SenderAddress: "tkahng@gmail.com",
		},
	}).ToSettings()
	wg := new(sync.WaitGroup)
	mockRoutineService := new(mockRoutineService)
	mockRoutineService.wg = wg
	app := &BaseAuthService{
		authStore:     mockStorage,
		token:         mockToken,
		password:      mockPassword,
		WorkerService: mockRoutineService,
		mail:          mockMailService,
		options:       settings,
	}

	testUserId := uuid.New()
	testEmail := "test@example.com"
	testPasswordStr := "password123"
	testHashedPassword := "hashedPassword123"

	testCases := []struct {
		name          string
		input         *shared.AuthenticationInput
		setupMocks    func()
		expectedError bool
		checkMail     bool
		checkWant     *mailer.AllEmailParams
	}{
		{
			name: "user does not exist, create user and account",
			input: &shared.AuthenticationInput{
				Email:    testEmail,
				Password: &testPasswordStr,
				Type:     shared.ProviderTypeCredentials,
			},
			setupMocks: func() {
				mockStorage.On("FindUserByEmail", ctx, testEmail).Return(nil, nil)
				mockStorage.On("CreateUser", ctx, mock.Anything).Return(&models.User{ID: testUserId, Email: testEmail}, nil)
				mockStorage.On("AssignUserRoles", ctx, testUserId, mock.Anything).Return(nil)
				mockPassword.On("HashPassword", testPasswordStr).Return(testHashedPassword, nil)
				mockStorage.On("LinkAccount", ctx, mock.Anything).Return(nil)
				mockStorage.On("SaveToken", ctx, mock.Anything).Return(nil)
				// mockStorage.On("FindUserAccountByUserIdAndProvider", ctx, testUserId, models.ProvidersCredentials).Return(nil, nil)
				// mockPassword.On("HashPassword", mock.Anything).Return(testHashedPassword, nil)
				// mockPassword.On("UpdateUserAccount", ctx, mock.Anything).Return(nil)
				// mockStorage.On("SaveToken", ctx, mock.Anything).Return(nil)
			},
			expectedError: false,
			checkMail:     true,
			checkWant: &mailer.AllEmailParams{
				SendMailParams: &mailer.SendMailParams{
					Type: string(EmailTypeVerify),
				},
				Message: &mailer.Message{
					From:    settings.Meta.SenderAddress,
					To:      testEmail,
					Subject: "TestApp - Verify your email address",
				},
			},
		},
		{
			name: "user exists, account exists, correct password",
			input: &shared.AuthenticationInput{
				Email:    testEmail,
				Password: &testPasswordStr,
				Type:     shared.ProviderTypeCredentials,
			},
			setupMocks: func() {
				mockStorage.On("FindUserByEmail", ctx, testEmail).Return(&models.User{ID: testUserId, Email: testEmail}, nil)
				mockStorage.On("FindUserAccountByUserIdAndProvider", ctx, testUserId, models.ProvidersCredentials).
					Return(&models.UserAccount{Password: &testHashedPassword}, nil)
				mockPassword.On("VerifyPassword", testHashedPassword, testPasswordStr).Return(true, nil)
			},
			expectedError: false,
		},
		{
			name: "user exists, account exists, incorrect password",
			input: &shared.AuthenticationInput{
				Email:    testEmail,
				Password: &testPasswordStr,
				Type:     shared.ProviderTypeCredentials,
			},
			setupMocks: func() {
				mockStorage.On("FindUserByEmail", ctx, testEmail).Return(&models.User{ID: testUserId, Email: testEmail}, nil)
				mockStorage.On("FindUserAccountByUserIdAndProvider", ctx, testUserId, models.ProvidersCredentials).
					Return(&models.UserAccount{Password: &testHashedPassword}, nil)
				mockPassword.On("VerifyPassword", testHashedPassword, testPasswordStr).Return(false, nil)
			},
			expectedError: true,
		},
		{
			name: "user exists, account does not exist, create account",
			input: &shared.AuthenticationInput{
				Email:    testEmail,
				Password: &testPasswordStr,
				Type:     shared.ProviderTypeCredentials,
			},
			setupMocks: func() {
				mockStorage.On("FindUserByEmail", ctx, testEmail).Return(&models.User{ID: testUserId, Email: testEmail}, nil)
				mockStorage.On("FindUserAccountByUserIdAndProvider", ctx, testUserId, models.ProvidersCredentials).Return(nil, nil)
				mockPassword.On("HashPassword", testPasswordStr).Return(testHashedPassword, nil)
				mockStorage.On("LinkAccount", ctx, mock.Anything).Return(nil)
				// mockStorage.On("FindUserAccountByUserIdAndProvider", ctx, testUserId, models.ProvidersCredentials).Return(nil)
			},
			expectedError: false,
			// checkMail:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStorage.ExpectedCalls = nil
			mockStorage.Calls = nil
			mockPassword.ExpectedCalls = nil
			mockPassword.Calls = nil
			mockMailService.param = nil

			if tc.setupMocks != nil {
				tc.setupMocks()
			}

			result, err := app.Authenticate(ctx, tc.input)
			wg.Wait()
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
			if tc.checkMail {
				param := mockMailService.param
				assert.NotNil(t, param)
				assert.Equal(t, param.Message.To, testEmail)
				assert.Equal(t, param.SendMailParams.Type, tc.checkWant.SendMailParams.Type)
			}

			mockStorage.AssertExpectations(t)
			mockPassword.AssertExpectations(t)
		})
	}
}
