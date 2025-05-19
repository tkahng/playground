package modules

// type Stores struct {
// 	Payment paymentmodule.PaymentStore
// 	Rbac    rbacmodule.RBACStore
// 	Team    teammodule.TeamStore
// 	Token   tokenmodule.TokenStore
// 	Account useraccountmodule.UserAccountStore
// 	User    usermodule.UserStore
// 	Auth    authmodule.AuthStore
// }

// func NewPostgresStores(dbx database.Dbx) *Stores {
// 	payment := paymentmodule.NewPostgresPaymentStore(dbx)
// 	rbac := rbacmodule.NewPostgresRBACStore(dbx)
// 	team := teammodule.NewPostgresTeamStore(dbx)
// 	token := tokenmodule.NewPostgresTokenStore(dbx)
// 	account := useraccountmodule.NewPostgresUserAccountStore(dbx)
// 	user := usermodule.NewPostgresUserStore(dbx)
// 	auth := authmodule.NewAuthStore(
// 		token,
// 		user,
// 		account,
// 	)
// 	return &Stores{
// 		Payment: payment,
// 		Rbac:    rbac,
// 		Team:    team,
// 		Token:   token,
// 		Account: account,
// 		User:    user,
// 		Auth:    auth,
// 	}
// }
