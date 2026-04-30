package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "playground",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Playground")
	},
}

func init() {
	rootCmd.AddCommand(NewServeCmd(), NewMigrateCmd(), NewSeedCmd(), NewSuperuserCmd(), NewStripeCmd())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
