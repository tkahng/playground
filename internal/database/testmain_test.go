//go:build !integration

package database_test

import (
	"os"
	"testing"

	"github.com/tkahng/playground/internal/database"
)

func TestMain(m *testing.M) {
	database.MustInitTestDB()
	os.Exit(m.Run())
}
