package core

import (
	"github.com/alexedwards/argon2id"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/security"
)

type CredentialService interface {
	CreateHash(password string) (hash string, err error)
	ComparePasswordAndHash(password string, hash string) (match bool, err error)
}

var _ CredentialService = (*CredentialServiceImpl)(nil)

type CredentialServiceImpl struct {
	opts *argon2id.Params
}

func NewCredentialService(args *shared.AuthenticateUserParams) CredentialService {
	return &CredentialServiceImpl{
		opts: argon2id.DefaultParams,
	}
}

func (cs *CredentialServiceImpl) CreateHash(password string) (hash string, err error) {
	return security.CreateHash(password, cs.opts)

}

func (cs *CredentialServiceImpl) ComparePasswordAndHash(password string, hash string) (match bool, err error) {
	return security.ComparePasswordAndHash(password, hash)
}
