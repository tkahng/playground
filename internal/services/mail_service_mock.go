package services

import (
	"github.com/tkahng/authgo/internal/tools/mailer"
)

type MockMailService struct {
	delegate MailService
	param    *mailer.AllEmailParams

	// Add a function field to override SendMail behavior
	SendMailOverride func(params *mailer.AllEmailParams) error
}

func NewMockMailService() *MockMailService {
	return &MockMailService{
		delegate: NewMailService(&mailer.LogMailer{}),
	}
}

// SendMail implements MailService.
func (m *MockMailService) SendMail(params *mailer.AllEmailParams) error {
	if m.SendMailOverride != nil {
		return m.SendMailOverride(params)
	}
	m.param = params
	return m.delegate.SendMail(params)
}

var _ MailService = (*MockMailService)(nil)
