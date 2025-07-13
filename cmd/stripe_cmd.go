package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/payment"
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

		dbx := database.CreateQueriesContext(ctx, dbconf.DatabaseUrl)
		adapter := stores.NewStorageAdapter(dbx)
		client := payment.NewPaymentClient(stripeconfig)
		service := services.NewPaymentService(client, adapter)

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

		dbx := database.CreateQueriesContext(ctx, dbconf.DatabaseUrl)
		adapter := stores.NewStorageAdapter(dbx)
		client := payment.NewPaymentClient(stripeconfig)
		service := services.NewPaymentService(client, adapter)
		// Create the payment service
		return service.SyncPerms(ctx)
	},
}
