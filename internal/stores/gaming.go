package stores

import (
	"github.com/tkahng/playground/internal/database"
)

type DBGamingStore struct {
	db database.Dbx
}

func NewDBGamingStore(db database.Dbx) *DBGamingStore {
	return &DBGamingStore{
		db: db,
	}
}

func (s *DBGamingStore) WithTx(db database.Dbx) *DBGamingStore {
	return &DBGamingStore{
		db: db,
	}
}
