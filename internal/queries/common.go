package queries

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stephenafamo/scan"
	"github.com/stephenafamo/scan/pgxscan"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/types"
)

func VarCollect[T any](args ...T) []T {
	return args
}

func Paginate(q squirrel.SelectBuilder, input *shared.PaginatedInput) squirrel.SelectBuilder {
	if input == nil {
		input = &shared.PaginatedInput{
			PerPage: 10,
			Page:    0,
		}
	}
	if input.PerPage == 0 {
		input.PerPage = 10
	}
	return q.Limit(uint64(input.PerPage)).Offset(uint64((input.Page) * input.PerPage))
}

func PaginateRepo(input *shared.PaginatedInput) (*int, *int) {
	if input == nil {
		input = &shared.PaginatedInput{
			PerPage: 10,
			Page:    0,
		}
	}
	if input.PerPage == 0 {
		input.PerPage = 10
	}
	return types.Pointer(int(input.PerPage)), types.Pointer(int((input.Page) * input.PerPage))
}

func OptionalRow[T any](record *T, err error) (*T, error) {
	if err == nil {
		return record, nil
	} else if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	} else if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else {
		return nil, err
	}
}

func ParseUUIDs(ids []string) []uuid.UUID {
	var uuids []uuid.UUID
	for _, id := range ids {
		parsed, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		uuids = append(uuids, parsed)
	}
	return uuids
}

func ReturnFirst[T any](args []*T) *T {
	if len(args) > 1 {
		return args[0]
	}
	return nil
}

type QueryBuilder interface {
	ToSql() (string, []any, error)
}

func QueryWithBuilder[T any](ctx context.Context, db Queryer, query QueryBuilder) ([]T, error) {
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	fmt.Println(sql, args)
	return QueryAll[T](ctx, db, sql, args...)
}

func QueryAll[T any](ctx context.Context, db Queryer, query string, args ...any) ([]T, error) {
	return pgxscan.All(ctx, db, scan.StructMapper[T](), query, args...)
}

func QueryOne[T any](ctx context.Context, db Queryer, query string, args ...any) (T, error) {
	return pgxscan.One(ctx, db, scan.StructMapper[T](), query, args...)
}

func Count(ctx context.Context, db Queryer, query string, args ...any) (int64, error) {
	return pgxscan.One(ctx, db, scan.SingleColumnMapper[int64], query, args...)
}

func Exec(ctx context.Context, db Queryer, query string, args ...any) error {
	_, err := db.Exec(ctx, query, args...)
	return err
}

func ExecWithBuilder(ctx context.Context, db Queryer, query QueryBuilder) error {
	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}
	return Exec(ctx, db, sql, args...)
}

func Identifier(name string) string {
	return fmt.Sprintf("\"%s\"", name)
}
