package repository_test

import (
	"context"
	"testing"

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
