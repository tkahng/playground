package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository[Model any] interface {
	Get(ctx context.Context, dbx DBTX, where *map[string]any, order *map[string]string, limit *int, skip *int) ([]*Model, error)
	GetOne(ctx context.Context, dbx DBTX, where *map[string]any) (*Model, error)
	Put(ctx context.Context, dbx DBTX, models []Model) ([]*Model, error)
	Post(ctx context.Context, dbx DBTX, models []Model) ([]*Model, error)
	DeleteReturn(ctx context.Context, dbx DBTX, where *map[string]any) ([]*Model, error)
	Delete(ctx context.Context, dbx DBTX, where *map[string]any) (int64, error)
	Count(ctx context.Context, dbx DBTX, where *map[string]any) (int64, error)
	Builder() SQLBuilderInterface
}

type Field struct {
	idx  int
	name string
}

type Relation struct {
	one          bool
	through      string
	throughField string
	endField     string
	src          string
	dest         string
	table        string
}

type SQLBuilder[Model any] struct {
	table       string
	keys        []string
	fields      []Field
	columnNames []string
	relations   map[string]Relation
	operations  map[string]func(string, ...string) string
	identifier  func(string) string
	parameter   func(reflect.Value, *[]any) string
	generator   func(reflect.StructField, *[]any) string
}

type SQLBuilderInterface interface {
	Table() string
	ColumnNames() []string
	Where(where *map[string]any, args *[]any, run func(string) []string) string
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

var registry = map[string]SQLBuilderInterface{}

func NewSQLBuilder[Model any]() *SQLBuilder[Model] {
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
	// Reflect on the Model type to extract metadata
	_type := reflect.TypeFor[Model]()

	table := strings.ToLower(_type.Name())
	fields := []Field{}
	columnNames := []string{}
	relations := map[string]Relation{}
	operations_ := map[string]func(string, ...string) string{}
	for idx := range _type.NumField() {
		_field := _type.Field(idx)

		if _field.Name == "_" {
			// Field named "_" is model information
			if tag := _field.Tag.Get("db"); tag != "" {
				table = strings.Split(tag, ",")[0]
			}
		} else {
			// Other fields are model attributes
			if tag := _field.Tag.Get("db"); tag != "" {
				if _field.Tag.Get("table") != "" {
					// Relation field detected
					through := _field.Tag.Get("through")
					throughs := strings.Split(through, ",")
					var throught, throughf, efield string
					if len(throughs) == 3 {
						throught = throughs[0]
						throughf = throughs[1]
						efield = throughs[2]
					}
					relations[tag] = Relation{
						one:          _field.Type.Kind() == reflect.Struct,
						src:          _field.Tag.Get("src"),
						dest:         _field.Tag.Get("dest"),
						table:        _field.Tag.Get("table"),
						through:      throught,
						throughField: throughf,
						endField:     efield,
					}
				} else {
					// Primitive fields detected
					name := strings.Split(tag, ",")[0]
					fields = append(fields, Field{idx, name})
					columnNames = append(columnNames, name)
					// Add base operations for the field
					for key, value := range operations {
						operations_[name+key] = value
					}

					// Check if the field has a method named "Operations"
					// Then add its custom defined operations for the field
					if _method, ok := _field.Type.MethodByName("Operations"); ok {
						var model Model
						value := reflect.ValueOf(model).FieldByName(_field.Name)
						operations := _method.Func.Call([]reflect.Value{value})[0].Interface()
						for key, value := range operations.(map[string]func(string, ...string) string) {
							operations_[name+key] = value
						}
					}
				}
			}
		}
	}

	slog.Debug("SQLBuilder initialized", slog.String("table", table), slog.Any("fields", fields), slog.Any("relations", relations))

	result := &SQLBuilder[Model]{
		table:      table,
		keys:       []string{fields[0].name},
		fields:     fields,
		relations:  relations,
		operations: operations_,
		identifier: identifier,
		parameter:  parameter,
		generator:  nil,
	}

	registry[table] = result

	return result
}
func (b *SQLBuilder[Model]) ColumnNames() []string {
	return b.columnNames
}

// Returns the table name with proper identifier formatting
func (b *SQLBuilder[Model]) Table() string {
	slog.Debug("Fetching table name", slog.String("table", b.table))
	return b.identifier(b.table)
}

// Returns a comma-separated list of field names with proper identifier formatting
func (b *SQLBuilder[Model]) Fields(prefix string) string {
	result := []string{}
	for _, field := range b.fields {
		result = append(result, prefix+b.identifier(field.name))
	}

	slog.Debug("Fetching fields", slog.Any("fields", result))
	return strings.Join(result, ",")
}

func (b *SQLBuilder[Model]) ValuesError(values *[]Model, args *[]any, keys *[]any) (fields string, vals string, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Error occurred during values generation", slog.Any("error", r),
				slog.String("table", b.table),
				slog.Any("values", values),
				slog.Any("args", args),
				slog.Any("keys", keys),
			)
			err = fmt.Errorf("error generating values for table %s", b.table)
		}
	}()
	fields, vals = b.Values(values, args, keys)
	return
}

// Constructs the VALUES clause for an INSERT query
func (b *SQLBuilder[Model]) Values(values *[]Model, args *[]any, keys *[]any) (string, string) {
	if values == nil {
		return "", ""
	}

	// Generate the field names for the VALUES clause
	fields := []string{}
	for idx, field := range b.fields {
		if idx == 0 {
			// The first field is the primary key
			if b.generator != nil {
				// If a generator function is provided, primary key will be generated
				fields = append(fields, b.identifier(field.name))
			}
		} else {
			// Other fields are added to the VALUES clause
			fields = append(fields, b.identifier(field.name))
		}
	}

	// Generate the field values for the VALUES clause
	result := []string{}
	for _, model := range *values {
		_type := reflect.TypeOf(model)
		_value := reflect.ValueOf(model)

		// Generate the values for the current model
		items := []string{}
		for idx, field := range b.fields {
			if idx == 0 {
				// The first field is the primary key
				if b.generator != nil {
					// If a generator function is provided, use it to generate the key
					items = append(items, b.generator(_type.Field(field.idx), keys))
				}
			} else {
				// Other fields are added to the VALUES clause
				items = append(items, b.parameter(_value.Field(field.idx), args))
			}
		}

		result = append(result, "("+strings.Join(items, ",")+")")
	}

	slog.Debug("Constructed VALUES clause", slog.Any("values", result))
	return strings.Join(fields, ","), strings.Join(result, ",")
}

func (b *SQLBuilder[Model]) SetError(set *Model, args *[]any, where *map[string]any) (ret string, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Error occurred during Set generation", slog.Any("error", r),
				slog.String("table", b.table),
				slog.Any("set", set),
				slog.Any("where", where),
			)
			err = fmt.Errorf("error generating Set for table %s", b.table)
		}
	}()
	return b.Set(set, args, where), nil
}

// Constructs the SET clause for an UPDATE query
func (b *SQLBuilder[Model]) Set(set *Model, args *[]any, where *map[string]any) string {
	if set == nil {
		return ""
	}

	_value := reflect.ValueOf(*set)

	// Generate the field names for the SET clause
	result := []string{}
	for idx, field := range b.fields {
		if idx == 0 {
			// The first field is the primary key
			// Use it to construct the WHERE clause
			if where != nil {
				// Get the field value
				_field := _value.Field(field.idx)
				for _field.Kind() == reflect.Pointer {
					_field = _field.Elem()
				}

				// Set the WHERE clause condition based on the field type
				switch _field.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					(*where)[field.name] = map[string]any{"_eq": fmt.Sprintf("%d", _field.Int())}
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					(*where)[field.name] = map[string]any{"_eq": fmt.Sprintf("%d", _field.Uint())}
				case reflect.Float32, reflect.Float64:
					(*where)[field.name] = map[string]any{"_eq": fmt.Sprintf("%f", _field.Float())}
				case reflect.Complex64, reflect.Complex128:
					(*where)[field.name] = map[string]any{"_eq": fmt.Sprintf("%f", _field.Complex())}
				case reflect.String:
					(*where)[field.name] = map[string]any{"_eq": _field.String()}
				default:
					u, ok := _field.Interface().(uuid.UUID)
					if ok {
						(*where)[field.name] = map[string]any{"_eq": u.String()}
					} else {
						panic("Invalid identifier type")
					}
				}
			}
		} else {
			// Other fields are added to the SET clause
			result = append(result, field.name+"="+b.parameter(_value.Field(field.idx), args))
		}
	}

	slog.Debug("Constructed SET clause", slog.String("set", strings.Join(result, ",")))
	return strings.Join(result, ",")
}

// Constructs the ORDER BY clause for a query
func (b *SQLBuilder[Model]) Order(order *map[string]string) string {
	if order == nil {
		return ""
	}

	// Generate the field names for the ORDER BY clause
	result := []string{}
	for key, val := range *order {
		if slices.Contains(b.columnNames, key) {
			result = append(result, fmt.Sprintf("%s %s", b.identifier(key), strings.ToUpper(val)))
		}
	}

	slog.Debug("Constructed ORDER BY clause", slog.Any("order", result))
	return strings.Join(result, ",")
}

func (b *SQLBuilder[Model]) WhereError(where *map[string]any, args *[]any, run func(string) []string) (ret string, err error) {
	if where == nil {
		return "", nil
	}
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Error occurred during where generation", slog.Any("error", r),
				slog.String("table", b.table),
				slog.Any("where", where),
			)
			err = fmt.Errorf("error generating where for table %s", b.table)
		}
	}()
	ret = b.Where(where, args, run)
	return
}

// Constructs the WHERE clause for a query
func (b *SQLBuilder[Model]) Where(where *map[string]any, args *[]any, run func(string) []string) string {
	fmt.Println("where for ", b.table)
	if where == nil {
		slog.Debug("No WHERE clause provided")
		fmt.Println("No WHERE clause provided")
		return ""
	}

	// Check for special conditions
	// _not, _and, and _or are used for logical operations
	if item, ok := (*where)["_not"]; ok {
		expr := item.(map[string]any)

		return "NOT (" + b.Where(&expr, args, run) + ")"
	} else if items, ok := (*where)["_and"]; ok {
		result := []string{}
		for _, item := range items.([]any) {
			expr := item.(map[string]any)
			result = append(result, b.Where(&expr, args, run))
		}

		return "(" + strings.Join(result, " AND ") + ")"
	} else if items, ok := (*where)["_or"]; ok {
		result := []string{}
		for _, item := range items.([]any) {
			expr := item.(map[string]any)
			result = append(result, b.Where(&expr, args, run))
		}

		return "(" + strings.Join(result, " OR ") + ")"
	}

	// Otherwise, construct the WHERE clause based on the field names and operations
	result := []string{}
	for key, item := range *where {
		fmt.Println(key, item)
		for op, value := range item.(map[string]any) {
			if handler, ok := b.operations[key+op]; ok {
				// Primitive field condition detected
				_value := reflect.ValueOf(value)

				if _value.Kind() == reflect.String {
					// String values are passed to operation handler as single parameter
					result = append(result, handler(b.identifier(key), b.parameter(_value, args)))
				} else if _value.Kind() == reflect.Slice || _value.Kind() == reflect.Array {
					// Slice or array values are passed to operation handler as a list of parameters
					items := []string{}
					for i := range _value.Len() {
						items = append(items, b.parameter(_value.Index(i), args))
					}

					result = append(result, handler(b.identifier(key), items...))
				}
			} else {
				// Relation field condition detected
				if relation, ok := b.relations[key]; ok {
					var builder SQLBuilderInterface
					// Get the target SQLBuilder for the relation
					if bld, ok := registry[relation.table]; !ok {
						continue
					} else {
						// Get the target SQLBuilder for the relation
						builder = bld
					}

					// Construct the sub-query for the related table
					args_ := []any{}
					where := item.(map[string]any)
					var query string
					if relation.through != "" {
						query = fmt.Sprintf(
							"SELECT %s FROM %s join %s on %s.%s = %s.%s",
							b.identifier(relation.dest),
							b.identifier(relation.through),
							builder.Table(),
							builder.Table(),
							b.identifier(relation.endField),
							b.identifier(relation.through),
							b.identifier(relation.throughField),
						)
					} else {
						query = fmt.Sprintf("SELECT %s FROM %s", b.identifier(relation.dest), builder.Table())
					}
					if expr := builder.Where(&where, &args_, run); expr != "" {
						query += fmt.Sprintf(" WHERE %s", expr)
					}

					if run == nil {
						*args = append(*args, args_...)
						if inop, ok := b.operations[relation.src+"_in"]; ok {
							result = append(result, inop(b.identifier(relation.src), query))
						}
						// If no run function is provided, sub-query is added to the main query
					} else {
						if inop, ok := b.operations[relation.src+"_in"]; ok {
							result = append(result, inop(b.identifier(relation.src), run(query)...))
						}
						// If a run function is provided, sub-query is executed and its result is added to the main query
					}
				} else {
					slog.Warn("Unknown field key or operation", slog.Any("field", key), slog.Any("operation", op), slog.Any("table", b.table))
				}
			}
		}
	}

	slog.Debug("Constructed WHERE clause", slog.Any("where", result))
	return strings.Join(result, " AND ")
}
