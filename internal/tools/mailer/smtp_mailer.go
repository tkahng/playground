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
	msg.SetBodyString(mail.TypeMultipartMixed, message.Body)
	return s.client.Send(msg)
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
		fmt.Printf("failed to create mail client: %s\n", err)
		os.Exit(1)
	}

	return &SmtpMailer{client: client}
}
