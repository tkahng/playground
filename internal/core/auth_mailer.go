package core

import (
	"fmt"
	"net/url"

	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mailer"
)

type AuthMailer interface {
	Client() mailer.Mailer
	SendVerificationEmail(tokenHash string, payload *OtpPayload, config *AppOptions) error
	SendPasswordResetEmail(tokenHash string, payload *OtpPayload, config *AppOptions) error
	SendSecurityPasswordResetEmail(tokenHash string, payload *OtpPayload, config *AppOptions) error
}

var _ AuthMailer = (*AuthMailerBase)(nil)

type AuthMailerBase struct {
	mailer mailer.Mailer
}

func (a *AuthMailerBase) Client() mailer.Mailer {
	return a.mailer
}

func (a *AuthMailerBase) SendVerificationEmail(tokenHash string, payload *OtpPayload, config *AppOptions) error {
	path, err := mailer.GetPath("/api/auth/verify", &mailer.EmailParams{
		Token:      tokenHash,
		Type:       string(shared.VerificationTokenType),
		RedirectTo: payload.RedirectTo,
	})
	if err != nil {
		return err
	}
	appUrl, err := url.Parse(config.Meta.AppURL)
	if err != nil {
		return err
	}
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
	return a.Client().Send(mailParams)
}

func (a *AuthMailerBase) SendPasswordResetEmail(tokenHash string, payload *OtpPayload, config *AppOptions) error {
	path, err := mailer.GetPath("/api/auth/confirm-password-reset", &mailer.EmailParams{
		Token:      tokenHash,
		Type:       string(shared.PasswordResetTokenType),
		RedirectTo: payload.RedirectTo,
	})
	if err != nil {
		return err
	}
	appUrl, err := url.Parse(config.Meta.AppURL)
	if err != nil {
		return err
	}
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
	return a.Client().Send(mailParams)
}

func (a *AuthMailerBase) SendSecurityPasswordResetEmail(tokenHash string, payload *OtpPayload, config *AppOptions) error {
	path, err := mailer.GetPath("/api/auth/confirm-password-reset", &mailer.EmailParams{
		Token:      tokenHash,
		Type:       string(shared.PasswordResetTokenType),
		RedirectTo: payload.RedirectTo,
	})
	if err != nil {
		return err
	}
	appUrl, err := url.Parse(config.Meta.AppURL)
	if err != nil {
		return err
	}
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
	return a.Client().Send(mailParams)
}
