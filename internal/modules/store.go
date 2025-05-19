package modules

import (
	"log/slog"

	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/modules/paymentmodule"
	"github.com/tkahng/authgo/internal/modules/rbacmodule"
	"github.com/tkahng/authgo/internal/modules/teammodule"
	"github.com/tkahng/authgo/internal/modules/tokenmodule"
	"github.com/tkahng/authgo/internal/modules/useraccountmodule"
	"github.com/tkahng/authgo/internal/modules/usermodule"
)

type Stores struct {
	Payment paymentmodule.PaymentStore
	Rbac    rbacmodule.RBACStore
	Team    teammodule.TeamStore
	Token   tokenmodule.TokenStore
	Account useraccountmodule.UserAccountStore
	User    usermodule.UserStore
}

func NewStores(dbx database.Dbx, logger *slog.Logger) *Stores {
	return &Stores{
		Payment: paymentmodule.NewPostgresPaymentStore(dbx),
		Rbac:    rbacmodule.NewPostgresRBACStore(dbx),
		Team:    teammodule.NewPostgresTeamStore(dbx),
		Token:   tokenmodule.NewPostgresTokenStore(dbx),
		Account: useraccountmodule.NewPostgresUserAccountStore(dbx),
		User:    usermodule.NewPostgresUserStore(dbx),
	}
}
