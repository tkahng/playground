package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
)

func MustCreate[T any](t *testing.T, repo Repository[T], db database.Dbx, arg *T) *T {
	t.Helper()
	res, err := repo.PostOne(context.Background(), db, arg)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func MustFindOne[T any](t *testing.T, repo Repository[T], db database.Dbx, arg *map[string]any) *T {
	t.Helper()
	res, err := repo.GetOne(context.Background(), db, arg)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func MustFindAll[T any](t *testing.T, repo Repository[T], db database.Dbx, arg *map[string]any) []*T {
	t.Helper()
	res, err := repo.Get(context.Background(), db, arg, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	return res
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
	return MustFindOne(t, Role, db, &map[string]any{"name": name})
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
	return MustFindOne(t, Permission, db, &map[string]any{"name": name})
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
	role := MustFindRoleByName(t, db, roleName)
	permission := MustFindPermissionByName(t, db, permissionName)
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
