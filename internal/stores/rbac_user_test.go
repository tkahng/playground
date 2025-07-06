package stores_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/test"
)

func TestListUserPermissionsSource(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		userStore := stores.NewDbUserStore(dbxx)
		rbacStore := stores.NewDbRBACStore(dbxx)
		// Create test user
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "test@test.com",
		})
		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}

		// Create test role and permission
		role, err := rbacStore.CreateRole(ctx, &stores.CreateRoleDto{
			Name: "test_role",
		})
		if err != nil {
			t.Fatalf("failed to create test role: %v", err)
		}

		perm, err := rbacStore.CreatePermission(ctx, "test_permission", nil)
		if err != nil {
			t.Fatalf("failed to create test permission: %v", err)
		}

		// Assign role to user and permission to role
		err = rbacStore.CreateUserRoles(ctx, user.ID, role.ID)
		if err != nil {
			t.Fatalf("failed to assign role to user: %v", err)
		}

		err = rbacStore.CreateRolePermissions(ctx, role.ID, perm.ID)
		if err != nil {
			t.Fatalf("failed to assign permission to role: %v", err)
		}

		type args struct {
			ctx    context.Context
			dbx    database.Dbx
			userId uuid.UUID
			limit  int64
			offset int64
		}
		tests := []struct {
			name    string
			args    args
			want    int
			wantErr bool
		}{
			{
				name: "list user permissions with valid user",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					userId: user.ID,
					limit:  10,
					offset: 0,
				},
				want:    1,
				wantErr: false,
			},
			{
				name: "list permissions for non-existent user",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					userId: uuid.New(),
					limit:  10,
					offset: 0,
				},
				want:    0,
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Create a new instance of the PostgresRBACStore
				store := stores.NewDbRBACStore(dbxx)
				got, err := store.ListUserPermissionsSource(tt.args.ctx, tt.args.userId, tt.args.limit, tt.args.offset)
				if (err != nil) != tt.wantErr {
					t.Errorf("ListUserPermissionsSource() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if len(got) != tt.want {
					t.Errorf("ListUserPermissionsSource() got %v permissions, want %v", len(got), tt.want)
				}
				if tt.want > 0 {
					if got[0].Name != "test_permission" {
						t.Errorf("ListUserPermissionsSource() got permission name = %v, want %v", got[0].Name, "test_permission")
					}
				}
			})
		}
		return errors.New("rollback")
	})
}

func TestCountUserPermissionSource(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		rbacstore := stores.NewDbRBACStore(dbxx)
		userstore := stores.NewDbUserStore(dbxx)
		// Create test user
		user, err := userstore.CreateUser(ctx, &models.User{
			Email: "test@test.com",
		})
		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}

		// Create test role and permission
		role, err := rbacstore.CreateRole(ctx, &stores.CreateRoleDto{
			Name: "test_role",
		})
		if err != nil {
			t.Fatalf("failed to create test role: %v", err)
		}

		perm, err := rbacstore.CreatePermission(ctx, "test_permission", nil)
		if err != nil {
			t.Fatalf("failed to create test permission: %v", err)
		}

		// Assign role to user and permission to role
		err = rbacstore.CreateUserRoles(ctx, user.ID, role.ID)
		if err != nil {
			t.Fatalf("failed to assign role to user: %v", err)
		}

		err = rbacstore.CreateRolePermissions(ctx, role.ID, perm.ID)
		if err != nil {
			t.Fatalf("failed to assign permission to role: %v", err)
		}

		type args struct {
			ctx    context.Context
			dbx    database.Dbx
			userId uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			want    int64
			wantErr bool
		}{
			{
				name: "count permissions for user with permissions",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					userId: user.ID,
				},
				want:    1,
				wantErr: false,
			},
			{
				name: "count permissions for non-existent user",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					userId: uuid.New(),
				},
				want:    0,
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				store := stores.NewDbRBACStore(dbxx)
				got, err := store.CountUserPermissionSource(tt.args.ctx, tt.args.userId)
				if (err != nil) != tt.wantErr {
					t.Errorf("CountUserPermissionSource() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("CountUserPermissionSource() = %v, want %v", got, tt.want)
				}
			})
		}
		return errors.New("rollback")
	})
}
func TestListUserNotPermissionsSource(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		// Create test user
		userstore := stores.NewDbUserStore(dbxx)
		rbacstore := stores.NewDbRBACStore(dbxx)
		user, err := userstore.CreateUser(ctx, &models.User{
			Email: "test@test.com",
		})
		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}

		// Create test role and permission that will be assigned
		role, err := rbacstore.CreateRole(ctx, &stores.CreateRoleDto{
			Name: "test_role",
		})
		if err != nil {
			t.Fatalf("failed to create test role: %v", err)
		}

		assignedPerm, err := rbacstore.CreatePermission(ctx, "assigned_permission", nil)
		if err != nil {
			t.Fatalf("failed to create assigned permission: %v", err)
		}

		// Create an unassigned permission
		unassignedPerm, err := rbacstore.CreatePermission(ctx, "unassigned_permission", nil)
		if err != nil {
			t.Fatalf("failed to create unassigned permission: %v", err)
		}

		// Assign role to user and permission to role
		err = rbacstore.CreateUserRoles(ctx, user.ID, role.ID)
		if err != nil {
			t.Fatalf("failed to assign role to user: %v", err)
		}

		err = rbacstore.CreateRolePermissions(ctx, role.ID, assignedPerm.ID)
		if err != nil {
			t.Fatalf("failed to assign permission to role: %v", err)
		}

		type args struct {
			ctx    context.Context
			dbx    database.Dbx
			userId uuid.UUID
			limit  int64
			offset int64
		}
		tests := []struct {
			name    string
			args    args
			want    int
			wantErr bool
		}{
			{
				name: "list unassigned permissions for user",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					userId: user.ID,
					limit:  10,
					offset: 0,
				},
				want:    1,
				wantErr: false,
			},
			{
				name: "list with non-existent user",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					userId: uuid.New(),
					limit:  10,
					offset: 0,
				},
				want:    2,
				wantErr: false,
			},
			{
				name: "list with zero limit",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					userId: user.ID,
					limit:  0,
					offset: 0,
				},
				want:    0,
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				store := stores.NewDbRBACStore(dbxx)
				got, err := store.ListUserNotPermissionsSource(tt.args.ctx, tt.args.userId, tt.args.limit, tt.args.offset)
				if (err != nil) != tt.wantErr {
					t.Errorf("ListUserNotPermissionsSource() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if len(got) != tt.want {
					t.Errorf("ListUserNotPermissionsSource() got %v permissions, want %v", len(got), tt.want)
				}
				if tt.name == "list unassigned permissions for user" && len(got) > 0 {
					if got[0].Name != unassignedPerm.Name {
						t.Errorf("ListUserNotPermissionsSource() got permission name = %v, want %v", got[0].Name, unassignedPerm.Name)
					}
				}
			})
		}
		return errors.New("rollback")
	})
}
func TestCountNotUserPermissionSource(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		// Create test user
		userstore := stores.NewDbUserStore(dbxx)
		rbacstore := stores.NewDbRBACStore(dbxx)
		user, err := userstore.CreateUser(ctx, &models.User{
			Email: "test@test.com",
		})
		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}

		// Create test role and permission that will be assigned
		role, err := rbacstore.CreateRole(ctx, &stores.CreateRoleDto{
			Name: "test_role",
		})
		if err != nil {
			t.Fatalf("failed to create test role: %v", err)
		}

		assignedPerm, err := rbacstore.CreatePermission(ctx, "assigned_permission", nil)
		if err != nil {
			t.Fatalf("failed to create assigned permission: %v", err)
		}

		// Create an unassigned permission
		_, err = rbacstore.CreatePermission(ctx, "unassigned_permission", nil)
		if err != nil {
			t.Fatalf("failed to create unassigned permission: %v", err)
		}

		// Assign role to user and permission to role
		err = rbacstore.CreateUserRoles(ctx, user.ID, role.ID)
		if err != nil {
			t.Fatalf("failed to assign role to user: %v", err)
		}

		err = rbacstore.CreateRolePermissions(ctx, role.ID, assignedPerm.ID)
		if err != nil {
			t.Fatalf("failed to assign permission to role: %v", err)
		}

		type args struct {
			ctx    context.Context
			dbx    database.Dbx
			userId uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			want    int64
			wantErr bool
		}{
			{
				name: "count unassigned permissions for user",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					userId: user.ID,
				},
				want:    1,
				wantErr: false,
			},
			{
				name: "count with non-existent user",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					userId: uuid.New(),
				},
				want:    2,
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				store := stores.NewDbRBACStore(dbxx)
				got, err := store.CountNotUserPermissionSource(tt.args.ctx, tt.args.userId)
				if (err != nil) != tt.wantErr {
					t.Errorf("CountNotUserPermissionSource() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("CountNotUserPermissionSource() = %v, want %v", got, tt.want)
				}
			})
		}
		return errors.New("rollback")
	})
}

func TestCreateUserRoles(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		userStore := stores.NewDbUserStore(dbxx)
		rbacStore := stores.NewDbRBACStore(dbxx)
		// Create a user
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Errorf("failed to create user: %v", err)
			return err
		}
		role, err := rbacStore.FindOrCreateRole(ctx, "basic")
		if err != nil {
			t.Errorf("failed to create role: %v", err)
			return err
		}
		type args struct {
			ctx     context.Context
			db      database.Dbx
			userId  uuid.UUID
			roleIds []uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "create user roles",
				args: args{
					ctx:     ctx,
					db:      dbxx,
					userId:  user.ID,
					roleIds: []uuid.UUID{role.ID},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := rbacStore.CreateUserRoles(tt.args.ctx, tt.args.userId, tt.args.roleIds...); (err != nil) != tt.wantErr {
					t.Errorf("createUserRoles() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
		return test.ErrEndTest
	})
}

func TestGetUserRoles(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		rbacStore := stores.NewDbRBACStore(dbxx)
		userStore := stores.NewDbUserStore(dbxx)
		err := rbacStore.EnsureRoleAndPermissions(ctx, "basic", "basic")
		if err != nil {
			return err
		}

		role, err := rbacStore.FindOrCreateRole(ctx, "basic")
		if err != nil {
			return err
		}
		user, err := userStore.CreateUser(
			ctx,
			&models.User{
				Email: "test@test.com",
			},
		)
		if err != nil {
			return err
		}

		err = rbacStore.CreateUserRoles(ctx, user.ID, role.ID)
		if err != nil {
			return err
		}

		type args struct {
			ctx     context.Context
			userIds []uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			want    [][]*models.Role
			wantErr bool
		}{
			{
				name: "get user roles",
				args: args{
					ctx:     ctx,
					userIds: []uuid.UUID{user.ID},
				},
				want: [][]*models.Role{
					{
						{
							Name: "basic",
						},
					},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := rbacStore.GetUserRoles(tt.args.ctx, tt.args.userIds...)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetUserRoles() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got[0][0].Name, tt.want[0][0].Name) {
					t.Errorf("GetUserRoles() = %v, want %v", got[0][0].Name, tt.want[0][0].Name)
				}
			})
		}
		return errors.New("rollback")
	})
}
