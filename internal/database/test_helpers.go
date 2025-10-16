package database

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"testing"

	"github.com/tkahng/playground/internal/conf"
)

// WithNewTx creates a new pool connection, runs the test within that transaciton, rolls back, and closes the pool.
func WithNewTx(t *testing.T, fn func(ctx context.Context, db Dbx)) {
	t.Helper()
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	ctx := context.Background()
	cfg := conf.ZeroEnvConfig()
	dbx := CreateNewQueriesContext(ctx, cfg.Db.GetDatabaseUrl())
	tx, err := dbx.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer dbx.Close()
	// nolint:errcheck
	defer tx.Rollback(ctx)
	// panic handle
	defer func() {
		if recErr := recover(); recErr != nil {
			slog.ErrorContext(ctx, "recovered from panic in transaction.", slog.Any("error", fmt.Sprint(recErr)), slog.Any("stacktrace", string(debug.Stack())))
			err := tx.Rollback(ctx)
			if err != nil {
				t.Error(err)
			}
			t.Fatal(fmt.Sprint(recErr))
		}
	}()
	fn(ctx, NewTxQueries(tx))
}
