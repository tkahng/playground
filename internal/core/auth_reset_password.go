package core

import (
	"context"
	"fmt"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
)

func (app *BaseApp) VerifyAndUsePasswordResetToken(ctx context.Context, db bob.Executor, verificationToken string) (*PasswordResetClaims, error) {
	opts := app.Settings().Auth
	jsond, err := ParseResetToken(verificationToken, opts.PasswordResetToken)
	if err != nil {
		return nil, fmt.Errorf("error at parsing verification token: %w", err)
	}
	token, err := repository.UseToken(ctx, db, jsond.Token)
	if err != nil {
		return nil, fmt.Errorf("error verifying refresh token: %w", err)
	}
	if token == nil {
		return nil, fmt.Errorf("token not found")
	}
	if token.Type != models.TokenTypesPasswordResetToken {
		return nil, fmt.Errorf("invalid token type. want verification_token, got  %v", token.Type)
	}
	return jsond, nil
}

// SendPasswordResetEmail implements App.

func (app *BaseApp) SendPasswordResetEmail(ctx context.Context, db bob.Executor, user *models.User, redirectTo string) error {
	opts := app.Settings().Auth
	config := app.Settings()
	client := app.NewMailClient()

	payload := app.TokenVerifier().CreateResetPasswordPayload(user, redirectTo)

	tokenHash, err := CreatePasswordResetToken(payload, opts.PasswordResetToken)
	if err != nil {
		return fmt.Errorf("error at creating verification token: %w", err)
	}

	err = PersistOtpToken(ctx, db, payload, opts.PasswordResetToken)
	if err != nil {
		return fmt.Errorf("error at storing verification token: %w", err)
	}
	mailParams, err := createPasswordResetMailParams(tokenHash, payload, config)
	client.Send(mailParams)
	if err != nil {
		return fmt.Errorf("error creating verification token: %w", err)
	}
	return nil
}

func (app *BaseApp) SendSecurityPasswordResetEmail(ctx context.Context, db bob.Executor, user *models.User, redirectTo string) error {
	opts := app.Settings().Auth
	config := app.Settings()
	client := app.NewMailClient()

	payload := app.TokenVerifier().CreateResetPasswordPayload(user, redirectTo)

	tokenHash, err := CreatePasswordResetToken(payload, opts.PasswordResetToken)
	if err != nil {
		return fmt.Errorf("error at creating security password reset token: %w", err)
	}

	err = PersistOtpToken(ctx, db, payload, opts.PasswordResetToken)
	if err != nil {
		return fmt.Errorf("error at storing security password reset token: %w", err)
	}
	mailParams, err := createSecurityPasswordResetMailParams(tokenHash, payload, config)
	client.Send(mailParams)
	if err != nil {
		return fmt.Errorf("error creating security password reset token: %w", err)
	}
	return nil
}
