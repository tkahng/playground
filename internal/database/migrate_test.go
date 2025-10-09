package database_test

import (
	"testing"

	"github.com/tkahng/playground/internal/database"
)

func TestNewMigrator(t *testing.T) {
	t.Run("Valid url should return migrator.", func(t *testing.T) {
		config := &database.MigratorConfig{
			DatabaseUrl: "postgres://postgres:postgres@localhost:5432/playground_migrator_test?sslmode=disable",
		}
		got := database.NewMigrator(config)
		if got == nil {
			t.Errorf("NewMigrator() = %v", got)
		}
	})
}

func TestDbmateMigrator_CreateAndMigrate(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		wantErr bool
	}{
		{
			name:    "create and migrate succeeded",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			var m database.Migrator
			m = database.NewMigrator(&database.MigratorConfig{
				DatabaseUrl: "postgres://postgres:postgres@localhost:5432/playground_migrator_test?sslmode=disable",
			})
			gotErr := m.CreateAndMigrate()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateAndMigrate() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateAndMigrate() succeeded unexpectedly")
			}
		})
	}
}
