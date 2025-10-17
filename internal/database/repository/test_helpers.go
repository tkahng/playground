package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
)

/* –––– */
func MustCreate[T any](t *testing.T, repo Repository[T], db database.Dbx, arg *T) *T {
	t.Helper()
	res, err := repo.PostOne(t.Context(), db, arg)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func MustFindOne[T any](t *testing.T, repo Repository[T], db database.Dbx, arg *map[string]any) *T {
	t.Helper()
	res, err := repo.GetOne(t.Context(), db, arg)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func MustCountAll[T any](t *testing.T, repo Repository[T], db database.Dbx, arg *map[string]any) int64 {
	t.Helper()
	res, err := repo.Count(t.Context(), db, arg)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func MustFindAll[T any](t *testing.T, repo Repository[T], db database.Dbx, arg *map[string]any) []*T {
	t.Helper()
	res, err := repo.Get(t.Context(), db, arg, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func MustCreateUserAndAccount(t *testing.T, db database.Dbx, fns ...func(*models.User, *models.UserAccount)) (*models.User, *models.UserAccount) {
	t.Helper()
	uid := uuid.NewString()
	user := &models.User{
		Email: uid + "@example.com",
		Name:  &uid,
	}
	pw := "Password123!"
	res := &models.UserAccount{
		Type:              models.ProviderTypeCredentials,
		ProviderAccountID: uid,
		Provider:          models.ProvidersCredentials,
		Password:          &pw,
	}
	for _, fn := range fns {
		fn(user, res)
	}
	user = MustCreate(t, User, db, user)
	res.UserID = user.ID
	res = MustCreate(t, UserAccount, db, res)
	return user, res
}

func MustCreateUser(t *testing.T, db database.Dbx, fns ...func(*models.User)) *models.User {
	t.Helper()
	uid := uuid.NewString()
	user := &models.User{
		Email: uid + "@example.com",
		Name:  &uid,
	}
	for _, fn := range fns {
		fn(user)
	}
	return MustCreate(t, User, db, user)
}
func MustCreateAccount(t *testing.T, db database.Dbx, fns ...func(*models.UserAccount)) *models.UserAccount {
	t.Helper()
	uid := uuid.NewString()
	pw := "Password123!"
	res := &models.UserAccount{
		Type:              models.ProviderTypeCredentials,
		ProviderAccountID: uid,
		Provider:          models.ProvidersCredentials,
		Password:          &pw,
	}
	for _, fn := range fns {
		fn(res)
	}
	return MustCreate(t, UserAccount, db, res)
}
func MustFindRoleByName(t *testing.T, db database.Dbx, name string) *models.Role {
	t.Helper()
	return MustFindOne(t, Role, db, &map[string]any{"name": map[string]any{"_eq": name}})
}
func MustCreateRoleByName(t *testing.T, db database.Dbx, name string) *models.Role {
	t.Helper()
	return MustCreate(t, Role, db, &models.Role{Name: name})
}
func MustFindOrCreateRoleByName(t *testing.T, db database.Dbx, name string) *models.Role {
	t.Helper()
	role := MustFindRoleByName(t, db, name)
	if role == nil {
		role = MustCreateRoleByName(t, db, name)
	}
	return role
}
func MustCreateUserRoleByName(t *testing.T, db database.Dbx, userId uuid.UUID, name string) *models.UserRole {
	t.Helper()
	role := MustFindRoleByName(t, db, name)
	return MustCreate(t, UserRole, db, &models.UserRole{UserID: userId, RoleID: role.ID})
}
func MustFindPermissionByName(t *testing.T, db database.Dbx, name string) *models.Permission {
	t.Helper()
	return MustFindOne(t, Permission, db, &map[string]any{"name": map[string]any{"_eq": name}})
}
func MustCreatePermissionByName(t *testing.T, db database.Dbx, name string) *models.Permission {
	t.Helper()
	return MustCreate(t, Permission, db, &models.Permission{Name: name})
}

func MustFindOrCreatePermissionByName(t *testing.T, db database.Dbx, name string) *models.Permission {
	t.Helper()
	role := MustFindPermissionByName(t, db, name)
	if role == nil {
		role = MustCreatePermissionByName(t, db, name)
	}
	return role
}
func MustCreateRolePermissionByName(t *testing.T, db database.Dbx, roleName, permissionName string) *models.RolePermission {
	t.Helper()
	role := MustFindOrCreateRoleByName(t, db, roleName)
	permission := MustFindOrCreatePermissionByName(t, db, permissionName)
	return MustCreate(t, RolePermission, db, &models.RolePermission{PermissionID: permission.ID, RoleID: role.ID})
}

func InitRbac(t *testing.T, db database.Dbx) {
	t.Helper()
	MustCreateRoleByName(t, db, "admin")
	MustCreateRoleByName(t, db, "advanced")
	MustCreateRoleByName(t, db, "pro")
	MustCreateRoleByName(t, db, "basic")

	MustCreatePermissionByName(t, db, "admin")
	MustCreatePermissionByName(t, db, "advanced")
	MustCreatePermissionByName(t, db, "pro")
	MustCreatePermissionByName(t, db, "basic")

	MustCreateRolePermissionByName(t, db, "admin", "admin")
	MustCreateRolePermissionByName(t, db, "admin", "advanced")
	MustCreateRolePermissionByName(t, db, "admin", "pro")
	MustCreateRolePermissionByName(t, db, "admin", "basic")
	MustCreateRolePermissionByName(t, db, "advanced", "advanced")
	MustCreateRolePermissionByName(t, db, "advanced", "pro")
	MustCreateRolePermissionByName(t, db, "advanced", "basic")
	MustCreateRolePermissionByName(t, db, "pro", "pro")
	MustCreateRolePermissionByName(t, db, "pro", "basic")
	MustCreateRolePermissionByName(t, db, "basic", "basic")
}

// CreateRolesAndPermissions creates roles, permissions, and the relations between them based of off a map of role names and slices of permission names
// for example:
//
//	CreateRolesAndPermissions(t, db, map[string][]string{
//		"admin":    {"admin", "advanced", "pro", "basic"},
//		"advanced": {"advanced", "pro", "basic"},
//		"pro":      {"pro", "basic"},
//		"basic":    {"basic"},
//	})
func CreateRolesAndPermissions(t *testing.T, db database.Dbx, rolePermissionsMap map[string][]string) {
	t.Helper()
	for roleName, permissionNames := range rolePermissionsMap {
		role := MustCreateRoleByName(t, db, roleName)
		for _, permissionName := range permissionNames {
			permission := MustFindOrCreatePermissionByName(t, db, permissionName)
			MustCreate(t, RolePermission, db, &models.RolePermission{PermissionID: permission.ID, RoleID: role.ID})
		}
	}
}
