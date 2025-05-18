package authmodule

import (
	"fmt"
	"net/url"

	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mailer"
)

type AuthMailer interface {
	Client() mailer.Mailer
	SendOtpEmail(emailType EmailType, tokenHash string, payload *shared.OtpPayload, config *conf.AppOptions) error
}

type EmailType string

const (
	EmailTypeVerify                EmailType = "verify"
	EmailTypeConfirmPasswordReset  EmailType = "confirm-password-reset"
	EmailTypeSecurityPasswordReset EmailType = "security-password-reset"
)

type SendMailParams struct {
	Subject      string
	Type         string
	TemplatePath string
	Template     string
}

var (
	EmailPathMap = map[EmailType]SendMailParams{
		EmailTypeVerify: {
			Subject:      "%s - Verify your email address",
			TemplatePath: "/api/auth/verify",
			Template:     mailer.DefaultConfirmationMail,
		},
		EmailTypeConfirmPasswordReset: {
			Subject:      "%s - Confirm your password reset",
			TemplatePath: "/password-reset",
			Template:     mailer.DefaultRecoveryMail,
		},
		EmailTypeSecurityPasswordReset: {
			Subject:      "%s - Reset your password",
			TemplatePath: "/password-reset",
			Template:     mailer.DefaultSecurityPasswordResetMail,
		},
	}
)

var _ AuthMailer = (*AuthMailerBase)(nil)

type AuthMailerBase struct {
	mailer mailer.Mailer
}

// SendOtpEmail implements AuthMailer.
func (a *AuthMailerBase) SendOtpEmail(emailType EmailType, tokenHash string, payload *shared.OtpPayload, config *conf.AppOptions) error {
	if payload == nil {
		return fmt.Errorf("payload is nil")
	}
	if config == nil {
		return fmt.Errorf("config is nil")
	}
	var params SendMailParams
	var ok bool
	if params, ok = EmailPathMap[emailType]; !ok {
		return fmt.Errorf("email type not found")
	}
	path, err := mailer.GetPath(params.TemplatePath, &mailer.EmailParams{
		Token:      tokenHash,
		Type:       string(payload.Type),
		RedirectTo: payload.RedirectTo,
	})
	if err != nil {
		return err
	}
	appUrl, err := url.Parse(config.Meta.AppUrl)
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
	bodyStr := mailer.GetTemplate("body", params.Template, param)
	mailParams := &mailer.Message{
		From:    config.Meta.SenderAddress,
		To:      payload.Email,
		Subject: fmt.Sprintf(params.Subject, config.Meta.AppName),
		Body:    bodyStr,
	}
	return a.Client().Send(mailParams)
}

func NewAuthMailer(mailer mailer.Mailer) *AuthMailerBase {
	return &AuthMailerBase{mailer: mailer}
}

func (a *AuthMailerBase) Client() mailer.Mailer {
	return a.mailer
}
