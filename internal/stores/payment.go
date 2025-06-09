package stores

import (
	"github.com/tkahng/authgo/internal/database"
)

type DbPaymentStore struct {
	*DbStripeStore
	*DbRbacStore
	*DbTeamStore
}

func NewDbPaymentStore(db database.Dbx) *DbPaymentStore {
	return &DbPaymentStore{
		DbStripeStore: NewDbStripeStore(db),
		DbRbacStore:   NewDbRBACStore(db),
		DbTeamStore:   NewDbTeamStore(db),
	}
}
