package mailer

import (
	"fmt"

	"github.com/tkahng/playground/internal/conf"
	"github.com/wneessen/go-mail"
)

type SmtpMailer struct {
	cfg conf.SmtpConfig
	// contains filtered or unexported fields
	client *mail.Client
}

// Send implements Mailer.
func (s *SmtpMailer) Send(message *Message) error {
	msg := mail.NewMsg()
	if err := msg.From(message.From); err != nil {
		return err
	}
	if err := msg.To(message.To); err != nil {
		return err
	}
	msg.Subject(message.Subject)
	msg.SetBodyString(mail.TypeTextHTML, message.Body)
	return s.client.DialAndSend(msg)
}

var _ Mailer = (*SmtpMailer)(nil)

func NewSmtpMailer(cfg conf.SmtpConfig) *SmtpMailer {
	client, err := mail.NewClient(
		cfg.Host,
		mail.WithTLSPortPolicy(mail.TLSMandatory),
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithUsername(cfg.Username),
		mail.WithPassword(cfg.EmailPass),
	)
	if err != nil {
		panic(fmt.Errorf("failed to create smtp client: %w", err))
	}

	return &SmtpMailer{cfg: cfg, client: client}
}
