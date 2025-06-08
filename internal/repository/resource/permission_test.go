package resource

import (
	"context"
	"fmt"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/test"
)

func TestNewPermissionQueryResource_FilterFunc(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		repo := NewPermissionQueryResource(db)

		filterFunc := repo.filter

		t.Run("nil filter returns empty map", func(t *testing.T) {
			qs := squirrel.Select(models.PermissionTable.Columns...).From(repo.builder.Table())
			where := filterFunc(qs, nil)
			sql, args, err := where.ToSql()
			fmt.Println("SQL:", sql)
			fmt.Println("Args:", args)
			assert.NoError(t, err)
			assert.NotNil(t, sql)
			assert.Equal(t, 0, len(args))
		})

		t.Run("Q filter", func(t *testing.T) {
			qs := squirrel.Select(models.PermissionTable.Columns...).From(repo.builder.Table())
			filter := &PermissionsListFilter{
				Q: "test",
			}
			where := filterFunc(qs, filter)
			sql, args, err := where.ToSql()
			assert.NoError(t, err)
			assert.NotNil(t, sql)
			assert.Equal(t, 2, len(args))
			expected := "SELECT id, name, description, created_at, updated_at FROM \"permissions\" WHERE (name ILIKE ? OR description ILIKE ?)"
			assert.Equal(t, expected, sql)
		})

		t.Run("Ids filter", func(t *testing.T) {
			qs := squirrel.Select(models.PermissionTable.Columns...).From(repo.builder.Table())
			id1 := uuid.New()
			id2 := uuid.New()
			filter := &PermissionsListFilter{
				Ids: []uuid.UUID{id1, id2},
			}
			where := filterFunc(qs, filter)
			sql, args, err := where.ToSql()
			assert.NoError(t, err)
			assert.NotNil(t, sql)
			assert.Equal(t, 2, len(args))
			expected := "SELECT id, name, description, created_at, updated_at FROM \"permissions\" WHERE id IN (?,?)"
			assert.Equal(t, expected, sql)
		})

		t.Run("Ids filter", func(t *testing.T) {
			qs := squirrel.Select(models.PermissionTable.Columns...).From(repo.builder.Table())
			id1 := uuid.New()
			id2 := uuid.New()
			filter := &PermissionsListFilter{
				Ids: []uuid.UUID{id1, id2},
			}
			where := filterFunc(qs, filter)
			sql, args, err := where.ToSql()
			assert.NoError(t, err)
			assert.NotNil(t, sql)
			assert.Equal(t, 2, len(args))
			expected := "SELECT id, name, description, created_at, updated_at FROM \"permissions\" WHERE id IN (?,?)"
			assert.Equal(t, expected, sql)
		})

		t.Run("Names filter", func(t *testing.T) {
			qs := squirrel.Select(models.PermissionTable.Columns...).From(repo.builder.Table())
			filter := &PermissionsListFilter{
				Names: []string{"read", "write"},
			}
			where := filterFunc(qs, filter)
			sql, args, err := where.ToSql()
			assert.NoError(t, err)
			assert.NotNil(t, sql)
			assert.Equal(t, 2, len(args))
			expected := "SELECT id, name, description, created_at, updated_at FROM \"permissions\" WHERE name IN (?,?)"
			assert.Equal(t, expected, sql)
		})

		t.Run("RoleId filter", func(t *testing.T) {
			qs := squirrel.Select(models.PermissionTable.Columns...).From(repo.builder.Table())
			roleId := uuid.New()
			filter := &PermissionsListFilter{
				RoleId: roleId,
			}
			where := filterFunc(qs, filter)
			sql, args, err := where.ToSql()
			assert.NoError(t, err)
			assert.NotNil(t, sql)
			assert.Equal(t, 2, len(args))
			expected := "SELECT id, name, description, created_at, updated_at FROM \"permissions\" JOIN role_permissions on permissions.id = role_permissions.permission_id and role_permissions.role_id = ? WHERE role_permissions.role_id = ?"
			assert.Equal(t, expected, sql)
		})
		t.Run("RoleId filter", func(t *testing.T) {
			qs := squirrel.Select(models.PermissionTable.Columns...).From(repo.builder.Table())
			roleId := uuid.New()
			filter := &PermissionsListFilter{
				RoleId: roleId,
			}
			where := filterFunc(qs, filter)
			sql, args, err := where.ToSql()
			assert.NoError(t, err)
			assert.NotNil(t, sql)
			assert.Equal(t, 2, len(args))
			expected := "SELECT id, name, description, created_at, updated_at FROM \"permissions\" JOIN role_permissions on permissions.id = role_permissions.permission_id and role_permissions.role_id = ? WHERE role_permissions.role_id = ?"
			assert.Equal(t, expected, sql)
		})
	})

}
