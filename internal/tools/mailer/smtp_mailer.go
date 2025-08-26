package mailer

import (
	"fmt"
	"os"

	"github.com/tkahng/playground/internal/conf"
	"github.com/wneessen/go-mail"
)

type SmtpMailer struct {
	// contains filtered or unexported fields
	client *mail.Client
}

func (s *SmtpMailer) SendEmail(params *Message) error {
	return nil
}

func NewSmtpMailer(cfg conf.SmtpConfig) *SmtpMailer {
	client, err := mail.NewClient(
		cfg.Host,
		mail.WithTLSPortPolicy(mail.TLSMandatory),
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithUsername(cfg.Username),
		mail.WithPassword(cfg.EmailPass),
	)
	if err != nil {
		fmt.Printf("failed to create mail client: %s\n", err)
		os.Exit(1)
	}

	return &SmtpMailer{client: client}
}
