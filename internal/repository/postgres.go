package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/stephenafamo/scan"
	"github.com/stephenafamo/scan/pgxscan"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/tools/types"
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
func (r *PostgresRepository[Model]) Get(ctx context.Context, db db.Dbx, where *map[string]any, order *map[string]string, limit *int, skip *int) ([]*Model, error) {
	args := []any{}
	query := fmt.Sprintf("SELECT %s FROM %s", r.builder.Fields(""), r.builder.Table())
	expr, err := r.builder.WhereError(where, &args, nil)
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
	if skip != nil {
		query += fmt.Sprintf(" OFFSET %d", *skip)
	}

	slog.Info("Executing Get query", slog.String("query", query), slog.Any("args", args))

	// Execute the query and scan the results
	result, err := pgxscan.All(ctx, db, scan.StructMapper[*Model](), query, args...)
	if err != nil {
		slog.Error("Error executing Get query", slog.String("query", query), slog.Any("args", args), slog.Any("error", err))
		return nil, err
	}

	return result, nil
}

// Put updates existing records in the database
func (r *PostgresRepository[Model]) Put(ctx context.Context, dbx db.Dbx, models []Model) ([]*Model, error) {
	result := []*Model{}

	for _, model := range models {
		args := []any{}
		where := map[string]any{}
		set, err := r.builder.SetError(&model, &args, &where)
		if err != nil {
			return nil, err
		}
		query := fmt.Sprintf("UPDATE %s SET %s", r.builder.Table(), set)
		if expr, err := r.builder.WhereError(&where, &args, nil); err != nil {
			return nil, err
		} else if expr != "" {
			query += fmt.Sprintf(" WHERE %s", expr)
		}
		query += fmt.Sprintf(" RETURNING %s", r.builder.Fields(""))

		slog.Info("Executing Put query", slog.String("query", query), slog.Any("args", args))

		items, err := pgxscan.All(ctx, dbx, scan.StructMapper[*Model](), query, args...)
		if err != nil {
			slog.Error("Error executing Put query", slog.String("query", query), slog.Any("args", args), slog.Any("error", err))
			// tx.Rollback(ctx)
			return nil, err
		}

		result = append(result, items...)
	}

	return result, nil
}

func (r *PostgresRepository[Model]) PutOne(ctx context.Context, dbx db.Dbx, model *Model) (*Model, error) {
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

func (r *PostgresRepository[Model]) GetOne(ctx context.Context, dbx db.Dbx, where *map[string]any) (*Model, error) {
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
func (r *PostgresRepository[Model]) Post(ctx context.Context, dbx db.Dbx, models []Model) ([]*Model, error) {
	args := []any{}
	query := fmt.Sprintf("INSERT INTO %s", r.builder.Table())
	if fields, values, err := r.builder.ValuesError(&models, &args, nil); err != nil {
		return nil, err
	} else if fields != "" && values != "" {
		query += fmt.Sprintf(" (%s) VALUES %s", fields, values)
	}
	query += fmt.Sprintf(" RETURNING %s", r.builder.Fields(""))

	slog.Info("Executing Post query", slog.String("query", query), slog.Any("args", args))

	// Execute the query and scan the results
	result, err := pgxscan.All(ctx, dbx, scan.StructMapper[*Model](), query, args...)
	if err != nil {
		slog.Error("Error executing Post query", slog.String("query", query), slog.Any("args", args), slog.Any("error", err))
		return nil, err
	}

	return result, nil
}

// Patch updates existing records in the database
func (r *PostgresRepository[Model]) PostOne(ctx context.Context, dbx db.Dbx, models *Model) (*Model, error) {
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
func (r *PostgresRepository[Model]) DeleteReturn(ctx context.Context, dbx db.Dbx, where *map[string]any) ([]*Model, error) {
	args := []any{}
	query := fmt.Sprintf("DELETE FROM %s", r.builder.Table())
	if expr, err := r.builder.WhereError(where, &args, nil); err != nil {
		return nil, err
	} else if expr != "" {
		query += fmt.Sprintf(" WHERE %s", expr)
	}
	query += fmt.Sprintf(" RETURNING %s", r.builder.Fields(""))

	slog.Info("Executing Delete query", slog.String("query", query), slog.Any("args", args))

	// Execute the query and scan the results
	result, err := pgxscan.All(ctx, dbx, scan.StructMapper[*Model](), query, args...)
	if err != nil {
		slog.Error("Error executing Delete query", slog.String("query", query), slog.Any("args", args), slog.Any("error", err))
		return nil, err
	}

	return result, nil
}

// DeleteReturn removes records from the database based on the provided filters
func (r *PostgresRepository[Model]) Delete(ctx context.Context, dbx db.Dbx, where *map[string]any) (int64, error) {
	args := []any{}
	query := fmt.Sprintf("DELETE FROM %s", r.builder.Table())
	if expr, err := r.builder.WhereError(where, &args, nil); err != nil {
		return 0, err
	} else if expr != "" {
		query += fmt.Sprintf(" WHERE %s", expr)
	}
	// query += fmt.Sprintf(" RETURNING %s", r.builder.Fields(""))

	slog.Info("Executing Delete query", slog.String("query", query), slog.Any("args", args))

	// Execute the query and scan the results
	result, err := dbx.Exec(ctx, query, args...)
	// result, err := pgxscan.All(ctx, dbx, scan.StructMapper[*Model](), query, args...)
	if err != nil {
		slog.Error("Error executing Delete query", slog.String("query", query), slog.Any("args", args), slog.Any("error", err))
		return 0, err
	}

	return result.RowsAffected(), nil
}

// Count returns the number of records that match the provided filters
func (r *PostgresRepository[Model]) Count(ctx context.Context, dbx db.Dbx, where *map[string]any) (int64, error) {
	args := []any{}
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", r.builder.Table())
	if expr, err := r.builder.WhereError(where, &args, nil); err != nil {
		return 0, err
	} else if expr != "" {
		query += fmt.Sprintf(" WHERE %s", expr)
	}

	slog.Info("Executing Get query", slog.String("query", query), slog.Any("args", args))

	// Execute the query and scan the results
	count, err := pgxscan.One(ctx, dbx, scan.SingleColumnMapper[int64], query, args...)

	// result, err := r.builder.Scan(dbx.Query(ctx, query, args...))
	if err != nil {
		slog.Error("Error executing Get query", slog.String("query", query), slog.Any("args", args), slog.Any("error", err))
		return 0, err
	}

	return count, nil
}
