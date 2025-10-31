package services

import (
	"errors"
)

var (
	ErrDelegateNil = errors.New("delegate is nil, cannot call method on nil")
)

type PasswordServiceDecorator struct {
	Delegate           HashService
	HashPasswordFunc   func(input string) (string, error)
	VerifyPasswordFunc func(val string, hash string) (match bool, err error)
}

func NewPasswordServiceDecorator() *PasswordServiceDecorator {
	return &PasswordServiceDecorator{
		Delegate: NewHashService(),
	}
}

func (p *PasswordServiceDecorator) Cleanup() {
	p.HashPasswordFunc = nil
	p.VerifyPasswordFunc = nil
}

func (p *PasswordServiceDecorator) Hash(input string) (string, error) {
	if p.HashPasswordFunc != nil {
		return p.HashPasswordFunc(input)
	}
	if p.Delegate == nil {
		return "", ErrDelegateNil
	}
	return p.Delegate.Hash(input)
}
func (p *PasswordServiceDecorator) Verify(value, hash string) (match bool, err error) {
	if p.VerifyPasswordFunc != nil {
		return p.VerifyPasswordFunc(value, hash)
	}
	if p.Delegate == nil {
		return false, ErrDelegateNil
	}
	return p.Delegate.Verify(value, hash)
}
