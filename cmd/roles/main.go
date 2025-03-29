package main

import (
	"context"
	"fmt"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/utils"
)

// var (
// 	Admin    = "admin"
// 	Advanced = "advanced"
// 	Pro      = "pro"
// 	Basic    = "basic"
// 	RoleTree = map[string][]string{
// 		Admin:    {Admin, Advanced, Pro, Basic},
// 		Advanced: {Advanced, Pro, Basic},
// 		Pro:      {Pro, Basic},
// 		Basic:    {Basic},
// 	}

// 	RoleArgs = map[string]*models.RoleSetter{
// 		Admin:    {Name: omit.From(Admin), Description: omitnull.From(Admin)},
// 		Advanced: {Name: omit.From(Advanced), Description: omitnull.From(Advanced)},
// 		Pro:      {Name: omit.From(Pro), Description: omitnull.From(Pro)},
// 		Basic:    {Name: omit.From(Basic), Description: omitnull.From(Basic)},
// 	}
// 	PermissionArgs = map[string]*models.PermissionSetter{
// 		Admin:    {Name: omit.From(Admin), Description: omitnull.From(Admin)},
// 		Advanced: {Name: omit.From(Advanced), Description: omitnull.From(Advanced)},
// 		Pro:      {Name: omit.From(Pro), Description: omitnull.From(Pro)},
// 		Basic:    {Name: omit.From(Basic), Description: omitnull.From(Basic)},
// 	}
// )

func main() {
	ctx := context.Background()
	conf := conf.AppConfigGetter()

	dbx := core.NewBobFromConf(ctx, conf.Db)
	tree := repository.GenerateRoleTreeDto()
	rolesMap, _ := repository.CreateRolesForTree(ctx, dbx, tree.RoleArgs)
	permissionsMap, _ := repository.CreatePermissionsForTree(ctx, dbx, tree.PermissionArgs)
	fmt.Println(permissionsMap)
	fmt.Println(rolesMap)

	result, _ := repository.SyncRolesAndPermissionsFromTree(ctx, dbx, tree.RoleTree, rolesMap, permissionsMap)
	utils.PrettyPrintJSON(result)
	NewFunction(ctx, dbx)
	// fmt.Println(result)

}

func NewFunction(ctx context.Context, db bob.DB) bool {
	user, err := repository.CreateUser(ctx, db, &shared.AuthenticateUserParams{
		Email: "tkahng@gmail.com",
	})

	if err != nil {
		fmt.Println(err)
		return true
	}
	roles, err := repository.FindRolesByNames(ctx, db, []string{"basic", "pro", "admin"})
	if err != nil {
		fmt.Println(err)
		return true
	}
	err = repository.AssignRoles(ctx, db, user, roles)
	if err != nil {
		fmt.Println(err)
		return true
	}
	return false
}

// func SyncRolesAndPermissions(ctx context.Context, rolesMap RolesMap, permissionsMap PermissionsMap, dbx db.AppQuery) RoleStructTree {

// 	dtos := make(RoleStructTree)
// 	for role, permissions := range RoleTree {
// 		var args []*db.CreateRolePermissionsParams
// 		for _, permission := range permissions {
// 			args = append(args, &db.CreateRolePermissionsParams{
// 				RoleID:       rolesMap[role].ID,
// 				PermissionID: permissionsMap[permission].ID,
// 			})
// 		}
// 		re, err := dbx.CreateRolePermissions(ctx, args)
// 		if err != nil {
// 			fmt.Println(err)
// 			continue
// 		}
// 		fmt.Println(re)
// 		perms, err := dbx.GetPermissionsByRole(ctx, role)
// 		if err != nil {
// 			fmt.Println(err)
// 			continue
// 		}
// 		dtos[role] = RoleDto{
// 			Role:        rolesMap[role],
// 			Permissions: perms,
// 		}
// 	}

// 	fmt.Println(dtos)
// 	return dtos
// }

// func CreatePermissions(ctx context.Context, dbx db.AppQuery, permissions map[string]*db.UpsertPermissionParams) PermissionsMap {
// 	rolesmap := make(PermissionsMap)
// 	for name, params := range permissions {
// 		role, err := dbx.UpsertPermission(ctx, params)
// 		if err != nil {
// 			fmt.Println(err)
// 			continue
// 		}
// 		rolesmap[name] = role
// 	}
// 	return rolesmap
// }

// func CreateRoles(ctx context.Context, dbx db.AppQuery, roles map[string]*db.UpsertRoleParams) RolesMap {
// 	rolesmap := make(RolesMap)
// 	for name, params := range roles {
// 		role, err := dbx.UpsertRole(ctx, params)
// 		if err != nil {
// 			fmt.Println(err)
// 			continue
// 		}
// 		rolesmap[name] = role
// 	}
// 	return rolesmap
// }
