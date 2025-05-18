package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/modules/paymentmodule"
	"github.com/tkahng/authgo/internal/modules/rbacmodule"
)

func NewStripeCmd() *cobra.Command {
	stripeCmd.AddCommand(stripeSyncCmd, stripeRolesCmd)
	return stripeCmd
}

var stripeCmd = &cobra.Command{
	Use:   "stripe",
	Short: "stripe",
}

var stripeSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "stripe sync",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		dbconf := conf.GetConfig[conf.DBConfig]()
		stripeconfig := conf.GetConfig[conf.StripeConfig]()

		dbx := database.CreateQueries(ctx, dbconf.DatabaseUrl)
		store := paymentmodule.NewPostgresPaymentStore(dbx)
		client := paymentmodule.NewPaymentClient(stripeconfig)
		rbacStore := rbacmodule.NewPostgresRBACStore(dbx)
		service := paymentmodule.NewPaymentService(client, store, rbacStore)

		return service.UpsertPriceProductFromStripe(ctx)
	},
}

var stripeRolesCmd = &cobra.Command{
	Use:   "role",
	Short: "stripe role",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		dbconf := conf.GetConfig[conf.DBConfig]()
		stripeconfig := conf.GetConfig[conf.StripeConfig]()

		dbx := database.CreateQueries(ctx, dbconf.DatabaseUrl)
		store := paymentmodule.NewPostgresPaymentStore(dbx)
		client := paymentmodule.NewPaymentClient(stripeconfig)
		rbacStore := rbacmodule.NewPostgresRBACStore(dbx)
		service := paymentmodule.NewPaymentService(client, store, rbacStore)
		return service.SyncPerms(ctx)
	},
}
