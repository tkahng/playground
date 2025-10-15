package test

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"testing"

	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/database"
)

const (
	TestDbUrl = "postgres://postgres:postgres@localhost:5432/playground_test?sslmode=disable"
)

var (
	ctxInstance context.Context
	ctxOnce     sync.Once
	dbx         *database.Queries
)

func SingletonDbSetup(t *testing.T) (context.Context, *database.Queries) {
	return singletonDbSetup(TestDbUrl)
}

func singletonDbSetup(url string) (context.Context, *database.Queries) {
	ctxOnce.Do(func() {
		ctxInstance = context.Background()
		dbx = database.CreateSingletonQueriesContext(ctxInstance, url)
	})
	return ctxInstance, dbx
}

func WithSingletonTx(t *testing.T, fn func(ctx context.Context, db database.Dbx)) {
	ctx, dbx := singletonDbSetup(TestDbUrl)
	tx, err := dbx.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)
	// panic handle
	defer func() {
		if recErr := recover(); recErr != nil {
			err := tx.Rollback(ctx)
			if err != nil {
				t.Error(err)
			}
			t.Fatal(recErr)
		}
	}()
	fn(ctx, database.NewTxQueries(tx))
}

// WithNewTx creates a new pool connection, runs the test within that transaciton, rolls back, and closes the pool.
func WithNewTx(t *testing.T, cfg *conf.EnvConfig, fn func(ctx context.Context, db database.Dbx)) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dbx := database.CreateNewQueriesContext(ctx, cfg.Db.GetDatabaseUrl())
	tx, err := dbx.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		dbx.Close()
	}()
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
	fn(ctx, database.NewTxQueries(tx))
}
