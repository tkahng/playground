package services

import (
	"github.com/alexedwards/argon2id"
	"github.com/tkahng/authgo/internal/tools/security"
)

type PasswordService interface {
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) (match bool, err error)
}

type passwordService struct {
}

func NewPasswordService() PasswordService {
	return &passwordService{}
}

// HashPassword implements PasswordManager.
func (b *passwordService) HashPassword(password string) (string, error) {
	hashedPassword, err := security.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hashedPassword, nil
}

// VerifyPassword implements PasswordManager.
func (b *passwordService) VerifyPassword(hashedPassword string, password string) (bool, error) {
	return security.ComparePasswordAndHash(password, hashedPassword)
}
