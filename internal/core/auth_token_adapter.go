package core

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/tools/security"
)

type TokenAdapter interface {
	CreateOtpTokenHash(payload *OtpPayload, config TokenOption) (string, error)
	ParseTokenString(tokenString string, config TokenOption, data any) error
	// VerifyTokenStorage(ctx context.Context, token string) error
	// GetToken(ctx context.Context, token string) (*shared.Token, error)
	// SaveToken(ctx context.Context, token *shared.CreateTokenDTO) error
	// DeleteToken(ctx context.Context, token string) error
}

var _ TokenAdapter = (*TokenAdapterBase)(nil)

func NewTokenAdapter(dbx db.Dbx) *TokenAdapterBase {
	// return &TokenAdapterBase{db: db, repo: repo}
	return &TokenAdapterBase{
		db: dbx,
		// repo: repo
	}
}

type TokenAdapterBase struct {
	db db.Dbx
	// repo repository.AppRepo
	// repo *AppRepo
}

// VerifyTokenStorage implements TokenAdapter.
// Verify if the token is stored in the database
// if it is, delete it
// if it is not, return an error

// CreateOtpTokenHash implements TokenAdapter.
func (a *TokenAdapterBase) CreateOtpTokenHash(payload *OtpPayload, config TokenOption) (string, error) {
	if payload == nil {
		return "", fmt.Errorf("payload is nil")
	}
	claims := OtpClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: config.ExpiresAt(),
		},
		OtpPayload: *payload,
	}
	token, err := security.NewJWTWithClaims(claims, config.Secret)
	if err != nil {
		return "", fmt.Errorf("error at creating verification token: %w", err)
	}
	return token, nil

}

func (a *TokenAdapterBase) ParseTokenString(token string, config TokenOption, data any) error {
	claims, err := security.ParseJWTMapClaims(token, config.Secret)
	if err != nil {
		return fmt.Errorf("error while parsing token string: %w", err)
	}
	if claimType, ok := claims["type"].(string); ok && claimType == string(config.Type) {
		_, err = security.MarshalToken(claims, data)
		if err != nil {
			return fmt.Errorf("error at error: %w", err)
		}
		return nil
	}
	return fmt.Errorf("invalid token type")
	// Convert the JSON to a struct
}
