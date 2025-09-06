package stores

import (
	"context"
	"time"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
)

type TokenStoreDecorator struct {
	Delegate                *DbTokenStore
	DeleteTokenFunc         func(ctx context.Context, token string) error
	GetTokenFunc            func(ctx context.Context, token string) (*models.Token, error)
	GetTokenByValueTypeFunc func(ctx context.Context, value string, tokenType models.TokenTypes) (*models.Token, error)
	SaveTokenFunc           func(ctx context.Context, token *CreateTokenDTO) error
	VerifyTokenStorageFunc  func(ctx context.Context, token string) error
	WithTxFunc              func(dbx database.Dbx) *TokenStoreDecorator
}

// GetTokenByValueTypeExpires implements DbTokenStoreInterface.
func (t *TokenStoreDecorator) GetTokenByValueTypeExpires(ctx context.Context, value string, tokenType models.TokenTypes, expires time.Time) (*models.Token, error) {
	if t.GetTokenByValueTypeFunc != nil {
		return t.GetTokenByValueTypeFunc(ctx, value, tokenType)
	}
	return t.Delegate.GetTokenByValueTypeExpires(ctx, value, tokenType, expires)
}

func NewTokenStoreDecorator(db database.Dbx) *TokenStoreDecorator {
	delegate := NewPostgresTokenStore(db)
	return &TokenStoreDecorator{
		Delegate: delegate,
	}
}

func (t *TokenStoreDecorator) Cleanup() {
	t.WithTxFunc = nil
	t.DeleteTokenFunc = nil
	t.GetTokenFunc = nil
	t.SaveTokenFunc = nil
	t.VerifyTokenStorageFunc = nil

}

func (t *TokenStoreDecorator) WithTx(dbx database.Dbx) *TokenStoreDecorator {
	if t.WithTxFunc != nil {
		return t.WithTxFunc(dbx)
	}
	return &TokenStoreDecorator{
		Delegate: t.Delegate.WithTx(dbx),
	}
}

// DeleteToken implements DbTokenStoreInterface.
func (t *TokenStoreDecorator) DeleteToken(ctx context.Context, token string) error {
	if t.DeleteTokenFunc != nil {
		return t.DeleteTokenFunc(ctx, token)
	}
	return t.Delegate.DeleteToken(ctx, token)

}

// GetToken implements DbTokenStoreInterface.
func (t *TokenStoreDecorator) GetToken(ctx context.Context, token string) (*models.Token, error) {
	if t.GetTokenFunc != nil {
		return t.GetTokenFunc(ctx, token)
	}
	return t.Delegate.GetToken(ctx, token)
}

// SaveToken implements DbTokenStoreInterface.
func (t *TokenStoreDecorator) SaveToken(ctx context.Context, token *CreateTokenDTO) error {
	if t.SaveTokenFunc != nil {
		return t.SaveTokenFunc(ctx, token)
	}
	return t.Delegate.SaveToken(ctx, token)
}

var _ DbTokenStoreInterface = (*TokenStoreDecorator)(nil)
