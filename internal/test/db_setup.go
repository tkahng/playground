package test

import (
	"context"
	"errors"
	"sync"

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
