package services

import (
	"errors"

	"github.com/tkahng/authgo/internal/tools/mailer"
)

type MailService interface {
	SendMail(params *mailer.AllEmailParams) error
}

type mailService struct {
	mailer mailer.Mailer
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
