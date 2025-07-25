package cmd

import (
	"fmt"
	"net/url"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/spf13/cobra"
	"github.com/tkahng/playground/internal/conf"
	database "github.com/tkahng/playground/internal/database"
)

func NewMigrateCmd() *cobra.Command {

	migrateCmd.AddCommand(upCmd)
	migrateCmd.AddCommand(testUpCmd)
	migrateCmd.AddCommand(resetCmd)
	return migrateCmd

}

// nolint:exhaustruct
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate",
}

// nolint:exhaustruct
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "migrate up",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := conf.GetConfig[conf.DBConfig]()
		return migrate(cfg.DatabaseUrl)
	},
}

// nolint:exhaustruct
var testUpCmd = &cobra.Command{
	Use:   "testup",
	Short: "migrate testup",
	RunE: func(cmd *cobra.Command, args []string) error {
		return migrate("postgres://postgres:postgres@localhost:5432/playground_test?sslmode=disable")
	},
}

func migrate(uri string) error {
	u, err := url.Parse(uri)
	if err != nil {
		return err
	}
	db := dbmate.New(u)
	db.FS = database.Migrations
	db.MigrationsDir = []string{"./migrations"}
	db.SchemaFile = "./internal/database/schema.sql"
	fmt.Println("Migrations:")
	migrations, err := db.FindMigrations()
	if err != nil {
		return err
	}
	for _, m := range migrations {
		fmt.Println(m.Version, m.FilePath)
	}
	fmt.Println("\nApplying...")
	err = db.CreateAndMigrate()
	if err != nil {
		return err
	}
	return nil
}

// nolint:exhaustruct
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
		db.SchemaFile = "./internal/database/schema.sql"
		fmt.Println("Migrations:")
		migrations, err := db.FindMigrations()
		if err != nil {
			return err
		}
		for _, m := range migrations {
			fmt.Println(m.Version, m.FilePath)
		}
		fmt.Println("\nApplying...")
		err = db.Drop()
		if err != nil {
			return err
		}
		err = db.CreateAndMigrate()
		if err != nil {
			return err
		}
		return nil
	},
}
