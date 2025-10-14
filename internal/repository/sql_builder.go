package repository

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
	"github.com/tkahng/playground/internal/tools/utils"
)

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
	// IsNot is the IS NOT operator
	IsNull    = "_isnull"
	IsNotNull = "_isnotnull"
)

var (
	nilOps = []string{
		IsNull, IsNotNull,
	}
	quoteIdentifierList = []string{
		"type", "interval",
	}
)

// OperatorSQLBuilderFunc is a function that returns the appropriate SQL expression for a given operator
type OperatorSQLBuilderFunc func(string, ...string) string

var operatorFuncMap = map[string]OperatorSQLBuilderFunc{
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
	IsNull:    func(key string, values ...string) string { return fmt.Sprintf("%s IS NULL", key) },
	IsNotNull: func(key string, values ...string) string { return fmt.Sprintf("%s IS NOT NULL", key) },
}

// Field represents a column of a table, or
// a selectable scalar field of a model.
// these
type Field struct {
	// Idx is the index of the field
	Idx int
	// Name is the raw name of the field. this might be formatted by the Identifier function
	Name string
	// QuoteIdentifier is true if the table name should be quoted
	QuoteIdentifier bool
}

func (f *Field) Identifier() string {
	if f.QuoteIdentifier {
		return DefaultQuoteIdentifierFunc(f.Name)
	}
	return f.Name
}

// Relation represents a models relational fields
//
// These contain information needed to traverse across tables
// through subqueries and joins.
type Relation struct {
	// table is the name of the related table
	table string

	one bool

	// src is the field of the current table that will be used for the join.
	// usually this is the primary key.
	src string

	// dest is the field of the related table that will be used for the join.
	//
	// usually this is the foreign key of the related table
	// that references the current table by its primary key.
	//
	// but for m2m relationships, this is the related tables primary key,
	// which references the through table by its foreign key.
	dest string

	// through is the name of the join table for m2m relationships.
	// this is only used for the through m2m relationship.
	through string

	// throughDest is the field of the through table that will be used for the join.
	throughDest string

	// throughSrc is the field of the through table that will be used for the join.
	throughSrc string
}

type SQLBuilder[Model any] struct {

	// schemaName is the name of the database schema. default "public"
	schemaName string

	// tableName is the name of the database table
	tableName string

	// quoteIdentifier is true if the table name should be quoted.
	// instead of `select * from schemaName.tableName`, use `select * from schemaName.\"tableName\"`
	quoteIdentifier bool

	// idColumnName is the name of the primary key column
	idColumnName string

	// fields are columns of the table. their order is by the order of the struct fields
	fields []*Field

	columnNames []string

	relations  map[string]*Relation
	operations map[string]func(string, ...string) string
	identifier func(string) string
	parameter  func(reflect.Value, *[]any) string
	generator  func(reflect.StructField, *[]any) (string, error)
	insertID   bool // If true, the id value read from the model will be insert into the database. default false
}

func (b *SQLBuilder[Model]) ReturningFields() string {
	// panic("unimplemented")
	var result []string
	for _, field := range b.fields {
		result = append(result, b.TableName()+"."+field.Identifier())
	}

	return strings.Join(result, ",")
}

func (b *SQLBuilder[Model]) ColumnNames() []string {
	// Returns the column names
	return b.columnNames
}

// Returns the column names with proper identifier formatting
func (b *SQLBuilder[Model]) Fields() []*Field {
	// Returns the fields with their indices and names
	return b.fields
}

func (b *SQLBuilder[Model]) ColumnNamesTablePrefix() []string {
	// Returns the column names with the table prefix
	var prefixedNames []string
	for _, field := range b.fields {
		prefixedNames = append(prefixedNames, b.TableName()+"."+field.Identifier())
	}
	return prefixedNames
}

// Returns the table name with proper identifier formatting
func (b *SQLBuilder[Model]) TableName() string {
	if b.quoteIdentifier {
		return DefaultQuoteIdentifierFunc(b.tableName)
	}
	return b.tableName
}

// Returns a comma-separated list of field names with proper identifier formatting
func (b *SQLBuilder[Model]) FieldString(prefix string) string {
	var result []string
	for _, field := range b.fields {
		result = append(result, prefix+b.Identifier(field.Name))
	}

	return strings.Join(result, ",")
}

// IdColumnName implements SQLBuilderInterface.
func (b *SQLBuilder[Model]) IdColumnName() string {
	return b.idColumnName
}
func (b *SQLBuilder[Model]) InsertID() bool {
	// Returns whether to skip inserting the primary key field
	return b.insertID
}
func (b *SQLBuilder[Model]) Generator() func(reflect.StructField, *[]any) (string, error) {
	// Returns the generator function for the primary key field
	return b.generator
}

func (b *SQLBuilder[Model]) Identifier(name string) string {
	if b.quoteIdentifier {
		return fmt.Sprintf("\"%s\"", name)
	}
	return name
}

func (b *SQLBuilder[Model]) Parameter(value reflect.Value, args *[]any) string {
	*args = append(*args, value.Interface())
	return fmt.Sprintf("$%d", len(*args))
}

type SQLBuilderInterface interface {
	Identifier(name string) string
	TableName() string
	ColumnNames() []string
	ColumnNamesTablePrefix() []string
	Fields() []*Field
	FieldString(prefix string) string
	Where(where *map[string]any, args *[]any, run func(string) []string) string
	WhereError(ctx context.Context, where *map[string]any, args *[]any, run func(string) []string) (ret string, err error)
	IdColumnName() string
	InsertID() bool
	Generator() func(reflect.StructField, *[]any) (string, error)

	Sort(filter Sortable) *map[string]string
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
	builder.insertID = true
	return nil
}

func InsertID[Model any](builder *SQLBuilder[Model]) error {
	if builder == nil {
		return errors.New("SQLBuilder cannot be nil")
	}
	builder.insertID = true
	return nil
}

func DefaultParameterFunc(value reflect.Value, args *[]any) string {
	*args = append(*args, value.Interface())
	return fmt.Sprintf("$%d", len(*args))
}

func DefaultQuoteIdentifierFunc(name string) string {
	if ok := slices.Contains(quoteIdentifierList, name); ok {
		return fmt.Sprintf("\"%s\"", name)
	}
	return name
}

func NewSQLBuilder[Model any](opts ...SQLBuilderOptions[Model]) *SQLBuilder[Model] {

	// Reflect on the Model type to extract metadata
	var _type reflect.Type = reflect.TypeFor[Model]()

	// default table name to lowercase model name
	var tableName string = strings.ToLower(_type.Name())

	var modelFields []*Field
	var modelColumnNames []string
	var modelRelations map[string]*Relation = map[string]*Relation{}
	var modelOperations map[string]func(string, ...string) string = map[string]func(string, ...string) string{}

	// iterate over the fields of the model type
	for idx := range _type.NumField() {
		_field := _type.Field(idx)

		// "_" named field is the model information
		if _field.Name == "_" {
			if dbTagValue := _field.Tag.Get("db"); dbTagValue != "" {
				fieldTag := ParseFieldTag(dbTagValue)
				if fieldTag == nil {
					panic("failed to parse db tag value")
				}

				tableName = fieldTag.Value
			}
		} else {
			// Other fields are model attributes
			if dbTagValue := _field.Tag.Get("db"); dbTagValue != "" {
				// split db tag value
				var fieldName string
				var fieldOptions []string
				var quoteIdentifier bool
				for idx, value := range strings.Split(dbTagValue, ",") {
					if idx == 0 {
						fieldName = value
					} else {
						fieldOptions = append(fieldOptions, value)
					}
				}
				if slices.Contains(fieldOptions, "quote") {
					quoteIdentifier = true
				}
				if fieldName == "" {
					panic(fmt.Sprintf("fieldName not set at struct field idx %d of %s", idx, tableName))
				}

				// if table tag is set, it's a relation
				if table := _field.Tag.Get("table"); table != "" {
					// Relation field detected
					var relation Relation
					relation.table = table

					if src := _field.Tag.Get("src"); src != "" {
						relation.src = src
					} else {
						panic(fmt.Sprintf("src not set at struct field name %s of %s", fieldName, tableName))
					}
					if dest := _field.Tag.Get("dest"); dest != "" {
						relation.dest = dest
					} else {
						panic(fmt.Sprintf("dest not set at struct field name %s of %s", fieldName, tableName))
					}
					if through := _field.Tag.Get("through"); through != "" {
						relation.through = through
						if throughDest := _field.Tag.Get("through_dest"); throughDest != "" {
							relation.throughDest = throughDest
						} else {
							panic(fmt.Sprintf("through_dest not set at struct field name %s of %s", fieldName, tableName))
						}
						if throughSrc := _field.Tag.Get("through_src"); throughSrc != "" {
							relation.throughSrc = throughSrc
						} else {
							panic(fmt.Sprintf("through_src not set at struct field name %s of %s", fieldName, tableName))
						}
					}

					modelRelations[fieldName] = &relation

				} else {
					// Primitive fields detected.
					// This are selectable columns of the table

					field := &Field{Idx: idx, Name: fieldName, QuoteIdentifier: quoteIdentifier}
					// get the fieldName of the field
					modelFields = append(modelFields, field)

					modelColumnNames = append(modelColumnNames, fieldName)

					// Add base operations for the field
					for key, value := range operatorFuncMap {
						modelOperations[fieldName+key] = value
					}

				}
			}
		}
	}
	result := &SQLBuilder[Model]{
		tableName:       tableName,
		columnNames:     modelColumnNames,
		fields:          modelFields,
		relations:       modelRelations,
		operations:      modelOperations,
		identifier:      DefaultQuoteIdentifierFunc,
		parameter:       DefaultParameterFunc,
		generator:       nil,
		insertID:        false,
		quoteIdentifier: false,
	}
	for _, opt := range opts {
		if err := opt(result); err != nil {
			slog.Error("Error applying SQLBuilder option", slog.Any("error", err))
			panic(fmt.Sprintf("Error applying SQLBuilder option: %v", err))
		}
	}

	registry[tableName] = result

	return result
}

var timestampNames = []string{"created_at", "updated_at"}

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
		slog.Info("Processing OR condition", slog.Any("ors", ors))
		result := []string{}

		orWheres, ok := ors.([]map[string]any)
		if ok {
			for _, item := range orWheres {
				expr := item
				slog.Info("Processing OR item", slog.Any("item", item))
				result = append(result, b.Where(&expr, args, run))
			}
		}

		return "(" + strings.Join(result, " OR ") + ")"
	}

	// Otherwise, construct the WHERE clause based on the field names and operations
	result := []string{}

	// iterate over the where map,
	// each key is a field or column name
	// each value will be a map of operators and values
	for whereField, whereFieldOperation := range *where {
		// fmt.Println("key", key, "item", item)

		// iterate over the map of operators and values,
		// each key is a operator code(_eq, _gt, etc)
		// each value will be a map of operators and values
		for whereOp, whereOpValue := range whereFieldOperation.(map[string]any) {
			// fmt.Println("operation", op, "value", value)

			// if this f
			if opFunc, ok := b.operations[whereField+whereOp]; ok {
				// Primitive field condition detected
				// slog.Info("Processing primitive field condition", slog.String("key", key), slog.String("operation", op), slog.Any("value", value))
				if whereOpValue == nil {
					// slog.Warn("Nil value detected for key", slog.String("key", key), slog.String("operation", op))
					if slices.Contains(nilOps, whereOp) {
						// slog.Info("Nil operation detected, adding to result", slog.String("key", key))
						// If the value is nil and the operation is a nil operation, send it
						result = append(result, opFunc(b.Identifier(whereField)))
					}
					continue // Skip nil values for non-nil operations
				}

				_value := reflect.ValueOf(whereOpValue)
				if !_value.IsValid() {
					slog.Info("value is invalid")
					continue
				}
				if _value.Kind() == reflect.Pointer && !_value.IsNil() {
					// If the value is a pointer, dereference it
					_value = _value.Elem()
				}
				if _value.Kind() == reflect.String {
					// String values are passed to operation handler as single parameter
					result = append(result, opFunc(b.Identifier(whereField), b.parameter(_value, args)))
				} else if it, ok := whereOpValue.(time.Time); ok {
					_newValue := reflect.ValueOf(it.Format(time.RFC3339Nano))
					result = append(result, opFunc(b.Identifier(whereField), b.parameter(_newValue, args)))
				} else if it, ok := whereOpValue.(fmt.Stringer); ok {
					_newValue := reflect.ValueOf(it.String())
					if _newValue.Kind() == reflect.String {
						// If the value implements fmt.Stringer, use its String method
						result = append(result, opFunc(b.Identifier(whereField), b.parameter(_newValue, args)))
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
					result = append(result, opFunc(b.Identifier(whereField), items...))
				}

			} else {
				// Relation field condition detected
				if relation, ok := b.relations[whereField]; ok {
					var relatedBuilder SQLBuilderInterface
					// Get the target SQLBuilder for the relation
					if bld, ok := registry[relation.table]; !ok {
						continue
					} else {
						// Get the target SQLBuilder for the relation
						relatedBuilder = bld
					}

					// Construct the sub-query for the related table
					where := whereFieldOperation.(map[string]any)

					// query is the subquery we need to generate.
					var query string

					var dest string = b.Identifier(relation.dest)
					var related string = relatedBuilder.TableName()
					// if through is not empty
					// it is a many-to-many relation
					if relation.through != "" {
						//goland:noinspection Annotator

						var through = b.Identifier(relation.through)
						var endField = b.Identifier(relation.throughSrc)
						throughField := b.Identifier(relation.throughDest)

						query = fmt.Sprintf(
							// SELECT dest FROM through join related on related.endField = through.throughField`
							// SELECT dest FROM through join related on related.endField = through.throughField`
							`SELECT %s FROM %s join %s on %s.%s = %s.%s`,
							// SELECT
							dest,
							// FROM
							through,
							related,
							related,
							endField,
							through,
							throughField,
						)
					} else {
						//goland:noinspection Annotator
						query = fmt.Sprintf("SELECT %s FROM %s", b.Identifier(relation.dest), relatedBuilder.TableName())
					}
					if expr := relatedBuilder.Where(&where, args, run); expr != "" {
						query += fmt.Sprintf(" WHERE %s", expr)
					}
					if run == nil {
						if inop, ok := b.operations[relation.src+"_in"]; ok {
							result = append(result, inop(b.Identifier(relation.src), query))
						}
						// If no run function is provided, sub-query is added to the main query
					} else {
						if inop, ok := b.operations[relation.src+"_in"]; ok {
							result = append(result, inop(b.Identifier(relation.src), run(query)...))
						}
						// If a run function is provided, sub-query is executed and its result is added to the main query
					}
				}
			}
		}
	}

	return strings.Join(result, " AND ")
}

func (b *SQLBuilder[Model]) Sort(filter Sortable) *map[string]string {
	if filter == nil {
		return nil
	}
	sortBy, sortOrder := filter.Sort()
	if sortBy != "" && slices.Contains(b.ColumnNames(), utils.Quote(sortBy)) {
		return &map[string]string{
			sortBy: sortOrder,
		}
	} else {
		slog.Info("sort by field not found in repository columns", "sortBy", sortBy, "sortOrder", sortOrder, "columns", b.ColumnNames())
		return nil // Return nil if the sortBy field is not found in the repository columns
	}
}

func (b *SQLBuilder[Model]) ValuesError(values *[]Model, args *[]any, keys *[]any) (fields string, vals string, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Error occurred during values generation", slog.Any("error", r),
				slog.String("table", b.tableName),
				slog.Any("values", values),
				slog.Any("args", args),
				slog.Any("keys", keys),
			)
			err = fmt.Errorf("error generating values for table %s", b.tableName)
		}
	}()
	fields, vals, err = b.Values(values, args, keys)
	return
}

// Constructs the VALUES clause for an INSERT query
func (b *SQLBuilder[Model]) Values(values *[]Model, args *[]any, keys *[]any) (fields string, vals string, err error) {
	if values == nil {
		err = fmt.Errorf("values cannot be nil")
		return
	}

	// Generate the field names for the VALUES clause
	var fieldsArray []string
	for idx, field := range b.fields {

		// first item is the primary key.
		if idx == 0 {
			// primary keys are often database generated,
			// in which case they are not included in the INSERT INTO clause
			// If insertID is true, we will provide the primary key,
			// so include it in the INSERT INTO clause
			if b.insertID {
				fieldsArray = append(fieldsArray, field.Identifier())
			}
		} else {
			// Skip timestamp fields
			if slices.Contains(timestampNames, field.Name) {
				continue
			}
			// Other fields are added to the VALUES clause
			fieldsArray = append(fieldsArray, field.Identifier())
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
			// first item is the primary key.
			if idx == 0 {
				// primary keys are often database generated,
				// in which case they are not included in the VALUES clause
				//
				// If insertID is true, we will provide the primary key,
				// so include it in the VALUES clause
				if b.insertID {
					//
					if b.generator != nil {
						// If a generator function is provided, use it to generate the key
						id, err := b.generator(_type.Field(field.Idx), keys)
						if err != nil {
							return "", "", fmt.Errorf("error generating primary key for field %s: %w", field.Name, err)
						}
						items = append(items, b.parameter(reflect.ValueOf(id), args))
					} else {
						items = append(items, b.parameter(_value.Field(field.Idx), args))
					}
				}
			} else {
				if slices.Contains(timestampNames, field.Name) {
					continue // Skip timestamp fields
				}
				// Other fields are added to the VALUES clause
				items = append(items, b.parameter(_value.Field(field.Idx), args))
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
				slog.String("table", b.tableName),
				slog.Any("set", set),
				slog.Any("where", where),
			)
			err = fmt.Errorf("error generating Set for table %s", b.tableName)
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
				_field := _value.Field(field.Idx)
				for _field.Kind() == reflect.Pointer {
					_field = _field.Elem()
				}

				// Set the WHERE clause condition based on the field type
				switch _field.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					(*where)[field.Name] = map[string]any{"_eq": fmt.Sprintf("%d", _field.Int())}
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					(*where)[field.Name] = map[string]any{"_eq": fmt.Sprintf("%d", _field.Uint())}
				case reflect.Float32, reflect.Float64:
					(*where)[field.Name] = map[string]any{"_eq": fmt.Sprintf("%f", _field.Float())}
				case reflect.Complex64, reflect.Complex128:
					(*where)[field.Name] = map[string]any{"_eq": fmt.Sprintf("%f", _field.Complex())}
				case reflect.String:
					(*where)[field.Name] = map[string]any{"_eq": _field.String()}
				default:
					if u, ok := _field.Interface().(uuid.UUID); ok {
						(*where)[field.Name] = map[string]any{"_eq": u.String()}
					} else if it, ok := _field.Interface().(time.Time); ok {
						_newValue := reflect.ValueOf(it.Format(time.RFC3339Nano))
						(*where)[field.Name] = map[string]any{"_eq": _newValue.String()}
					} else if it, ok := _field.Interface().(fmt.Stringer); ok {
						_newValue := reflect.ValueOf(it.String())
						if _newValue.Kind() == reflect.String {
							// If the value implements fmt.Stringer, use its String method
							(*where)[field.Name] = map[string]any{"_eq": _newValue.String()}
						}
					} else {
						panic("Invalid identifier type")
					}
				}
			}
		} else {
			// Other fields are added to the SET clause
			result = append(result, b.Identifier(field.Name)+"="+b.parameter(_value.Field(field.Idx), args))
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
			result = append(result, fmt.Sprintf("%s %s", b.Identifier(key), strings.ToUpper(val)))
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
				slog.String("table", b.tableName),
				slog.Any("where", where),
			)
			err = fmt.Errorf("error generating where for table %s. check your filters", b.tableName)
		}
	}()
	ret = b.Where(where, args, run)
	return
}
