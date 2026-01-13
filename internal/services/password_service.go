package services

import (
	"github.com/alexedwards/argon2id"
	"github.com/tkahng/playground/internal/tools/security"
)

type HashService interface {
	Hash(input string) (string, error)
	Verify(value, hash string) (match bool, err error)
}

type hashService struct {
}

func NewHashService() HashService {
	return &hashService{}
}

func (b *hashService) Hash(input string) (string, error) {
	return security.CreateHash(input, argon2id.DefaultParams)
}

func (b *hashService) Verify(value, hash string) (bool, error) {
	return security.ComparePasswordAndHash(value, hash)
}
