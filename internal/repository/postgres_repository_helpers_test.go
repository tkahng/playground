package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/repository"
)

func MustCreate[T any](t *testing.T, repo repository.Repository[T], db database.Dbx, arg *T) *T {
	t.Helper()
	res, err := repo.PostOne(context.Background(), db, arg)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func MustFindOne[T any](t *testing.T, repo repository.Repository[T], db database.Dbx, arg *map[string]any) *T {
	t.Helper()
	res, err := repo.GetOne(context.Background(), db, arg)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func MustFindAll[T any](t *testing.T, repo repository.Repository[T], db database.Dbx, arg *map[string]any) []*T {
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
	return MustCreate(t, repository.User, db, user)
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
	return MustCreate(t, repository.UserAccount, db, res)
}
func MustFindRoleByName(t *testing.T, db database.Dbx, name string) *models.Role {
	t.Helper()
	return MustFindOne(t, repository.Role, db, &map[string]any{"name": name})
}
func MustCreateUserRoleByName(t *testing.T, db database.Dbx, userId uuid.UUID, name string) *models.UserRole {
	t.Helper()
	role := MustFindRoleByName(t, db, name)
	return MustCreate(t, repository.UserRole, db, &models.UserRole{UserID: userId, RoleID: role.ID})
}
func MustFindPermissionByName(t *testing.T, db database.Dbx, name string) *models.Permission {
	t.Helper()
	return MustFindOne(t, repository.Permission, db, &map[string]any{"name": name})
}
func MustCreateRolePermissionByName(t *testing.T, db database.Dbx, roleName, permissionName string) *models.RolePermission {
	t.Helper()
	role := MustFindRoleByName(t, db, roleName)
	permission := MustFindPermissionByName(t, db, permissionName)
	return MustCreate(t, repository.RolePermission, db, &models.RolePermission{PermissionID: permission.ID, RoleID: role.ID})
}

func InitRbac(t *testing.T, db database.Dbx) {
	t.Helper()
	var adminRole, advRole, proRole, basicRole *models.Role
	adminRole = MustCreate(t, repository.Role, db, &models.Role{Name: "admin"})
	advRole = MustCreate(t, repository.Role, db, &models.Role{Name: "advanced"})
	proRole = MustCreate(t, repository.Role, db, &models.Role{Name: "pro"})
	basicRole = MustCreate(t, repository.Role, db, &models.Role{Name: "basic"})

	var adminPermission, advPermission, proPermission, basicPermission *models.Permission
	adminPermission = MustCreate(t, repository.Permission, db, &models.Permission{Name: "admin"})
	advPermission = MustCreate(t, repository.Permission, db, &models.Permission{Name: "advanced"})
	proPermission = MustCreate(t, repository.Permission, db, &models.Permission{Name: "pro"})
	basicPermission = MustCreate(t, repository.Permission, db, &models.Permission{Name: "basic"})

	MustCreate(t, repository.RolePermission, db, &models.RolePermission{RoleID: adminRole.ID, PermissionID: adminPermission.ID})
	MustCreate(t, repository.RolePermission, db, &models.RolePermission{RoleID: adminRole.ID, PermissionID: advPermission.ID})
	MustCreate(t, repository.RolePermission, db, &models.RolePermission{RoleID: adminRole.ID, PermissionID: proPermission.ID})
	MustCreate(t, repository.RolePermission, db, &models.RolePermission{RoleID: adminRole.ID, PermissionID: basicPermission.ID})
	MustCreate(t, repository.RolePermission, db, &models.RolePermission{RoleID: advRole.ID, PermissionID: advPermission.ID})
	MustCreate(t, repository.RolePermission, db, &models.RolePermission{RoleID: advRole.ID, PermissionID: proPermission.ID})
	MustCreate(t, repository.RolePermission, db, &models.RolePermission{RoleID: advRole.ID, PermissionID: basicPermission.ID})
	MustCreate(t, repository.RolePermission, db, &models.RolePermission{RoleID: proRole.ID, PermissionID: proPermission.ID})
	MustCreate(t, repository.RolePermission, db, &models.RolePermission{RoleID: proRole.ID, PermissionID: basicPermission.ID})
	MustCreate(t, repository.RolePermission, db, &models.RolePermission{RoleID: basicRole.ID, PermissionID: basicPermission.ID})
}
