package repository_test

import (
	"context"
	"slices"
	"testing"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/types"
)

var (
	knownRoleNames, knwonPermissionNames                     = []string{"superuser", "advanced", "pro", "basic"}, []string{"superuser", "advanced", "pro", "basic"}
	knownRoleNamesPermissionsMap         map[string][]string = map[string][]string{
		"basic":     {"basic"},
		"pro":       {"basic", "pro"},
		"advanced":  {"basic", "pro", "advanced"},
		"superuser": {"basic", "pro", "advanced", "superuser"},
	}
)

func TestAuth_UserAccountRbac(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, dbx database.Dbx) {

		var (
			roles             *[]*models.Role                = &[]*models.Role{}
			permissions       *[]*models.Permission          = &[]*models.Permission{}
			roleNameMap       *map[string]*models.Role       = &map[string]*models.Role{}
			permissionNameMap *map[string]*models.Permission = &map[string]*models.Permission{}
		)

		// init rbac
		t.Run("initiating rbac. should not panic", func(t *testing.T) {
			repository.CreateRolesAndPermissions(t, dbx, knownRoleNamesPermissionsMap)
		})

		// find all 4 roles and 4 permissions
		// map role names to roles, map permission names to permissions
		t.Run("find all 4 roles and 4 permissions", func(t *testing.T) {
			// find all roles
			*roles = repository.MustFindAll(t, repository.Role, dbx, nil)

			tempRoleNameMap := *roleNameMap
			// map role names to roles
			for _, role := range *roles {
				tempRoleNameMap[role.Name] = role
			}
			// update the roleNameMap
			*roleNameMap = tempRoleNameMap
			// get the number of keys in the roleNameMap
			if len(tempRoleNameMap) != 4 {
				t.Errorf("expected 4 roles, got %d", len(tempRoleNameMap))
			}

			// find all permissions
			*permissions = repository.MustFindAll(t, repository.Permission, dbx, nil)

			tempPermissionNameMap := *permissionNameMap
			// map permission names to permissions
			for _, permission := range *permissions {
				tempPermissionNameMap[permission.Name] = permission
			}
			// update the permissionNameMap
			*permissionNameMap = tempPermissionNameMap

			if len(tempPermissionNameMap) != 4 {
				t.Errorf("expected 4 permissions, got %d", len(tempPermissionNameMap))
			}
		})

		// populate all roles.permissions relation by querying the permission's roles names
		// compare the roles.permissions relation with the knownRoleNamesPermissionsMap
		t.Run("find each role's permissions by querying the permission's roles names, then assign them to the role's permissions relation field.", func(t *testing.T) {
			// iterate over all roles
			for _, role := range *roles {
				// find each role's permissions by querying the permission's roles names
				rolePermissions := repository.MustFindAll(
					t,
					repository.Permission,
					dbx,
					&map[string]any{
						"roles": map[string]any{
							"name": map[string]any{
								"_eq": role.Name,
							},
						},
					},
				)
				// assigned the found permissions to the role
				role.Permissions = rolePermissions
			}
			// check each role has at least one permission
			for _, role := range *roles {
				if len(role.Permissions) == 0 {
					t.Fatalf("expected more than one permission per role, got 0")
				}
			}
		})

		// verify each role has the correct number and specific permissions
		t.Run("verify each role has the correct number of permissions based on knownRoleNamesPermissionsMap", func(t *testing.T) {
			tempRoleNameMap := *roleNameMap
			for roleName, permissionNames := range knownRoleNamesPermissionsMap {
				if role, ok := tempRoleNameMap[roleName]; ok {
					if len(role.Permissions) != len(permissionNames) {
						t.Errorf("expected role %s to have %d permissions, got %d", role.Name, len(permissionNames), len(role.Permissions))
					}
					for _, permissionName := range permissionNames {
						if !slices.ContainsFunc(role.Permissions, func(permission *models.Permission) bool {
							return permission.Name == permissionName
						}) {
							t.Errorf("expected role %s to have permission %s", role.Name, permissionName)
						}
					}
				} else {
					t.Fatalf("expected role %s to exist", roleName)
				}
			}
		})

		t.Run("find roles with basic permission", func(t *testing.T) {
			// find roles with basic permission.
			// there should be 4, each role should have basic permission
			rolesWithBasicPermission := repository.MustFindAll(t, repository.Role, dbx, &map[string]any{
				"permissions": map[string]any{
					"name": map[string]any{
						"_eq": "basic",
					},
				},
			})
			// should be 4
			if len(rolesWithBasicPermission) != 4 {
				t.Errorf("expected 4 roles with basic permissions, got %d", len(rolesWithBasicPermission))
			}
			// every role name should be known
			test.TestSliceEveryFunc(t, "every role name should be known", rolesWithBasicPermission, func(role *models.Role) bool {
				return slices.Contains(knownRoleNames, role.Name)
			})
			// every role name should be unique
			test.TestSliceEveryUniqueFunc(t, "all role names should be unique", rolesWithBasicPermission, func(role *models.Role) string {
				return role.Name
			})

		})
		t.Run("roles with advanced permission or with names basic and pro", func(t *testing.T) {
			// roles with advanced permission and with names in basic and pro
			//
			// 2 roles with advanced permission(admin and advanced)
			// 2 roles with names in basic and pro
			basicProNames := []string{"basic", "pro"}
			permAdvNamesInBasicProRoles := repository.MustFindAll(t, repository.Role, dbx, &map[string]any{
				"_or": []map[string]any{
					{
						"permissions": map[string]any{
							"name": map[string]any{
								"_eq": "advanced",
							},
						},
					},
					{
						"name": map[string]any{
							"_in": basicProNames,
						},
					},
				},
			})
			// total 4
			if len(permAdvNamesInBasicProRoles) != 4 {
				t.Errorf("expected 4 roles with permissions, got %d", len(permAdvNamesInBasicProRoles))
			}

			var rolesWithAdvPerms, rolesWithBasicOrProName []*models.Role
			for _, role := range permAdvNamesInBasicProRoles {
				if slices.Contains(basicProNames, role.Name) {
					rolesWithBasicOrProName = append(rolesWithBasicOrProName, role)
				} else {
					rolesWithAdvPerms = append(rolesWithAdvPerms, role)
				}
			}

			rolesWithBasicOrProName = repository.MustFindAll(t, repository.Role, dbx, &map[string]any{
				"_or": []map[string]any{
					{
						"name": map[string]any{
							"_eq": "superuser",
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
			if len(rolesWithBasicOrProName) != 3 {
				t.Errorf("expected 2 roles with permissions, got %d", len(rolesWithBasicOrProName))
			}
		})
	})
}

func TestMustCreateUserAndAccount(t *testing.T) {
	t.Parallel()
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		tests := []struct {
			name string // description of this test case
			// Named input parameters for target function.
			db        database.Dbx
			fns       []func(*models.User, *models.UserAccount)
			predicate func(*models.User, *models.UserAccount)
		}{
			{
				name: "Create user and credential account",
				db:   db,
				fns: []func(user *models.User, account *models.UserAccount){
					func(user *models.User, account *models.UserAccount) {
						user.Name = types.Pointer("credential")
						account.Type = models.ProviderTypeCredentials
						account.Provider = models.ProvidersCredentials
					},
				},
				predicate: func(user *models.User, account *models.UserAccount) {
					if user == nil {
						t.Errorf("expected user, got nil")
					}
					if account == nil {
						t.Errorf("expected account, got nil")
					}
					if *user.Name != "credential" {
						t.Errorf("expected user name to be credential, got %s", *user.Name)
					}
					if account.Type != models.ProviderTypeCredentials {
						t.Errorf("expected account type to be credentials, got %s", account.Type)
					}
					if account.Provider != models.ProvidersCredentials {
						t.Errorf("expected account provider to be credentials, got %s", account.Provider)
					}
				},
			},
			{
				name: "create google account",
				db:   db,
				fns: []func(user *models.User, account *models.UserAccount){
					func(user *models.User, account *models.UserAccount) {
						user.Name = types.Pointer("google")
						account.Type = models.ProviderTypeOAuth
						account.Provider = models.ProvidersGoogle
					},
				},
				predicate: func(user *models.User, account *models.UserAccount) {
					if user == nil {
						t.Errorf("expected user, got nil")
					}
					if account == nil {
						t.Errorf("expected account, got nil")
					}
					if *user.Name != "google" {
						t.Errorf("expected user name to be google, got %s", *user.Name)
					}
					if account.Type != models.ProviderTypeOAuth {
						t.Errorf("expected account type to be oauth, got %s", account.Type)
					}
					if account.Provider != models.ProvidersGoogle {
						t.Errorf("expected account provider to be google, got %s", account.Provider)
					}
				},
			},
			{
				name: "create github account",
				db:   db,
				fns: []func(user *models.User, account *models.UserAccount){
					func(user *models.User, account *models.UserAccount) {
						user.Name = types.Pointer("github")
						account.Type = models.ProviderTypeOAuth
						account.Provider = models.ProvidersGithub
					},
				},
				predicate: func(user *models.User, account *models.UserAccount) {
					if user == nil {
						t.Errorf("expected user, got nil")
					}
					if account == nil {
						t.Errorf("expected account, got nil")
					}
					if *user.Name != "github" {
						t.Errorf("expected user name to be github, got %s", *user.Name)
					}
					if account.Type != models.ProviderTypeOAuth {
						t.Errorf("expected account type to be oauth, got %s", account.Type)
					}
					if account.Provider != models.ProvidersGithub {
						t.Errorf("expected account provider to be github, got %s", account.Provider)
					}
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				user, account := repository.MustCreateUserAndAccount(t, tt.db, tt.fns...)
				tt.predicate(user, account)
			})
		}
	})
}

func TestMustCreateUserAndAccount_Randomize(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		var count int64 = 100
		for range count {
			repository.MustCreateUserAndAccount(t, db)
		}
		userCount := repository.MustCountAll(t, repository.User, db, &map[string]any{})
		if userCount != count {
			t.Errorf("expected at least 10 users, got %d", userCount)
		}
		countAcc := repository.MustCountAll(t, repository.User, db, &map[string]any{})
		if countAcc != count {
			t.Errorf("expected at least 10 accounts, got %d", countAcc)
		}
	})
}

// func TestPost(t *testing.T) {
// 	database.WithNewTx(t, func(ctx context.Context, db database.Dbx) {
// 		testfunc := func() {
// 			var userInput *models.User = new(models.User)
// 			var userInputs []*models.User = []*models.User{userInput}
// 			res, err := repository.post(t, repository.User, db, userInput)
// 		}
// 	})
// }

// func RepositoryPostTestFunc[T any](t testing.TB, repo repository.Repository[any], db database.Dbx, arg *map[string]any)
