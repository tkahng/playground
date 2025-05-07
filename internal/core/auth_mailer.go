package core

import (
	"github.com/tkahng/authgo/internal/tools/mailer"
)

type AuthMailer interface {
	Client() mailer.Mailer
	SendOtpEmail(emailType EmailType, tokenHash string, payload *OtpPayload, config *AppOptions) error
}
