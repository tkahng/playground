package core

import (
	"context"
	"fmt"
	"net/url"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mailer"
)

func (app *BaseApp) VerifyAndUseVerificationToken(ctx context.Context, db bob.Executor, verificationToken string) (*EmailVerificationClaims, error) {
	opts := app.Settings().Auth
	jsond, err := app.TokenVerifier().ParseVerificationToken(verificationToken, opts.VerificationToken)
	if err != nil {
		return nil, fmt.Errorf("error at parsing verification token: %w", err)
	}
	err = app.TokenStorage().UseVerificationTokenAndUpdateUser(ctx, db, jsond.Token)
	if err != nil {
		return nil, fmt.Errorf("error verifying refresh token: %w", err)
	}
	return jsond, nil
}

// SendVerificationEmail implements App.
func (app *BaseApp) SendVerificationEmail(ctx context.Context, db bob.Executor, user *models.User, redirectTo string) error {
	opts := app.Settings().Auth
	config := app.Settings()
	client := app.NewMailClient()
	payload := app.TokenVerifier().CreateVerificationPayload(user, redirectTo)

	tokenHash, err := app.TokenVerifier().CreateVerificationToken(payload, opts.VerificationToken)
	if err != nil {
		return fmt.Errorf("error at creating verification token: %w", err)
	}

	err = app.TokenStorage().PersistVerificationToken(ctx, db, payload, opts.VerificationToken)
	if err != nil {
		return fmt.Errorf("error at storing verification token: %w", err)
	}
	mailParams, err := createVerificationMailParams(tokenHash, payload, config)
	client.Send(mailParams)
	if err != nil {
		return fmt.Errorf("error creating verification token: %w", err)
	}
	return nil
}

func createVerificationMailParams(tokenHash string, payload *OtpPayload, config *AppOptions) (*mailer.Message, error) {
	path, err := mailer.GetPath("/api/auth/confirm-verification", &mailer.EmailParams{
		Token:      tokenHash,
		Type:       string(shared.VerificationTokenType),
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
	bodyStr := mailer.GetTemplate("body", mailer.DefaultConfirmationMail, param)
	mailParams := &mailer.Message{
		From:    config.Meta.SenderAddress,
		To:      payload.Email,
		Subject: fmt.Sprintf("%s - Verify your email address", config.Meta.AppName),
		Body:    bodyStr,
	}
	return mailParams, err
}
