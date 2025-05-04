package test

import (
	"context"

	"github.com/tkahng/authgo/internal/db"
)

func DbSetup() (context.Context, *db.Queries) {
	var (
		ctx context.Context = context.Background()
		dbx *db.Queries     = db.CreatePool(ctx, "postgres://postgres@localhost:5432/authgo_test?sslmode=disable")
	)

	return ctx, dbx
}
