package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/services/payment"
	"github.com/tkahng/authgo/internal/services/rbac"
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
		store := payment.NewPostgresPaymentStore(dbx)
		client := payment.NewPaymentClient(stripeconfig)
		rbacStore := rbac.NewPostgresRBACStore(dbx)
		service := payment.NewPaymentService(client, store, rbacStore)

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
		store := payment.NewPostgresPaymentStore(dbx)
		client := payment.NewPaymentClient(stripeconfig)
		rbacStore := rbac.NewPostgresRBACStore(dbx)
		service := payment.NewPaymentService(client, store, rbacStore)
		return service.SyncPerms(ctx)
	},
}
