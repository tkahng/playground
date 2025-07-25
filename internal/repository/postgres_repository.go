package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/types"
)

// PostgresRepository provides CRUD operations for Postgres
type PostgresRepository[Model any] struct {
	builder *SQLBuilder[Model]
}

var _ Repository[models.User] = (*PostgresRepository[models.User])(nil)

// NewPostgresRepository initializes a new PostgresRepository

func NewPostgresRepository[Model any](builder *SQLBuilder[Model]) *PostgresRepository[Model] {
	return &PostgresRepository[Model]{
		builder: builder,
	}
}

func (r *PostgresRepository[Model]) Builder() SQLBuilderInterface {
	return r.builder
}

// Get retrieves records from the database based on the provided filters
func (r *PostgresRepository[Model]) Get(ctx context.Context, db database.Dbx, where *map[string]any, order *map[string]string, limit *int, offset *int) ([]*Model, error) {
	var args []any
	//goland:noinspection Annotator
	query := fmt.Sprintf("SELECT %s FROM %s", r.builder.FieldString(""), r.builder.Table())
	expr, err := r.builder.WhereError(ctx, where, &args, nil)
	if err != nil {
		return nil, err
	}
	if expr != "" {
		query += fmt.Sprintf(" WHERE %s", expr)
	}
	if orderexpr := r.builder.Order(order); orderexpr != "" {
		query += fmt.Sprintf(" ORDER BY %s", orderexpr)
	}
	if limit != nil {
		query += fmt.Sprintf(" LIMIT %d", *limit)
	}
	if offset != nil {
		query += fmt.Sprintf(" OFFSET %d", *offset)
	}

	// Execute the query and scan the results
	// slog.Info("query and args", slog.String("query", query), slog.Any("args", args))
	items, err := database.QueryAll[*Model](
		ctx,
		db,
		query,
		args...,
	)
	if err != nil {
		slog.ErrorContext(ctx, "Error executing Get query", slog.String("query", query), slog.Any("args", args), slog.Any("error", err))
		return nil, err
	}

	return items, nil
}

// Put updates existing records in the database
func (r *PostgresRepository[Model]) Put(ctx context.Context, dbx database.Dbx, models []Model) ([]*Model, error) {
	result := []*Model{}

	for _, model := range models {
		args := []any{}
		where := map[string]any{}
		set, err := r.builder.SetError(&model, &args, &where)
		if err != nil {
			return nil, err
		}
		//goland:noinspection Annotator
		query := fmt.Sprintf("UPDATE %s SET %s", r.builder.Table(), set)
		if expr, err := r.builder.WhereError(ctx, &where, &args, nil); err != nil {
			return nil, err
		} else if expr != "" {
			query += fmt.Sprintf(" WHERE %s", expr)
		}
		query += fmt.Sprintf(" RETURNING %s", r.builder.FieldString(""))

		items, err := database.QueryAll[*Model](
			ctx,
			dbx,
			query,
			args...,
		)
		if err != nil {
			slog.ErrorContext(ctx, "Error executing Put query", slog.String("query", query), slog.Any("args", args), slog.Any("error", err))
			return nil, err
		}

		result = append(result, items...)
	}

	return result, nil
}

func (r *PostgresRepository[Model]) PutOne(ctx context.Context, dbx database.Dbx, model *Model) (*Model, error) {
	if model == nil {
		return nil, nil
	}
	result, err := r.Put(ctx, dbx, []Model{*model})
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	re := result[0]
	return re, nil
}

func (r *PostgresRepository[Model]) GetOne(ctx context.Context, dbx database.Dbx, where *map[string]any) (*Model, error) {
	result, err := r.Get(ctx, dbx, where, nil, types.Pointer(1), nil)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	re := result[0]
	return re, nil
}

// Post inserts new records into the database
func (r *PostgresRepository[Model]) PostExec(ctx context.Context, dbx database.Dbx, models []Model) (int64, error) {
	args := []any{}
	//goland:noinspection Annotator
	query := fmt.Sprintf("INSERT INTO %s", r.builder.Table())
	if fields, values, err := r.builder.ValuesError(&models, &args, nil); err != nil {
		return 0, err
	} else if fields != "" && values != "" {
		query += fmt.Sprintf(" (%s) VALUES %s", fields, values)
	}
	// Execute the query and scan the results
	// fmt.Println("query", query, "args", args)
	result, err := database.Exec(
		ctx,
		dbx,
		query,
		args...,
	)
	if err != nil {
		slog.ErrorContext(ctx, "Error executing Post query", slog.String("query", query), slog.Any("args", args), slog.Any("error", err))
		return 0, err
	}

	return result, nil
}

// Post inserts new records into the database
func (r *PostgresRepository[Model]) Post(ctx context.Context, dbx database.Dbx, models []Model) ([]*Model, error) {
	args := []any{}
	//goland:noinspection Annotator
	query := fmt.Sprintf("INSERT INTO %s", r.builder.Table())
	if fields, values, err := r.builder.ValuesError(&models, &args, nil); err != nil {
		return nil, err
	} else if fields != "" && values != "" {
		query += fmt.Sprintf(" (%s) VALUES %s", fields, values)
	}
	query += fmt.Sprintf(" RETURNING %s", r.builder.FieldString(""))

	// Execute the query and scan the results
	// fmt.Println("query", query, "args", args)
	result, err := database.QueryAll[*Model](
		ctx,
		dbx,
		query,
		args...,
	)
	if err != nil {
		slog.ErrorContext(ctx, "Error executing Post query", slog.String("query", query), slog.Any("args", args), slog.Any("error", err))
		return nil, err
	}

	return result, nil
}

// Patch updates existing records in the database
func (r *PostgresRepository[Model]) PostOne(ctx context.Context, dbx database.Dbx, models *Model) (*Model, error) {
	data, err := r.Post(ctx, dbx, []Model{*models})
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return data[0], nil
}

// DeleteReturn removes records from the database based on the provided filters
func (r *PostgresRepository[Model]) DeleteReturn(ctx context.Context, dbx database.Dbx, where *map[string]any) ([]*Model, error) {
	args := []any{}
	//goland:noinspection Annotator
	query := fmt.Sprintf("DELETE FROM %s", r.builder.Table())
	if expr, err := r.builder.WhereError(ctx, where, &args, nil); err != nil {
		return nil, err
	} else if expr != "" {
		query += fmt.Sprintf(" WHERE %s", expr)
	}
	query += fmt.Sprintf(" RETURNING %s", r.builder.FieldString(""))

	// Execute the query and scan the results
	result, err := database.QueryAll[*Model](
		ctx,
		dbx,
		query,
		args...,
	)
	if err != nil {
		slog.ErrorContext(ctx, "Error executing Delete query", slog.String("query", query), slog.Any("args", args), slog.Any("error", err))
		return nil, err
	}

	return result, nil
}

// DeleteReturn removes records from the database based on the provided filters
func (r *PostgresRepository[Model]) Delete(ctx context.Context, dbx database.Dbx, where *map[string]any) (int64, error) {
	args := []any{}
	//goland:noinspection Annotator
	query := fmt.Sprintf("DELETE FROM %s", r.builder.Table())
	if expr, err := r.builder.WhereError(ctx, where, &args, nil); err != nil {
		return 0, err
	} else if expr != "" {
		query += fmt.Sprintf(" WHERE %s", expr)
	}
	// query += fmt.Sprintf(" RETURNING %s", r.builder.Fields(""))

	// Execute the query and scan the results
	result, err := database.Exec(
		ctx,
		dbx,
		query,
		args...,
	)
	if err != nil {
		slog.ErrorContext(ctx, "Error executing Delete query", slog.String("query", query), slog.Any("args", args), slog.Any("error", err))
		return 0, err
	}

	return result, nil
}

// Count returns the number of records that match the provided filters
func (r *PostgresRepository[Model]) Count(ctx context.Context, dbx database.Dbx, where *map[string]any) (int64, error) {
	args := []any{}
	//goland:noinspection Annotator
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", r.builder.Table())
	if expr, err := r.builder.WhereError(ctx, where, &args, nil); err != nil {
		return 0, err
	} else if expr != "" {
		query += fmt.Sprintf(" WHERE %s", expr)
	}

	// Execute the query and scan the results
	// fmt.Println("query", query, "args", args)
	count, err := database.Count(ctx, dbx, query, args...)

	// result, err := r.builder.Scan(dbx.Query(ctx, query, args...))
	if err != nil {
		slog.ErrorContext(ctx, "Error executing Get query", slog.String("query", query), slog.Any("args", args), slog.Any("error", err))
		return 0, err
	}

	return count, nil
}
