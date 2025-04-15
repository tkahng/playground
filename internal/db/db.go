package db

import (
	"context"
	"database/sql"
	"database/sql/driver"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/scan"
)

type DBTX interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	// BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	// Ping(ctx context.Context) error
	// Executor() *DbTx
}

var _ DBTX = (*pgxpool.Pool)(nil)

type DBTx struct {
	pool DBTX
}

func NewDBTx(pool DBTX) *DBTx {
	return &DBTx{pool: pool}
}

var _ bob.Executor = (*DBTx)(nil)

type rows struct {
	pgx.Rows
}

func (r rows) Close() error {
	r.Rows.Close()
	return nil
}

func (r rows) Columns() ([]string, error) {
	fields := r.FieldDescriptions()
	cols := make([]string, len(fields))

	for i, field := range fields {
		cols[i] = field.Name
	}

	return cols, nil
}

func (v *DBTx) QueryContext(ctx context.Context, query string, args ...any) (scan.Rows, error) {
	r, err := v.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return rows{r}, nil
}

func (v *DBTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	tag, err := v.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return driver.RowsAffected(tag.RowsAffected()), err
}
