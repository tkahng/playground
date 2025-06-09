package stores

import (
	"github.com/tkahng/authgo/internal/database"
)

type DbAuthStore struct {
	db database.Dbx
	*DbAccountStore
	*DbUserStore
	*DbTokenStore
}

func (s *DbAuthStore) WithTx(dbx database.Dbx) *DbAuthStore {
	return &DbAuthStore{
		db:             dbx,
		DbAccountStore: s.DbAccountStore,
		DbUserStore:    s.DbUserStore,
		DbTokenStore:   s.DbTokenStore,
	}
}

func NewDbAuthStore(db database.Dbx) *DbAuthStore {
	return &DbAuthStore{
		db:             db,
		DbAccountStore: NewDbAccountStore(db),
		DbUserStore:    NewDbUserStore(db),
		DbTokenStore:   NewPostgresTokenStore(db),
	}
}

func (s *DbAuthStore) RunInTransaction(
	fn func(store *DbAuthStore) error,
) error {
	return s.db.RunInTx(func(tx database.Dbx) error {
		store := s.WithTx(tx)
		return fn(store)
	})
}
