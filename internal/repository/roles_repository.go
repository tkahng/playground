package repository

import (
	"context"
	"log"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/tkahng/authgo/internal/db/models"
)

func EnsureRoleAndPermissions(ctx context.Context, db bob.DB, roleName string, permissionNames ...string) error {
	// find superuser role
	role, err := FindOrCreateRole(ctx, db, roleName)
	if err != nil {
		return err
	}
	for _, permissionName := range permissionNames {
		perm, err := FindOrCreatePermission(ctx, db, permissionName)
		if err != nil {
			continue
		}

		err = role.AttachPermissions(ctx, db, perm)
		if err != nil && !IsUniqConstraintErr(err) {
			log.Println(err)
		}
	}
	return nil
}

type CreateRoleDto struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func FindOrCreateRole(ctx context.Context, dbx bob.Executor, roleName string) (*models.Role, error) {
	role, err := FindRoleByName(ctx, dbx, roleName)
	if err != nil {
		return nil, err
	}
	if role == nil {
		role, err = CreateRole(ctx, dbx, &CreateRoleDto{Name: roleName})
		if err != nil {
			return nil, err
		}
	}
	return role, nil
}

func CreateRole(ctx context.Context, dbx bob.Executor, role *CreateRoleDto) (*models.Role, error) {
	data, err := models.Roles.Insert(
		&models.RoleSetter{
			Name:        omit.From(role.Name),
			Description: omitnull.FromPtr(role.Description),
		},
		im.Returning("*"),
	).One(ctx, dbx)
	return data, err
}

func FindRolesByNames(ctx context.Context, dbx bob.Executor, params []string) ([]*models.Role, error) {
	return models.Roles.Query(models.SelectWhere.Roles.Name.In(params...)).All(ctx, dbx)
}

func FindRoleByName(ctx context.Context, dbx bob.Executor, name string) (*models.Role, error) {
	data, err := models.Roles.Query(models.SelectWhere.Roles.Name.EQ(name)).One(ctx, dbx)
	return OptionalRow(data, err)
}

type CreatePermissionDto struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func FindOrCreatePermission(ctx context.Context, dbx bob.Executor, permissionName string) (*models.Permission, error) {
	permission, err := FindPermissionByName(ctx, dbx, permissionName)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		permission, err = CreatePermission(ctx, dbx, &CreatePermissionDto{Name: permissionName})
		if err != nil {
			return nil, err
		}
	}
	return permission, nil
}

func CreatePermission(ctx context.Context, dbx bob.Executor, permission *CreatePermissionDto) (*models.Permission, error) {
	data, err := models.Permissions.Insert(
		&models.PermissionSetter{
			Name:        omit.From(permission.Name),
			Description: omitnull.FromPtr(permission.Description),
		},
		im.Returning("*"),
	).One(ctx, dbx)
	return data, err
}

func FindPermissionByName(ctx context.Context, dbx bob.Executor, params string) (*models.Permission, error) {
	data, err := models.Permissions.Query(
		models.SelectWhere.Permissions.Name.EQ(params),
	).One(ctx, dbx)
	return OptionalRow(data, err)
}
func FindPermissionsByNames(ctx context.Context, dbx bob.Executor, params []string) ([]*models.Permission, error) {
	return models.Permissions.Query(models.SelectWhere.Permissions.Name.In(params...)).All(ctx, dbx)
}
