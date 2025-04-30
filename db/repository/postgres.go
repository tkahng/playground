package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
)

// PostgresRepository provides CRUD operations for Postgres
type PostgresRepository[Model any] struct {
	db      *sql.DB
	builder *SQLBuilder[Model]
}

// NewPostgresRepository initializes a new PostgresRepository
func NewPostgresRepository[Model any](db *sql.DB) *PostgresRepository[Model] {
	// Define SQL operators and helper functions for query building
	operations := map[string]func(string, ...string) string{
		"_eq":     func(key string, values ...string) string { return fmt.Sprintf("%s = %s", key, values[0]) },
		"_neq":    func(key string, values ...string) string { return fmt.Sprintf("%s != %s", key, values[0]) },
		"_gt":     func(key string, values ...string) string { return fmt.Sprintf("%s > %s", key, values[0]) },
		"_gte":    func(key string, values ...string) string { return fmt.Sprintf("%s >= %s", key, values[0]) },
		"_lt":     func(key string, values ...string) string { return fmt.Sprintf("%s < %s", key, values[0]) },
		"_lte":    func(key string, values ...string) string { return fmt.Sprintf("%s <= %s", key, values[0]) },
		"_like":   func(key string, values ...string) string { return fmt.Sprintf("%s LIKE %s", key, values[0]) },
		"_nlike":  func(key string, values ...string) string { return fmt.Sprintf("%s NOT LIKE %s", key, values[0]) },
		"_ilike":  func(key string, values ...string) string { return fmt.Sprintf("%s ILIKE %s", key, values[0]) },
		"_nilike": func(key string, values ...string) string { return fmt.Sprintf("%s NOT ILIKE %s", key, values[0]) },
		"_in": func(key string, values ...string) string {
			return fmt.Sprintf("%s IN (%s)", key, strings.Join(values, ","))
		},
		"_nin": func(key string, values ...string) string {
			return fmt.Sprintf("%s NOT IN (%s)", key, strings.Join(values, ","))
		},
	}
	identifier := func(name string) string {
		return fmt.Sprintf("\"%s\"", name)
	}
	parameter := func(value reflect.Value, args *[]any) string {
		*args = append(*args, value.Interface())
		return fmt.Sprintf("$%d", len(*args))
	}

	return &PostgresRepository[Model]{
		db:      db,
		builder: NewSQLBuilder[Model](operations, identifier, parameter, nil),
	}
}

// Get retrieves records from the database based on the provided filters
func (r *PostgresRepository[Model]) Get(ctx context.Context, where *map[string]any, order *map[string]any, limit *int, skip *int) ([]Model, error) {
	args := []any{}
	query := fmt.Sprintf("SELECT %s FROM %s", r.builder.Fields(""), r.builder.Table())
	if expr := r.builder.Where(where, &args, nil); expr != "" {
		query += fmt.Sprintf(" WHERE %s", expr)
	}
	if expr := r.builder.Order(order); expr != "" {
		query += fmt.Sprintf(" ORDER BY %s", expr)
	}
	if *limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", *limit)
	}
	if *skip > 0 {
		query += fmt.Sprintf(" OFFSET %d", *skip)
	}

	slog.Info("Executing Get query", slog.String("query", query), slog.Any("args", args))

	// Execute the query and scan the results
	result, err := r.builder.Scan(r.db.QueryContext(ctx, query, args...))
	if err != nil {
		slog.Error("Error executing Get query", slog.String("query", query), slog.Any("args", args), slog.Any("error", err))
		return nil, err
	}

	return result, nil
}

// Put updates existing records in the database
func (r *PostgresRepository[Model]) Put(ctx context.Context, models *[]Model) ([]Model, error) {
	result := []Model{}

	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("Error starting transaction for Put", slog.Any("error", err))
		return nil, err
	}

	// Update each model in the database
	for _, model := range *models {
		args := []any{}
		where := map[string]any{}
		query := fmt.Sprintf("UPDATE %s SET %s", r.builder.Table(), r.builder.Set(&model, &args, &where))
		if expr := r.builder.Where(&where, &args, nil); expr != "" {
			query += fmt.Sprintf(" WHERE %s", expr)
		}
		query += fmt.Sprintf(" RETURNING %s", r.builder.Fields(""))

		slog.Info("Executing Put query", slog.String("query", query), slog.Any("args", args))

		items, err := r.builder.Scan(tx.QueryContext(ctx, query, args...))
		if err != nil {
			slog.Error("Error executing Put query", slog.String("query", query), slog.Any("args", args), slog.Any("error", err))
			tx.Rollback()
			return nil, err
		}

		result = append(result, items...)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		slog.Error("Error committing transaction for Put", slog.Any("error", err))
		return nil, err
	}

	return result, nil
}

// Post inserts new records into the database
func (r *PostgresRepository[Model]) Post(ctx context.Context, models *[]Model) ([]Model, error) {
	args := []any{}
	query := fmt.Sprintf("INSERT INTO %s", r.builder.Table())
	if fields, values := r.builder.Values(models, &args, nil); fields != "" && values != "" {
		query += fmt.Sprintf(" (%s) VALUES %s", fields, values)
	}
	query += fmt.Sprintf(" RETURNING %s", r.builder.Fields(""))

	slog.Info("Executing Post query", slog.String("query", query), slog.Any("args", args))

	// Execute the query and scan the results
	result, err := r.builder.Scan(r.db.QueryContext(ctx, query, args...))
	if err != nil {
		slog.Error("Error executing Post query", slog.String("query", query), slog.Any("args", args), slog.Any("error", err))
		return nil, err
	}

	return result, nil
}

// Delete removes records from the database based on the provided filters
func (r *PostgresRepository[Model]) Delete(ctx context.Context, where *map[string]any) ([]Model, error) {
	args := []any{}
	query := fmt.Sprintf("DELETE FROM %s", r.builder.Table())
	if expr := r.builder.Where(where, &args, nil); expr != "" {
		query += fmt.Sprintf(" WHERE %s", expr)
	}
	query += fmt.Sprintf(" RETURNING %s", r.builder.Fields(""))

	slog.Info("Executing Delete query", slog.String("query", query), slog.Any("args", args))

	// Execute the query and scan the results
	result, err := r.builder.Scan(r.db.QueryContext(ctx, query, args...))
	if err != nil {
		slog.Error("Error executing Delete query", slog.String("query", query), slog.Any("args", args), slog.Any("error", err))
		return nil, err
	}

	return result, nil
}
