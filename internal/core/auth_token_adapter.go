package core

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tkahng/authgo/internal/crud/crudModels"
	"github.com/tkahng/authgo/internal/crud/crudrepo"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/security"
)

type TokenAdapter interface {
	CreateOtpTokenHash(payload *OtpPayload, config TokenOption) (string, error)
	ParseTokenString(tokenString string, config TokenOption, data any) error
	VerifyTokenStorage(ctx context.Context, token string) error
	GetToken(ctx context.Context, token string) (*shared.Token, error)
	SaveToken(ctx context.Context, token *shared.CreateTokenDTO) error
	DeleteToken(ctx context.Context, token string) error
}

var _ TokenAdapter = (*TokenAdapterBase)(nil)

type TokenAdapterBase struct {
	db *pgxpool.Pool
	// repo repository.AppRepo
	// repo *AppRepo
}

// VerifyTokenStorage implements TokenAdapter.
// Verify if the token is stored in the database
// if it is, delete it
// if it is not, return an error
func (a *TokenAdapterBase) VerifyTokenStorage(ctx context.Context, token string) error {
	res, err := a.GetToken(ctx, token)
	if err != nil {
		return fmt.Errorf("error at getting token: %w", err)
	}
	if res == nil {
		return fmt.Errorf("token not found")
	}
	err = a.DeleteToken(ctx, token)
	if err != nil {
		return fmt.Errorf("error at deleting token: %w", err)
	}
	return nil
}

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

func NewTokenAdapter(dbx *pgxpool.Pool) *TokenAdapterBase {
	// return &TokenAdapterBase{db: db, repo: repo}
	return &TokenAdapterBase{
		db: dbx,
		// repo: repo
	}
}

func (a *TokenAdapterBase) ParseTokenString(token string, config TokenOption, data any) error {
	claims, err := security.ParseJWTMapClaims(token, config.Secret)
	if err != nil {
		return fmt.Errorf("error while parsing token string: %w", err)
	}
	if !checkTokenType(claims, config.Type) {
		return fmt.Errorf("invalid token type")
	}
	// Convert the JSON to a struct
	_, err = security.MarshalToken(claims, data)
	if err != nil {
		return fmt.Errorf("error at error: %w", err)
	}
	return nil
}

func checkTokenType(claims jwt.MapClaims, tokenType shared.TokenType) bool {
	if claimType, ok := claims["type"].(string); ok && claimType == string(tokenType) {
		return true
	} else {
		return false
	}
}

func (a *TokenAdapterBase) GetToken(ctx context.Context, token string) (*shared.Token, error) {
	res, err := crudrepo.Token.GetOne(ctx,
		a.db,
		&map[string]any{
			"token": map[string]any{
				"_eq": token,
			},
			"expires": map[string]any{
				"_gt": time.Now(),
			},
		})
	if err != nil {
		return nil, fmt.Errorf("error at getting token: %w", err)
	}
	return &shared.Token{
		Type:       shared.TokenType(res.Type),
		Identifier: res.Identifier,
		Expires:    res.Expires,
		Token:      res.Token,
		ID:         res.ID,
		UserID:     res.UserID,
		Otp:        res.Otp,
	}, nil
}

func (a *TokenAdapterBase) SaveToken(ctx context.Context, token *shared.CreateTokenDTO) error {
	_, err := crudrepo.Token.PostOne(ctx, a.db, &crudModels.Token{
		Type:       crudModels.TokenTypes(token.Type),
		Identifier: token.Identifier,
		Expires:    token.Expires,
		Token:      token.Token,
		UserID:     token.UserID,
		Otp:        token.Otp,
	})

	if err != nil {
		return fmt.Errorf("error at saving token: %w", err)
	}
	return nil
}

func (a *TokenAdapterBase) DeleteToken(ctx context.Context, token string) error {
	_, err := crudrepo.Token.Delete(ctx, a.db, &map[string]any{
		"token": map[string]any{
			"_eq": token,
		},
	})
	if err != nil {
		return fmt.Errorf("error at deleting token: %w", err)
	}
	return nil
}
