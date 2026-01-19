package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

func PgxQueryWithBuilder[T any](ctx context.Context, db Dbx, query QueryBuilder) ([]*T, error) {
	sql, args, err := query.ToSql()
	if err != nil {
		slog.ErrorContext(
			ctx,
			"failed to build query.",
			slog.Any("error", err),
			slog.String("query", sql),
			slog.Any("args", args),
		)
		return nil, fmt.Errorf("error building query: %w", err)
	}
	return PgxQuery[T](ctx, db, sql, args)
}

// PgxQuery returns a slice of address of a T scanned from rows from [Dbx.Query]
func PgxQuery[T any](ctx context.Context, db Dbx, query string, args ...any) ([]*T, error) {
	ctxDbx := GetContextOrDefaultDbx(ctx, db)
	slog.DebugContext(ctx, "starting query:", slog.String("query", query), slog.Any("args", args))
	r, err := ctxDbx.Query(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"error at PgxQuery.",
			slog.Any("error", err),
			slog.String("query", query),
			slog.Any("args", args),
		)

		return nil, fmt.Errorf("error during query: %w", err)
	}
	return pgx.CollectRows(r, pgx.RowToAddrOfStructByNameLax[T])
}

// func PgxQuery
func PgxQuerySingleScalar[T comparable](ctx context.Context, db Dbx, query QueryBuilder) (T, error) {
	ctxDbx := GetContextOrDefaultDbx(ctx, db)
	var zero T
	sql, args, err := query.ToSql()
	if err != nil {
		return zero, err
	}
	slog.DebugContext(ctx, "PgxQuerySingleScalar:", slog.String("query", sql), slog.Any("args", args))
	r, err := ctxDbx.Query(ctx, sql, args...)
	if err != nil {
		return zero, err
	}
	res, err := pgx.CollectRows(r, pgx.RowTo[T])
	if err != nil {
		return zero, err
	}
	if len(res) == 0 {
		return zero, nil
	}
	return res[0], nil
}
