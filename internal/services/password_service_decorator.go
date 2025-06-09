package services

import (
	"errors"
)

var (
	ErrDelegateNil = errors.New("delegate is nil, cannot call method on nil")
)

type PasswordServiceDecorator struct {
	Delegate           PasswordService
	HashPasswordFunc   func(password string) (string, error)
	VerifyPasswordFunc func(hashedPassword string, password string) (match bool, err error)
}

func NewPasswordServiceDecorator() *PasswordServiceDecorator {
	return &PasswordServiceDecorator{
		Delegate: NewPasswordService(),
	}
}

func (p *PasswordServiceDecorator) Cleanup() {
	p.HashPasswordFunc = nil
	p.VerifyPasswordFunc = nil
}

func (p *PasswordServiceDecorator) HashPassword(password string) (string, error) {
	if p.HashPasswordFunc != nil {
		return p.HashPasswordFunc(password)
	}
	if p.Delegate == nil {
		return "", ErrDelegateNil
	}
	return p.Delegate.HashPassword(password)
}
func (p *PasswordServiceDecorator) VerifyPassword(hashedPassword string, password string) (match bool, err error) {
	if p.VerifyPasswordFunc != nil {
		return p.VerifyPasswordFunc(hashedPassword, password)
	}
	if p.Delegate == nil {
		return false, ErrDelegateNil
	}
	return p.Delegate.VerifyPassword(hashedPassword, password)
}
