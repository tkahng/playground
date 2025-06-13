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
func QueryWithBuilderSingle[T any](ctx context.Context, db Dbx, query QueryBuilder) ([]T, error) {
	sql, args, err := query.ToSql()
	// fmt.Println("query", sql, "args", args)
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

func Exec(ctx context.Context, db Dbx, query string, args ...any) (int64, error) {
	ctxDbx := GetContextOrDefaultDbx(ctx, db)
	result, err := ctxDbx.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func One[T any](ctx context.Context, db Dbx, query string, args ...any) (T, error) {
	var result T
	res, err := QueryAll[T](
		ctx,
		db,
		query,
		args...,
	)
	if err != nil {
		return result, err
	}
	if len(res) == 0 {
		return result, nil
	}
	return res[0], nil
}

type CountOutput struct {
	Count int64
}
