package stores

import (
	"context"
	"fmt"
	"time"

	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type PostgresTokenStore struct {
	db database.Dbx
}

func NewPostgresTokenStore(db database.Dbx) *PostgresTokenStore {
	return &PostgresTokenStore{
		db: db,
	}
}
func (p *PostgresTokenStore) WithTx(tx database.Dbx) *PostgresTokenStore {
	return &PostgresTokenStore{
		db: tx,
	}
}

// var _ services. = &PostgresTokenStore{}

func (a *PostgresTokenStore) GetToken(ctx context.Context, token string) (*models.Token, error) {
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

func (a *PostgresTokenStore) SaveToken(ctx context.Context, token *shared.CreateTokenDTO) error {
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

func (a *PostgresTokenStore) DeleteToken(ctx context.Context, token string) error {
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

func (a *PostgresTokenStore) VerifyTokenStorage(ctx context.Context, token string) error {
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
