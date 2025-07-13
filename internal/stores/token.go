package stores

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/repository"
	"github.com/tkahng/playground/internal/shared"
)

type CreateTokenDTO struct {
	Type       models.TokenTypes `db:"type" json:"type"`
	Identifier string            `db:"identifier" json:"identifier"`
	Expires    time.Time         `db:"expires" json:"expires"`
	Token      string            `db:"token" json:"token"`
	ID         *uuid.UUID        `db:"id" json:"id"`
	UserID     *uuid.UUID        `db:"user_id" json:"user_id"`
	Otp        *string           `db:"otp" json:"otp"`
}

type DbTokenStoreInterface interface {
	GetToken(ctx context.Context, token string) (*models.Token, error)
	SaveToken(ctx context.Context, token *CreateTokenDTO) error
	DeleteToken(ctx context.Context, token string) error
	VerifyTokenStorage(ctx context.Context, token string) error
}

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

func (a *DbTokenStore) SaveToken(ctx context.Context, token *CreateTokenDTO) error {
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
