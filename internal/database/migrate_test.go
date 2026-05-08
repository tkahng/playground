package database_test

import (
	"testing"

	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/database"
)

func TestNewMigrator(t *testing.T) {
	t.Run("Valid url should return migrator.", func(t *testing.T) {
		config := &database.MigratorConfig{
			DatabaseURL: "postgres://postgres:postgres@localhost:5432/playground_migrator_test?sslmode=disable",
		}
		got := database.NewMigrator(config)
		if got == nil {
			t.Errorf("NewMigrator() = %v", got)
		}
	})
}

func TestDbmateMigrator_CreateAndMigrate_Status_Drop(t *testing.T) {
	t.Run("create and migrate, status and drop succeeded", func(t *testing.T) {
		dbName := uuid.NewString()
		cfg := conf.GetConfig[conf.DBConfig]()
		cfg.Db = dbName
		mConfig := database.MigratorConfig{
			AutoDumpSchema: false,
			DatabaseURL:    cfg.GetDatabaseURL(),
		}
		m := database.NewMigrator(&mConfig)
		defer func() {
			err := m.Drop()
			if err != nil {
				t.Errorf("Drop() failed: %v", err)
			}
		}()

		gotErr := m.CreateAndMigrate()
		if gotErr != nil {
			t.Errorf("Create() failed: %v", gotErr)
		}
		status, gotErr := m.Status()
		if gotErr != nil {
			t.Errorf("Status() failed: %v", gotErr)
		}
		if status != 0 {
			t.Errorf("Status() = %v, want 0", status)
		}
	})
}
