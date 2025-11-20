package cmd

import (
	"errors"

	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/spf13/cobra"
	"github.com/tkahng/playground/internal/conf"
	database "github.com/tkahng/playground/internal/database"
)

func NewMigrateCmd() *cobra.Command {
	migrateCmd.AddCommand(upCmd)
	migrateCmd.AddCommand(resetCmd)
	migrateCmd.AddCommand(makeSchema)
	migrateCmd.AddCommand(rollBack)
	migrateCmd.AddCommand(newMigrations)
	return migrateCmd
}

//nolint:exhaustruct
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
			DatabaseUrl:    cfg.GetDatabaseUrl(),
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
			DatabaseUrl:    cfg.GetDatabaseUrl(),
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
			DatabaseUrl:    cfg.GetDatabaseUrl(),
		}

		migrator := database.NewMigrator(&mConfig)
		return migrator.DumpSchema()
	},
}
var rollBack = &cobra.Command{
	Use:   "rollback",
	Short: "migrate rollback",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := conf.GetConfig[conf.DBConfig]()
		mConfig := database.MigratorConfig{
			AutoDumpSchema: false,
			DatabaseUrl:    cfg.GetDatabaseUrl(),
		}

		migrator := database.NewMigrator(&mConfig)
		return migrator.Rollback()
	},
}
var newMigrations = &cobra.Command{
	Use:   "new",
	Short: "migrate new",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("missing migration name")
		}
		name := args[0]
		cfg := conf.GetConfig[conf.DBConfig]()
		mConfig := database.MigratorConfig{
			AutoDumpSchema: false,
			DatabaseUrl:    cfg.GetDatabaseUrl(),
		}

		migrator := database.NewMigrator(&mConfig)
		return migrator.NewMigration(name)
	},
}
