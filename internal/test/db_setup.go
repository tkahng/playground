package test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/tkahng/authgo/internal/database"
)

var (
	EndTestErr  = errors.New("end test. rollback transaction")
	ctxInstance context.Context
	ctxOnce     sync.Once
	dbx         *database.Queries
)

func DbSetup() (context.Context, *database.Queries) {
	ctxOnce.Do(func() {
		ctxInstance = context.Background()
		dbx = database.CreateQueries(ctxInstance, "postgres://postgres:postgres@localhost:5432/authgo_test?sslmode=disable")
	})
	return ctxInstance, dbx
}

func WithTx(t *testing.T, fn func(ctx context.Context, db database.Dbx)) {
	ctx := context.Background()
	tx, err := dbx.Pool().Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	// panic handle
	defer func() {
		if err := recover(); err != nil {
			err := tx.Rollback(ctx)
			if err != nil {
				return
			}
			t.Fatal(err)
		}
	}()
	fn(ctx, database.NewTxQueries(tx))
}
