package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/jobs"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mailer"
	"github.com/tkahng/authgo/internal/tools/types"
	"github.com/tkahng/authgo/internal/tools/utils"
	"github.com/tkahng/authgo/internal/workers"
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
	dispatcher := jobs.NewDispatchDecorator()

	// Log registered handlers for debugging
	dispatcher.SetHandlerFunc = func(kind string, handler func(context.Context, *models.JobRow) error) {
		fmt.Printf("Handler registered for kind: %s\n", kind)
		dispatcher.Delegate.SetHandler(kind, handler)
	}

	// Log dispatched job kind for debugging
	dispatcher.DispatchFunc = func(ctx context.Context, row *models.JobRow) error {
		fmt.Printf("Dispatching job with kind: %s\n", row.Kind)
		return dispatcher.Delegate.Dispatch(ctx, row)
	}

	enqueuer := jobs.DBEnqueuerDecorator{}
	enqueuer.EnqueueFunc = func(ctx context.Context, args jobs.JobArgs, uniqueKey *string, runAfter time.Time, maxAttempts int) error {
		payload, err := json.Marshal(args)
		if err != nil {
			return fmt.Errorf("marshal args: %w", err)
		}

		// Generate time-ordered UUIDv7 for better database performance
		id, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("generate uuid: %w", err)
		}
		job := models.JobRow{
			ID:          id,
			Kind:        args.Kind(),
			UniqueKey:   uniqueKey,
			Payload:     payload,
			Status:      "pending",
			RunAfter:    runAfter,
			Attempts:    0,
			MaxAttempts: int64(maxAttempts),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		enqueuer.Jobs = append(enqueuer.Jobs, &job)
		return err

	}
	jobStore := jobs.NewJobStoreDecorator()

	jobStore.ClaimPendingJobsFunc = func(ctx context.Context, limit int) ([]models.JobRow, error) {
		if len(enqueuer.Jobs) > 0 {
			jobStore.Job = enqueuer.Jobs[0]
			if len(enqueuer.Jobs) > 0 {
				jobd := enqueuer.Jobs[0]
				if jobd != nil {
					return []models.JobRow{*jobd}, nil
				}
			}
		}
		return []models.JobRow{}, nil
	}

	jobStore.MarkDoneFunc = func(ctx context.Context, id uuid.UUID) error {
		if jobStore.Job != nil {
			jobStore.Job.Status = "done"
		}
		return nil
	}
	jobStore.MarkFailedFunc = func(ctx context.Context, id uuid.UUID, reason string) error {
		if jobStore.Job != nil {
			jobStore.Job.Status = "failed"
		}
		return nil
	}
	jobStore.RescheduleJobFunc = func(ctx context.Context, id uuid.UUID, delay time.Duration) error {
		if jobStore.Job != nil {
			jobStore.Job.Status = "pending"
			jobStore.Job.RunAfter = time.Now().Add(delay)
		}
		return nil
	}
	jobStore.RunInTxFunc = func(ctx context.Context, fn func(js jobs.JobStore) error) error {
		return fn(jobStore)
	}
	poller := jobs.NewPollerDecorator(jobStore, dispatcher)
	mockStorage := new(MockAuthStore)
	mockStorage2 := new(MockAuthStore)
	mockStorage.Tx = mockStorage2
	mockToken := NewJwtService()
	mockPassword := new(MockPasswordService)
	mockMailService := &MockMailService{
		delegate: NewMailService(&mailer.LogMailer{}),
		SendMailOverride: func(params *mailer.AllEmailParams) error {
			bs, _ := json.Marshal(params)
			fmt.Printf("Mock SendMail called with params: %+v\n", string(bs))
			return nil
		},
	}

	settings := conf.NewSettings()
	wg := new(sync.WaitGroup)
	mockRoutineService := &MockRoutineService{
		Wg: wg,
	}
	app := &BaseAuthService{
		authStore: mockStorage,
		token:     mockToken,
		password:  mockPassword,
		routine:   mockRoutineService,
		mail:      mockMailService,
		options:   settings,
		enqueuer:  &enqueuer,
	}
	mailWorker := NewWorkerDecorator(app)
	mailWorker.WorkFunc = func(ctx context.Context, job *jobs.Job[workers.OtpEmailJobArgs]) error {
		fmt.Println("mail worker")
		utils.PrettyPrintJSON(job)
		return nil
	}
	jobs.RegisterWorker(dispatcher, mailWorker)
	testUserId := uuid.New()
	testEmail := "test@example.com"
	testPasswordStr := "password123"
	testHashedPassword := "hashedPassword123"

	testCases := []struct {
		name            string
		input           *shared.AuthenticationInput
		setupMocks      func()
		expectedError   bool
		checkMail       bool
		checkWant       *mailer.AllEmailParams
		checkJob        bool
		checkJobWant    *models.JobRow
		checkJobPayload string
	}{
		{
			name: "user does not exist, create user and account",
			input: &shared.AuthenticationInput{
				Email:    testEmail,
				Password: &testPasswordStr,
				Type:     shared.ProviderTypeCredentials,
			},
			setupMocks: func() {
				mockStorage.On("FindUser", mock.Anything, mock.Anything).Return(nil, nil)
				// mockStorage.On("RunInTransaction", ctx, mock.Anything).Return(nil)
				// mockStorage.On("WithTx", ctx, mock.Anything).Return(mockStorage2)
				mockStorage.On("CreateUser", mock.Anything, mock.Anything).Return(&models.User{ID: testUserId, Email: testEmail}, nil)
				// mockStorage.On("AssignUserRoles", ctx, testUserId, mock.Anything).Return(nil)
				mockPassword.On("HashPassword", testPasswordStr).Return(testHashedPassword, nil)
				mockStorage.On("CreateUserAccount", mock.Anything, mock.Anything).Return(&models.UserAccount{
					Password: &testHashedPassword,
					UserID:   testUserId,
					Provider: models.ProvidersCredentials,
					Type:     models.ProviderTypeCredentials,
				}, nil)
				// mockStorage.On("FindUser", mock.Anything, mock.Anything).Return(&models.User{ID: testUserId, Email: testEmail}, nil)
				// mockStorage.On("SaveToken", mock.Anything, mock.Anything).Return(nil)
				// mockStorage.On("SaveToken", ctx, mock.Anything).Return(nil)
				// mockStorage.On("FindUserAccountByUserIdAndProvider", ctx, testUserId, models.ProvidersCredentials).Return(nil, nil)
				// mockPassword.On("HashPassword", mock.Anything).Return(testHashedPassword, nil)
				// mockPassword.On("UpdateUserAccount", ctx, mock.Anything).Return(nil)
				// mockStorage.On("SaveToken", ctx, mock.Anything).Return(nil)
			},
			expectedError: false,
			checkJob:      true,
			checkJobWant: &models.JobRow{
				Status: "done",
				Kind:   "otp_email",
			},
			checkJobPayload: "verify",
		},
		{
			name: "user exists, account exists, correct password",
			input: &shared.AuthenticationInput{
				Email:    testEmail,
				Password: &testPasswordStr,
				Type:     shared.ProviderTypeCredentials,
			},
			setupMocks: func() {
				mockStorage.On("FindUser", mock.Anything, mock.Anything).Return(&models.User{ID: testUserId, Email: testEmail}, nil)
				mockStorage.On("FindUserAccountByUserIdAndProvider", mock.Anything, testUserId, models.ProvidersCredentials).
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
				mockStorage.On("FindUser", mock.Anything, mock.Anything).Return(&models.User{ID: testUserId, Email: testEmail}, nil)
				mockStorage.On("FindUserAccountByUserIdAndProvider", mock.Anything, testUserId, models.ProvidersCredentials).
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
				mockStorage.On("FindUser", mock.Anything, mock.Anything).Return(&models.User{ID: testUserId, Email: testEmail}, nil)
				mockStorage.On("FindUserAccountByUserIdAndProvider", mock.Anything, testUserId, models.ProvidersCredentials).Return(nil, nil)
				mockPassword.On("HashPassword", testPasswordStr).Return(testHashedPassword, nil)
				mockStorage.On("CreateUserAccount", mock.Anything, mock.Anything).Return(&models.UserAccount{
					Password: &testHashedPassword,
					UserID:   testUserId,
					Provider: models.ProvidersCredentials,
					Type:     models.ProviderTypeCredentials,
				}, nil)
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
				mockStorage.On("FindUser", mock.Anything, mock.Anything).Return(&models.User{ID: testUserId, Email: testEmail}, nil)
				mockStorage.On("FindUserAccountByUserIdAndProvider", mock.Anything, testUserId, mock.Anything).Return(nil, nil)

				mockStorage.On("CreateUserAccount", mock.Anything, mock.Anything).Return(&models.UserAccount{
					UserID:   testUserId,
					Provider: models.ProvidersGoogle,
					Type:     models.ProviderTypeOAuth,
				}, nil)

				mockStorage.On("UpdateUser", mock.Anything, mock.Anything).Return(nil)
				// mockStorage.On("FindUser", mock.Anything, mock.Anything).Return(&models.User{ID: testUserId, Email: testEmail}, nil)
				// mockStorage.On("SaveToken", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: false,
			checkJob:      true,
			checkJobWant: &models.JobRow{
				Status: "done",
				Kind:   "otp_email",
			},
			checkJobPayload: "security-password-reset",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			enqueuer.Jobs = nil
			jobStore.Job = nil
			mockStorage.ExpectedCalls = nil
			mockStorage.Calls = nil
			mockStorage2.ExpectedCalls = nil
			mockStorage2.Calls = nil
			mockPassword.ExpectedCalls = nil
			mockPassword.Calls = nil
			mockMailService.param = nil

			if tc.setupMocks != nil {
				tc.setupMocks()
			}

			result, err := app.Authenticate(ctx, tc.input)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			if tc.checkJob {
				err = poller.PollOnce(ctx)
				if err != nil {
					t.Fatalf("poller error: %v", err)
				}

				job := jobStore.Job
				assert.NotNil(t, job)
				assert.Contains(t, string(job.Payload), tc.checkJobPayload)
			}

			mockStorage.AssertExpectations(t)
			mockPassword.AssertExpectations(t)
		})
	}
}
