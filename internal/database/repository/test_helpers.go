package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
)

type RepositoryScenario[T any] interface {
	Get(t testing.TB, opts ...QueryOptionFunc)
	Post(t testing.TB, models []*T)
	Put(t testing.TB, models []*T)
	Delete(t testing.TB, where *map[string]any)
}

// GetScenario is the test environment for the repository get method.
// it should contain
type GetScenario[T any] struct {
	OptFuncs []QueryOptionFunc
	testFunc func(t testing.TB, res []*T)
}

type PostScenario[T any] struct {
	Args     []*T
	testFunc func(t testing.TB, args, res []*T)
}

// func DefaultPostScenarioTestFunc[T any](t testing.TB, args, res []*T, fieldNames ...string) {
// 	t.Helper()
// 	if len(args) != len(res) {
// 		t.Errorf("PostOne() got = %d, want %d", len(args), len(res))
// 	}
// 	for i, arg := range args {
// 		if arg.ID != res[i].ID {
// 			t.Errorf("PostOne() got = %s, want %s", arg.ID, res[i].ID)
// 		}
// 	}
// }

type RepositoryScenarioImpl[T any] struct {
	repo Repository[T]
	db   database.Dbx
}

// func (r RepositoryScenarioImpl[T]) test(t testing.TB) {

// }

func MustCreate[T any](t testing.TB, repo Repository[T], db database.Dbx, arg *T) *T {
	t.Helper()
	res, err := repo.PostOne(t.Context(), db, arg)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func MustFindOne[T any](t testing.TB, repo Repository[T], db database.Dbx, arg *map[string]any) *T {
	t.Helper()
	res, err := repo.GetOne(t.Context(), db, arg)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func MustCountAll[T any](t testing.TB, repo Repository[T], db database.Dbx, arg *map[string]any) int64 {
	t.Helper()
	res, err := repo.Count(t.Context(), db, arg)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func MustFindAll[T any](t testing.TB, repo Repository[T], db database.Dbx, arg *map[string]any) []*T {
	t.Helper()
	res, err := repo.Get(t.Context(), db, arg, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func MustCreateUserAndAccount(t testing.TB, db database.Dbx, fns ...func(*models.User, *models.UserAccount)) (*models.User, *models.UserAccount) {
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

func MustCreateUser(t testing.TB, db database.Dbx, fns ...func(*models.User)) *models.User {
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
func MustCreateAccount(t testing.TB, db database.Dbx, fns ...func(*models.UserAccount)) *models.UserAccount {
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
func MustFindRoleByName(t testing.TB, db database.Dbx, name string) *models.Role {
	t.Helper()
	return MustFindOne(t, Role, db, &map[string]any{"name": map[string]any{"_eq": name}})
}
func MustCreateRoleByName(t testing.TB, db database.Dbx, name string) *models.Role {
	t.Helper()
	return MustCreate(t, Role, db, &models.Role{Name: name})
}

func MustFindPermissionByName(t testing.TB, db database.Dbx, name string) *models.Permission {
	t.Helper()
	return MustFindOne(t, Permission, db, &map[string]any{"name": map[string]any{"_eq": name}})
}
func MustCreatePermissionByName(t testing.TB, db database.Dbx, name string) *models.Permission {
	t.Helper()
	return MustCreate(t, Permission, db, &models.Permission{Name: name})
}

func MustFindOrCreatePermissionByName(t testing.TB, db database.Dbx, name string) *models.Permission {
	t.Helper()
	role := MustFindPermissionByName(t, db, name)
	if role == nil {
		role = MustCreatePermissionByName(t, db, name)
	}
	return role
}

// CreateRolesAndPermissions creates roles, permissions, and the relations between them based of off a map of role names and slices of permission names
// for example:
//
//	CreateRolesAndPermissions(t, db, map[string][]string{
//		"superuser":    {"superuser", "advanced", "pro", "basic"},
//		"advanced": {"advanced", "pro", "basic"},
//		"pro":      {"pro", "basic"},
//		"basic":    {"basic"},
//	})
func CreateRolesAndPermissions(t testing.TB, db database.Dbx, rolePermissionsMap map[string][]string) {
	t.Helper()
	for roleName, permissionNames := range rolePermissionsMap {
		role := MustCreateRoleByName(t, db, roleName)
		for _, permissionName := range permissionNames {
			permission := MustFindOrCreatePermissionByName(t, db, permissionName)
			MustCreate(t, RolePermission, db, &models.RolePermission{PermissionID: permission.ID, RoleID: role.ID})
		}
	}
}
