package cmd

import (
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/spf13/cobra"
	"github.com/tkahng/playground/internal/conf"
	database "github.com/tkahng/playground/internal/database"
)

func NewMigrateCmd() *cobra.Command {
	migrateCmd.PersistentFlags().Bool("test", false, "is for test?")
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
func migrateUp(cmd *cobra.Command, args []string) error {
	cfg := conf.GetConfig[conf.DBConfig]()
	isTest, err := cmd.Flags().GetBool("test")
	if err != nil {
		return err
	}
	mConfig := database.MigratorConfig{
		AutoDumpSchema: false,
	}
	if isTest {
		mConfig.DatabaseUrl = cfg.TestDatabaseUrl
	} else {
		mConfig.DatabaseUrl = cfg.DatabaseUrl
	}

	migrator := database.NewMigrator(&mConfig)
	return migrator.CreateAndMigrate()
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "migrate up",
	RunE: migrateUp,
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "migrate reset",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := conf.GetConfig[conf.DBConfig]()
		isTest, err := cmd.Flags().GetBool("test")
		if err != nil {
			return err
		}
		mConfig := database.MigratorConfig{
			AutoDumpSchema:  false,
		}
		if isTest {
			mConfig.DatabaseUrl = cfg.TestDatabaseUrl
		} else {
			mConfig.DatabaseUrl = cfg.DatabaseUrl
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
		isTest, err := cmd.Flags().GetBool("test")
		if err != nil {
			return err
		}
		mConfig := database.MigratorConfig{
			AutoDumpSchema:  true,
		}
		if isTest {
			mConfig.DatabaseUrl = cfg.TestDatabaseUrl
		} else {
			mConfig.DatabaseUrl = cfg.DatabaseUrl
		}
		
		migrator := database.NewMigrator(&mConfig)
		return migrator.DumpSchema()
	},
}



