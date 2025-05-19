package services

import "github.com/tkahng/authgo/internal/tools/mailer"

var (
	TeamEmailPathMap = map[EmailType]SendMailParams{
		EmailTypeTeamInvite: {
			Subject:      "%s - You are invited to join a team",
			TemplatePath: "/team-invitation",
			Template:     mailer.DefaultTeamInviteMail,
		},
	}
)

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

type EmailType string

const (
	EmailTypeVerify                EmailType = "verify"
	EmailTypeConfirmPasswordReset  EmailType = "confirm-password-reset"
	EmailTypeSecurityPasswordReset EmailType = "security-password-reset"
	EmailTypeTeamInvite            EmailType = "team-invite"
	EmailTypeInvite                EmailType = "invite"
)
