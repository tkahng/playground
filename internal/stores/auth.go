package stores

import (
	"github.com/tkahng/authgo/internal/database"
)

type AuthStoreAdapterInterface interface {
	// WithTx(dbx database.Dbx) AuthStore
	// RunInTransaction(fn func(store AuthStore) error) error
	User() DbAuthStore
	Account() DbAccountStore
	Token() DbTokenStore
}

type AuthStoreAdapter struct {
	db      database.Dbx
	user    *DbUserStore
	account *DbAccountStore
	token   *DbTokenStore
}

func NewAuthStoreAdapter(db database.Dbx) *AuthStoreAdapter {
	return &AuthStoreAdapter{
		db:      db,
		user:    NewDbUserStore(db),
		account: NewDbAccountStore(db),
		token:   NewPostgresTokenStore(db),
		// teamGroup:
	}
}

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
