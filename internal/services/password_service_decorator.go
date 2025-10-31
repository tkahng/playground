package services

import (
	"errors"
)

var (
	ErrDelegateNil = errors.New("delegate is nil, cannot call method on nil")
)

type HashServiceDecorator struct {
	Delegate   HashService
	HashFunc   func(input string) (string, error)
	VerifyFunc func(val string, hash string) (match bool, err error)
}

func NewHashServiceDecorator() *HashServiceDecorator {
	return &HashServiceDecorator{
		Delegate: NewHashService(),
	}
}

func (p *HashServiceDecorator) Cleanup() {
	p.HashFunc = nil
	p.VerifyFunc = nil
}

func (p *HashServiceDecorator) Hash(input string) (string, error) {
	if p.HashFunc != nil {
		return p.HashFunc(input)
	}
	if p.Delegate == nil {
		return "", ErrDelegateNil
	}
	return p.Delegate.Hash(input)
}
func (p *HashServiceDecorator) Verify(value, hash string) (match bool, err error) {
	if p.VerifyFunc != nil {
		return p.VerifyFunc(value, hash)
	}
	if p.Delegate == nil {
		return false, ErrDelegateNil
	}
	return p.Delegate.Verify(value, hash)
}
