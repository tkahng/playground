package repository_test

import (
	"context"
	"testing"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/test"
)

func TestInitRbac(t *testing.T) {
	t.Parallel()
	test.WithTx(t, func(ctx context.Context, dbx database.Dbx) {
		repository.InitRbac(t, dbx)
		roles := repository.MustFindAll(t, repository.Role, dbx, nil)
		if len(roles) != 4 {
			t.Errorf("expected 4 roles, got %d", len(roles))
		}
		permissions := repository.MustFindAll(t, repository.Permission, dbx, nil)
		if len(permissions) != 4 {
			t.Errorf("expected 4 permissions, got %d", len(permissions))
		}
	})
}
