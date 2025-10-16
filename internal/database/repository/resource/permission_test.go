// nolint:exhaustruct
package resource

import (
	"context"
	"fmt"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
)

func TestNewPermissionQueryResource_FilterFunc(t *testing.T) {

	database.WithNewTx(t, func(ctx context.Context, db database.Dbx) {
		repo := NewPermissionQueryResource(db)

		filterFunc := repo.filter

		t.Run("nil filter returns empty map", func(t *testing.T) {
			qs := squirrel.Select(repository.PermissionBuilder.QualifiedColumnNames()...).From(repo.builder.TableName())
			where := filterFunc(qs, nil)
			sql, args, err := where.ToSql()
			fmt.Println("SQL:", sql)
			fmt.Println("Args:", args)
			assert.NoError(t, err)
			assert.NotNil(t, sql)
			assert.Equal(t, 0, len(args))
		})

		t.Run("Q filter", func(t *testing.T) {
			qs := squirrel.Select(repository.PermissionBuilder.QualifiedColumnNames()...).From(repo.builder.TableName())
			filter := &PermissionsFilter{
				Q: "test",
			}
			where := filterFunc(qs, filter)
			sql, args, err := where.ToSql()
			assert.NoError(t, err)
			assert.NotNil(t, sql)
			assert.Equal(t, 2, len(args))
			expected := "SELECT auth.permissions.id, auth.permissions.name, auth.permissions.description, auth.permissions.created_at, auth.permissions.updated_at FROM auth.permissions WHERE (auth.permissions.name ILIKE ? OR auth.permissions.description ILIKE ?)"
			assert.Equal(t, expected, sql)
		})

		t.Run("Ids filter", func(t *testing.T) {
			qs := squirrel.Select(repository.PermissionBuilder.QualifiedColumnNames()...).From(repo.builder.TableName())
			id1 := uuid.New()
			id2 := uuid.New()
			filter := &PermissionsFilter{
				Ids: []uuid.UUID{id1, id2},
			}
			where := filterFunc(qs, filter)
			sql, args, err := where.ToSql()
			assert.NoError(t, err)
			assert.NotNil(t, sql)
			assert.Equal(t, 2, len(args))
			expected := "SELECT auth.permissions.id, auth.permissions.name, auth.permissions.description, auth.permissions.created_at, auth.permissions.updated_at FROM auth.permissions WHERE auth.permissions.id IN (?,?)"
			assert.Equal(t, expected, sql)
		})

		t.Run("Ids filter", func(t *testing.T) {
			qs := squirrel.Select(repository.PermissionBuilder.QualifiedColumnNames()...).From(repo.builder.TableName())
			id1 := uuid.New()
			id2 := uuid.New()
			filter := &PermissionsFilter{
				Ids: []uuid.UUID{id1, id2},
			}
			where := filterFunc(qs, filter)
			sql, args, err := where.ToSql()
			assert.NoError(t, err)
			assert.NotNil(t, sql)
			assert.Equal(t, 2, len(args))
			expected := "SELECT auth.permissions.id, auth.permissions.name, auth.permissions.description, auth.permissions.created_at, auth.permissions.updated_at FROM auth.permissions WHERE auth.permissions.id IN (?,?)"
			assert.Equal(t, expected, sql)
		})

		t.Run("Names filter", func(t *testing.T) {
			qs := squirrel.Select(repository.PermissionBuilder.QualifiedColumnNames()...).From(repo.builder.TableName())
			filter := &PermissionsFilter{
				Names: []string{"read", "write"},
			}
			where := filterFunc(qs, filter)
			sql, args, err := where.ToSql()
			assert.NoError(t, err)
			assert.NotNil(t, sql)
			assert.Equal(t, 2, len(args))
			expected := "SELECT auth.permissions.id, auth.permissions.name, auth.permissions.description, auth.permissions.created_at, auth.permissions.updated_at FROM auth.permissions WHERE auth.permissions.name IN (?,?)"
			assert.Equal(t, expected, sql)
		})

		t.Run("RoleId filter", func(t *testing.T) {
			qs := squirrel.Select(repository.PermissionBuilder.QualifiedColumnNames()...).From(repo.builder.TableName())
			roleId := uuid.New()
			filter := &PermissionsFilter{
				RoleId: roleId,
			}
			where := filterFunc(qs, filter)
			sql, args, err := where.ToSql()
			assert.NoError(t, err)
			assert.NotNil(t, sql)
			assert.Equal(t, 2, len(args))
			expected := "SELECT auth.permissions.id, auth.permissions.name, auth.permissions.description, auth.permissions.created_at, auth.permissions.updated_at FROM auth.permissions JOIN auth.role_permissions on auth.permissions.id = auth.role_permissions.permission_id and auth.role_permissions.role_id = ? WHERE auth.role_permissions.role_id = ?"
			assert.Equal(t, expected, sql)
		})
		t.Run("RoleId filter", func(t *testing.T) {
			qs := squirrel.Select(repository.PermissionBuilder.QualifiedColumnNames()...).From(repo.builder.TableName())
			roleId := uuid.New()
			filter := &PermissionsFilter{
				RoleId: roleId,
			}
			where := filterFunc(qs, filter)
			sql, args, err := where.ToSql()
			assert.NoError(t, err)
			assert.NotNil(t, sql)
			assert.Equal(t, 2, len(args))
			expected := "SELECT auth.permissions.id, auth.permissions.name, auth.permissions.description, auth.permissions.created_at, auth.permissions.updated_at FROM auth.permissions JOIN auth.role_permissions on auth.permissions.id = auth.role_permissions.permission_id and auth.role_permissions.role_id = ? WHERE auth.role_permissions.role_id = ?"
			assert.Equal(t, expected, sql)
		})
	})

}
