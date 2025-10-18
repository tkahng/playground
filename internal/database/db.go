package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Executor interface {
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

type Dbx interface {
	Executor

	Begin(ctx context.Context) (pgx.Tx, error)

	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)

	// Close
	//
	// Close calls close on the underlying pool.
	Close()
	// RunInTx
	//
	// Deprecated: use RunInTxContext
	RunInTx(fn func(Dbx) error) error
	RunInTxContext(ctx context.Context, fn func(context.Context) error) error
}

// type TxFunc

var _ Dbx = (*Queries)(nil)

type Queries struct {
	db *pgxpool.Pool
}

// Close calls close on the underlying pool.
func (v *Queries) Close() {
	slog.Info("close called on Queries. calling close on pool.")
	v.db.Close()

}

// Acquire implements Dbx.
func (v *Queries) Acquire(ctx context.Context) (c *pgxpool.Conn, err error) {
	return v.db.Acquire(ctx)
}

func (v *Queries) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return v.db.QueryRow(ctx, sql, args...)
}

func (v *Queries) Begin(ctx context.Context) (pgx.Tx, error) {
	return v.BeginTx(ctx, pgx.TxOptions{})
}

// BeginTx acquires a connection from the Pool and starts a transaction with pgx.TxOptions determining the transaction mode.
// Unlike database/sql, the context only affects the begin command. i.e. there is no auto-rollback on context cancellation.
func (v *Queries) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	return v.db.BeginTx(ctx, opts)
}

// SendBatch implements Dbx.
func (v *Queries) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return v.db.SendBatch(ctx, b)
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
	return WithTxContext(ctx, v, fn)
}

var _ Dbx = (*txQueries)(nil)

type txQueries struct {
	db pgx.Tx
}

func NewTxQueries(tx pgx.Tx) *txQueries {
	return &txQueries{db: tx}
}

// BeginTx for txQueries will simply call the Begin, start
func (v *txQueries) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	return v.db.Begin(ctx)
}

// Close implements Dbx.
func (v *txQueries) Close() {
	slog.Info("close called on txQueries, nothing to do.")
}

// RunInTxContext implements Dbx.
func (v *txQueries) RunInTxContext(ctx context.Context, fn func(context.Context) error) error {
	return WithTxContext(ctx, v, fn)
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

func (v *txQueries) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return v.db.Query(ctx, sql, args...)
}

func (v *txQueries) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return v.db.Exec(ctx, sql, args...)
}

// RunInTx
//
// Deprecated: use RunInTxContext
func (v *txQueries) RunInTx(fn func(Dbx) error) error {
	return WithTx(v, fn)
}

// WithTx
//
// Deprecated: use WithTxContext
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

// WithTxContext
//
// creates a new transaction from the dbx, and embeds it in the provided context ctx.
// it is then passed to the given function fn, which can check the context for the embedded transaction
// and use it as needed.
func WithTxContext(ctx context.Context, dbx Dbx, fn func(context.Context) error) (returnErr error) {
	db := GetContextOrDefaultDbx(ctx, dbx)
	tx, beginErr := db.Begin(ctx)
	if beginErr != nil {
		slog.Error("error starting transaction", slog.Any("error", beginErr))
		returnErr = beginErr
		return
	}

	defer func() {
		if recErr := recover(); recErr != nil {
			slog.ErrorContext(ctx, "recovered from panic in transaction.", slog.String("errorMessage", fmt.Sprint(recErr)), slog.Any("stacktrace", string(debug.Stack())))
			rollBackErr := tx.Rollback(context.Background())
			if rollBackErr != nil {
				slog.ErrorContext(ctx, "error rolling back transaction", slog.Any("error", rollBackErr))
				returnErr = errors.New("there was an error while recovering from a failure")
				return
			}
		}
	}()
	txCtx := setContextTx(ctx, &txQueries{db: tx})
	fnErr := fn(txCtx)
	if fnErr == nil {
		if commitErr := tx.Commit(ctx); commitErr != nil {
			slog.ErrorContext(ctx, "error committing transaction", slog.Any("error", commitErr))
			returnErr = commitErr
			return
		}
	} else {
		slog.ErrorContext(ctx, "error in transaction function. rolling back.", slog.Any("error", beginErr))
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			slog.ErrorContext(ctx, "error rolling back transaction", slog.Any("error", rollbackErr))
			returnErr = rollbackErr
			return
		}
	}

	return beginErr
}
