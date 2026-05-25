package database

import (
	"context"
	"testing"
)

// TestWithNewDatabase_IsConnected verifies that WithNewDatabase produces a
// working connection to a freshly-cloned database.
func TestWithNewDatabase_IsConnected(t *testing.T) {
	WithNewDatabase(t, func(ctx context.Context, db Dbx) {
		var n int
		if err := db.QueryRow(ctx, "SELECT 1").Scan(&n); err != nil {
			t.Fatalf("SELECT 1: %v", err)
		}
		if n != 1 {
			t.Errorf("SELECT 1 = %d, want 1", n)
		}
	})
}

// TestWithNewDatabase_HasMigratedSchema verifies that all migrations ran on the
// template database by checking that the ledger schema and seeded system accounts exist.
func TestWithNewDatabase_HasMigratedSchema(t *testing.T) {
	WithNewDatabase(t, func(ctx context.Context, db Dbx) {
		// Verify ledger schema exists.
		var schemaCount int
		if err := db.QueryRow(ctx,
			`SELECT count(*) FROM information_schema.schemata WHERE schema_name = 'ledger'`,
		).Scan(&schemaCount); err != nil {
			t.Fatalf("schema query: %v", err)
		}
		if schemaCount != 1 {
			t.Errorf("ledger schema count = %d, want 1 (run migrations on playground_test)", schemaCount)
		}

		// Verify seeded system accounts exist.
		var accountCount int
		if err := db.QueryRow(ctx,
			`SELECT count(*) FROM ledger.accounts WHERE code IN ($1, $2)`,
			"system:points_issuance", "system:game_escrow",
		).Scan(&accountCount); err != nil {
			t.Fatalf("system accounts query: %v", err)
		}
		if accountCount != 2 {
			t.Errorf("system account count = %d, want 2", accountCount)
		}
	})
}
