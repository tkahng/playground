package cmd

import (
	"fmt"
	"net/url"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/spf13/cobra"
	database "github.com/tkahng/authgo/db"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/db/migrator"
)

func init() {
	migrateCmd.AddCommand(upCmd)
	migrateCmd.AddCommand(resetCmd)
}

func NewMigrateCmd() *cobra.Command {

	migrateCmd.AddCommand(upCmd)
	migrateCmd.AddCommand(resetCmd)
	return migrateCmd

}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate",
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "migrate up",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := conf.AppConfigGetter()
		u, err := url.Parse(cfg.Db.DatabaseUrl)
		if err != nil {
			return err
		}
		db := dbmate.New(u)
		db.FS = database.Migrations
		db.MigrationsDir = []string{"./migrations"}
		fmt.Println("Migrations:")
		migrations, err := db.FindMigrations()
		if err != nil {
			return fmt.Errorf("error at error: %w", err)
		}
		for _, m := range migrations {
			fmt.Println(m.Version, m.FilePath)
		}
		fmt.Println("\nApplying...")
		err = db.CreateAndMigrate()
		if err != nil {
			return fmt.Errorf("error at error: %w", err)
		}
		return nil

	},
}
var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "migrate reset",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := conf.AppConfigGetter()
		migrator.Migrate(cfg.Db.DatabaseUrl)
		return nil
	},
}

// func NewUpCmd() *cobra.Command {
// 	command := &cobra.Command{
// 		Use:   "up",
// 		Short: "migrate up",
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			cfg := conf.AppConfigGetter()
// 			u, err := url.Parse(cfg.Db.DatabaseUrl)
// 			if err != nil {
// 				return err
// 			}
// 			db := dbmate.New(u)
// 			db.FS = database.Migrations
// 			db.MigrationsDir = []string{"./migrations"}
// 			fmt.Println("Migrations:")
// 			migrations, err := db.FindMigrations()
// 			if err != nil {
// 				return fmt.Errorf("error at error: %w", err)
// 			}
// 			for _, m := range migrations {
// 				fmt.Println(m.Version, m.FilePath)
// 			}
// 			fmt.Println("\nApplying...")
// 			err = db.CreateAndMigrate()
// 			if err != nil {
// 				return fmt.Errorf("error at error: %w", err)
// 			}
// 			return nil

// 		},
// 	}
// 	return command
// }

// func ResetCmd() *cobra.Command {
// 	command := &cobra.Command{
// 		Use:   "reset",
// 		Short: "migrate reset",
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			cfg := conf.AppConfigGetter()
// 			migrator.Migrate(cfg.Db.DatabaseUrl)
// 			return nil
// 		},
// 	}
// 	return command
// }
