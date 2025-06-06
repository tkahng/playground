package crudrepo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

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
	table        string
	keys         []string
	fields       []Field
	columnNames  []string
	relations    map[string]Relation
	operations   map[string]func(string, ...string) string
	identifier   func(string) string
	parameter    func(reflect.Value, *[]any) string
	generator    func(reflect.StructField, *[]any) (string, error)
	skipIdInsert bool // Skip inserting the primary key field if it is generated
}

type SQLBuilderInterface interface {
	Table() string
	ColumnNames() []string
	Where(where *map[string]any, args *[]any, run func(string) []string) string
}

var registry = map[string]SQLBuilderInterface{}

type SQLBuilderOptions[Model any] func(*SQLBuilder[Model]) error

func UuidV7Generator[Model any](builder *SQLBuilder[Model]) error {
	if builder == nil {
		return errors.New("SQLBuilder cannot be nil")
	}
	builder.generator = func(field reflect.StructField, keys *[]any) (string, error) {
		id, err := uuid.NewV7()
		if err != nil {
			slog.Error("Error generating UUID v7", slog.Any("error", err), slog.String("field", field.Name))
			return "", fmt.Errorf("error generating UUID v7 for field %s: %w", field.Name, err)
		}
		return id.String(), nil
	}
	return nil
}

const (
	// Eq is the equality operator
	Eq = "_eq"
	// Neq is the inequality operator
	Neq = "_neq"
	// Gt is the greater than operator
	Gt = "_gt"
	// Gte is the greater than or equal to operator
	Gte = "_gte"
	// Lt is the less than operator
	Lt = "_lt"
	// Lte is the less than or equal to operator
	Lte = "_lte"
	// Like is the LIKE operator
	Like = "_like"
	// Nlike is the NOT LIKE operator
	Nlike = "_nlike"
	// Ilike is the ILIKE operator (case-insensitive LIKE)
	Ilike = "_ilike"

	// Nilike is the NOT ILIKE operator (case-insensitive NOT LIKE)
	Nilike = "_nilike"
	// In is the IN operator
	In = "_in"
	// Nin is the NOT IN operator
	Nin = "_nin"
)

func NewSQLBuilder[Model any](opts ...SQLBuilderOptions[Model]) *SQLBuilder[Model] {
	operations := map[string]func(string, ...string) string{
		Eq:     func(key string, values ...string) string { return fmt.Sprintf("%s = %s", key, values[0]) },
		Neq:    func(key string, values ...string) string { return fmt.Sprintf("%s != %s", key, values[0]) },
		Gt:     func(key string, values ...string) string { return fmt.Sprintf("%s > %s", key, values[0]) },
		Gte:    func(key string, values ...string) string { return fmt.Sprintf("%s >= %s", key, values[0]) },
		Lt:     func(key string, values ...string) string { return fmt.Sprintf("%s < %s", key, values[0]) },
		Lte:    func(key string, values ...string) string { return fmt.Sprintf("%s <= %s", key, values[0]) },
		Like:   func(key string, values ...string) string { return fmt.Sprintf("%s LIKE %s", key, values[0]) },
		Nlike:  func(key string, values ...string) string { return fmt.Sprintf("%s NOT LIKE %s", key, values[0]) },
		Ilike:  func(key string, values ...string) string { return fmt.Sprintf("%s ILIKE %s", key, values[0]) },
		Nilike: func(key string, values ...string) string { return fmt.Sprintf("%s NOT ILIKE %s", key, values[0]) },
		In: func(key string, values ...string) string {
			return fmt.Sprintf("%s IN (%s)", key, strings.Join(values, ","))
		},
		Nin: func(key string, values ...string) string {
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

	result := &SQLBuilder[Model]{
		table:        table,
		columnNames:  columnNames,
		keys:         []string{fields[0].name},
		fields:       fields,
		relations:    relations,
		operations:   operations_,
		identifier:   identifier,
		parameter:    parameter,
		generator:    nil,
		skipIdInsert: true,
	}
	for _, opt := range opts {
		if err := opt(result); err != nil {
			slog.Error("Error applying SQLBuilder option", slog.Any("error", err))
			panic(fmt.Sprintf("Error applying SQLBuilder option: %v", err))
		}
	}

	registry[table] = result

	return result
}
func (b *SQLBuilder[Model]) ColumnNames() []string {
	return b.columnNames
}

// Returns the table name with proper identifier formatting
func (b *SQLBuilder[Model]) Table() string {
	return b.identifier(b.table)
}

// Returns a comma-separated list of field names with proper identifier formatting
func (b *SQLBuilder[Model]) Fields(prefix string) string {
	result := []string{}
	for _, field := range b.fields {
		result = append(result, prefix+b.identifier(field.name))
	}

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
	fields, vals, err = b.Values(values, args, keys)
	return
}

var timestampNames = []string{"created_at", "updated_at"}

// Constructs the VALUES clause for an INSERT query
func (b *SQLBuilder[Model]) Values(values *[]Model, args *[]any, keys *[]any) (fields string, vals string, err error) {
	if values == nil {
		err = fmt.Errorf("values cannot be nil")
		return
	}

	// Generate the field names for the VALUES clause
	fieldsArray := []string{}
	for idx, field := range b.fields {
		if idx == 0 {
			// The first field is the primary key
			if b.generator != nil {
				// If a generator function is provided, primary key will be generated
				fieldsArray = append(fieldsArray, b.identifier(field.name))
			}
			if b.skipIdInsert {
				// If skipIdInsert is true, skip inserting the primary key field
				continue
			}
			// Otherwise, add the primary key field to the VALUES clause
			fieldsArray = append(fieldsArray, b.identifier(field.name))
		} else {
			if slices.Contains(timestampNames, field.name) {
				continue // Skip timestamp fields
			}
			// Other fields are added to the VALUES clause
			fieldsArray = append(fieldsArray, b.identifier(field.name))
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
					id, err := b.generator(_type.Field(field.idx), keys)
					if err != nil {
						return "", "", fmt.Errorf("error generating primary key for field %s: %w", field.name, err)
					}
					items = append(items, b.parameter(reflect.ValueOf(id), args))
				}
				if b.skipIdInsert {
					// If skipIdInsert is true, skip inserting the primary key field
					continue
				}
				items = append(items, b.parameter(_value.Field(field.idx), args))
			} else {
				if slices.Contains(timestampNames, field.name) {
					continue // Skip timestamp fields
				}
				// Other fields are added to the VALUES clause
				items = append(items, b.parameter(_value.Field(field.idx), args))
			}
		}

		result = append(result, "("+strings.Join(items, ",")+")")
	}

	fields = strings.Join(fieldsArray, ",")
	vals = strings.Join(result, ",")
	return fields, vals, nil
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
	ret = b.Set(set, args, where)
	return
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
			result = append(result, b.identifier(field.name)+"="+b.parameter(_value.Field(field.idx), args))
		}
	}

	return strings.Join(result, ",")
}

// Constructs the ORDER BY clause for a query
func (b *SQLBuilder[Model]) Order(order *map[string]string) string {
	// fmt.Println("order", order)
	if order == nil {
		return ""
	}

	// Generate the field names for the ORDER BY clause
	result := []string{}
	// fmt.Println("columnnames", b.columnNames)
	for key, val := range *order {
		if slices.Contains(b.columnNames, key) {
			result = append(result, fmt.Sprintf("%s %s", b.identifier(key), strings.ToUpper(val)))
		}
	}

	return strings.Join(result, ",")
}

func (b *SQLBuilder[Model]) WhereError(ctx context.Context, where *map[string]any, args *[]any, run func(string) []string) (ret string, err error) {
	if where == nil {
		return "", nil
	}
	defer func() {
		if r := recover(); r != nil {
			slog.ErrorContext(ctx, "Error occurred during where generation", slog.Any("error", r),
				slog.String("table", b.table),
				slog.Any("where", where),
			)
			err = fmt.Errorf("error generating where for table %s. check your filters", b.table)
		}
	}()
	ret = b.Where(where, args, run)
	return
}

// Constructs the WHERE clause for a query
func (b *SQLBuilder[Model]) Where(where *map[string]any, args *[]any, run func(string) []string) string {
	if where == nil {
		return ""
	}

	// Check for special conditions
	// _not, _and, and _or are used for logical operations
	if item, ok := (*where)["_not"]; ok {
		expr, ok := item.(map[string]any)
		if ok {
			return "NOT (" + b.Where(&expr, args, run) + ")"
		}
	} else if items, ok := (*where)["_and"]; ok {
		result := []string{}
		ands, ok := items.([]map[string]any)
		if ok {
			for _, item := range ands {
				expr := item
				result = append(result, b.Where(&expr, args, run))
			}
		}

		return "(" + strings.Join(result, " AND ") + ")"
	} else if ors, ok := (*where)["_or"]; ok {
		result := []string{}

		orWheres, ok := ors.([]map[string]any)
		if ok {
			for _, item := range orWheres {
				expr := item
				result = append(result, b.Where(&expr, args, run))
			}
		}

		return "(" + strings.Join(result, " OR ") + ")"
	}

	// Otherwise, construct the WHERE clause based on the field names and operations
	result := []string{}
	for key, item := range *where {
		// fmt.Println(key, item)
		for op, value := range item.(map[string]any) {
			if handler, ok := b.operations[key+op]; ok {
				// Primitive field condition detected
				_value := reflect.ValueOf(value)

				if _value.Kind() == reflect.String {
					// String values are passed to operation handler as single parameter
					result = append(result, handler(b.identifier(key), b.parameter(_value, args)))
				} else if it, ok := _value.Interface().(time.Time); ok {
					_newValue := reflect.ValueOf(it.Format(time.RFC3339Nano))
					result = append(result, handler(b.identifier(key), b.parameter(_newValue, args)))
				} else if it, ok := _value.Interface().(fmt.Stringer); ok {
					_newValue := reflect.ValueOf(it.String())
					if _newValue.Kind() == reflect.String {
						// If the value implements fmt.Stringer, use its String method
						result = append(result, handler(b.identifier(key), b.parameter(_newValue, args)))
					}
				} else if _value.Kind() == reflect.Slice || _value.Kind() == reflect.Array {
					// Slice or array values are passed to operation handler as a list of parameters
					items := []string{}
					for i := range _value.Len() {
						if _value.Index(i).Kind() == reflect.String {
							items = append(items, b.parameter(_value.Index(i), args))
						} else if it, ok := _value.Index(i).Interface().(fmt.Stringer); ok {
							// If the value implements fmt.Stringer, use its String method
							items = append(items, b.parameter(reflect.ValueOf(it.String()), args))
						}
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
					if expr := builder.Where(&where, args, run); expr != "" {
						query += fmt.Sprintf(" WHERE %s", expr)
					}
					if run == nil {
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
				}
			}
		}
	}

	return strings.Join(result, " AND ")
}
