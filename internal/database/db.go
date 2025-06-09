package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Dbx interface {
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Begin(ctx context.Context) (pgx.Tx, error)
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	RunInTransaction(ctx context.Context, fn func(Dbx) error) error
	RunInTx(fn func(Dbx) error) error
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

// type TxFunc

var _ Dbx = (*Queries)(nil)

type Queries struct {
	db *pgxpool.Pool
}

func (v *Queries) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return v.db.QueryRow(ctx, sql, args...)
}

// Begin implements Dbx.
func (v *Queries) Begin(ctx context.Context) (pgx.Tx, error) {
	return v.db.Begin(ctx)
}

// SendBatch implements Dbx.
func (v *Queries) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return v.db.SendBatch(ctx, b)
}

func (v *Queries) Pool() *pgxpool.Pool {
	return v.db
}

func NewQueries(pool *pgxpool.Pool) *Queries {
	return &Queries{db: pool}
}

func (v *Queries) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return v.db.Query(ctx, sql, args...)
}

func (v *Queries) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return v.db.Exec(ctx, sql, args...)
}

func (v *Queries) RunInTransaction(ctx context.Context, fn func(Dbx) error) error {
	return WithTx(v, fn)
}

// RunInTx implements Dbx.
func (v *Queries) RunInTx(fn func(Dbx) error) error {
	return WithTx(v, fn)
}
