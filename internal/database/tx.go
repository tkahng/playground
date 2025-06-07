package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var _ Dbx = (*txQueries)(nil)

type txQueries struct {
	db pgx.Tx
}

// QueryRow implements Dbx.
func (v *txQueries) QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row {
	return v.db.QueryRow(ctx, sql, arguments...)
}

// Begin implements Dbx.
func (v *txQueries) Begin(ctx context.Context) (pgx.Tx, error) {
	return v.db.Begin(ctx)
}

// SendBatch implements Dbx.
func (v *txQueries) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return v.db.SendBatch(ctx, b)
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

func (v *txQueries) RunInTransaction(ctx context.Context, fn func(Dbx) error) error {

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

func WithTx(ctx context.Context, dbx Dbx, fn func(tx Dbx) error) error {
	tx, err := dbx.Begin(ctx)
	if err != nil {
		slog.Error("error starting transaction", slog.Any("error", err))
		return fmt.Errorf("error starting transaction: %w", err)
	}
	// Ensure the transaction will be rolled back if not committed
	defer tx.Rollback(ctx)

	err = fn(&txQueries{db: tx})
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			slog.Error("error committing transaction", slog.Any("error", err))
			return fmt.Errorf("error committing transaction: %w", err)
		}
	}

	return err
}
