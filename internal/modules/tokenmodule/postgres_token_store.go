package tokenmodule

import (
	"context"
	"fmt"
	"time"

	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type tokenStore struct {
	db database.Dbx
}

var _ TokenStore = (*tokenStore)(nil)

func NewTokenStore(db database.Dbx) TokenStore {
	return &tokenStore{
		db: db,
	}
}

func (a *tokenStore) GetToken(ctx context.Context, token string) (*models.Token, error) {
	res, err := crudrepo.Token.GetOne(ctx,
		a.db,
		&map[string]any{
			"token": map[string]any{
				"_eq": token,
			},
		})
	if err != nil {
		return nil, fmt.Errorf("error at getting token: %w", err)
	}
	if res == nil {
		return nil, shared.ErrTokenNotFound
	}
	if res.Expires.Before(time.Now()) {
		return nil, shared.ErrTokenExpired
	}
	return res, nil
}

func (a *tokenStore) SaveToken(ctx context.Context, token *shared.CreateTokenDTO) error {
	_, err := crudrepo.Token.PostOne(ctx, a.db, &models.Token{
		Type:       models.TokenTypes(token.Type),
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

func (a *tokenStore) DeleteToken(ctx context.Context, token string) error {
	_, err := crudrepo.Token.DeleteReturn(ctx, a.db, &map[string]any{
		"token": map[string]any{
			"_eq": token,
		},
	})
	if err != nil {
		return fmt.Errorf("error at deleting token: %w", err)
	}
	return nil
}

func (a *tokenStore) VerifyTokenStorage(ctx context.Context, token string) error {
	res, err := a.GetToken(ctx, token)
	if err != nil {
		return err
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
