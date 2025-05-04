package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var _ Dbx = (*txQueries)(nil)

type txQueries struct {
	db pgx.Tx
}

func (v *txQueries) Commit(ctx context.Context) error {
	return v.db.Commit(ctx)
}

func NewTxQueries(tx pgx.Tx) *txQueries {
	return &txQueries{db: tx}
}

func (v *txQueries) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return v.db.Query(ctx, sql, args...)
}

func (v *txQueries) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return v.db.Exec(ctx, sql, args...)
}

func (v *txQueries) RunInTransaction(ctx context.Context, fn TxFunc) error {
	err := fn(v)
	return err
}
