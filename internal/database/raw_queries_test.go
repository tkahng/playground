//go:build !integration

package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/tools/security"
)

func TestCreateDatabaseWithTemplate(t *testing.T) {
	cfg := conf.ZeroEnvConfig()
	cfg.Db.Db = "postgres"
	dbx, err := CreatePool(context.Background(), cfg.Db.GetDatabaseURL())
	if err != nil {
		t.Fatal(err)
	}
	defer dbx.Close()
	name := fmt.Sprintf("playground_test_%s", security.RandomString(10))
	err = CreateDatabaseWithTemplate(context.Background(), dbx, name, "playground_test")
	if err != nil {
		t.Fatal(err)
	}
	err = DeleteDatabase(context.Background(), dbx, name)
	if err != nil {
		t.Fatal(err)
	}
}
