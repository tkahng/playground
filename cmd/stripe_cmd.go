package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/core"
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

		_, dbx := core.NewPoolAndBobFromConf(ctx, dbconf)
		service := core.NewStripeServiceFromConf(stripeconfig)

		return service.UpsertPriceProductFromStripe(ctx, dbx)
	},
}

var stripeRolesCmd = &cobra.Command{
	Use:   "role",
	Short: "stripe role",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		dbconf := conf.GetConfig[conf.DBConfig]()
		stripeconfig := conf.GetConfig[conf.StripeConfig]()

		_, db := core.NewPoolAndBobFromConf(ctx, dbconf)
		service := core.NewStripeServiceFromConf(stripeconfig)
		return service.SyncRoles(ctx, db)
	},
}
