package test

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/pool"
)

func DbSetup() (context.Context, bob.DB, *pgxpool.Pool) {
	// migrator.Migrate()

	var (
		ctx context.Context = context.Background()
		pl  *pgxpool.Pool   = pool.CreatePool(ctx, "postgres://postgres@localhost:5432/authgo_test?sslmode=disable")

		dbx bob.DB = bob.NewDB(stdlib.OpenDBFromPool(pl))
	)

	return ctx, dbx, pl
}
