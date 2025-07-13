package services

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/tkahng/playground/internal/conf"
)

var _ JwtService = &JwtServiceDecorator{}

func NewJwtServiceDecorator() *JwtServiceDecorator {
	return &JwtServiceDecorator{
		Delegate: NewJwtService(),
	}
}

func (j *JwtServiceDecorator) Cleanup() {
	j.CreateJwtTokenFunc = nil
	j.ParseTokenFunc = nil
}

type JwtServiceDecorator struct {
	Delegate           JwtService
	CreateJwtTokenFunc func(payload jwt.Claims, signingKey string) (string, error)
	ParseTokenFunc     func(token string, config conf.TokenOption, data any) error
}

// CreateJwtToken implements JwtService.
func (j *JwtServiceDecorator) CreateJwtToken(payload jwt.Claims, signingKey string) (string, error) {
	if j.CreateJwtTokenFunc != nil {
		return j.CreateJwtTokenFunc(payload, signingKey)
	}
	return j.Delegate.CreateJwtToken(payload, signingKey)
}

// ParseToken implements JwtService.
func (j *JwtServiceDecorator) ParseToken(token string, config conf.TokenOption, data any) error {
	if j.ParseTokenFunc != nil {
		return j.ParseTokenFunc(token, config, data)
	}
	return j.Delegate.ParseToken(token, config, data)
}
