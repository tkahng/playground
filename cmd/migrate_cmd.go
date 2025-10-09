package cmd

import (
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/spf13/cobra"
	"github.com/tkahng/playground/internal/conf"
	database "github.com/tkahng/playground/internal/database"
)

func NewMigrateCmd() *cobra.Command {
	migrateCmd.AddCommand(upCmd)
	migrateCmd.AddCommand(resetCmd)
	migrateCmd.AddCommand(makeSchema)
	return migrateCmd

}

// nolint:exhaustruct
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate",
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "migrate up",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := conf.GetConfig[conf.DBConfig]()
		mConfig := database.MigratorConfig{
			AutoDumpSchema: false,
			DatabaseUrl:    cfg.GetUrl(),
		}
		migrator := database.NewMigrator(&mConfig)
		return migrator.CreateAndMigrate()
	},
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "migrate reset",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := conf.GetConfig[conf.DBConfig]()
		mConfig := database.MigratorConfig{
			AutoDumpSchema: false,
			DatabaseUrl:    cfg.GetUrl(),
		}
		migrator := database.NewMigrator(&mConfig)
		return migrator.Reset()
	},
}
var makeSchema = &cobra.Command{
	Use:   "schema",
	Short: "migrate schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := conf.GetConfig[conf.DBConfig]()
		mConfig := database.MigratorConfig{
			AutoDumpSchema: false,
			DatabaseUrl:    cfg.GetUrl(),
		}

		migrator := database.NewMigrator(&mConfig)
		return migrator.DumpSchema()
	},
}
