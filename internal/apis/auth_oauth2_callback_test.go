package apis_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/auth"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/tools/types"
)

func TestOAuth2Signin_Success(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			app := core.NewTestBaseApp(conf.ZeroEnvConfig(), db)
			param := &auth.OAuth2SigninInput{
				Email:             "test@example.com",
				EmailVerifiedAt:   types.Pointer(time.Now()),
				Provider:          models.ProvidersApple,
				ProviderAccountID: "provider_account_id",
				AccessToken:       types.Pointer("access_token"),
				RefreshToken:      types.Pointer("refresh_token"),
				RedirectTo:        "",
				Expiry:            time.Now(),
			}
			got, gotErr := apis.OAuth2Signin(ctx, app, param)
			if gotErr != nil {
				t.Errorf("AuthServiceImpl.OAuth2Signin() error = %v", gotErr)
			}
			if got == nil {
				t.Errorf("AuthServiceImpl.OAuth2Signin() got = %v", got)
			}
			user := repository.MustFindOneCtx(t, ctx, repository.User, db, nil)
			if user == nil {
				t.Errorf("AuthServiceImpl.OAuth2Signin() user = %v", user)
			}
			if user.EmailVerifiedAt == nil {
				t.Errorf("AuthServiceImpl.OAuth2Signin() user.EmailVerifiedAt = %v", user.EmailVerifiedAt)
			}
			customer := repository.MustFindOneCtx(t, ctx, repository.StripeCustomer, db, nil)
			if customer == nil {
				t.Errorf("AuthServiceImpl.OAuth2Signin() customer = %v", customer)
			}
			if customer.UserID == nil {
				t.Errorf("AuthServiceImpl.OAuth2Signin() customer.UserID = %v", customer.UserID)
			}
			if *customer.UserID != user.ID {
				t.Errorf("AuthServiceImpl.OAuth2Signin() customer.UserID = %v, want %v", *customer.UserID, user.ID)
			}
		})
	})
}
func TestOAuth2Signin_Fail_Existing_Provider(t *testing.T) {
	t.Run("fail: existing provider", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			app := core.NewTestBaseApp(conf.ZeroEnvConfig(), db)
			existingUser := core.CreateUserWithOptions(
				t,
				app,
				core.UserWithProvider(models.ProvidersGoogle),
				core.UserWithVerified(time.Now()),
				core.UserWithProviderType(models.ProviderTypeOAuth),
			)
			param := &auth.OAuth2SigninInput{
				Email:             existingUser.User.Email,
				EmailVerifiedAt:   types.Pointer(time.Now()),
				Provider:          models.ProvidersGoogle,
				ProviderAccountID: "provider_account_id",
				AccessToken:       types.Pointer("access_token"),
				RefreshToken:      types.Pointer("refresh_token"),
				RedirectTo:        "",
				Expiry:            time.Now(),
			}
			_, gotErr := apis.OAuth2Signin(ctx, app, param)
			if !errors.Is(gotErr, shared.ErrAccountProviderConflict) {
				t.Errorf("AuthServiceImpl.OAuth2Signin() error = %v", gotErr)
			}
		})
	})
}
func TestOAuth2Signin_Fail_unknown_error_during_customer_creation_rollback_everything(t *testing.T) {
	t.Run("fail: unknown error during customer creation. rollback everything.", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			app := core.NewTestBaseApp(conf.ZeroEnvConfig(), db)
			paymentClient := core.ExtractTestPaymentClient(t, app)
			paymentClient.CreateCustomerFunc = func(email string, name *string, metadata *map[string]string) (*stripe.Customer, error) {
				return nil, errors.New("unknown error")
			}
			param := &auth.OAuth2SigninInput{
				Email:             "test@example.com",
				EmailVerifiedAt:   types.Pointer(time.Now()),
				Provider:          models.ProvidersApple,
				ProviderAccountID: "provider_account_id",
				AccessToken:       types.Pointer("access_token"),
				RefreshToken:      types.Pointer("refresh_token"),
				RedirectTo:        "",
				Expiry:            time.Now(),
			}
			_, gotErr := apis.OAuth2Signin(ctx, app, param)
			if gotErr == nil {
				t.Errorf("Expected error, got nil")
			}
			if gotErr.Error() != "unknown error" {
				t.Errorf("AuthServiceImpl.OAuth2Signin() error = %v", gotErr)
			}
			userCount := repository.MustCountAllCtx(t, ctx, repository.User, db, nil)
			if userCount != 0 {
				t.Errorf("AuthServiceImpl.OAuth2Signin() userCount = %v", userCount)
			}
			accountCount := repository.MustCountAllCtx(t, ctx, repository.UserAccount, db, nil)
			if accountCount != 0 {
				t.Errorf("AuthServiceImpl.OAuth2Signin() accountCount = %v", accountCount)
			}
			customerCount := repository.MustCountAllCtx(t, ctx, repository.StripeCustomer, db, nil)
			if customerCount != 0 {
				t.Errorf("AuthServiceImpl.OAuth2Signin() customerCount = %v", customerCount)
			}
		})
	})
}

func TestOAuth2Signin_Success_Existing_Credential_Unverified(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		app := core.NewTestBaseApp(conf.ZeroEnvConfig(), db)
		mailer := core.ExtractTestMailer(t, app)
		existingUser := core.CreateUserWithOptions(
			t,
			app,
			core.UserWithPassword("password"),
			core.UserWithProvider(models.ProvidersCredentials),
		)
		param := &auth.OAuth2SigninInput{
			Email:             existingUser.User.Email,
			EmailVerifiedAt:   types.Pointer(time.Now()),
			Provider:          models.ProvidersApple,
			ProviderAccountID: "provider_account_id",
			AccessToken:       types.Pointer("access_token"),
			RefreshToken:      types.Pointer("refresh_token"),
			RedirectTo:        "",
			Expiry:            time.Now(),
		}
		_, gotErr := apis.OAuth2Signin(ctx, app, param)
		if gotErr != nil {
			t.Errorf("AuthServiceImpl.OAuth2Signin() error = %v", gotErr)
		}
		pollErr := app.JobManager().PollOnce(ctx)
		if pollErr != nil {
			t.Errorf("AuthServiceImpl.OAuth2Signin() poll error = %v", pollErr)
		}
		user := repository.MustFindOneCtx(t, ctx, repository.User, db, nil)
		if user == nil {
			t.Errorf("AuthServiceImpl.OAuth2Signin() user = %v", user)
		}
		if user.EmailVerifiedAt == nil {
			t.Errorf("Expected emailVerifiedAt to not be nil")
		}
		customer := repository.MustFindOneCtx(t, ctx, repository.StripeCustomer, db, nil)
		if customer == nil {
			t.Errorf("AuthServiceImpl.OAuth2Signin() customer = %v", customer)
		}
		if customer.Email != existingUser.User.Email {
			t.Errorf("AuthServiceImpl.OAuth2Signin() customer.Email = %v", customer.Email)
		}
		if customer.UserID == nil {
			t.Errorf("AuthServiceImpl.OAuth2Signin() customer.userId = %v", customer.UserID)
		}
		if len(mailer.Messages) != 1 {
			t.Errorf("Expected 1 email to be sent, got %d", len(mailer.Messages))
		}
		mail := mailer.Messages[0]
		if !strings.Contains(mail.Body, "Reset password") {
			t.Errorf("Expected reset password email to be sent, got %s", mail.Body)
		}
	})

}
