package repository_test

import (
	"fmt"
	"testing"
)

func TestStringBuilder(t *testing.T) {
	query := testSubQueryBuilder(t)
	t.Log(query)
}
func testSubQueryBuilder(t *testing.T) string {
	query := fmt.Sprintf(
		`SELECT %s FROM %s join %s on %s.%s = %s.%s`,
		"dest",
		"through",
		"model",
		"model",
		"endField",
		"through",
		"throughField",
	)
	return query
}
