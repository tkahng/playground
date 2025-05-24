package jobs_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Get test database URL from environment
	// dbURL := os.Getenv("TEST_DATABASE_URL")
	dbURL := "postgres://postgres:postgres@localhost:5432/authgo_test?sslmode=disable"
	// if dbURL == "" {
	// }

	// Create connection pool
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		panic(err)
	}

	testPool, err = pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		panic(err)
	}
	defer testPool.Close()

	os.Exit(m.Run())
}

func withTx(t *testing.T, fn func(ctx context.Context, db *pgxpool.Pool)) {
	ctx := context.Background()
	tx, err := testPool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	fn(ctx, testPool)
}

// migrateDB ensures test database schema is ready

// setupTestTx creates a transaction for each test
func setupTestTx(ctx context.Context, t *testing.T) pgx.Tx {
	tx, err := testPool.Begin(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	t.Cleanup(func() {
		if t.Failed() {
			t.Log("Test failed, rolling back transaction")
			tx.Rollback(ctx)
		} else {
			tx.Rollback(ctx) // Always rollback to discard changes
		}
	})

	return tx
}
