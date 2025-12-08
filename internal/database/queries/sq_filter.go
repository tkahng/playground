package queries

import (
	"fmt"

	"github.com/Masterminds/squirrel"
)

type FilterOperator string

const (
	FilterOperatorEq      FilterOperator = "eq"
	FilterOperatorGt      FilterOperator = "gt"
	FilterOperatorGte     FilterOperator = "gte"
	FilterOperatorLt      FilterOperator = "lt"
	FilterOperatorLte     FilterOperator = "lte"
	FilterOperatorNeq     FilterOperator = "neq"
	FilterOperatorNotNull FilterOperator = "not_null"
	FilterOperatorNull    FilterOperator = "null"
)

func ToSquirrelOp(sq squirrel.SelectBuilder, op FilterOperator, key string, value any) squirrel.SelectBuilder {
	switch op {
	case FilterOperatorEq:
		return sq.Where(fmt.Sprintf("%s = ?", key), value)
	case FilterOperatorGt:
		return sq.Where(fmt.Sprintf("%s > ?", key), value)
	case FilterOperatorGte:
		return sq.Where(fmt.Sprintf("%s >= ?", key), value)
	case FilterOperatorLt:
		return sq.Where(fmt.Sprintf("%s < ?", key), value)
	case FilterOperatorLte:
		return sq.Where(fmt.Sprintf("%s <= ?", key), value)
	case FilterOperatorNeq:
		return sq.Where(fmt.Sprintf("%s != ?", key), value)
	case FilterOperatorNotNull:
		return sq.Where(fmt.Sprintf("%s IS NOT NULL", key))
	case FilterOperatorNull:
		return sq.Where(fmt.Sprintf("%s IS NULL", key))
	default:
		return sq
	}
}
