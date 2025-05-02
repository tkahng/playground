package queries

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/scan"
	"github.com/stephenafamo/scan/pgxscan"
	"github.com/tkahng/authgo/internal/shared"
)

func VarCollect[T any](args ...T) []T {
	return args
}

func ViewApplyPagination[T any, Ts ~[]T](view *psql.ViewQuery[T, Ts], input *shared.PaginatedInput) {
	if input == nil {
		input = &shared.PaginatedInput{
			PerPage: 10,
			Page:    0,
		}
	}
	if input.PerPage == 0 {
		input.PerPage = 10
	}
	view.Apply(
		sm.Limit(psql.Arg(input.PerPage)),
		sm.Offset(psql.Arg((input.Page)*input.PerPage)),
	)
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

func CountExec[T any, Ts ~[]T](ctx context.Context, db Queryer, v *psql.ViewQuery[T, Ts]) (int64, error) {
	data, err := v.Count(ctx, db)
	if err != nil {
		return 0, err
	}
	return data, nil
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

type QueryBuilder interface {
	ToSql() (string, []interface{}, error)
}

func ExecQuery[T any](ctx context.Context, db Queryer, query QueryBuilder) ([]T, error) {
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	return pgxscan.All(ctx, db, scan.StructMapper[T](), sql, args...)
}
