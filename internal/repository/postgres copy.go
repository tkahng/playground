package repository

// import (
// 	"context"
// 	"fmt"
// 	"log/slog"
// 	"reflect"
// 	"strings"

// 	"github.com/Masterminds/squirrel"
// 	sc "github.com/Masterminds/squirrel"
// )

// // PostgresRepository provides CRUD operations for Postgres
// type PostgresRepository[Model any] struct {
// 	db      DBTX
// 	builder *SQLBuilder[Model]
// }

// // NewPostgresRepository initializes a new PostgresRepository
// func NewPostgresRepository[Model any](db DBTX) *PostgresRepository[Model] {
// 	// Define SQL operators and helper functions for query building
// 	operations := map[string]func(string, ...string) string{
// 		"_eq":     func(key string, values ...string) string { return fmt.Sprintf("%s = %s", key, values[0]) },
// 		"_neq":    func(key string, values ...string) string { return fmt.Sprintf("%s != %s", key, values[0]) },
// 		"_gt":     func(key string, values ...string) string { return fmt.Sprintf("%s > %s", key, values[0]) },
// 		"_gte":    func(key string, values ...string) string { return fmt.Sprintf("%s >= %s", key, values[0]) },
// 		"_lt":     func(key string, values ...string) string { return fmt.Sprintf("%s < %s", key, values[0]) },
// 		"_lte":    func(key string, values ...string) string { return fmt.Sprintf("%s <= %s", key, values[0]) },
// 		"_like":   func(key string, values ...string) string { return fmt.Sprintf("%s LIKE %s", key, values[0]) },
// 		"_nlike":  func(key string, values ...string) string { return fmt.Sprintf("%s NOT LIKE %s", key, values[0]) },
// 		"_ilike":  func(key string, values ...string) string { return fmt.Sprintf("%s ILIKE %s", key, values[0]) },
// 		"_nilike": func(key string, values ...string) string { return fmt.Sprintf("%s NOT ILIKE %s", key, values[0]) },
// 		"_in": func(key string, values ...string) string {
// 			return fmt.Sprintf("%s IN (%s)", key, strings.Join(values, ","))
// 		},
// 		"_nin": func(key string, values ...string) string {
// 			return fmt.Sprintf("%s NOT IN (%s)", key, strings.Join(values, ","))
// 		},
// 	}
// 	identifier := func(name string) string {
// 		return fmt.Sprintf("\"%s\"", name)
// 	}
// 	parameter := func(value reflect.Value, args *[]any) string {
// 		*args = append(*args, value.Interface())
// 		return fmt.Sprintf("$%d", len(*args))
// 	}

// 	return &PostgresRepository[Model]{
// 		db:      db,
// 		builder: NewSQLBuilder[Model](operations, identifier, parameter, nil),
// 	}
// }

// func (r *PostgresRepository[Model]) Get(ctx context.Context, where *map[string]any, order *map[string]any, limit *uint64, skip *uint64) ([]Model, error) {
// 	// query := fmt.Sprintf("SELECT %s FROM %s", r.builder.Fields(""), r.builder.Table())
// 	query := sc.Select(r.builder.Fields("")).From(r.builder.Table())
// 	query = r.builder.Where(where, query)
// 	// if expr := r.builder.Order(order); expr != "" {
// 	// 	query += fmt.Sprintf(" ORDER BY %s", expr)
// 	// }
// 	query = r.builder.Order(order, query)
// 	if limit != nil {
// 		query = query.Limit(*limit)
// 	}
// 	if skip != nil {
// 		query = query.Offset(*skip)
// 	}
// 	sql, args, err := query.PlaceholderFormat(sc.Dollar).ToSql()
// 	if err != nil {
// 		slog.Error("Error executing Get query", slog.String("query", sql), slog.Any("args", args), slog.Any("error", err))
// 		return nil, err
// 	}
// 	slog.Info("Executing Get query", slog.String("query", sql), slog.Any("args", args))

// 	// Execute the query and scan the results
// 	result, err := r.builder.Scan(r.db.Query(ctx, sql, args...))
// 	if err != nil {
// 		slog.Error("Error executing Get query", slog.String("query", sql), slog.Any("args", args), slog.Any("error", err))
// 		return nil, err
// 	}

// 	return result, nil
// }

// // Put updates existing records in the database
// func (r *PostgresRepository[Model]) Put(ctx context.Context, models *[]Model) ([]Model, error) {
// 	result := []Model{}

// 	// Begin a transaction
// 	tx, err := r.db.Begin(ctx)
// 	if err != nil {
// 		slog.Error("Error starting transaction for Put", slog.Any("error", err))
// 		return nil, err
// 	}
// 	defer tx.Rollback(ctx)
// 	// Update each model in the database
// 	for _, model := range *models {
// 		where := map[string]any{}
// 		q := squirrel.Update(r.builder.Table())
// 		q = r.builder.Set(&model, nil, &where, q)
// 		// query := fmt.Sprintf("UPDATE %s SET %s", r.builder.Table(), r.builder.Set(&model, &args, &where))
// 		// if expr := r.builder.Where(&where, &args, nil); expr != "" {
// 		// 	query += fmt.Sprintf(" WHERE %s", expr)
// 		// }
// 		q = r.builder.WhereUpdate(&where, q)
// 		// query += fmt.Sprintf(" RETURNING %s", r.builder.Fields(""))
// 		sql, args, err := q.PlaceholderFormat(sc.Dollar).ToSql()
// 		if err != nil {
// 			slog.Error("Error executing Put query", slog.String("query", sql), slog.Any("args", args), slog.Any("error", err))
// 			return nil, err
// 		}
// 		slog.Info("Executing Put query", slog.String("query", sql), slog.Any("args", args))

// 		items, err := r.builder.Scan(tx.Query(ctx, sql, args...))
// 		if err != nil {
// 			slog.Error("Error executing Put query", slog.String("query", sql), slog.Any("args", args), slog.Any("error", err))
// 			return nil, err
// 		}

// 		result = append(result, items...)
// 	}

// 	// Commit the transaction
// 	if err := tx.Commit(ctx); err != nil {
// 		slog.Error("Error committing transaction for Put", slog.Any("error", err))
// 		return nil, err
// 	}

// 	return result, nil
// }

// // Post inserts new records into the database
// func (r *PostgresRepository[Model]) Post(ctx context.Context, models *[]Model) ([]Model, error) {
// 	args := []any{}
// 	query := fmt.Sprintf("INSERT INTO %s", r.builder.Table())
// 	if fields, values := r.builder.Values(models, &args, nil); fields != "" && values != "" {
// 		query += fmt.Sprintf(" (%s) VALUES %s", fields, values)
// 	}
// 	query += fmt.Sprintf(" RETURNING %s", r.builder.Fields(""))

// 	slog.Info("Executing Post query", slog.String("query", query), slog.Any("args", args))

// 	// Execute the query and scan the results
// 	result, err := r.builder.Scan(r.db.Query(ctx, query, args...))
// 	if err != nil {
// 		slog.Error("Error executing Post query", slog.String("query", query), slog.Any("args", args), slog.Any("error", err))
// 		return nil, err
// 	}

// 	return result, nil
// }

// // Delete removes records from the database based on the provided filters
// func (r *PostgresRepository[Model]) Delete(ctx context.Context, where *map[string]any) ([]Model, error) {
// 	q := squirrel.Delete(r.builder.Table())
// 	q = r.builder.WhereDelete(where, q)
// 	query := fmt.Sprintf("RETURNING %s", r.builder.Fields(""))
// 	q = q.Suffix(query)
// 	sql, args, err := q.PlaceholderFormat(sc.Dollar).ToSql()
// 	if err != nil {
// 		slog.Error("Error executing Delete query", slog.String("query", sql), slog.Any("args", args), slog.Any("error", err))
// 		return nil, err
// 	}
// 	slog.Info("Executing Delete query", slog.String("query", sql), slog.Any("args", args))

// 	// Execute the query and scan the results
// 	result, err := r.builder.Scan(r.db.Query(ctx, sql, args...))
// 	if err != nil {
// 		slog.Error("Error executing Delete query", slog.String("query", query), slog.Any("args", args), slog.Any("error", err))
// 		return nil, err
// 	}

// 	return result, nil
// }
