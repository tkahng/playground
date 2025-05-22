package services

import (
	"github.com/stretchr/testify/mock"
)

type MockPasswordService struct {
	mock.Mock
}

// HashPassword implements PasswordManager.
func (m *MockPasswordService) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

// VerifyPassword implements PasswordManager.
func (m *MockPasswordService) VerifyPassword(hashedPassword string, password string) (match bool, err error) {
	args := m.Called(hashedPassword, password)
	return args.Bool(0), args.Error(1)
}

var _ PasswordService = (*MockPasswordService)(nil)
