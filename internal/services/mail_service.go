package services

import (
	"errors"

	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/tools/mailer"
)

var (
	TeamEmailPathMap = map[EmailType]mailer.SendMailParams{
		EmailTypeTeamInvite: {
			Subject:      "%s - You are invited to join a team",
			TemplatePath: "/team-invitation",
			Template:     mailer.DefaultTeamInviteMail,
		},
	}
)

var (
	EmailPathMap = map[EmailType]mailer.SendMailParams{
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

type EmailType string

const (
	EmailTypeVerify                EmailType = "verify"
	EmailTypeConfirmPasswordReset  EmailType = "confirm-password-reset"
	EmailTypeSecurityPasswordReset EmailType = "security-password-reset"
	EmailTypeTeamInvite            EmailType = "team-invite"
	EmailTypeInvite                EmailType = "invite"
)

type MailService interface {
	SendMail(params *mailer.AllEmailParams) error
}

type mailService struct {
	mailer  mailer.Mailer
	options *conf.AppOptions
}

func (m *mailService) SendMail(params *mailer.AllEmailParams) error {
	if params == nil || params.Message == nil {
		return errors.New("params or message is nil")
	}
	return m.mailer.Send(params.Message)
}

var _ MailService = (*mailService)(nil)

func NewMailService(mailer mailer.Mailer) MailService {
	return &mailService{
		mailer: mailer,
	}
}
