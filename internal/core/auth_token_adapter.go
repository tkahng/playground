package core

import (
	"context"
	"fmt"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/security"
)

type TokenAdapter interface {
	ParseTokenString(ctx context.Context, tokenString string, config TokenOption, data any) error
	GetToken(ctx context.Context, token string) (*shared.Token, error)
	SaveToken(ctx context.Context, token *shared.CreateTokenDTO) error
	DeleteToken(ctx context.Context, token string) error
}

var _ TokenAdapter = (*TokenAdapterBase)(nil)

type TokenAdapterBase struct {
	db bob.Executor
}

func NewTokenAdapter(db bob.Executor) *TokenAdapterBase {
	return &TokenAdapterBase{db: db}
}

func (a *TokenAdapterBase) ParseTokenString(ctx context.Context, token string, config TokenOption, data any) error {
	claims, err := security.ParseJWTMapClaims(token, config.Secret)
	if err != nil {
		return fmt.Errorf("error while parsing token string: %w", err)
	}
	if !CheckTokenType(claims, config.Type) {
		return fmt.Errorf("invalid token type")
	}
	// Convert the JSON to a struct
	_, err = security.MarshalToken(claims, data)
	if err != nil {
		return fmt.Errorf("error at error: %w", err)
	}
	return nil
}

func (a *TokenAdapterBase) GetToken(ctx context.Context, token string) (*shared.Token, error) {
	res, err := repository.GetToken(ctx, a.db, token)
	if err != nil {
		return nil, fmt.Errorf("error at getting token: %w", err)
	}
	return shared.ToToken(res), nil
}

func (a *TokenAdapterBase) SaveToken(ctx context.Context, token *shared.CreateTokenDTO) error {
	tt := shared.ToModelTokenType(token.Type)
	_, err := repository.CreateToken(ctx, a.db, &repository.TokenDTO{
		Type:       tt,
		Identifier: token.Identifier,
		Expires:    token.Expires,
		Token:      token.Token,
		ID:         token.ID,
		UserID:     token.UserID,
		Otp:        token.Otp,
	})
	if err != nil {
		return fmt.Errorf("error at saving token: %w", err)
	}
	return nil
}

func (a *TokenAdapterBase) DeleteToken(ctx context.Context, token string) error {
	err := repository.DeleteToken(ctx, a.db, token)
	if err != nil {
		return fmt.Errorf("error at deleting token: %w", err)
	}
	return nil
}
