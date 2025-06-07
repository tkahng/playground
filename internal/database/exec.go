package database

import (
	"context"

	"github.com/stephenafamo/scan"
	"github.com/stephenafamo/scan/pgxscan"
)

type QueryBuilder interface {
	ToSql() (string, []any, error)
}

func QueryWithBuilder[T any](ctx context.Context, db Dbx, query QueryBuilder) ([]T, error) {
	sql, args, err := query.ToSql()
	// fmt.Println("query", sql, "args", args)
	if err != nil {
		return nil, err
	}
	return QueryAll[T](ctx, db, sql, args...)
}

func ExecWithBuilder(ctx context.Context, db Dbx, query QueryBuilder) error {
	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, sql, args...)
	return err
}

func QueryAll[T any](ctx context.Context, db Dbx, query string, args ...any) ([]T, error) {
	return pgxscan.All(ctx, db, scan.StructMapper[T](), query, args...)
}

func Count(ctx context.Context, db Dbx, query string, args ...any) (int64, error) {
	return pgxscan.One(ctx, db, scan.SingleColumnMapper[int64], query, args...)
}

type CountOutput struct {
	Count int64
}
