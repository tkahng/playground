package stores

import (
	"context"

	"github.com/tkahng/authgo/internal/database"

	"github.com/tkahng/authgo/internal/services"
)

type PostgresAuthStore struct {
	db database.Dbx
	*PostgresAccountStore
	*PostgresUserStore
	*PostgresTokenStore
}

func (s *PostgresAuthStore) WithTx(dbx database.Dbx) services.AuthStore {
	return &PostgresAuthStore{
		db:                   dbx,
		PostgresAccountStore: s.PostgresAccountStore,
		PostgresUserStore:    s.PostgresUserStore,
		PostgresTokenStore:   s.PostgresTokenStore,
	}
}

func NewPostgresAuthStore(db database.Dbx) *PostgresAuthStore {
	return &PostgresAuthStore{
		db:                   db,
		PostgresAccountStore: NewPostgresUserAccountStore(db),
		PostgresUserStore:    NewPostgresUserStore(db),
		PostgresTokenStore:   NewPostgresTokenStore(db),
	}
}

var _ services.AuthStore = (*PostgresAuthStore)(nil)

func (s *PostgresAuthStore) RunInTransaction(
	ctx context.Context,
	fn func(store services.AuthStore) error,
) error {
	return s.db.RunInTransaction(ctx, func(tx database.Dbx) error {
		store := s.WithTx(tx)
		return fn(store)
	})
}
