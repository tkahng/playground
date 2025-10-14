package test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/tkahng/playground/internal/database"
)

var (
	ErrEndTest  = errors.New("end test. rollback transaction")
	ctxInstance context.Context
	ctxOnce     sync.Once
	dbx         *database.Queries
)

func SingletonDbSetupTest(t *testing.T) (context.Context, *database.Queries) {
	t.Helper()
	return dbSetup()
}

func dbSetup() (context.Context, *database.Queries) {
	ctxOnce.Do(func() {
		ctxInstance = context.Background()
		dbx = database.CreateSingletonQueriesContext(ctxInstance, "postgres://postgres:postgres@localhost:5432/playground_test?sslmode=disable")
	})
	return ctxInstance, dbx
}

func WithSingletonTx(t *testing.T, fn func(ctx context.Context, db database.Dbx)) {
	t.Helper()
	ctx, dbx := dbSetup()
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
