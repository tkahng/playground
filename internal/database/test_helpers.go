package database

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"testing"

	"github.com/tkahng/playground/internal/conf"
)

// WithNewTestTx creates a new pool connection, runs the test within that transaciton, rolls back, and closes the pool.
func WithNewTestTx(t *testing.T, fn func(ctx context.Context, db Dbx)) {
	t.Helper()
	// TODO: add context timeout
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	ctx := t.Context()
	cfg := conf.ZeroEnvConfig()
	dbx := CreateNewQueriesContext(ctx, cfg.Db.GetDatabaseUrl())
	tx, beginErr := dbx.Begin(ctx)
	if beginErr != nil {
		t.Fatal(beginErr)
	}
	defer dbx.Close()
	// nolint:errcheck
	defer tx.Rollback(context.Background())
	// panic handle
	defer func() {
		if recErr := recover(); recErr != nil {
			slog.ErrorContext(ctx, "recovered from panic in transaction.", slog.Any("error", fmt.Sprint(recErr)), slog.Any("stacktrace", string(debug.Stack())))
			rollBackErr := tx.Rollback(context.Background())
			if rollBackErr != nil {
				slog.ErrorContext(ctx, "recovered from panic in transaction.", slog.Any("error", fmt.Sprint(recErr)), slog.Any("stacktrace", string(debug.Stack())))
				t.Error(rollBackErr)
			}
			t.Fatal(fmt.Sprint(recErr))
		}
	}()
	fn(ctx, NewTxQueries(tx))
}
