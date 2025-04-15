package cmd

import (
	"fmt"
	"net/url"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/spf13/cobra"
	"github.com/tkahng/authgo/internal/conf"
	database "github.com/tkahng/authgo/internal/db"
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
		cfg := conf.GetConfig[conf.DBConfig]()
		u, err := url.Parse(cfg.DatabaseUrl)
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
		cfg := conf.GetConfig[conf.DBConfig]()
		u, err := url.Parse(cfg.DatabaseUrl)
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
		db.Drop()
		err = db.CreateAndMigrate()
		if err != nil {
			return fmt.Errorf("error at error: %w", err)
		}
		return nil
	},
}
