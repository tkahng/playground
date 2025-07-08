package database

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Dbx interface {
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Begin(ctx context.Context) (pgx.Tx, error)
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	RunInTx(fn func(Dbx) error) error
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
	RunInTxContext(ctx context.Context, fn func(context.Context) error) error
}

// type TxFunc

var _ Dbx = (*Queries)(nil)

type Queries struct {
	db *pgxpool.Pool
}

func GetContextOrDefaultDbx(ctx context.Context, dbx Dbx) Dbx {
	tx := getContextTx(ctx)
	if tx != nil {
		return tx
	}
	return dbx
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

func (v *Queries) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return v.db.Query(ctx, sql, args...)
}

func (v *Queries) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return v.db.Exec(ctx, sql, args...)
}

// RunInTx implements Dbx.
func (v *Queries) RunInTx(fn func(Dbx) error) error {
	return WithTx(v, fn)
}

func (v *Queries) RunInTxContext(ctx context.Context, fn func(context.Context) error) error {
	return WithTxContext(v, ctx, fn)
}

var _ Dbx = (*txQueries)(nil)

type txQueries struct {
	db pgx.Tx
}

// RunInTxContext implements Dbx.
func (v *txQueries) RunInTxContext(ctx context.Context, fn func(context.Context) error) error {
	return WithTxContext(v, ctx, fn)
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

func (v *txQueries) RunInTx(fn func(Dbx) error) error {
	return WithTx(v, fn)
}

func WithTx(dbx Dbx, fn func(tx Dbx) error) error {
	ctx := context.Background() // Use the appropriate context as needed
	tx, err := dbx.Begin(ctx)
	if err != nil {
		slog.Error("error starting transaction", slog.Any("error", err))
		return err
	}

	defer func() {
		if err := recover(); err != nil {
			err := tx.Rollback(ctx)
			if err != nil {
				return
			}
		}
	}()

	err = fn(&txQueries{db: tx})
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			slog.ErrorContext(ctx, "error committing transaction", slog.Any("error", err))
			return err
		}
	} else {
		slog.ErrorContext(ctx, "error in transaction function", slog.Any("error", err))
		err := tx.Rollback(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "error rolling back transaction", slog.Any("error", err))
			return err
		}
	}

	return err
}
func WithTxContext(dbx Dbx, ctx context.Context, fn func(context.Context) error) error {
	tx, err := dbx.Begin(ctx)
	if err != nil {
		slog.Error("error starting transaction", slog.Any("error", err))
		return err
	}

	defer func() {
		if err := recover(); err != nil {
			err := tx.Rollback(ctx)
			if err != nil {
				return
			}
		}
	}()
	txCtx := setContextTx(ctx, &txQueries{db: tx})
	err = fn(txCtx)
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			slog.ErrorContext(ctx, "error committing transaction", slog.Any("error", err))
			return err
		}
	} else {
		slog.ErrorContext(ctx, "error in transaction function", slog.Any("error", err))
		err := tx.Rollback(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "error rolling back transaction", slog.Any("error", err))
			return err
		}
	}

	return err
}
