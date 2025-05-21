package stores

import (
	"github.com/tkahng/authgo/internal/database"

	"github.com/tkahng/authgo/internal/services"
)

type PaymentStore struct {
	*PostgresStripeStore
	*PostgresRBACStore
	*PostgresTeamStore
}

func NewPostgresPaymentStore(db database.Dbx) *PaymentStore {
	return &PaymentStore{
		PostgresStripeStore: NewPostgresStripeStore(db),
		PostgresRBACStore:   NewPostgresRBACStore(db),
		PostgresTeamStore:   NewPostgresTeamStore(db),
	}
}

var _ services.PaymentStore = (*PaymentStore)(nil)
var _ services.PaymentTeamStore = (*PostgresTeamStore)(nil)
var _ services.PaymentRbacStore = (*PostgresRBACStore)(nil)
