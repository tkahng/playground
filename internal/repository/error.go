package repository

import "strings"

func IsUniqConstraintErr(err error) bool {
	return strings.Contains(err.Error(), `(SQLSTATE 23505)`)
}
