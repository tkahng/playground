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

var _ GamingStore = (*DBGamingStore)(nil)

type GamingStore interface {
	// friendships
	GamingFriendshipStore
	// players
	GamingPlayerStore
	// games
	RpsGameStore
	WithTx(db database.Dbx) *DBGamingStore
}
