package mailer

import (
	"github.com/wneessen/go-mail"
)

type SmtpMailer struct {
	// contains filtered or unexported fields
	Client mail.Client
}

func (s *SmtpMailer) SendEmail(params *Message) error {
	return nil
}
