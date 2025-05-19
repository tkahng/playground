package stores

import (
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/modules/authmodule"
)

type PostgresAuthStore struct {
	*PostgresAccountStore
	*PostgresUserStore
	*PostgresTokenStore
}

func NewPostgresAuthStore(db database.Dbx) *PostgresAuthStore {
	return &PostgresAuthStore{
		PostgresAccountStore: NewPostgresUserAccountStore(db),
		PostgresUserStore:    NewPostgresUserStore(db),
		PostgresTokenStore:   NewPostgresTokenStore(db),
	}
}

var _ authmodule.AuthStore = (*PostgresAuthStore)(nil)
