package stores

import (
	"github.com/tkahng/authgo/internal/database"

	"github.com/tkahng/authgo/internal/services"
)

type DbPaymentStore struct {
	*DbStripeStore
	*DbRBACStore
	*DbTeamStore
}

func NewDbPaymentStore(db database.Dbx) *DbPaymentStore {
	return &DbPaymentStore{
		DbStripeStore: NewDbStripeStore(db),
		DbRBACStore:   NewDbRBACStore(db),
		DbTeamStore:   NewDbTeamStore(db),
	}
}

var _ services.PaymentStore = (*DbPaymentStore)(nil)
var _ services.PaymentTeamStore = (*DbTeamStore)(nil)
var _ services.PaymentRbacStore = (*DbRBACStore)(nil)
