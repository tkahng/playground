package database_test

import (
	"testing"

	"github.com/tkahng/playground/internal/database"
)

func TestNewMigrator(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		config *database.MigratorConfig
		want   database.Migrator
	}{
		struct {
			name   string
			config *database.MigratorConfig
			want   database.Migrator
		}{
			name: "Valid url should return migrator.",
			config: &database.MigratorConfig{
				DatabaseUrl: "postgres://postgres:postgres@localhost:5432/playground_migrator_test?sslmode=disable",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := database.NewMigrator(tt.config)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("NewMigrator() = %v, want %v", got, tt.want)
			}
		})
	}
}
