package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/mailer"
	"github.com/tkahng/authgo/internal/tools/types"
)

func TestHandleRefreshToken(t *testing.T) {
	ctx := context.Background()
	// adapter := new(MockAuthStore)
	adapter := stores.NewAdapterDecorators()
	// adatper := resource.NewResourceDecoratorAdapter()
	mockToken := NewJwtServiceDecorator()
	app := &BaseAuthService{
		token:   mockToken,
		adapter: adapter,
		// adapter:   adatper,
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
				adapter.TokenFunc.Cleanup()
				mockToken.Cleanup()
				adapter.TokenFunc.GetTokenFunc = func(ctx context.Context, token string) (*models.Token, error) {
					return &models.Token{
						Type:  models.TokenTypesRefreshToken,
						Token: "valid.token.here",
					}, nil
				}
				adapter.TokenFunc.DeleteTokenFunc = func(ctx context.Context, token string) error {
					return nil
				}
				mockToken.ParseTokenFunc = func(token string, config conf.TokenOption, data any) error {
					return nil
				}
				// mockToken.On("ParseToken", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockToken.CreateJwtTokenFunc = func(payload jwt.Claims, signingKey string) (string, error) {
					return "new.valid.token.here", nil
				}

				adapter.UserFunc.GetUserInfoFunc = func(ctx context.Context, email string) (*shared.UserInfo, error) {
					return &shared.UserInfo{
						User: shared.User{
							ID:    uuid.New(),
							Email: "test@example.com",
						},
					}, nil
				}
				adapter.TokenFunc.SaveTokenFunc = func(ctx context.Context, token *shared.CreateTokenDTO) error {
					return nil
				}
			},
			expectedError: false,
		},
		{
			name:  "invalid token format",
			token: "invalid-token",
			setupMocks: func() {
				mockToken.Cleanup()
				mockToken.ParseTokenFunc = func(token string, config conf.TokenOption, data any) error {
					return errors.New("invalid token format")
				}
			},
			expectedError: true,
		},
		{
			name:  "token verification fails",
			token: "valid.token.here",
			setupMocks: func() {
				mockToken.Cleanup()
				mockToken.ParseTokenFunc = func(token string, config conf.TokenOption, data any) error {
					return errors.New("token verification failed")
				}
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

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

			// mockStorage.AssertExpectations(t)
		})
	}
}
func TestResetPassword(t *testing.T) {
	ctx := context.Background()
	storageDecorators := stores.NewAdapterDecorators()
	// adapter := resource.NewResourceDecoratorAdapter()
	// mockStorage := new(MockAuthStore)
	passwordManager := NewPasswordService()
	app := &BaseAuthService{
		adapter:  storageDecorators,
		password: passwordManager,
		// adapter:   adapter,
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
				storageDecorators.Cleanup()
				storageDecorators.UserAccountFunc.FindUserAccountByUserIdAndProviderFunc = func(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error) {
					return &models.UserAccount{
						Password: &testHashedPassword,
						UserID:   userId,
						Provider: models.ProvidersCredentials,
					}, nil
				}
				// mockStorage.On("FindUserAccountByUserIdAndProvider", ctx, testUserId, models.ProvidersCredentials).
				// 	Return(&models.UserAccount{
				// 		Password: &testHashedPassword,
				// 		UserID:   testUserId,
				// 		Provider: models.ProvidersCredentials,
				// 		Type:     models.ProviderTypeCredentials,
				// 	}, nil)
				storageDecorators.UserAccountFunc.UpdateUserAccountFunc = func(ctx context.Context, account *models.UserAccount) error {
					return nil
				}
				// mockStorage.On("UpdateUserAccount", ctx, mock.Anything).Return(nil)
			},
			expectedError: false,
		},
		{
			name:        "user account not found",
			userId:      testUserId,
			oldPassword: testOldPassword,
			newPassword: testNewPassword,
			setupMocks: func() {
				storageDecorators.Cleanup()
				storageDecorators.UserAccountFunc.FindUserAccountByUserIdAndProviderFunc = func(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error) {
					return nil, nil
				}
				// mockStorage.On("FindUserAccountByUserIdAndProvider", ctx, testUserId, models.ProvidersCredentials).
				// 	Return(nil, nil)
			},
			expectedError: true,
		},
		{
			name:        "incorrect old password",
			userId:      testUserId,
			oldPassword: "wrongPassword",
			newPassword: testNewPassword,
			setupMocks: func() {
				storageDecorators.Cleanup()
				storageDecorators.UserAccountFunc.FindUserAccountByUserIdAndProviderFunc = func(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error) {
					return &models.UserAccount{Password: &testHashedPassword}, nil
				}
				// mockStorage.On("FindUserAccountByUserIdAndProvider", ctx, testUserId, models.ProvidersCredentials).
				// 	Return(&models.UserAccount{Password: &testHashedPassword}, nil)
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// mockStorage.ExpectedCalls = nil
			// mockStorage.Calls = nil

			if tc.setupMocks != nil {
				tc.setupMocks()
			}

			err := app.ResetPassword(ctx, tc.userId, tc.oldPassword, tc.newPassword)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// mockStorage.AssertExpectations(t)
		})
	}
}

func TestAuthenticate(t *testing.T) {
	ctx := context.Background()
	storeDecorator := stores.NewAdapterDecorators()
	// adapter := resource.NewResourceDecoratorAdapter()
	mockToken := NewJwtService()
	mockPassword := NewPasswordServiceDecorator()
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
	mockRoutineService := &MockRoutineService{
		Wg: wg,
	}
	app := &BaseAuthService{
		// adapter:   adapter,
		adapter:  storeDecorator,
		token:    mockToken,
		password: mockPassword,
		routine:  mockRoutineService,
		mail:     mockMailService,
		options:  settings,
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
				storeDecorator.Cleanup()
				mockPassword.Cleanup()
				storeDecorator.UserFunc.FindUserFunc = func(ctx context.Context, user *stores.UserFilter) (*models.User, error) {
					return nil, nil // Simulate user not found
				}
				// mockStorage.On("FindUser", ctx, mock.Anything).Return(nil, nil)

				storeDecorator.UserFunc.CreateUserFunc = func(ctx context.Context, user *models.User) (*models.User, error) {
					return &models.User{ID: testUserId, Email: testEmail}, nil // Simulate user creation
				}
				// mockStorage.On("CreateUser", ctx, mock.Anything).Return(&models.User{ID: testUserId, Email: testEmail}, nil)
				// mockPassword.On("HashPassword", testPasswordStr).Return(testHashedPassword, nil)
				mockPassword.HashPasswordFunc = func(password string) (string, error) {
					return testHashedPassword, nil // Simulate password hashing
				}
				storeDecorator.UserAccountFunc.CreateUserAccountFunc = func(ctx context.Context, account *models.UserAccount) (*models.UserAccount, error) {
					return &models.UserAccount{
						Password: &testHashedPassword,
						UserID:   testUserId,
						Provider: models.ProvidersCredentials,
						Type:     models.ProviderTypeCredentials,
					}, nil // Simulate account creation
				}
				// mockStorage.On("CreateUserAccount", ctx, mock.Anything).Return(&models.UserAccount{
				// 	Password: &testHashedPassword,
				// 	UserID:   testUserId,
				// 	Provider: models.ProvidersCredentials,
				// 	Type:     models.ProviderTypeCredentials,
				// }, nil)
				storeDecorator.TokenFunc.SaveTokenFunc = func(ctx context.Context, token *shared.CreateTokenDTO) error {
					return nil // Simulate token saving
				}
				// mockStorage.On("SaveToken", ctx, mock.Anything).Return(nil)

			},
			expectedError: false,
			checkMail:     true,
			checkWant: &mailer.AllEmailParams{
				SendMailParams: &mailer.SendMailParams{
					Type: string(mailer.EmailTypeVerify),
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
				storeDecorator.Cleanup()
				mockPassword.Cleanup()
				storeDecorator.UserFunc.FindUserFunc = func(ctx context.Context, user *stores.UserFilter) (*models.User, error) {
					return &models.User{ID: testUserId, Email: testEmail}, nil // Simulate user found
				}
				// mockStorage.On("FindUser", ctx, mock.Anything).Return(&models.User{ID: testUserId, Email: testEmail}, nil)
				storeDecorator.UserAccountFunc.FindUserAccountByUserIdAndProviderFunc = func(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error) {
					return &models.UserAccount{Password: &testHashedPassword}, nil // Simulate account found
				}
				// mockStorage.On("FindUserAccountByUserIdAndProvider", ctx, testUserId, models.ProvidersCredentials).
				// 	Return(&models.UserAccount{Password: &testHashedPassword}, nil)
				mockPassword.VerifyPasswordFunc = func(hashedPassword string, password string) (bool, error) {
					return true, nil // Simulate password verification success
				}
				// mockPassword.On("VerifyPassword", testHashedPassword, testPasswordStr).Return(true, nil)
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
				storeDecorator.Cleanup()
				mockPassword.Cleanup()
				storeDecorator.UserFunc.FindUserFunc = func(ctx context.Context, user *stores.UserFilter) (*models.User, error) {
					return &models.User{ID: testUserId, Email: testEmail}, nil // Simulate user found
				}
				// mockStorage.On("FindUser", ctx, mock.Anything).Return(&models.User{ID: testUserId, Email: testEmail}, nil)
				storeDecorator.UserAccountFunc.FindUserAccountByUserIdAndProviderFunc = func(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error) {
					return &models.UserAccount{Password: &testHashedPassword}, nil // Simulate account found
				}
				// mockStorage.On("FindUserAccountByUserIdAndProvider", ctx, testUserId, models.ProvidersCredentials).
				// 	Return(&models.UserAccount{Password: &testHashedPassword}, nil)
				mockPassword.VerifyPasswordFunc = func(hashedPassword string, password string) (bool, error) {
					return false, nil // Simulate password verification failure
				}
				// mockPassword.On("VerifyPassword", testHashedPassword, testPasswordStr).Return(false, nil)
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
				storeDecorator.Cleanup()
				mockPassword.Cleanup()
				storeDecorator.UserFunc.FindUserFunc = func(ctx context.Context, user *stores.UserFilter) (*models.User, error) {
					return &models.User{ID: testUserId, Email: testEmail}, nil // Simulate user found
				}
				// mockStorage.On("FindUser", ctx, mock.Anything).Return(&models.User{ID: testUserId, Email: testEmail}, nil)
				storeDecorator.UserAccountFunc.FindUserAccountByUserIdAndProviderFunc = func(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error) {
					return nil, nil // Simulate account not found
				}
				// mockStorage.On("FindUserAccountByUserIdAndProvider", ctx, testUserId, models.ProvidersCredentials).Return(nil, nil)
				mockPassword.HashPasswordFunc = func(password string) (string, error) {
					return testHashedPassword, nil // Simulate password hashing
				}
				// mockPassword.On("HashPassword", testPasswordStr).Return(testHashedPassword, nil)
				storeDecorator.UserAccountFunc.CreateUserAccountFunc = func(ctx context.Context, account *models.UserAccount) (*models.UserAccount, error) {
					return &models.UserAccount{
						Password: &testHashedPassword,
						UserID:   testUserId,
						Provider: models.ProvidersCredentials,
						Type:     models.ProviderTypeCredentials,
					}, nil // Simulate account creation
				}
				// mockStorage.On("CreateUserAccount", ctx, mock.Anything).Return(&models.UserAccount{
				// 	Password: &testHashedPassword,
				// 	UserID:   testUserId,
				// 	Provider: models.ProvidersCredentials,
				// 	Type:     models.ProviderTypeCredentials,
				// }, nil)
			},
			expectedError: false,
		},
		{
			name: "user exists, account does not exist, create account password reset",
			input: &shared.AuthenticationInput{
				Email:           testEmail,
				Provider:        shared.Providers(models.ProvidersGoogle),
				Type:            shared.ProviderTypeOAuth,
				EmailVerifiedAt: types.Pointer(time.Now()),
			},
			setupMocks: func() {
				storeDecorator.Cleanup()
				// mockStorage.On("FindUser", ctx, mock.Anything).Return(&models.User{ID: testUserId, Email: testEmail}, nil)
				// mockStorage.On("FindUserAccountByUserIdAndProvider", ctx, testUserId, mock.Anything).Return(nil, nil)

				storeDecorator.UserFunc.FindUserFunc = func(ctx context.Context, user *stores.UserFilter) (*models.User, error) {
					return &models.User{ID: testUserId, Email: testEmail}, nil // Simulate user found
				}
				storeDecorator.UserAccountFunc.FindUserAccountByUserIdAndProviderFunc = func(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error) {
					return nil, nil // Simulate account not found
				}
				storeDecorator.UserAccountFunc.CreateUserAccountFunc = func(ctx context.Context, account *models.UserAccount) (*models.UserAccount, error) {
					return &models.UserAccount{
						UserID:   testUserId,
						Provider: models.ProvidersGoogle,
						Type:     models.ProviderTypeOAuth,
					}, nil // Simulate account creation
				}
				// mockStorage.On("CreateUserAccount", ctx, mock.Anything).Return(&models.UserAccount{
				// 	UserID:   testUserId,
				// 	Provider: models.ProvidersGoogle,
				// 	Type:     models.ProviderTypeOAuth,
				// }, nil)
				storeDecorator.UserFunc.UpdateUserFunc = func(ctx context.Context, user *models.User) error {
					return nil // Simulate user update
				}
				// mockStorage.On("UpdateUser", ctx, mock.Anything).Return(nil)
				storeDecorator.TokenFunc.SaveTokenFunc = func(ctx context.Context, token *shared.CreateTokenDTO) error {
					return nil // Simulate token saving
				}
				// mockStorage.On("SaveToken", ctx, mock.Anything).Return(nil)
			},
			expectedError: false,
			checkMail:     false,
			checkWant: &mailer.AllEmailParams{
				SendMailParams: &mailer.SendMailParams{
					Type: string(mailer.EmailTypeSecurityPasswordReset),
				},
				Message: &mailer.Message{
					From: settings.Meta.SenderAddress,
					To:   testEmail,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// mockStorage.ExpectedCalls = nil
			// mockStorage.Calls = nil
			// mockStorage2.ExpectedCalls = nil
			// mockStorage2.Calls = nil
			// mockPassword.ExpectedCalls = nil
			// mockPassword.Calls = nil
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
				assert.Equal(t, param.To, testEmail)
				assert.Equal(t, param.Type, tc.checkWant.Type)
			}

			// mockStorage.AssertExpectations(t)
			// mockPassword.AssertExpectations(t)
		})
	}
}
