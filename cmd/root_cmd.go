package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "hugo",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hugo Static Site Generator v0.9 -- HEAD")
	},
}

func init() {
	rootCmd.AddCommand(NewServeCmd(), NewMigrateCmd(), NewSeedCmd(), NewSuperuserCmd(), NewStripeCmd())
}

func Execute() {
	rootCmd.Execute()
}
