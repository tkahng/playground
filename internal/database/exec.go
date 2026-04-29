package database

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
)

type QueryBuilder interface {
	ToSql() (string, []any, error)
}

func QueryWithBuilder[T any](ctx context.Context, db Dbx, query QueryBuilder) ([]T, error) {
	sql, args, err := query.ToSql()
	slog.DebugContext(ctx, "Query With Builder:", slog.String("query", sql), slog.Any("args", args))
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
	slog.DebugContext(ctx, "Exec With Builder:", slog.String("query", sql), slog.Any("args", args))
	return Exec(ctx, db, sql, args...)
}

func QueryAll[T any](ctx context.Context, db Dbx, query string, args ...any) ([]T, error) {
	ctxDbx := GetContextOrDefaultDbx(ctx, db)
	rows, err := ctxDbx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, rowToStructByDBTag[T])
}

// QueryOne executes a query and scans the single result row into T using db struct tags.
// Returns pgx.ErrNoRows if the query returns no rows.
func QueryOne[T any](ctx context.Context, db Dbx, query string, args ...any) (T, error) {
	ctxDbx := GetContextOrDefaultDbx(ctx, db)
	rows, err := ctxDbx.Query(ctx, query, args...)
	if err != nil {
		var zero T
		return zero, err
	}
	return pgx.CollectOneRow(rows, rowToStructByDBTag[T])
}

func Count(ctx context.Context, db Dbx, query string, args ...any) (int64, error) {
	ctxDbx := GetContextOrDefaultDbx(ctx, db)
	rows, err := ctxDbx.Query(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return pgx.CollectOneRow(rows, pgx.RowTo[int64])
}

func QueryOneSingleColumn[T any](ctx context.Context, db Dbx, query string, args ...any) (T, error) {
	ctxDbx := GetContextOrDefaultDbx(ctx, db)
	rows, err := ctxDbx.Query(ctx, query, args...)
	if err != nil {
		var zero T
		return zero, err
	}
	return pgx.CollectOneRow(rows, pgx.RowTo[T])
}

func QueryManySingleColumn[T any](ctx context.Context, db Dbx, query string, args ...any) ([]T, error) {
	ctxDbx := GetContextOrDefaultDbx(ctx, db)
	rows, err := ctxDbx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[T])
}

func Exec(ctx context.Context, db Dbx, query string, args ...any) (int64, error) {
	ctxDbx := GetContextOrDefaultDbx(ctx, db)
	result, err := ctxDbx.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// CountOutput is used by QueryWithBuilder-based count queries.
type CountOutput struct {
	Count int64 `db:"count"`
}

func PgxQueryRowsToStruct[T any](ctx context.Context, db Dbx, query QueryBuilder) ([]*T, error) {
	ctxDbx := GetContextOrDefaultDbx(ctx, db)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	slog.DebugContext(ctx, "PgxQueryRowsToStruct:", slog.String("query", sql), slog.Any("args", args))
	r, err := ctxDbx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(r, pgx.RowToAddrOfStructByNameLax[T])
}

func PgxQuerySingleScalar[T comparable](ctx context.Context, db Dbx, query QueryBuilder) (T, error) {
	ctxDbx := GetContextOrDefaultDbx(ctx, db)
	var zero T
	sql, args, err := query.ToSql()
	if err != nil {
		return zero, err
	}
	slog.DebugContext(ctx, "PgxQuerySingleScalar:", slog.String("query", sql), slog.Any("args", args))
	r, err := ctxDbx.Query(ctx, sql, args...)
	if err != nil {
		return zero, err
	}
	res, err := pgx.CollectRows(r, pgx.RowTo[T])
	if err != nil {
		return zero, err
	}
	if len(res) == 0 {
		return zero, nil
	}
	return res[0], nil
}

// dbTagCache caches column-name → field-index mappings keyed by struct reflect.Type.
// All db-tagged fields are included (both scalar and relation fields).
// Relation fields are used as intermediate navigation nodes for dot-path column aliases.
var dbTagCache sync.Map

// buildDBTagIndex returns a map from db tag name to struct field index for t.
// Skips the blank info field ("_") and fields with no/empty/"-" db tag.
// Relation fields (those with a "table" tag) are included so they can serve
// as intermediate nodes when resolving dotted column aliases like "price.id".
func buildDBTagIndex(t reflect.Type) map[string]int {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if cached, ok := dbTagCache.Load(t); ok {
		return cached.(map[string]int)
	}
	m := make(map[string]int, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Name == "_" {
			continue
		}
		tag := f.Tag.Get("db")
		if tag == "" || tag == "-" {
			continue
		}
		if comma := strings.IndexByte(tag, ','); comma != -1 {
			tag = tag[:comma]
		}
		if tag != "" {
			m[tag] = i
		}
	}
	dbTagCache.Store(t, m)
	return m
}

// resolveFieldAddr follows path into v (an addressable struct Value), allocating any
// nil pointer structs it traverses, and returns the scan destination address.
func resolveFieldAddr(v reflect.Value, path []string) (any, bool) {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, false
	}

	tagIndex := buildDBTagIndex(v.Type())
	fieldIdx, ok := tagIndex[path[0]]
	if !ok {
		return nil, false
	}
	fieldVal := v.Field(fieldIdx)

	if len(path) == 1 {
		return fieldVal.Addr().Interface(), true
	}
	return resolveFieldAddr(fieldVal, path[1:])
}

// rowToStructByDBTag is a pgx.RowToFunc that scans a row into T using "db" struct tags.
// T may be a struct value or a pointer to a struct.
// Dot-separated column aliases (e.g. "stripe_customer.id") are resolved as nested struct paths.
func rowToStructByDBTag[T any](row pgx.CollectableRow) (T, error) {
	tType := reflect.TypeOf((*T)(nil)).Elem()

	isPtr := tType.Kind() == reflect.Ptr
	structType := tType
	if isPtr {
		structType = tType.Elem()
	}
	if structType.Kind() != reflect.Struct {
		var zero T
		return zero, fmt.Errorf("rowToStructByDBTag: T must be a struct or *struct, got %v", tType)
	}

	structVal := reflect.New(structType).Elem()

	descs := row.FieldDescriptions()
	dests := make([]any, len(descs))
	for i, desc := range descs {
		parts := strings.Split(desc.Name, ".")
		addr, ok := resolveFieldAddr(structVal, parts)
		if ok {
			dests[i] = addr
		} else {
			var discard any
			dests[i] = &discard
		}
	}

	if err := row.Scan(dests...); err != nil {
		var zero T
		return zero, err
	}

	if isPtr {
		ptr := reflect.New(structType)
		ptr.Elem().Set(structVal)
		result, ok := ptr.Interface().(T)
		if !ok {
			var zero T
			return zero, fmt.Errorf("rowToStructByDBTag: type assertion to %T failed", zero)
		}
		return result, nil
	}

	result, ok := structVal.Interface().(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("rowToStructByDBTag: type assertion to %T failed", zero)
	}
	return result, nil
}
