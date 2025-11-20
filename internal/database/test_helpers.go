package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"testing"

	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/tools/logger"
)

var (
	ErrEndTest  = errors.New("end test. rollback transaction")
	ctxInstance context.Context
	ctxOnce     sync.Once
	dbx         *Queries
)

func DbSetup(cfg *conf.EnvConfig) (context.Context, *Queries) {
	ctxOnce.Do(func() {
		ctxInstance = context.Background()
		dbx = CreateNewQueriesContext(ctxInstance, cfg.Db.GetDatabaseUrl())
	})
	return ctxInstance, dbx
}

func WithSingletonTestTx(t *testing.T, fn func(ctx context.Context, db Dbx)) {
	DbSetup(conf.ZeroEnvConfig())
	ctx := ctxInstance
	tx, err := dbx.BeginTx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	//nolint:errcheck
	defer tx.Rollback(ctx)
	// panic handle
	defer func() {
		if recErr := recover(); recErr != nil {
			slog.ErrorContext(ctx, "recovered from panic in transaction.", slog.Any("error", fmt.Sprint(recErr)), slog.Any("stacktrace", string(debug.Stack())))
			rollBackErr := tx.Rollback(ctx)
			if rollBackErr != nil {
				slog.ErrorContext(ctx, "recovered from panic in transaction.", slog.Any("error", fmt.Sprint(recErr)), slog.Any("stacktrace", string(debug.Stack())))
				t.Error(rollBackErr)
			}
			t.Fatal(fmt.Sprint(recErr))
		}
	}()
	fn(ctx, NewTxQueries(tx))
}

// WithNewTestTx creates a new pool connection, runs the test within that transactions, rolls back, and closes the pool.
func WithNewTestTx(t *testing.T, fn func(ctx context.Context, db Dbx)) {
	t.Helper()
	// TODO: add context timeout
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// ctx := t.Context()
	_ = logger.GetDefaultLogger()
	ctx := context.Background()
	cfg := conf.ZeroEnvConfig()
	dbx := CreateNewQueriesContext(ctx, cfg.Db.GetDatabaseUrl())
	tx, beginErr := dbx.BeginTx(ctx)
	if beginErr != nil {
		t.Fatal(beginErr)
	}
	defer dbx.Close()
	//nolint:errcheck
	defer tx.Rollback(ctx)
	// panic handle
	defer func() {
		if recErr := recover(); recErr != nil {
			slog.ErrorContext(ctx, "recovered from panic in transaction.", slog.Any("error", fmt.Sprint(recErr)), slog.Any("stacktrace", string(debug.Stack())))
			rollBackErr := tx.Rollback(ctx)
			if rollBackErr != nil {
				slog.ErrorContext(ctx, "recovered from panic in transaction.", slog.Any("error", fmt.Sprint(recErr)), slog.Any("stacktrace", string(debug.Stack())))
				t.Error(rollBackErr)
			}
			t.Fatal(fmt.Sprint(recErr))
		}
	}()
	fn(ctx, NewTxQueries(tx))
}
