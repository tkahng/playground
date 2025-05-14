package core

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/tools/security"
)

type TokenManager interface {
	ParseToken(token string, config conf.TokenOption, data any) error
	CreateJwtToken(payload jwt.Claims, signingKey string) (string, error)
}

type BaseTokenManager struct {
}

func NewTokenManager() TokenManager {
	return &BaseTokenManager{}
}

func (tm *BaseTokenManager) ParseToken(token string, config conf.TokenOption, data any) error {
	claims, err := security.ParseJWTMapClaims(token, config.Secret)
	if err != nil {
		return fmt.Errorf("error while parsing token string: %w", err)
	}
	if claimType, ok := claims["type"].(string); ok && claimType == string(config.Type) {
		_, err = security.MarshalToken(claims, data)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("invalid token type")
}

func (tm *BaseTokenManager) CreateJwtToken(payload jwt.Claims, signingKey string) (string, error) {
	return security.NewJWTWithClaims(payload, signingKey)
}
