package repository

import (
	"context"
	"fmt"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/tkahng/authgo/internal/db/models"
)

func CreatePermissionsForTree(ctx context.Context, db bob.Executor, rolePermissionsMap map[string]*models.PermissionSetter) (PermissionsMap, error) {

	rolesmap := make(PermissionsMap)
	for roleName, params := range rolePermissionsMap {
		permissions, err := models.Permissions.Insert(params, im.Returning("*")).One(ctx, db)
		if err != nil {
			return nil, fmt.Errorf("error creating permission: %w", err)
		}
		rolesmap[roleName] = permissions
	}
	return rolesmap, nil
}

func CreateRolesForTree(ctx context.Context, dbx bob.Executor, roles map[string]*models.RoleSetter) (RolesMap, error) {
	rolesmap := make(RolesMap)
	for name, params := range roles {
		role, err := models.Roles.Insert(params, im.Returning("*")).One(ctx, dbx)
		if err != nil {
			return nil, fmt.Errorf("error creating role: %w", err)
		}
		rolesmap[name] = role
	}
	return rolesmap, nil
}

func SyncRolesAndPermissionsFromTree(ctx context.Context, dbx bob.Executor, roleTree map[string][]string, rolesMap RolesMap, permissionsMap PermissionsMap) (RoleStructTree, error) {
	dtos := make(RoleStructTree)
	for roleName, permissions := range roleTree {
		var role = rolesMap[roleName]
		var args []*models.Permission
		for _, permission := range permissions {
			args = append(args, permissionsMap[permission])
		}
		err := role.AttachPermissions(ctx, dbx, args...)
		if err != nil {
			return nil, fmt.Errorf("error creating role: %w", err)
		}
		// fmt.Println(re)
		perms, err := role.Permissions().All(ctx, dbx)
		if err != nil {
			return nil, fmt.Errorf("error creating role: %w", err)
		}
		dtos[roleName] = RoleDto{
			Role:        role,
			Permissions: perms,
		}
	}

	fmt.Println(dtos)
	return dtos, nil
}

type RoleTreeDto struct {
	Admin          string                              `json:"admin"`
	Manager        string                              `json:"manager"`
	Basic          string                              `json:"basic"`
	RoleTree       map[string][]string                 `json:"role_tree"`
	RoleArgs       map[string]*models.RoleSetter       `json:"role_args"`
	PermissionArgs map[string]*models.PermissionSetter `json:"permission_args"`
	// Advanced       string                              `json:"advanced"`
	// Pro            string                              `json:"pro"`
}

func GenerateRoleTreeDto() RoleTreeDto {
	var (
		Admin   = "admin"
		Manager = "manager"
		Basic   = "basic"
	)
	return RoleTreeDto{
		Admin:   "admin",
		Manager: "manager",
		Basic:   "basic",
		RoleTree: map[string][]string{
			Admin:   {Admin, Manager, Basic},
			Manager: {Manager, Basic},
			Basic:   {Basic},
		},

		RoleArgs: map[string]*models.RoleSetter{
			Admin:   {Name: omit.From(Admin), Description: omitnull.From(Admin)},
			Manager: {Name: omit.From(Manager), Description: omitnull.From(Manager)},
			Basic:   {Name: omit.From(Basic), Description: omitnull.From(Basic)},
		},
		PermissionArgs: map[string]*models.PermissionSetter{
			Admin:   {Name: omit.From(Admin), Description: omitnull.From(Admin)},
			Manager: {Name: omit.From(Manager), Description: omitnull.From(Manager)},
			Basic:   {Name: omit.From(Basic), Description: omitnull.From(Basic)},
		},
	}
}

func PopulateRolesFromTree(ctx context.Context, dbx bob.Executor) error {
	tree := GenerateRoleTreeDto()
	rolesMap, err := CreateRolesForTree(ctx, dbx, tree.RoleArgs)
	if err != nil {
		return fmt.Errorf("error creating roles: %w", err)
	}
	permissionsMap, err := CreatePermissionsForTree(ctx, dbx, tree.PermissionArgs)
	if err != nil {
		return fmt.Errorf("error creating permissions: %w", err)
	}

	_, err = SyncRolesAndPermissionsFromTree(ctx, dbx, tree.RoleTree, rolesMap, permissionsMap)
	if err != nil {
		return fmt.Errorf("error syncing roles and permissions: %w", err)
	}
	return nil
}

func InitRolesFromTree(ctx context.Context, dbx bob.Executor) {

	cnt, err := models.Roles.Query().Count(ctx, dbx)
	if err != nil {
		panic(err)
	}
	if cnt == 0 {
		PopulateRolesFromTree(ctx, dbx)
	}
	a, err := models.Roles.Query(models.SelectWhere.Roles.Name.EQ("admin")).One(ctx, dbx)
	a, err = OptionalRow(a, err)
	if err != nil {
		panic(err)
	}
	if a == nil {
		role, err := models.Roles.Insert(&models.RoleSetter{Name: omit.From("admin"), Description: omitnull.From("admin")}, im.Returning("*")).One(ctx, dbx)
		_, err = OptionalRow(role, err)
		if err != nil {
			panic(err)
		}
		// fmt.Println(role)
	}
}
