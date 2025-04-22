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

type AuthMailer interface {
	SendVerificationEmail(ctx context.Context, db bob.Executor, user *models.User, redirectTo string) error
	SendPasswordResetEmail(ctx context.Context, db bob.Executor, user *models.User, redirectTo string) error
	SendSecurityPasswordResetEmail(ctx context.Context, db bob.Executor, user *models.User, redirectTo string) error
}

func createVerificationMailParams(tokenHash string, payload *OtpPayload, config *AppOptions) (*mailer.Message, error) {
	path, err := mailer.GetPath("/api/auth/verify", &mailer.EmailParams{
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

func createPasswordResetMailParams(tokenHash string, payload *OtpPayload, config *AppOptions) (*mailer.Message, error) {
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

func createSecurityPasswordResetMailParams(tokenHash string, payload *OtpPayload, config *AppOptions) (*mailer.Message, error) {
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
	bodyStr := mailer.GetTemplate("body", mailer.DefaultSecurityPasswordResetMail, param)
	mailParams := &mailer.Message{
		From:    config.Meta.SenderAddress,
		To:      payload.Email,
		Subject: fmt.Sprintf("%s - Reset your password", config.Meta.AppName),
		Body:    bodyStr,
	}
	return mailParams, err
}
