package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/core"
)

func NewStripeCmd() *cobra.Command {
	stripeCmd.AddCommand(stripeRolesCmd)
	return stripeCmd
}

var stripeCmd = &cobra.Command{
	Use:   "stripe",
	Short: "stripe",
}

var stripeRolesCmd = &cobra.Command{
	Use:   "sync",
	Short: "stripe sync",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		conf := conf.AppConfigGetter()

		dbx := core.NewBobFromConf(ctx, conf.Db)
		app := core.NewBaseApp(dbx, conf)

		return app.Payment().UpsertPriceProductFromStripe(ctx, dbx)
	},
}
