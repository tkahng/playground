package repository

import (
	"context"

	"github.com/aarondl/opt/null"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/types"
	"github.com/stephenafamo/scan"
)

type Param[T any] struct {
	Value types.JSON[T]
}

func GetParams[T any](ctx context.Context, dbx bob.Executor, key string) (*Param[T], error) {
	query := psql.Select(
		sm.Columns("value"),
		sm.From("app_params"),
		sm.Where(psql.Quote("name").EQ(psql.Arg(key))),
		sm.Limit(1),
	)
	param, err := bob.One(ctx, dbx, query, scan.StructMapper[*Param[T]]())
	// return param, err
	return OptionalRow(param, err)
}

func SetParams[T any](ctx context.Context, dbx bob.Executor, key string, data T) error {
	query := psql.Insert(
		im.Into("app_params", "name", "value"),
		im.Values(
			psql.Arg(key),
			psql.Arg(null.From(types.NewJSON(data))),
		),
		im.OnConflict("name").DoUpdate(
			im.SetCol("value").To(
				psql.Raw("EXCLUDED.value"),
			),
		),
	)
	_, err := bob.Exec(ctx, dbx, query)

	return err
}
