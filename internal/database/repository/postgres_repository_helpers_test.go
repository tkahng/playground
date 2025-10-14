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
		// init
		repository.InitRbac(t, dbx)
		// find all
		roles := repository.MustFindAll(t, repository.Role, dbx, nil)
		if len(roles) != 4 {
			t.Errorf("expected 4 roles, got %d", len(roles))
		}
		permissions := repository.MustFindAll(t, repository.Permission, dbx, nil)
		if len(permissions) != 4 {
			t.Errorf("expected 4 permissions, got %d", len(permissions))
		}
		// find roles with permissions
		rolesWithPermissions := repository.MustFindAll(t, repository.Role, dbx, &map[string]any{
			"permissions": map[string]any{
				"name": map[string]any{
					"_eq": "admin",
				},
			},
		})
		if len(rolesWithPermissions) != 1 && rolesWithPermissions[0].Name != "admin" {
			t.Errorf("expected 1 roles with permissions, got %d", len(rolesWithPermissions))
		}
		rolesWithPermissions = repository.MustFindAll(t, repository.Role, dbx, &map[string]any{
			"permissions": map[string]any{
				"name": map[string]any{
					"_eq": "basic",
				},
			},
		})
		if len(rolesWithPermissions) != 4 {
			t.Errorf("expected 4 roles with permissions, got %d", len(rolesWithPermissions))
		}
		rolesWithPermissions = repository.MustFindAll(t, repository.Role, dbx, &map[string]any{
			"_or": []map[string]any{
				{
					"name": map[string]any{
						"_eq": "admin",
					},
				},
				{
					"permissions": map[string]any{
						"name": map[string]any{
							"_eq": "pro",
						},
					},
				},
			},
		})
		if len(rolesWithPermissions) != 3 {
			t.Errorf("expected 2 roles with permissions, got %d", len(rolesWithPermissions))
		}
	})
}
