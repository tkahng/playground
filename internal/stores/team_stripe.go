package stores

import (
	"github.com/tkahng/authgo/internal/database"
)

type DbTeamStripeStore struct {
	*DbTeamStore
	*DbStripeStore
}

func NewDbTeamStripeStore(db database.Dbx) *DbTeamStripeStore {
	return &DbTeamStripeStore{
		DbTeamStore:   NewDbTeamStore(db),
		DbStripeStore: NewDbStripeStore(db),
	}
}

func (p *DbTeamStripeStore) WithTx(tx database.Dbx) *DbTeamStripeStore {
	return &DbTeamStripeStore{
		DbTeamStore:   p.DbTeamStore.WithTx(tx),
		DbStripeStore: p.DbStripeStore.WithTx(tx),
	}
}
