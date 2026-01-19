package database

import (
	"context"
	"log/slog"

	"github.com/stephenafamo/scan"
	"github.com/stephenafamo/scan/pgxscan"
)

type QueryBuilder interface {
	ToSql() (string, []any, error)
}

func QueryWithBuilder[T any](ctx context.Context, db Dbx, query QueryBuilder) ([]T, error) {
	sql, args, err := query.ToSql()
	// fmt.Println("query", sql, "args", args)

	slog.DebugContext(ctx, "Query With Builder:", slog.String("query", sql), slog.Any("args", args))
	if err != nil {
		return nil, err
	}
	return QueryAll[T](ctx, db, sql, args...)
}

func ExecWithBuilder(ctx context.Context, db Dbx, query QueryBuilder) (int64, error) {
	sql, args, err := query.ToSql()
	if err != nil {
		return 0, err
	}
	slog.DebugContext(ctx, "Exec With Builder:", slog.String("query", sql), slog.Any("args", args))
	result, err := Exec(ctx, db, sql, args...)
	return result, err
}

func QueryAll[T any](ctx context.Context, db Dbx, query string, args ...any) ([]T, error) {
	ctxDbx := GetContextOrDefaultDbx(ctx, db)
	return pgxscan.All(ctx, ctxDbx, scan.StructMapper[T](), query, args...)
}

func Count(ctx context.Context, db Dbx, query string, args ...any) (int64, error) {
	ctxDbx := GetContextOrDefaultDbx(ctx, db)
	return pgxscan.One(ctx, ctxDbx, scan.SingleColumnMapper[int64], query, args...)
}

func QueryOneSingleColumn[T any](ctx context.Context, db Dbx, query string, args ...any) (T, error) {
	ctxDbx := GetContextOrDefaultDbx(ctx, db)
	return pgxscan.One(ctx, ctxDbx, scan.SingleColumnMapper[T], query, args...)
}

func QueryManySingleColumn[T any](ctx context.Context, db Dbx, query string, args ...any) ([]T, error) {
	ctxDbx := GetContextOrDefaultDbx(ctx, db)
	return pgxscan.All(ctx, ctxDbx, scan.SingleColumnMapper[T], query, args...)
}

func Exec(ctx context.Context, db Dbx, query string, args ...any) (int64, error) {
	ctxDbx := GetContextOrDefaultDbx(ctx, db)
	result, err := ctxDbx.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

type CountOutput struct {
	Count int64
}
