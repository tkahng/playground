package mailer

import (
	"log/slog"

	"github.com/resend/resend-go/v2"
	"github.com/tkahng/playground/internal/conf"
)

type Message struct {
	From    string
	To      string
	Subject string
	Body    string
}

type Mailer interface {
	Send(message *Message) error
}

var _ Mailer = (*ResendMailer)(nil)

type ResendMailer struct {
	config *conf.ResendConfig
	client *resend.Client
}

func NewResendMailer(cfg conf.ResendConfig) *ResendMailer {
	return &ResendMailer{
		config: &cfg,
		client: resend.NewClient(cfg.ResendApiKey),
	}
}

func (m *ResendMailer) Send(params *Message) error {
	_, err := m.client.Emails.Send(&resend.SendEmailRequest{
		From:    params.From,
		ReplyTo: "Your Name <tkahng@gmail.com>",
		To:      []string{params.To},
		Subject: params.Subject,
		Html:    params.Body,
	})
	if err != nil {
		slog.Error(
			"Failed to send email",
			slog.Any("error", err),
			slog.String("from", params.From),
			slog.String("to", params.To),
			slog.String("subject", params.Subject),
			slog.String("body", params.Body),
		)
	}
	return err
}
