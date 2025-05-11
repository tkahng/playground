package test

import (
	"context"
	"sync"

	"github.com/tkahng/authgo/internal/db"
)

var (
	ctxInstance context.Context
	ctxOnce     sync.Once
	dbx         *db.Queries
)

func DbSetup() (context.Context, *db.Queries) {
	ctxOnce.Do(func() {
		ctxInstance = context.Background()
		dbx = db.CreateQueries(ctxInstance, "postgres://postgres@localhost:5432/authgo_test?sslmode=disable")
	})
	return ctxInstance, dbx
}
