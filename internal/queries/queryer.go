package queries

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stephenafamo/scan"
)

type Queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (scan.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}
