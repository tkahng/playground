package core

import (
	"context"
	"fmt"
	"net/url"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mailer"
)

func (app *BaseApp) VerifyAndUsePasswordResetToken(ctx context.Context, db bob.DB, verificationToken string) (*PasswordResetClaims, error) {
	opts := app.AuthOptions()
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

func (app *BaseApp) SendPasswordResetEmail(ctx context.Context, db bob.DB, user *models.User, redirectTo string) error {
	opts := app.AuthOptions()
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

func createPasswordResetMailParams(tokenHash string, payload *OtpPayload, config *Settings) (*mailer.Message, error) {
	path, err := mailer.GetPath("/api/auth/confirm-password-reset", &mailer.EmailParams{
		Token:      tokenHash,
		Type:       string(shared.PasswordResetTokenType),
		RedirectTo: payload.RedirectTo,
	})
	appUrl, _ := url.Parse(config.Meta.AppURL)
	param := &mailer.CommonParams{
		SiteURL:         appUrl.String(),
		ConfirmationURL: appUrl.ResolveReference(path).String(),
		Email:           payload.Email,
		Token:           payload.Otp,
		TokenHash:       tokenHash,
		RedirectTo:      payload.RedirectTo,
	}
	bodyStr := mailer.GetTemplate("body", mailer.DefaultRecoveryMail, param)
	mailParams := &mailer.Message{
		From:    config.Meta.SenderAddress,
		To:      payload.Email,
		Subject: fmt.Sprintf("%s - Verify your email address", config.Meta.AppName),
		Body:    bodyStr,
	}
	return mailParams, err
}
