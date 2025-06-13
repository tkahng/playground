package stores_test

import (
	"context"
	"errors"

	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"

	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/test"
)

func TestListRoles(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(tx database.Dbx) error {
		// Create test roles and permissions
		rbacstore := stores.NewDbRBACStore(tx)
		err := rbacstore.EnsureRoleAndPermissions(
			ctx,
			shared.PermissionNameAdmin,
			shared.PermissionNameAdmin,
			shared.PermissionNameBasic,
		)
		if err != nil {
			t.Fatalf("failed to ensure role and permissions: %v", err)
		}
		type args struct {
			ctx   context.Context
			db    database.Dbx
			input *stores.RoleListFilter
		}
		tests := []struct {
			name      string
			args      args
			wantCount int
			wantErr   bool
		}{
			{
				name: "Test List Roles",
				args: args{
					ctx: ctx,
					db:  tx,
					input: &stores.RoleListFilter{
						Names: []string{
							shared.PermissionNameAdmin,
						},
						PaginatedInput: stores.PaginatedInput{
							Page:    0,
							PerPage: 20,
						},
					},
				},
				wantCount: 1,
				wantErr:   false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				store := stores.NewDbRBACStore(tx)
				got, err := store.ListRoles(tt.args.ctx, tt.args.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("ListRoles() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(len(got), tt.wantCount) {
					t.Errorf("ListRoles() = %v, want %v", len(got), tt.wantCount)
				}
			})
		}
		return test.ErrEndTest
	})
}

func TestCountRoles(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(tx database.Dbx) error {
		rbacstore := stores.NewDbRBACStore(tx)
		err := rbacstore.EnsureRoleAndPermissions(
			ctx,
			shared.PermissionNameAdmin,
			shared.PermissionNameAdmin,
			shared.PermissionNameBasic,
		)

		if err != nil {
			t.Fatalf("failed to ensure role and permissions: %v", err)
		}
		err = rbacstore.EnsureRoleAndPermissions(
			ctx,
			shared.PermissionNameBasic,
			shared.PermissionNameBasic,
		)
		if err != nil {
			t.Fatalf("failed to ensure role and permissions: %v", err)
		}
		tests := []struct {
			name    string
			filter  *stores.RoleListFilter
			want    int64
			wantErr bool
		}{
			{
				name: "Count all roles",
				filter: &stores.RoleListFilter{
					Names: []string{
						shared.PermissionNameAdmin,
						shared.PermissionNameBasic,
					},
				},
				want:    2,
				wantErr: false,
			},
			{
				name: "Count filtered roles",
				filter: &stores.RoleListFilter{
					Names: []string{
						shared.PermissionNameAdmin,
					},
				},
				want:    1,
				wantErr: false,
			},
			{
				name:    "Count with nil filter",
				filter:  nil,
				want:    2,
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				store := stores.NewDbRBACStore(tx)
				got, err := store.CountRoles(ctx, tt.filter)
				if (err != nil) != tt.wantErr {
					t.Errorf("CountRoles() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("CountRoles() = %v, want %v", got, tt.want)
				}
			})
		}
		return test.ErrEndTest
	})
}

func TestLoadRolePermissions(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		rbacStore := stores.NewDbRBACStore(dbxx)
		err := rbacStore.EnsureRoleAndPermissions(ctx, "basic", "basic")
		if err != nil {
			t.Fatalf("failed to ensure role and permissions: %v", err)
		}
		role, err := rbacStore.FindOrCreateRole(ctx, "basic")
		if err != nil {
			t.Fatalf("failed to find or create role: %v", err)
		}
		type args struct {
			ctx     context.Context
			roleIds []uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			want    [][]*models.Permission
			wantErr bool
		}{
			{
				name: "basic role",
				args: args{
					ctx:     ctx,
					roleIds: []uuid.UUID{role.ID},
				},
				want: [][]*models.Permission{
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
				got, err := rbacStore.LoadRolePermissions(tt.args.ctx, tt.args.roleIds...)
				if (err != nil) != tt.wantErr {
					t.Errorf("LoadRolePermissions() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got[0][0].Name, tt.want[0][0].Name) {
					t.Errorf("LoadRolePermissions() = %v, want %v", got[0][0].Name, tt.want[0][0].Name)
				}
			})
		}
		return errors.New("rollback")
	})
}

// func TestGetUserPermissions(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		rbacStore := stores.NewDbRBACStore(dbxx)
// 		userStore := stores.NewDbUserStore(dbxx)
// 		permission, err := rbacStore.FindOrCreatePermission(ctx, "basic")
// 		if err != nil {
// 			t.Fatalf("failed to find or create permission: %v", err)
// 		}
// 		user, err := userStore.CreateUser(
// 			ctx,
// 			&models.User{
// 				Email: "test@test.com",
// 			},
// 		)
// 		if err != nil {
// 			return err
// 		}
// 		err = rbacStore.CreateUserPermissions(ctx, user.ID, permission.ID)
// 		if err != nil {
// 			return err
// 		}
// 		type args struct {
// 			ctx     context.Context
// 			db      database.Dbx
// 			userIds []uuid.UUID
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			want    [][]*models.Permission
// 			wantErr bool
// 		}{
// 			{
// 				name: "get user permissions",
// 				args: args{
// 					ctx: ctx,

// 					userIds: []uuid.UUID{user.ID},
// 				},
// 				want: [][]*models.Permission{
// 					{
// 						{
// 							Name: "basic",
// 						},
// 					},
// 				},
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				got, err := rbacStore.GetUserPermissions(tt.args.ctx, tt.args.userIds...)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("GetUserPermissions() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if !reflect.DeepEqual(len(got[0]), len(tt.want[0])) {
// 					t.Errorf("GetUserPermissions() = %v, want %v", len(got[0]), len(tt.want[0]))
// 				}
// 				if !reflect.DeepEqual(got[0][0].Name, tt.want[0][0].Name) {
// 					t.Errorf("GetUserPermissions() = %v, want %v", got[0][0].Name, tt.want[0][0].Name)
// 				}
// 			})
// 		}
// 		return errors.New("rollback")
// 	})
// }

func TestFindOrCreateRole(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		rbacStore := stores.NewDbRBACStore(dbxx)
		type args struct {
			ctx      context.Context
			roleName string
		}
		tests := []struct {
			name    string
			args    args
			want    string
			wantErr bool
		}{
			{
				name: "create new role",
				args: args{
					ctx: ctx,

					roleName: "test_role",
				},
				want:    "test_role",
				wantErr: false,
			},
			{
				name: "find existing role",
				args: args{
					ctx: ctx,

					roleName: "test_role",
				},
				want:    "test_role",
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := rbacStore.FindOrCreateRole(tt.args.ctx, tt.args.roleName)
				if (err != nil) != tt.wantErr {
					t.Errorf("FindOrCreateRole() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got.Name != tt.want {
					t.Errorf("FindOrCreateRole() = %v, want %v", got.Name, tt.want)
				}
			})
		}
		return errors.New("rollback")
	})
}
func TestCreateRole(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		rbacStore := stores.NewDbRBACStore(dbxx)
		type args struct {
			ctx  context.Context
			role *shared.CreateRoleDto
		}
		tests := []struct {
			name    string
			args    args
			want    string
			wantErr bool
		}{
			{
				name: "create role with name only",
				args: args{
					ctx: ctx,
					role: &shared.CreateRoleDto{
						Name: "test_role",
					},
				},
				want:    "test_role",
				wantErr: false,
			},
			{
				name: "create role with description",
				args: args{
					ctx: ctx,

					role: &shared.CreateRoleDto{
						Name:        "test_role_2",
						Description: new(string),
					},
				},
				want:    "test_role_2",
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := rbacStore.CreateRole(tt.args.ctx, tt.args.role)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateRole() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got.Name != tt.want {
					t.Errorf("CreateRole() = %v, want %v", got.Name, tt.want)
				}
			})
		}
		return errors.New("rollback")
	})
}

func TestUpdateRole(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		// Create initial role to update
		rbacStore := stores.NewDbRBACStore(dbxx)
		role, err := rbacStore.CreateRole(ctx, &shared.CreateRoleDto{
			Name: "initial_role",
		})
		if err != nil {
			t.Fatalf("failed to create initial role: %v", err)
		}

		description := "updated description"

		type args struct {
			ctx     context.Context
			id      uuid.UUID
			roledto *shared.UpdateRoleDto
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "update existing role",
				args: args{
					ctx: ctx,

					id: role.ID,
					roledto: &shared.UpdateRoleDto{
						Name:        "updated_role",
						Description: &description,
					},
				},
				wantErr: false,
			},
			{
				name: "update non-existent role",
				args: args{
					ctx: ctx,

					id: uuid.New(),
					roledto: &shared.UpdateRoleDto{
						Name: "test_role",
					},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := rbacStore.UpdateRole(tt.args.ctx, tt.args.id, tt.args.roledto)
				if (err != nil) != tt.wantErr {
					t.Errorf("UpdateRole() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if tt.name == "update existing role" {
					// Verify the update
					updatedRole, err := repository.Role.GetOne(ctx,
						dbxx,
						&map[string]any{
							"id": map[string]any{
								"_eq": tt.args.id.String(),
							},
						})
					if err != nil {
						t.Errorf("Failed to get updated role: %v", err)
						return
					}
					if updatedRole.Name != tt.args.roledto.Name {
						t.Errorf("Role name = %v, want %v", updatedRole.Name, tt.args.roledto.Name)
					}
					if *updatedRole.Description != *tt.args.roledto.Description {
						t.Errorf("Role description = %v, want %v", *updatedRole.Description, *tt.args.roledto.Description)
					}
				}
			})
		}
		return errors.New("rollback")
	})
}

func TestDeleteRole(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		// Create a role to delete
		rbacStore := stores.NewDbRBACStore(dbxx)
		role, err := rbacStore.CreateRole(ctx, &shared.CreateRoleDto{
			Name: "role_to_delete",
		})
		if err != nil {
			t.Fatalf("failed to create test role: %v", err)
		}

		type args struct {
			ctx context.Context
			id  uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "delete existing role",
				args: args{
					ctx: ctx,

					id: role.ID,
				},
				wantErr: false,
			},
			{
				name: "delete non-existent role",
				args: args{
					ctx: ctx,

					id: uuid.New(),
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := rbacStore.DeleteRole(tt.args.ctx, tt.args.id)
				if (err != nil) != tt.wantErr {
					t.Errorf("DeleteRole() error = %v, wantErr %v", err, tt.wantErr)
				}

				if tt.name == "delete existing role" {
					// Verify the role was deleted
					deletedRole, err := repository.Role.GetOne(ctx,
						dbxx,
						&map[string]any{
							"id": map[string]any{
								"_eq": tt.args.id.String(),
							},
						})
					if err != nil {
						t.Errorf("Failed to check deleted role: %v", err)
						return
					}
					if deletedRole != nil {
						t.Errorf("Role still exists after deletion")
					}
				}
			})
		}
		return errors.New("rollback")
	})
}

func TestDeletePermission(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		rbacStore := stores.NewDbRBACStore(dbxx)
		// Create a permission to delete
		permission, err := rbacStore.CreatePermission(ctx, "permission_to_delete", nil)
		if err != nil {
			t.Fatalf("failed to create test permission: %v", err)
		}

		type args struct {
			ctx context.Context
			id  uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "delete existing permission",
				args: args{
					ctx: ctx,

					id: permission.ID,
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := rbacStore.DeletePermission(tt.args.ctx, tt.args.id)
				if (err != nil) != tt.wantErr {
					t.Errorf("DeletePermission() error = %v, wantErr %v", err, tt.wantErr)
				}

				if tt.name == "delete existing permission" {
					// Verify the permission was deleted
					deletedPermission, err := repository.Permission.GetOne(ctx,
						dbxx,
						&map[string]any{
							"id": map[string]any{
								"_eq": tt.args.id.String(),
							},
						})
					if err != nil {
						t.Errorf("Failed to check deleted permission: %v", err)
						return
					}
					if deletedPermission != nil {
						t.Errorf("Permission still exists after deletion")
					}
				}
			})
		}
		return errors.New("rollback")
	})
}
