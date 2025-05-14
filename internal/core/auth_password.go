package core

import (
	"github.com/alexedwards/argon2id"
	"github.com/tkahng/authgo/internal/tools/security"
)

type PasswordManager interface {
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) (match bool, err error)
}

type BasePasswordManager struct {
	secret string
}

func NewPasswordManager() PasswordManager {
	return &BasePasswordManager{}
}

// HashPassword implements PasswordManager.
func (b *BasePasswordManager) HashPassword(password string) (string, error) {
	hashedPassword, err := security.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hashedPassword, nil
}

// VerifyPassword implements PasswordManager.
func (b *BasePasswordManager) VerifyPassword(hashedPassword string, password string) (bool, error) {
	return security.ComparePasswordAndHash(hashedPassword, password)
}
