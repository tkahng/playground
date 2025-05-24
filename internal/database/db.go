package database

import (
	"context"
	"fmt"

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
}

// type TxFunc

var _ Dbx = (*Queries)(nil)

type Queries struct {
	db *pgxpool.Pool
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

	tx, err := v.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	// Ensure the transaction will be rolled back if not committed
	defer tx.Rollback(ctx)

	err = fn(&txQueries{db: tx})
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("error committing transaction: %w", err)
		}
	}

	return err
}
