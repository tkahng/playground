package services

import (
	"github.com/tkahng/authgo/internal/tools/mailer"
)

type MockMailService struct {
	delegate MailService
	param    *mailer.AllEmailParams
}

func NewMockMailService() *MockMailService {
	return &MockMailService{
		delegate: NewMailService(&mailer.LogMailer{}),
	}
}

// SendMail implements MailService.
func (m *MockMailService) SendMail(params *mailer.AllEmailParams) error {
	m.param = params
	return m.delegate.SendMail(params)
}

var _ MailService = (*MockMailService)(nil)
