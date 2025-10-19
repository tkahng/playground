package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string

const (
	txContextKey contextKey = "tx_context_key"
)

func setContextTx(ctx context.Context, tx Dbx) context.Context {
	return context.WithValue(ctx, txContextKey, tx)
}
func getContextTx(ctx context.Context) Dbx {
	if tx, ok := ctx.Value(txContextKey).(Dbx); ok {
		return tx
	} else {
		return nil
	}
}
func GetContextOrDefaultDbx(ctx context.Context, dbx Dbx) Dbx {
	tx := getContextTx(ctx)
	if tx != nil {
		return tx
	}
	return dbx
}

type Executor interface {
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

type Dbx interface {
	Executor

	Begin(ctx context.Context) (pgx.Tx, error)

	BeginTx(ctx context.Context, opts ...func(*pgx.TxOptions)) (pgx.Tx, error)

	// Close
	//
	// Close calls close on the underlying pool.
	Close()
	// RunInTx
	//
	// Deprecated: use RunInTxCtx
	RunInTx(fn func(Dbx) error) error
	RunInTxCtx(ctx context.Context, fn func(context.Context) error) error
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

func (v *Queries) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return v.db.QueryRow(ctx, sql, args...)
}

func (v *Queries) Begin(ctx context.Context) (pgx.Tx, error) {
	opt := &pgx.TxOptions{}
	return v.db.BeginTx(ctx, *opt)
}

// BeginTx acquires a connection from the Pool and starts a transaction with pgx.TxOptions determining the transaction mode.
// Unlike database/sql, the context only affects the begin command. i.e. there is no auto-rollback on context cancellation.
func (v *Queries) BeginTx(ctx context.Context, opts ...func(*pgx.TxOptions)) (pgx.Tx, error) {
	opt := &pgx.TxOptions{}
	for _, f := range opts {
		f(opt)
	}
	return v.db.BeginTx(ctx, *opt)
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

func (v *Queries) RunInTxCtx(ctx context.Context, fn func(context.Context) error) error {
	return WithCtxTx(ctx, v, fn)
}

var _ Dbx = (*txQueries)(nil)

type txQueries struct {
	db pgx.Tx
}

func NewTxQueries(tx pgx.Tx) *txQueries {
	return &txQueries{db: tx}
}

// BeginTx for txQueries will simply call the Begin, starting a pseudo nested transaction.
// the opts will be ignored
func (v *txQueries) BeginTx(ctx context.Context, opts ...func(*pgx.TxOptions)) (pgx.Tx, error) {
	return v.db.Begin(ctx)
}

// Begin for txQueries will simply call the Begin, starting a pseudo nested transaction.
// the opts will be ignored
func (v *txQueries) Begin(ctx context.Context) (pgx.Tx, error) {
	return v.db.Begin(ctx)
}

// Close is a no-op. this is here to implement Dbx
func (v *txQueries) Close() {
	slog.Info("close called on txQueries, nothing to do.")
}

// RunInTxCtx implements Dbx.
func (v *txQueries) RunInTxCtx(ctx context.Context, fn func(context.Context) error) error {
	return WithCtxTx(ctx, v, fn)
}

// QueryRow implements Dbx.
func (v *txQueries) QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row {
	return v.db.QueryRow(ctx, sql, arguments...)
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
// Deprecated: use RunInTxCtx
func (v *txQueries) RunInTx(fn func(Dbx) error) error {
	return WithTx(v, fn)
}

// WithTxWithContext
//
// WithTxWithContext is like WithTx, but it takes a context as a parameter.
func WithTxWithContext(ctx context.Context, dbx Dbx, fn func(tx Dbx) error, opts ...func(*pgx.TxOptions)) (returnErr error) {
	contextOrDefaultDb := GetContextOrDefaultDbx(ctx, dbx)
	tx, beginErr := contextOrDefaultDb.BeginTx(ctx, opts...)
	if beginErr != nil {
		slog.Error("error starting transaction", slog.Any("error", beginErr))
		returnErr = errors.New("there was an error starting a transaction")
		return
	}

	defer func() {
		if recErr := recover(); recErr != nil {
			slog.ErrorContext(ctx, "error in transaction function. rolling back.", slog.Any("error", beginErr))
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				slog.ErrorContext(ctx, "error rolling back transaction", slog.Any("error", rollbackErr))
				returnErr = errors.New("there was an error while recovering from a failure")
				return
			}
			returnErr = errors.New(fmt.Sprint(recErr))
			return
		}
	}()
	// txCtx := setContextTx(ctx, &txQueries{db: tx})
	if fnErr := fn(NewTxQueries(tx)); fnErr != nil {
		slog.ErrorContext(ctx, "error in transaction function. rolling back.", slog.Any("error", beginErr))
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			slog.ErrorContext(ctx, "error rolling back transaction", slog.Any("error", rollbackErr))
			returnErr = errors.New("there was an error while recovering from a failure")
			return
		}
		returnErr = fnErr
		return
	} else {
		if commitErr := tx.Commit(context.Background()); commitErr != nil {
			slog.ErrorContext(ctx, "error committing transaction", slog.Any("error", commitErr))
			returnErr = errors.New("there was an error while committing a transaction")
			return
		}
	}
	return
}

// WithTx
//
// Deprecated: use WithTxContext
func WithTx(dbx Dbx, fn func(tx Dbx) error) (returnErr error) {
	ctx := context.Background() // Use the appropriate context as needed
	contextOrDefaultDb := GetContextOrDefaultDbx(ctx, dbx)
	tx, beginErr := contextOrDefaultDb.BeginTx(ctx)
	if beginErr != nil {
		slog.Error("error starting transaction", slog.Any("error", beginErr))
		returnErr = errors.New("there was an error starting a transaction")
		return
	}

	defer func() {
		if recErr := recover(); recErr != nil {
			slog.ErrorContext(ctx, "error in transaction function. rolling back.", slog.Any("error", beginErr))
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				slog.ErrorContext(ctx, "error rolling back transaction", slog.Any("error", rollbackErr))
				returnErr = errors.New("there was an error while recovering from a failure")
				return
			}
			returnErr = errors.New(fmt.Sprint(recErr))
			return
		}
	}()
	// txCtx := setContextTx(ctx, &txQueries{db: tx})
	if fnErr := fn(NewTxQueries(tx)); fnErr != nil {
		slog.ErrorContext(ctx, "error in transaction function. rolling back.", slog.Any("error", beginErr))
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			slog.ErrorContext(ctx, "error rolling back transaction", slog.Any("error", rollbackErr))
			returnErr = errors.New("there was an error while recovering from a failure")
			return
		}
		returnErr = fnErr
		return
	} else {
		if commitErr := tx.Commit(context.Background()); commitErr != nil {
			slog.ErrorContext(ctx, "error committing transaction", slog.Any("error", commitErr))
			returnErr = errors.New("there was an error while committing a transaction")
			return
		}
	}
	return
}

// WithCtxTx
//
// creates a new transaction from the dbx, and embeds it in the provided context ctx.
// it is then passed to the given function fn, which can check the context for the embedded transaction
// and use it as needed.
func WithCtxTx(ctx context.Context, dbx Dbx, fn func(context.Context) error, opts ...func(*pgx.TxOptions)) (returnErr error) {
	contextOrDefaultDb := GetContextOrDefaultDbx(ctx, dbx)
	tx, beginErr := contextOrDefaultDb.BeginTx(ctx, opts...)
	if beginErr != nil {
		slog.Error("error starting transaction", slog.Any("error", beginErr))
		returnErr = errors.New("there was an error starting a transaction")
		return
	}

	defer func() {
		if recErr := recover(); recErr != nil {
			slog.ErrorContext(ctx, "error in transaction function. rolling back.", slog.Any("error", beginErr))
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				slog.ErrorContext(ctx, "error rolling back transaction", slog.Any("error", rollbackErr))
				returnErr = errors.New("there was an error while recovering from a failure")
				return
			}
			returnErr = errors.New(fmt.Sprint(recErr))
			return
		}
	}()
	txCtx := setContextTx(ctx, &txQueries{db: tx})
	if fnErr := fn(txCtx); fnErr != nil {
		slog.ErrorContext(ctx, "error in transaction function. rolling back.", slog.Any("error", beginErr))
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			slog.ErrorContext(ctx, "error rolling back transaction", slog.Any("error", rollbackErr))
			returnErr = errors.New("there was an error while recovering from a failure")
			return
		}
		returnErr = fnErr
		return
	} else {
		if commitErr := tx.Commit(context.Background()); commitErr != nil {
			slog.ErrorContext(ctx, "error committing transaction", slog.Any("error", commitErr))
			returnErr = errors.New("there was an error while committing a transaction")
			return
		}
	}
	return
}

// WithCtxTx
//
// creates a new transaction from the dbx, and embeds it in the provided context ctx.
// it is then passed to the given function fn, which can check the context for the embedded transaction
// and use it as needed.
func WithCtxAndTx(ctx context.Context, dbx Dbx, fn func(context.Context, Dbx) error, opts ...func(*pgx.TxOptions)) (returnErr error) {
	contextOrDefaultDb := GetContextOrDefaultDbx(ctx, dbx)
	tx, beginErr := contextOrDefaultDb.BeginTx(ctx, opts...)
	if beginErr != nil {
		slog.Error("error starting transaction", slog.Any("error", beginErr))
		returnErr = errors.New("there was an error starting a transaction")
		return
	}

	defer func() {
		if recErr := recover(); recErr != nil {
			slog.ErrorContext(ctx, "error in transaction function. rolling back.", slog.Any("error", beginErr))
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				slog.ErrorContext(ctx, "error rolling back transaction", slog.Any("error", rollbackErr))
				returnErr = errors.New("there was an error while recovering from a failure")
				return
			}
			returnErr = errors.New(fmt.Sprint(recErr))
			return
		}
	}()
	// txCtx := setContextTx(ctx, &txQueries{db: tx})
	if fnErr := fn(ctx, NewTxQueries(tx)); fnErr != nil {
		slog.ErrorContext(ctx, "error in transaction function. rolling back.", slog.Any("error", beginErr))
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			slog.ErrorContext(ctx, "error rolling back transaction", slog.Any("error", rollbackErr))
			returnErr = errors.New("there was an error while recovering from a failure")
			return
		}
		returnErr = fnErr
		return
	} else {
		if commitErr := tx.Commit(context.Background()); commitErr != nil {
			slog.ErrorContext(ctx, "error committing transaction", slog.Any("error", commitErr))
			returnErr = errors.New("there was an error while committing a transaction")
			return
		}
	}
	return
}
