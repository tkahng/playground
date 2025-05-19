package authmodule

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/tkahng/authgo/internal/conf"
)

type JwtService interface {
	ParseToken(token string, config conf.TokenOption, data any) error
	CreateJwtToken(payload jwt.Claims, signingKey string) (string, error)
}

// import (
// 	"fmt"

// 	"github.com/golang-jwt/jwt/v5"
// 	"github.com/tkahng/authgo/internal/conf"
// 	"github.com/tkahng/authgo/internal/tools/security"
// )

// type JwtService interface {
// 	ParseToken(token string, config conf.TokenOption, data any) error
// 	CreateJwtToken(payload jwt.Claims, signingKey string) (string, error)
// }

// type jwtService struct {
// }

// func NewJwtService() JwtService {
// 	return &jwtService{}
// }

// func (tm *jwtService) ParseToken(token string, config conf.TokenOption, data any) error {
// 	claims, err := security.ParseJWTMapClaims(token, config.Secret)
// 	if err != nil {
// 		return fmt.Errorf("error while parsing token string: %w", err)
// 	}
// 	if claimType, ok := claims["type"].(string); ok && claimType == string(config.Type) {
// 		_, err = security.MarshalToken(claims, data)
// 		if err != nil {
// 			return err
// 		}
// 		return nil
// 	}
// 	return fmt.Errorf("invalid token type")
// }

// func (tm *jwtService) CreateJwtToken(payload jwt.Claims, signingKey string) (string, error) {
// 	return security.NewJWTWithClaims(payload, signingKey)
// }
