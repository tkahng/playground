package auth

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mailer"
	"github.com/tkahng/playground/internal/workers"
)

type PasswordManager interface {
	// ConfirmPasswordReset finalizes the password reset process by providing a valid token and a new password
	ConfirmPasswordReset(ctx context.Context, token, password string) error
	// CheckPasswordResetToken checks if the token is valid
	CheckPasswordResetToken(ctx context.Context, token string) error
	// RequestPasswordReset sends password reset email
	RequestPasswordReset(ctx context.Context, email string) error
	// UpdatePassword updates user's password given that the current password is provided correctly
	UpdatePassword(ctx context.Context, userId uuid.UUID, oldPassword, newPassword string) error
}

func (a *AuthServiceImpl) ConfirmPasswordReset(ctx context.Context, token string, password string) error {
	email, err := a.token.ValidateToken(ctx, token, models.TokenTypesPasswordResetToken)
	if err != nil {
		return fmt.Errorf("error getting token: %w", err)
	}

	user, err := a.adapter.User().FindUser(
		ctx,
		&stores.UserFilter{
			Emails: []string{email},
		},
	)
	if err != nil {
		return fmt.Errorf("error getting user by email: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	account, err := a.adapter.UserAccount().FindUserAccount(ctx, &stores.UserAccountFilter{
		UserIds:   []uuid.UUID{user.ID},
		Providers: []models.Providers{models.ProvidersCredentials},
	})
	if err != nil {
		return fmt.Errorf("error getting user account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("user account not found")
	}
	if account.Password == nil {
		a.logger.ErrorContext(
			ctx,
			"user account does not have password",
			slog.String("user_id", user.ID.String()),
			slog.String("account_id", account.ID.String()),
			slog.String("provider", account.Provider.String()),
		)
		return fmt.Errorf("user account does not have password")
	}
	hash, err := a.hash.Hash(password)
	if err != nil {
		return fmt.Errorf("error at hashing password: %w", err)
	}
	account.Password = &hash
	err = a.adapter.UserAccount().UpdateUserAccount(ctx, account)
	if err != nil {
		return fmt.Errorf("error updating user password: %w", err)
	}
	return nil
}

func (a *AuthServiceImpl) CheckPasswordResetToken(ctx context.Context, token string) error {
	return a.token.CheckToken(ctx, token, models.TokenTypesPasswordResetToken)
}

func (a *AuthServiceImpl) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := a.adapter.User().FindUser(
		ctx,
		&stores.UserFilter{
			Emails: []string{email},
		},
	)
	if err != nil {
		return fmt.Errorf("error getting user by email: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	account, err := a.adapter.UserAccount().FindUserAccount(ctx, &stores.UserAccountFilter{
		UserIds:   []uuid.UUID{user.ID},
		Providers: []models.Providers{models.ProvidersCredentials},
	})
	if err != nil {
		return fmt.Errorf("error getting user account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("user account not found")
	}
	err = a.job.EnqueueOtpMailJob(ctx, &workers.OtpEmailJobArgs{
		UserID: user.ID,
		Type:   mailer.EmailTypeConfirmPasswordReset,
	})
	if err != nil {
		return fmt.Errorf("error sending password reset email: %w", err)
	}
	return nil
}

func (a *AuthServiceImpl) UpdatePassword(ctx context.Context, userId uuid.UUID, oldPassword string, newPassword string) error {
	account, err := a.adapter.UserAccount().FindUserAccount(ctx, &stores.UserAccountFilter{
		UserIds:   []uuid.UUID{userId},
		Providers: []models.Providers{models.ProvidersCredentials},
	})
	if err != nil {
		return err
	}
	if account == nil {
		return fmt.Errorf("user account not found")
	}
	if account.Password == nil {
		return fmt.Errorf("user account does not have password")
	}
	if match, err := a.hash.Verify(oldPassword, *account.Password); err != nil {
		return fmt.Errorf("error at comparing password: %w", err)
	} else if !match {
		return fmt.Errorf("password is incorrect")
	}
	hash, err := a.hash.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("error at hashing password: %w", err)
	}
	account.Password = &hash
	err = a.adapter.UserAccount().UpdateUserAccount(ctx, account)
	if err != nil {
		return fmt.Errorf("error updating user password: %w", err)
	}
	return nil
}
