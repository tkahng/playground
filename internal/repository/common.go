package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/tkahng/authgo/internal/shared"
)

func VarCollect[T any](args ...T) []T {
	return args
}

func ViewApplyPagination[T any, Ts ~[]T](view *psql.ViewQuery[T, Ts], input *shared.PaginatedInput) {
	if input == nil {
		input = &shared.PaginatedInput{
			// PerPage: shared.From(10),
			// Page:    shared.From(1),
			PerPage: 10,
			Page:    1,
		}
	}
	// if input.Page.IsUnset() {
	if input.Page == 0 {
		// input.Page = shared.From(1)
		input.Page = 1
	}
	if input.PerPage == 0 {
		// input.PerPage = shared.From(10)
		input.PerPage = 10
	}
	view.Apply(
		sm.Limit(psql.Arg(input.PerPage)),
		sm.Offset(psql.Arg((input.Page-1)*input.PerPage)),
		// sm.Offset(psql.Arg((input.Page.MustGet()-1)*input.PerPage.MustGet())),
	)
}

func CountExec[T any, Ts ~[]T](ctx context.Context, db bob.Executor, v *psql.ViewQuery[T, Ts]) (int64, error) {
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
