package test

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tkahng/authgo/internal/db"
)

func DbSetup() (context.Context, *db.Queries, *pgxpool.Pool) {
	var (
		ctx context.Context = context.Background()
		pl  *pgxpool.Pool   = db.CreatePool(ctx, "postgres://postgres@localhost:5432/authgo_test?sslmode=disable")
		dbx *db.Queries     = db.NewQueries(pl)
	)

	return ctx, dbx, pl
}
