package services

import (
	"github.com/alexedwards/argon2id"
	"github.com/tkahng/playground/internal/tools/security"
)

type HashService interface {
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) (match bool, err error)
}

type hashService struct {
}

func NewHashService() HashService {
	return &hashService{}
}

// HashPassword implements PasswordManager.
func (b *hashService) HashPassword(password string) (string, error) {
	hashedPassword, err := security.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hashedPassword, nil
}

// VerifyPassword implements PasswordManager.
func (b *hashService) VerifyPassword(hashedPassword string, password string) (bool, error) {
	return security.ComparePasswordAndHash(password, hashedPassword)
}
