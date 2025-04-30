package repository

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type Repository[Model any] interface {
	Get(ctx context.Context, where *map[string]any, order *map[string]any, limit *int, skip *int) ([]Model, error)
	Put(ctx context.Context, models *[]Model) ([]Model, error)
	Post(ctx context.Context, models *[]Model) ([]Model, error)
	Delete(ctx context.Context, where *map[string]any) ([]Model, error)
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
	table      string
	keys       []string
	fields     []Field
	relations  map[string]Relation
	operations map[string]func(string, ...string) string
	identifier func(string) string
	parameter  func(reflect.Value, *[]any) string
	generator  func(reflect.StructField, *[]any) string
}

type SQLBuilderInterface interface {
	Table() string
	Where(where *map[string]any, selq sq.SelectBuilder) sq.SelectBuilder
	WhereUpdate(where *map[string]any, selq sq.UpdateBuilder) sq.UpdateBuilder
}

var registry = map[string]SQLBuilderInterface{}

func NewSQLBuilder[Model any](operations map[string]func(string, ...string) string, identifier func(string) string, parameter func(reflect.Value, *[]any) string, generator func(reflect.StructField, *[]any) string) *SQLBuilder[Model] {
	// Reflect on the Model type to extract metadata
	_type := reflect.TypeFor[Model]()

	table := strings.ToLower(_type.Name())
	fields := []Field{}
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
				if _field.Tag.Get("json") == "-" {
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
		generator:  generator,
	}

	registry[table] = result

	return result
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

// Constructs the SET clause for an UPDATE query
func (b *SQLBuilder[Model]) Set(set *Model, args *[]any, where *map[string]any, q sq.UpdateBuilder) sq.UpdateBuilder {
	if set == nil {
		return q
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
					panic("Invalid identifier type")
				}
			}
		} else {
			// Other fields are added to the SET clause
			q = q.Set(b.identifier(field.name), _value.Field(field.idx).Interface())
			// result = append(result, field.name+"="+b.parameter(_value.Field(field.idx), args))
		}
	}

	slog.Debug("Constructed SET clause", slog.String("set", strings.Join(result, ",")))
	return q
}

// Constructs the ORDER BY clause for a query
func (b *SQLBuilder[Model]) Order(order *map[string]any, query sq.SelectBuilder) sq.SelectBuilder {
	if order == nil {
		return query
	}
	// Generate the field names for the ORDER BY clause
	result := []string{}
	for key, val := range *order {
		result = append(result, fmt.Sprintf("%s %s", b.identifier(key), val))
	}
	slog.Debug("Constructed ORDER BY clause", slog.Any("order", result))
	query = query.OrderBy(result...)
	return query
}

type baba interface {
	From(from string) baba
	Columns(columns ...string) baba
}

// Constructs the WHERE clause for a query
func (b *SQLBuilder[Model]) Where(where *map[string]any, selq sq.SelectBuilder) sq.SelectBuilder {
	q := selq
	if where == nil {
		return q
	}

	// Check for special conditions
	// _not, _and, and _or are used for logical operations
	// if item, ok := (*where)["_not"]; ok {
	// 	expr := item.(map[string]any)
	// 	query, args, err := b.Where2(&expr, selq)
	// 	if err != nil {
	// 		return "", nil, err
	// 	}
	// 	return "NOT (" + query + ")", args, nil
	// } else if items, ok := (*where)["_and"]; ok {
	// 	result := []string{}
	// 	for _, item := range items.([]any) {
	// 		expr := item.(map[string]any)
	// 		query, args, err := b.Where2(&expr, args, )
	// 		if err != nil {
	// 			return "", nil, err
	// 		}
	// 		result = append(result, query)
	// 	}

	// 	return "(" + strings.Join(result, " AND ") + ")", nil, nil
	// } else if items, ok := (*where)["_or"]; ok {
	// 	result := []string{}
	// 	for _, item := range items.([]any) {
	// 		expr := item.(map[string]any)
	// 		result = append(result, b.Where(&expr, args, run))
	// 	}

	// 	return "(" + strings.Join(result, " OR ") + ")", nil, nil
	// }

	// Otherwise, construct the WHERE clause based on the field names and operations
	result := []string{}
	for key, item := range *where {
		fmt.Println(key, item)
		for op, value := range item.(map[string]any) {
			fmt.Println("operation", key+op)
			if handler, ok := b.operations[key+op]; ok {
				// Primitive field condition detected
				_value := reflect.ValueOf(value)

				if _value.Kind() == reflect.String {
					qs := handler(b.identifier(key), "?")
					fmt.Println(qs)
					// fmt.Println(qs, _value.String())
					q = q.Where(qs, _value.String())
					// String values are passed to operation handler as single parameter
					// result = append(result, handler(b.identifier(key), b.parameter(_value, nil)))
				} else if _value.Kind() == reflect.Slice || _value.Kind() == reflect.Array {
					// Slice or array values are passed to operation handler as a list of parameters
					items := []string{}
					vals := []any{}
					for i := range _value.Len() {
						items = append(items, "?")
						vals = append(vals, _value.Index(i).String())
					}
					qs := handler(b.identifier(key), items...)

					q = q.Where(qs, vals...)
				}
			} else {
				// Relation field condition detected
				if relation, ok := b.relations[key]; ok {
					builder := registry[relation.table]

					newWhere := item.(map[string]any)
					fmt.Println("newWhere", newWhere)
					var newQuery sq.SelectBuilder
					if relation.through != "" {
						fmt.Println("through", relation.through, "throughField", relation.throughField, "dest", relation.dest, "src", relation.src)
						// "roles on roles.id = user_roles.role_id",
						on := builder.Table() + " on " + builder.Table() + "." + b.identifier(relation.endField) + " = " + b.identifier(relation.through) + "." + b.identifier(relation.throughField)
						fmt.Println(on)
						newQuery = sq.Select(b.identifier(relation.dest)).From(b.identifier(relation.through)).Join(
							builder.Table() + " on " + builder.Table() + "." + b.identifier(relation.endField) + " = " + b.identifier(relation.through) + "." + b.identifier(relation.throughField),
						)
					} else {
						newQuery = sq.Select(b.identifier(relation.dest)).From(builder.Table())
					}
					newQuery = builder.Where(&newWhere, newQuery)
					newqs, newqargs, err := newQuery.ToSql()
					if err != nil {
						slog.Error("Error during query execution", slog.Any("error", err))
						return q
					}
					q = q.Where(b.identifier(relation.src)+" in ("+newqs+")", newqargs...)
				}
			}
		}
	}

	slog.Debug("Constructed WHERE clause", slog.Any("where", result))
	return q
}
func (b *SQLBuilder[Model]) WhereUpdate(where *map[string]any, selq sq.UpdateBuilder) sq.UpdateBuilder {
	q := selq
	if where == nil {
		return q
	}

	// Check for special conditions
	// _not, _and, and _or are used for logical operations
	// if item, ok := (*where)["_not"]; ok {
	// 	expr := item.(map[string]any)
	// 	query, args, err := b.Where2(&expr, selq)
	// 	if err != nil {
	// 		return "", nil, err
	// 	}
	// 	return "NOT (" + query + ")", args, nil
	// } else if items, ok := (*where)["_and"]; ok {
	// 	result := []string{}
	// 	for _, item := range items.([]any) {
	// 		expr := item.(map[string]any)
	// 		query, args, err := b.Where2(&expr, args, )
	// 		if err != nil {
	// 			return "", nil, err
	// 		}
	// 		result = append(result, query)
	// 	}

	// 	return "(" + strings.Join(result, " AND ") + ")", nil, nil
	// } else if items, ok := (*where)["_or"]; ok {
	// 	result := []string{}
	// 	for _, item := range items.([]any) {
	// 		expr := item.(map[string]any)
	// 		result = append(result, b.Where(&expr, args, run))
	// 	}

	// 	return "(" + strings.Join(result, " OR ") + ")", nil, nil
	// }

	// Otherwise, construct the WHERE clause based on the field names and operations
	result := []string{}
	for key, item := range *where {
		fmt.Println(key, item)
		for op, value := range item.(map[string]any) {
			fmt.Println("operation", key+op)
			if handler, ok := b.operations[key+op]; ok {
				// Primitive field condition detected
				_value := reflect.ValueOf(value)

				if _value.Kind() == reflect.String {
					qs := handler(b.identifier(key), "?")
					fmt.Println(qs)
					// fmt.Println(qs, _value.String())
					q = q.Where(qs, _value.String())
					// String values are passed to operation handler as single parameter
					// result = append(result, handler(b.identifier(key), b.parameter(_value, nil)))
				} else if _value.Kind() == reflect.Slice || _value.Kind() == reflect.Array {
					// Slice or array values are passed to operation handler as a list of parameters
					items := []string{}
					vals := []any{}
					for i := range _value.Len() {
						items = append(items, "?")
						vals = append(vals, _value.Index(i).String())
					}
					qs := handler(b.identifier(key), items...)

					q = q.Where(qs, vals...)
				}
			} else {
				// Relation field condition detected
				if relation, ok := b.relations[key]; ok {
					builder := registry[relation.table]

					newWhere := item.(map[string]any)
					fmt.Println("newWhere", newWhere)
					var newQuery sq.SelectBuilder
					if relation.through != "" {
						fmt.Println("through", relation.through, "throughField", relation.throughField, "dest", relation.dest, "src", relation.src)
						// "roles on roles.id = user_roles.role_id",
						on := builder.Table() + " on " + builder.Table() + "." + b.identifier(relation.endField) + " = " + b.identifier(relation.through) + "." + b.identifier(relation.throughField)
						fmt.Println(on)
						newQuery = sq.Select(b.identifier(relation.dest)).From(b.identifier(relation.through)).Join(
							builder.Table() + " on " + builder.Table() + "." + b.identifier(relation.endField) + " = " + b.identifier(relation.through) + "." + b.identifier(relation.throughField),
						)
					} else {
						newQuery = sq.Select(b.identifier(relation.dest)).From(builder.Table())
					}
					newQuery = builder.Where(&newWhere, newQuery)
					newqs, newqargs, err := newQuery.ToSql()
					if err != nil {
						slog.Error("Error during query execution", slog.Any("error", err))
						return q
					}
					q = q.Where(b.identifier(relation.src)+" in ("+newqs+")", newqargs...)
				}
			}
		}
	}

	slog.Debug("Constructed WHERE clause", slog.Any("where", result))
	return q
}
func (b *SQLBuilder[Model]) WhereDelete(where *map[string]any, selq sq.DeleteBuilder) sq.DeleteBuilder {
	q := selq
	if where == nil {
		return q
	}

	// Check for special conditions
	// _not, _and, and _or are used for logical operations
	// if item, ok := (*where)["_not"]; ok {
	// 	expr := item.(map[string]any)
	// 	query, args, err := b.Where2(&expr, selq)
	// 	if err != nil {
	// 		return "", nil, err
	// 	}
	// 	return "NOT (" + query + ")", args, nil
	// } else if items, ok := (*where)["_and"]; ok {
	// 	result := []string{}
	// 	for _, item := range items.([]any) {
	// 		expr := item.(map[string]any)
	// 		query, args, err := b.Where2(&expr, args, )
	// 		if err != nil {
	// 			return "", nil, err
	// 		}
	// 		result = append(result, query)
	// 	}

	// 	return "(" + strings.Join(result, " AND ") + ")", nil, nil
	// } else if items, ok := (*where)["_or"]; ok {
	// 	result := []string{}
	// 	for _, item := range items.([]any) {
	// 		expr := item.(map[string]any)
	// 		result = append(result, b.Where(&expr, args, run))
	// 	}

	// 	return "(" + strings.Join(result, " OR ") + ")", nil, nil
	// }

	// Otherwise, construct the WHERE clause based on the field names and operations
	result := []string{}
	for key, item := range *where {
		fmt.Println(key, item)
		for op, value := range item.(map[string]any) {
			fmt.Println("operation", key+op)
			if handler, ok := b.operations[key+op]; ok {
				// Primitive field condition detected
				_value := reflect.ValueOf(value)

				if _value.Kind() == reflect.String {
					qs := handler(b.identifier(key), "?")
					fmt.Println(qs)
					// fmt.Println(qs, _value.String())
					q = q.Where(qs, _value.String())
					// String values are passed to operation handler as single parameter
					// result = append(result, handler(b.identifier(key), b.parameter(_value, nil)))
				} else if _value.Kind() == reflect.Slice || _value.Kind() == reflect.Array {
					// Slice or array values are passed to operation handler as a list of parameters
					items := []string{}
					vals := []any{}
					for i := range _value.Len() {
						items = append(items, "?")
						vals = append(vals, _value.Index(i).String())
					}
					qs := handler(b.identifier(key), items...)

					q = q.Where(qs, vals...)
				}
			} else {
				// Relation field condition detected
				if relation, ok := b.relations[key]; ok {
					builder := registry[relation.table]

					newWhere := item.(map[string]any)
					fmt.Println("newWhere", newWhere)
					var newQuery sq.SelectBuilder
					if relation.through != "" {
						fmt.Println("through", relation.through, "throughField", relation.throughField, "dest", relation.dest, "src", relation.src)
						// "roles on roles.id = user_roles.role_id",
						on := builder.Table() + " on " + builder.Table() + "." + b.identifier(relation.endField) + " = " + b.identifier(relation.through) + "." + b.identifier(relation.throughField)
						fmt.Println(on)
						newQuery = sq.Select(b.identifier(relation.dest)).From(b.identifier(relation.through)).Join(
							builder.Table() + " on " + builder.Table() + "." + b.identifier(relation.endField) + " = " + b.identifier(relation.through) + "." + b.identifier(relation.throughField),
						)
					} else {
						newQuery = sq.Select(b.identifier(relation.dest)).From(builder.Table())
					}
					newQuery = builder.Where(&newWhere, newQuery)
					newqs, newqargs, err := newQuery.ToSql()
					if err != nil {
						slog.Error("Error during query execution", slog.Any("error", err))
						return q
					}
					q = q.Where(b.identifier(relation.src)+" in ("+newqs+")", newqargs...)
				}
			}
		}
	}

	slog.Debug("Constructed WHERE clause", slog.Any("where", result))
	return q
}

// Scans the rows returned by a query into a slice of Model
func (b *SQLBuilder[Model]) Scan(rows pgx.Rows, err error) ([]Model, error) {
	if err != nil {
		slog.Error("Error during query execution", slog.Any("error", err))
		return nil, err
	}
	defer rows.Close()

	// Iterate over the rows and scan each one into a Model instance
	result := []Model{}
	for rows.Next() {
		var model Model
		_value := reflect.ValueOf(&model).Elem()

		// Create a slice of addresses to scan the values into
		_addrs := []any{}
		for _, field := range b.fields {
			_addrs = append(_addrs, _value.Field(field.idx).Addr().Interface())
		}

		// Scan the row into the addresses
		if err := rows.Scan(_addrs...); err != nil {
			return nil, err
		}

		result = append(result, model)
	}

	if err = rows.Err(); err != nil {
		slog.Error("Error during row iteration", slog.Any("error", err))
		return nil, err
	}

	slog.Debug("Scan completed", slog.Any("result", result))
	return result, nil
}
