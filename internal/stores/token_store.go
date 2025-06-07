package stores

import (
	"context"
	"fmt"
	"time"

	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
)

type DbTokenStore struct {
	db database.Dbx
}

func NewPostgresTokenStore(db database.Dbx) *DbTokenStore {
	return &DbTokenStore{
		db: db,
	}
}
func (p *DbTokenStore) WithTx(tx database.Dbx) *DbTokenStore {
	return &DbTokenStore{
		db: tx,
	}
}

// var _ services. = &PostgresTokenStore{}

func (a *DbTokenStore) GetToken(ctx context.Context, token string) (*models.Token, error) {
	res, err := repository.Token.GetOne(ctx,
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

func (a *DbTokenStore) SaveToken(ctx context.Context, token *shared.CreateTokenDTO) error {
	_, err := repository.Token.PostOne(ctx, a.db, &models.Token{
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

func (a *DbTokenStore) DeleteToken(ctx context.Context, token string) error {
	_, err := repository.Token.DeleteReturn(ctx, a.db, &map[string]any{
		"token": map[string]any{
			"_eq": token,
		},
	})
	if err != nil {
		return fmt.Errorf("error at deleting token: %w", err)
	}
	return nil
}

func (a *DbTokenStore) VerifyTokenStorage(ctx context.Context, token string) error {
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
