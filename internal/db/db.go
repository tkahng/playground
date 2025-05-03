package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBTX interface {
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

var _ DBTX = (*Queries)(nil)

type Queries struct {
	pool DBTX
}

func NewQueries(pool DBTX) *Queries {
	return &Queries{pool: pool}
}

func (v *Queries) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return v.pool.Query(ctx, sql, args...)
}

func (v *Queries) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return v.pool.Exec(ctx, sql, args...)
}
