package repository

import (
	"context"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/tkahng/authgo/internal/db/models"
)

type CreateRoleDto struct {
	Name        string  `json:"name"`
	DisplayName *string `json:"displayName,omitempty"`
}

func CreateRole(ctx context.Context, dbx bob.Executor, role *CreateRoleDto) (*models.Role, error) {
	data, err := models.Roles.Insert(
		&models.RoleSetter{
			Name:        omit.From(role.Name),
			Description: omitnull.FromPtr(role.DisplayName),
		},
		im.Returning("*"),
	).One(ctx, dbx)

	return OptionalRow(data, err)
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

func CreatePermission(ctx context.Context, dbx bob.Executor, permission *CreatePermissionDto) (*models.Permission, error) {
	data, err := models.Permissions.Insert(
		&models.PermissionSetter{
			Name:        omit.From(permission.Name),
			Description: omitnull.FromPtr(permission.Description),
		},
		im.Returning("*"),
	).One(ctx, dbx)
	return OptionalRow(data, err)
}

func FindPermissionsByName(ctx context.Context, dbx bob.Executor, params []string) ([]*models.Permission, error) {
	return models.Permissions.Query(models.SelectWhere.Permissions.Name.In(params...)).All(ctx, dbx)
}
