package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"runtime/debug"
	"slices"
	"sort"
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
)

// OperatorSQLBuilderFunc is a function that returns the appropriate SQL expression for a given operator.
//
// for the operator _gte, the function will return "column >= value"
//
// for the operator _in, the function will return "column IN (value1, value2, value3)"
type OperatorSQLBuilderFunc func(col string, values ...string) string

// operatorFuncMap is a map operator names to their corresponding OperatorSQLBuilderFunc
var operatorFuncMap = map[string]OperatorSQLBuilderFunc{
	Eq:     func(col string, values ...string) string { return fmt.Sprintf("%s = %s", col, values[0]) },
	Neq:    func(col string, values ...string) string { return fmt.Sprintf("%s != %s", col, values[0]) },
	Gt:     func(col string, values ...string) string { return fmt.Sprintf("%s > %s", col, values[0]) },
	Gte:    func(col string, values ...string) string { return fmt.Sprintf("%s >= %s", col, values[0]) },
	Lt:     func(col string, values ...string) string { return fmt.Sprintf("%s < %s", col, values[0]) },
	Lte:    func(col string, values ...string) string { return fmt.Sprintf("%s <= %s", col, values[0]) },
	Like:   func(col string, values ...string) string { return fmt.Sprintf("%s LIKE %s", col, values[0]) },
	Nlike:  func(col string, values ...string) string { return fmt.Sprintf("%s NOT LIKE %s", col, values[0]) },
	Ilike:  func(col string, values ...string) string { return fmt.Sprintf("%s ILIKE %s", col, values[0]) },
	Nilike: func(col string, values ...string) string { return fmt.Sprintf("%s NOT ILIKE %s", col, values[0]) },
	In: func(col string, values ...string) string {
		return fmt.Sprintf("%s IN (%s)", col, strings.Join(values, ","))
	},
	Nin: func(col string, values ...string) string {
		return fmt.Sprintf("%s NOT IN (%s)", col, strings.Join(values, ","))
	},
	IsNull:    func(col string, values ...string) string { return fmt.Sprintf("%s IS NULL", col) },
	IsNotNull: func(col string, values ...string) string { return fmt.Sprintf("%s IS NOT NULL", col) },
}

// Field represents a column of a table, or
// a selectable scalar field of a model.
type Field struct {
	// Idx is the index of the field in its struct. used for reflection.
	Idx int

	// Name is the name of the field of the model.
	//
	// eg: id, type, email
	Name string

	// ColumnName is formatted name of the column of the table
	//
	// eg: id, "type", email
	ColumnName string

	// QualifiedColumnName is formatted name of the column of the table
	//
	// eg: schema.table.id, schema.table."type", schema.table.email
	QualifiedColumnName string

	// IsID is true if the field is the primary key
	IsID bool

	// QuoteIdentifier is true if the table name should be quoted
	QuoteIdentifier bool
}

// Relation represents a models relational fields
//
// These contain information needed to traverse across tables
// through subqueries and joins.
type Relation struct {
	// Idx is the index of the field in its struct. used for reflection.
	Idx int

	fieldName string
	// table is the name of the related table
	table string

	// src is the field of the current table that will be used for the join.
	// this is usually the primary key of the current table.
	// this is the field that will perform the IN filter result of the subquery
	src string

	// dest is the field of the related table that will be used for the join.
	//
	// usually this is the foreign key of the related table
	// that references the current table by its primary key.
	// during filters this field is selected from the related table in the subquery
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

type SQLBuilderInterface interface {
	TableName() string
	FieldNames() []string
	ColumnNames() []string
	QualifiedColumnNames() []string
	Fields() []*Field
	GetFieldByName(name string) *Field
	MustGetFieldByName(name string) *Field
	Where(where *map[string]any, args *[]any) (string, error)
	WhereError(ctx context.Context, where *map[string]any, args *[]any) (string, error)
	IdFieldName() string
	InsertID() bool
	Generator() func(reflect.StructField, *[]any) (string, error)

	Sort(filter Sortable) *map[string]string
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

	relations  map[string]*Relation
	operations map[string]func(string, ...string) string
	generator  func(reflect.StructField, *[]any) (string, error)

	insertID bool // If true, the id value read from the model will be insert into the database. default false
}

// GetRelationByName returns the relation with the given name.
// returns nil if the relation is not found
func (b *SQLBuilder[Model]) GetRelationByName(name string) *Relation {
	for _, field := range b.relations {
		if field.fieldName == name {
			return field
		}
	}
	return nil
}

// GetFieldByName returns the field with the given name.
// returns nil if the field is not found
func (b *SQLBuilder[Model]) GetFieldByName(name string) *Field {
	for _, field := range b.fields {
		if field.Name == name {
			return field
		}
	}
	return nil
}

// MustGetFieldByName returns the field with the given field name.
// name is the field name of the model, without its schema, table, or quotes.
// panics if the field is not found
func (b *SQLBuilder[Model]) MustGetFieldByName(name string) *Field {
	var field *Field = b.GetFieldByName(name)
	if field == nil {
		panic(fmt.Sprintf("could not get field by name: field %s not found for model %s", name, b.tableName))
	}
	return field
}

// Fields returns the fields
func (b *SQLBuilder[Model]) Fields() []*Field {
	return b.fields
}

// FieldNames returns the names of the fields, not the qualified column names.
func (b *SQLBuilder[Model]) FieldNames() []string {
	fieldNames := []string{}
	for _, field := range b.fields {
		fieldNames = append(fieldNames, field.Name)
	}
	return fieldNames
}

// ColumnNames returns the names of the columns, not the qualified column names.
func (b *SQLBuilder[Model]) ColumnNames() []string {
	fieldNames := []string{}
	for _, field := range b.fields {
		fieldNames = append(fieldNames, field.ColumnName)
	}
	return fieldNames
}

// ColumnNames returns the names of the columns, not the qualified column names.
func (b *SQLBuilder[Model]) ColumnNamesJoined() string {
	return strings.Join(b.ColumnNames(), ",")
}

// QualifiedColumnNames returns the formatted column names with the table prefix.
func (b *SQLBuilder[Model]) QualifiedColumnNames() []string {
	// Returns the column names with the table prefix
	fieldNames := []string{}
	for _, field := range b.fields {
		fieldNames = append(fieldNames, field.QualifiedColumnName)
	}
	return fieldNames
}

// QualifiedColumnNames returns the formatted column names with the table prefix.
func (b *SQLBuilder[Model]) QualifiedColumnNamesJoined() string {
	return strings.Join(b.QualifiedColumnNames(), ",")
}

// TableName Returns the table name with proper identifier formatting
func (b *SQLBuilder[Model]) TableName() string {
	if b.quoteIdentifier {
		return b.schemaName + "." + utils.Quote(b.tableName)
	}
	return b.schemaName + "." + b.tableName
}

// IdFieldName implements SQLBuilderInterface.
func (b *SQLBuilder[Model]) IdFieldName() string {
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

// registry holds all of the SQLBuilderInterface created during buildtime.
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

func NewSQLBuilder[Model any](opts ...SQLBuilderOptions[Model]) *SQLBuilder[Model] {
	// Reflect on the Model type to extract metadata
	var _type = reflect.TypeFor[Model]()

	// default table name to lowercase model name
	var tableName = strings.ToLower(_type.Name())

	// default quote identifier to true
	var quoteIdentifier = false

	var modelFields []*Field
	var modelRelations = map[string]*Relation{}
	var modelOperations = map[string]func(string, ...string) string{}

	var schemaName = "public"
	// iterate over the fields of the model type
	for idx := range _type.NumField() {
		_field := _type.Field(idx)

		// first field is the info field _ struct{}
		if idx == 0 {
			if _field.Name == "_" {
				// tag should be db with table name
				// or else panic
				if dbTagValue := _field.Tag.Get("db"); dbTagValue != "" {
					fieldTag := ParseFieldTag(dbTagValue)
					if fieldTag == nil {
						panic("failed to parse info db tag value")
					}
					tableName = fieldTag.Value

					if quote := fieldTag.GetOptionValue("quote"); quote == "true" {
						quoteIdentifier = true
					}
					if schemaTagValue := _field.Tag.Get("schema"); schemaTagValue != "" {
						schemaName = schemaTagValue
					}
				} else {
					panic("db info value not set")
				}
			} else {
				panic("first field must be info field")
			}
		} else {
			// Other fields are model attributes
			if dbTagValue := _field.Tag.Get("db"); dbTagValue != "" {
				// split db tag value
				var fieldName string
				var columnName string
				var isId bool

				fieldTag := ParseFieldTag(dbTagValue)
				if fieldTag == nil {
					panic("failed to parse info db tag value")
				}
				fieldName = fieldTag.Value
				if quote := fieldTag.GetOptionValue("quote"); quote == "true" {
					columnName = utils.Quote(fieldName)
				} else {
					columnName = fieldName
				}
				if idIdTag := fieldTag.GetOptionValue("pk"); idIdTag == "true" {
					isId = true
				}

				if fieldName == "" {
					panic(fmt.Sprintf("fieldName not set at struct field idx %d of %s", idx, tableName))
				}

				// if table tag is set, it's a relation
				if table := _field.Tag.Get("table"); table != "" {
					// Relation field detected
					var relation Relation
					relation.table = table
					relation.fieldName = fieldName
					relation.Idx = idx

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

					modelRelations[relation.fieldName] = &relation
				} else {
					// Primitive fields detected.
					// This are selectable columns of the table

					field := &Field{
						Idx:                 idx,
						Name:                fieldName,
						ColumnName:          columnName,
						QualifiedColumnName: schemaName + "." + tableName + "." + columnName,
						IsID:                isId,
					}
					// get the fieldName of the field
					modelFields = append(modelFields, field)

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
		fields:          modelFields,
		relations:       modelRelations,
		operations:      modelOperations,
		generator:       nil,
		insertID:        false,
		quoteIdentifier: quoteIdentifier,
		schemaName:      schemaName,
	}
	for _, opt := range opts {
		if err := opt(result); err != nil {
			slog.Error("Error applying SQLBuilder option", slog.Any("error", err))
			panic(fmt.Sprintf("Error applying SQLBuilder option: %v", err))
		}
	}
	registry[schemaName+"."+tableName] = result

	return result
}

var timestampNames = []string{"created_at", "updated_at"}

func (b *SQLBuilder[Model]) WhereError(ctx context.Context, where *map[string]any, args *[]any) (ret string, err error) {
	if where == nil {
		return "", nil
	}
	defer func() {
		if r := recover(); r != nil {
			slog.ErrorContext(ctx, "Error occurred during where generation", slog.Any("error", r),
				slog.String("table", b.tableName),
				slog.Any("where", where),
				slog.Any("stacktrace", string(debug.Stack())),
			)
			err = fmt.Errorf("error generating where for table %s. check your filters", b.tableName)
		}
	}()
	return b.Where(where, args)
}

// GenerateParameterPlaceholder returns the numbered parameter placeholder,
// e.g. $1, $2, etc,
// their numbers are incremented for each call as it adds them to the args slice.
// the total number of placeholders created is equal to the length of the args slice.
//
// value is reflect.Value,
// args is a slice of any.
//
// every time build the parameter placeholder, you call this function
// with the reflect.Value of the value to insert, and a pointer to a slice of args.
// for each call, the function adds the underlying value of value and adds it to the args slice.
// then the numer in the placeholder is generated from the length of the args slice,
// which is also the current input value's placeholder number, and its index in the args slice.–
func GenerateParameterPlaceholder(value reflect.Value, args *[]any) string {
	*args = append(*args, value.Interface())
	return fmt.Sprintf("$%d", len(*args))
}

// Where constructs the WHERE clause for a query.
//
// where is a map[string]any:
//
//	var where = map[string]any{
//		"name": map[string]any{
//			"_eq": "John",
//		},
//		"roles": map[string]any{
//			"name": map[string]any{
//				"_in": []string{"superuser", "user"},
//			},
//		},
//	}
//
// args is a slice of any:
//
//	var args = []any{"John", []string{"superuser", "user"}}
type fieldIdx struct {
	Name string
	Idx  int
}

func (b *SQLBuilder[Model]) getSortedFields(where *map[string]any) []*fieldIdx {
	fields := []*fieldIdx{}
	for key := range *where {
		if field := b.GetFieldByName(key); field != nil {
			fields = append(fields, &fieldIdx{
				Name: key,
				Idx:  field.Idx,
			})
		} else if relation := b.GetRelationByName(key); relation != nil {
			fields = append(fields, &fieldIdx{
				Name: key,
				Idx:  relation.Idx,
			})
		} else {
			slog.Error("Unknown field or relation", slog.String("field", key))
			panic(fmt.Sprintf("Unknown field or relation: %s", key))
		}
	}
	if len(fields) > 1 {
		sort.Slice(fields, func(i, j int) bool {
			return fields[i].Idx < fields[j].Idx
		})
	}
	return fields
}

func (b *SQLBuilder[Model]) Where(where *map[string]any, args *[]any) (ret string, retErr error) {
	if where == nil {
		return ret, retErr
	}

	// Check for special conditions
	// _not, _and, and _or are used for logical operations
	if item, ok := (*where)["_not"]; ok {
		// in case of _not, the value is a map[string]any
		if expr, ok := item.(map[string]any); ok {
			notExpr, err := b.Where(&expr, args)
			if err != nil {
				return "", err
			}
			ret = "NOT (" + notExpr + ")"
			return ret, retErr
		}
	} else if items, ok := (*where)["_and"]; ok {
		// in case of _and, the value is a []map[string]any
		result := []string{}
		if ands, ok := items.([]map[string]any); ok {
			for _, item := range ands {
				expr := item
				andExpr, err := b.Where(&expr, args)
				if err != nil {
					return "", err
				}
				result = append(result, andExpr)
			}
		}

		ret = "(" + strings.Join(result, " AND ") + ")"
		return ret, retErr
	} else if ors, ok := (*where)["_or"]; ok {
		// in case of _and, the value is a []map[string]any
		result := []string{}
		orWheres, ok := ors.([]map[string]any)
		if ok {
			for _, item := range orWheres {
				expr := item
				orExpr, err := b.Where(&expr, args)
				if err != nil {
					return "", err
				}
				result = append(result, orExpr)
			}
		}

		ret = "(" + strings.Join(result, " OR ") + ")"
		return ret, retErr
	}

	// Otherwise, construct the WHERE clause based on the field names and operations
	result := []string{}

	fieldIdxs := b.getSortedFields(where)
	for _, fieldIdx := range fieldIdxs {
		var whereFieldName = fieldIdx.Name
		if whereFieldValue, ok := (*where)[whereFieldName]; ok {
			expr, ok := whereFieldValue.(map[string]any)
			if ok {
				for whereOp, whereOpValue := range expr {
					// if this field and operation is registered, go ahead.
					// if not, it might be a relational field
					if opFunc, ok := b.operations[whereFieldName+whereOp]; ok {
						var whereField = b.GetFieldByName(whereFieldName)
						if whereField == nil {
							return "", fmt.Errorf("could not get field by name: field %s not found for model %s", whereFieldName, b.tableName)
						}
						// Primitive field condition detected
						//
						// if the value is nil, it should use the _isNil, _isNotNil operations
						if whereOpValue == nil {
							if slices.Contains(nilOps, whereOp) {
								// If the value is nil and the operation is a nil operation, send it
								result = append(result, opFunc(whereField.ColumnName))
							} else {
								// If the value is nil and the operation is not a nil operation, ignore it
								return "", fmt.Errorf("nil value for non-nil operation %s for model %s", whereOp, b.tableName)
							}
							continue
						}

						_value := reflect.ValueOf(whereOpValue)

						// If the value is a pointer, dereference it
						if _value.Kind() == reflect.Pointer && !_value.IsNil() {
							_value = _value.Elem()
						}

						if !_value.IsValid() {
							return "", fmt.Errorf("invalid value for field %s and operation %s for model %s", whereFieldName, whereOp, b.TableName())
						}
						// String values are passed
						// time values are formatted as strings
						// fmt.Stringer values are called cast then called String method
						if (_value.Kind() == reflect.Slice || _value.Kind() == reflect.Array) && _value.Type().Elem().Kind() != reflect.Uint8 {
							// Slice or array values are iterated and check for string types.
							items := []string{}
							for i := range _value.Len() {
								_valueItem := _value.Index(i)
								item := convert(_valueItem)
								items = append(items, GenerateParameterPlaceholder(reflect.ValueOf(item), args))
							}
							result = append(result, opFunc(whereField.QualifiedColumnName, items...))
						} else {
							item := convert(_value)
							result = append(result, opFunc(whereField.QualifiedColumnName, GenerateParameterPlaceholder(reflect.ValueOf(item), args)))
						}
					} else {
						// this field name and opation is not registered.
						// Relation field condition detected
						if relation := b.GetRelationByName(whereFieldName); relation != nil {
							var relatedBuilder SQLBuilderInterface
							// Get the target SQLBuilder for the relation
							if bld, ok := registry[relation.table]; !ok {
								return "", fmt.Errorf("relation %s not found for model %s", relation.table, b.tableName)
							} else {
								// Get the target SQLBuilder for the relation
								relatedBuilder = bld
							}

							// Construct the sub-query for the related table
							relationWhere := expr

							// query is the subquery we need to generate.
							var query string

							var relatedDestField = relatedBuilder.GetFieldByName(relation.dest)
							if relatedDestField == nil {
								return "", fmt.Errorf("could not get field by name: dest field %s not found for model %s for relation to %s", relation.dest, relatedBuilder.TableName(), b.tableName)
							}
							var srcField = b.GetFieldByName(relation.src)
							if srcField == nil {
								panic(fmt.Sprintf("could not get field by name: src field %s not found for model %s", relation.src, b.tableName))
							}
							// if through is not empty
							// it is a many-to-many relation
							if relation.through != "" {
								var throughBuilder SQLBuilderInterface
								// Get the target SQLBuilder for the relation
								if bld, ok := registry[relation.through]; !ok {
									return "", fmt.Errorf("relation through %s not found for model %s", relation.through, b.tableName)
								} else {
									// Get the target SQLBuilder for the relation
									throughBuilder = bld
								}

								var throughTableName = throughBuilder.TableName()
								var throughSrcField = throughBuilder.GetFieldByName(relation.throughSrc)
								if throughSrcField == nil {
									return "", fmt.Errorf("could not get field by name: through src field %s not found for model %s for relation to %s", relation.dest, throughBuilder.TableName(), b.tableName)
								}
								var throughDestField = throughBuilder.GetFieldByName(relation.throughDest)
								if throughDestField == nil {
									return "", fmt.Errorf("could not get field by name: dest field %s not found for model %s for relation to %s", relation.dest, relatedBuilder.TableName(), relatedBuilder.TableName())
								}
								query = fmt.Sprintf(
									`SELECT %s FROM %s join %s on %s.%s = %s.%s`,
									// SELECT
									throughSrcField.QualifiedColumnName,
									// FROM
									throughTableName,
									// join
									relatedBuilder.TableName(),
									// on
									relatedBuilder.TableName(),
									relatedDestField.ColumnName,
									// =
									throughTableName,
									throughDestField.ColumnName,
								)
							} else {
								query = fmt.Sprintf("SELECT %s FROM %s", relatedDestField.QualifiedColumnName, relatedBuilder.TableName())
							}
							if expr, err := relatedBuilder.Where(&relationWhere, args); err != nil {
								return "", err
							} else if expr != "" {
								query += fmt.Sprintf(" WHERE %s", expr)
							}
							if inop, ok := b.operations[srcField.Name+"_in"]; ok {
								result = append(result, inop(srcField.QualifiedColumnName, query))
							}
						}
					}
				}
			}
		}
	}

	ret = strings.Join(result, " AND ")
	return ret, retErr
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
		return fields, vals, err
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
			if b.InsertID() {
				// use field.ColumnName for insert. qualifying the column name
				// with the table name will result in error.
				fieldsArray = append(fieldsArray, field.ColumnName)
			}
		} else {
			// Skip timestamp fields
			if slices.Contains(timestampNames, field.Name) {
				continue
			}
			// Other fields are added to the VALUES clause
			fieldsArray = append(fieldsArray, field.ColumnName)
		}
	}

	// Generate the field values for the VALUES clause
	valuesArray := []string{}
	for _, model := range *values {
		_type := reflect.TypeOf(model)
		_value := reflect.ValueOf(model)

		// Generate the values for the current model
		values := []string{}
		for idx, field := range b.fields {
			// first item is the primary key.
			if idx == 0 {
				// primary keys are often database generated,
				// in which case they are not included in the VALUES clause
				//
				// If insertID is true, we will provide the primary key,
				// so include it in the VALUES clause
				if b.InsertID() {
					//
					if b.generator != nil {
						// If a generator function is provided, use it to generate the key
						id, err := b.generator(_type.Field(field.Idx), keys)
						if err != nil {
							return "", "", fmt.Errorf("error generating primary key for field %s: %w", field.Name, err)
						}
						values = append(values, GenerateParameterPlaceholder(reflect.ValueOf(id), args))
					} else {
						values = append(values, GenerateParameterPlaceholder(_value.Field(field.Idx), args))
					}
				}
			} else {
				if slices.Contains(timestampNames, field.Name) {
					continue // Skip timestamp fields
				}
				// Other fields are added to the VALUES clause
				values = append(values, GenerateParameterPlaceholder(_value.Field(field.Idx), args))
			}
		}

		valuesArray = append(valuesArray, "("+strings.Join(values, ",")+")")
	}

	fields = strings.Join(fieldsArray, ",")
	vals = strings.Join(valuesArray, ",")
	return fields, vals, nil
}

func (b *SQLBuilder[Model]) SetError(set *Model, args *[]any, where *map[string]any) (ret string, returnErr error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Error occurred during Set generation", slog.Any("error", r),
				slog.String("table", b.tableName),
				slog.Any("set", set),
				slog.Any("where", where),
			)
			returnErr = fmt.Errorf("error generating Set for table %s", b.tableName)
		}
	}()
	ret = b.Set(set, args, where)
	return
}

// convert converts a reflect.Value to a string.
//
// most sql drivers can handle most types, but some cases are ambiguous, or need more refinement.
//
// time.Time cannot be used as its fmt.Stringer, but instead needs to be RFC3339Nano formatted.
// so we want to explicitly convert them
//
// uuid.UUID underlying type is [16]byte, making it ambiguous from just []byte, which is a common column type.
// in this case we first check for strict []byte, them uuid.UUID, then either string([]byte) or (uuid.UUID).String().
func convert(_field reflect.Value) string {
	if _field.Kind() == reflect.Pointer {
		_field = _field.Elem()
	}
	switch _field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", _field.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", _field.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", _field.Float())
	case reflect.Complex64, reflect.Complex128:
		return fmt.Sprintf("%f", _field.Complex())
	case reflect.String:
		return _field.String()
	case reflect.Bool:
		return fmt.Sprintf("%t", _field.Bool())
	default:
		var val string
		if it, ok := _field.Interface().([]byte); ok {
			val = string(it)
		} else if it, ok := _field.Interface().(uuid.UUID); ok {
			val = it.String()
		} else if it, ok := _field.Interface().(time.Time); ok {
			_newValue := reflect.ValueOf(it.Format(time.RFC3339Nano))
			val = _newValue.String()
		} else if it, ok := _field.Interface().(fmt.Stringer); ok {
			val = it.String()
		} else {
			panic("Invalid identifier type")
		}
		return val
	}
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
				val := convert(_field)
				(*where)[field.Name] = map[string]any{"_eq": val}
			}
		} else {
			// Other fields are added to the SET clause
			result = append(result, field.ColumnName+"="+GenerateParameterPlaceholder(_value.Field(field.Idx), args))
		}
	}

	return strings.Join(result, ",")
}

func (b *SQLBuilder[Model]) SortError(filter Sortable) (sort *map[string]string, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Error occurred during Sort generation", slog.Any("error", r),
				slog.String("table", b.tableName),
				slog.Any("fitler", filter),
			)
			err = fmt.Errorf("error generating Set for table %s", b.tableName)
		}
	}()
	sort = b.Sort(filter)
	return
}

func (b *SQLBuilder[Model]) Sort(filter Sortable) *map[string]string {
	if filter == nil {
		return nil
	}
	sortBy, sortOrder := filter.Sort()
	sortBy = strings.TrimSpace(sortBy)
	sortOrder = strings.TrimSpace(sortOrder)
	if sortBy == "" || sortOrder == "" {
		return nil
	}
	if slices.Contains(b.FieldNames(), sortBy) {
		return &map[string]string{
			sortBy: sortOrder,
		}
	}
	slog.Warn("sort by field not found in repository columns", "sortBy", sortBy, "sortOrder", sortOrder, "columns", b.FieldNames())
	return nil // Return nil if the sortBy field is not found in the repository columns
}

// OrderError is a wrapper around Order that recovers from panics
func (b *SQLBuilder[Model]) OrderError(order *map[string]string) (res string, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Error occurred during Sort generation", slog.Any("error", r),
				slog.String("table", b.tableName),
				slog.Any("order", order),
			)
			err = fmt.Errorf("error generating Order for table %s", b.tableName)
		}
	}()
	res, err = b.orderError(order)
	return
}

// Constructs the ORDER BY clause for a query
// order is a map[string]string{"field1": "asc", "field2": "desc"}
// which will be converted to "field1 ASC, field2 DESC"
func (b *SQLBuilder[Model]) orderError(order *map[string]string) (string, error) {
	// fmt.Println("order", order)
	if order == nil {
		return "", nil
	}

	// Generate the field names for the ORDER BY clause
	result := []string{}

	// fmt.Println("columnnames", b.columnNames)

	for key, val := range *order {
		// if key is in FieldNames, add it to the ORDER BY clause
		if slices.Contains(b.FieldNames(), key) {
			field := b.GetFieldByName(key)
			if field == nil {
				return "", fmt.Errorf("orderable field %s not found for model %s", key, b.tableName)
			}
			result = append(result, fmt.Sprintf("%s %s", field.QualifiedColumnName, strings.ToUpper(val)))
		}
	}

	return strings.Join(result, ","), nil
}

// Constructs the ORDER BY clause for a query
// order is a map[string]string{"field1": "asc", "field2": "desc"}
// which will be converted to "field1 ASC, field2 DESC"
func (b *SQLBuilder[Model]) Order(order *map[string]string) string {
	// fmt.Println("order", order)
	if order == nil {
		return ""
	}

	// Generate the field names for the ORDER BY clause
	result := []string{}

	// fmt.Println("columnnames", b.columnNames)

	for key, val := range *order {
		// if key is in FieldNames, add it to the ORDER BY clause
		if slices.Contains(b.FieldNames(), key) {
			field := b.MustGetFieldByName(key)
			result = append(result, fmt.Sprintf("%s %s", field.ColumnName, strings.ToUpper(val)))
		}
	}

	return strings.Join(result, ",")
}
