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

func TestFindPermissionsByIds(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
		// Create test permissions
		rbacstore := stores.NewDbRBACStore(dbxx)
		// userstore := stores.NewPostgresUserStore(dbxx)
		perm1, err := rbacstore.CreatePermission(ctx, "test_perm_1", nil)
		if err != nil {
			t.Fatalf("failed to create test permission 1: %v", err)
		}

		perm2, err := rbacstore.CreatePermission(ctx, "test_perm_2", nil)
		if err != nil {
			t.Fatalf("failed to create test permission 2: %v", err)
		}

		type args struct {
			ctx    context.Context
			dbx    database.Dbx
			params []uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			want    int
			wantErr bool
		}{
			{
				name: "find multiple permissions by ids",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					params: []uuid.UUID{perm1.ID, perm2.ID},
				},
				want:    2,
				wantErr: false,
			},
			{
				name: "find single permission by id",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					params: []uuid.UUID{perm1.ID},
				},
				want:    1,
				wantErr: false,
			},
			{
				name: "find with non-existent ids",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					params: []uuid.UUID{uuid.New()},
				},
				want:    0,
				wantErr: false,
			},
			{
				name: "find with empty id list",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					params: []uuid.UUID{},
				},
				want:    0,
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				store := stores.NewDbRBACStore(dbxx)
				got, err := store.FindPermissionsByIds(tt.args.ctx, tt.args.params)
				if (err != nil) != tt.wantErr {
					t.Errorf("FindPermissionsByIds() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if len(got) != tt.want {
					t.Errorf("FindPermissionsByIds() got %v permissions, want %v", len(got), tt.want)
				}

				// For cases with expected results, verify permissions are returned in ascending name order
				if len(got) > 1 {
					for i := 1; i < len(got); i++ {
						if got[i-1].Name > got[i].Name {
							t.Errorf("FindPermissionsByIds() results not sorted by name in ascending order")
						}
					}
				}
			})
		}
		return errors.New("rollback")
	})
}

func TestListPermissions(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(tx database.Dbx) error {
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
			input *shared.PermissionsListParams
		}
		tests := []struct {
			name      string
			args      args
			wantCount int
			wantErr   bool
		}{
			{
				name: "Test Case 1",
				args: args{
					ctx: ctx,
					db:  tx,
					input: &shared.PermissionsListParams{
						PermissionsListFilter: shared.PermissionsListFilter{
							// Q: "super",
							Names: []string{
								shared.PermissionNameAdmin,
							},
						},
						PaginatedInput: shared.PaginatedInput{
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
				got, err := store.ListPermissions(tt.args.ctx, tt.args.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("ListPermissions() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(len(got), tt.wantCount) {
					t.Errorf("ListPermissions() = %v, want %v", len(got), tt.wantCount)
				}
			})
		}
		return test.ErrEndTest
	})
}

func TestCountPermissions(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(tx database.Dbx) error {
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

		tests := []struct {
			name    string
			filter  *shared.PermissionsListFilter
			want    int64
			wantErr bool
		}{
			{
				name: "Count all permissions",
				filter: &shared.PermissionsListFilter{
					Names: []string{
						shared.PermissionNameAdmin,
						shared.PermissionNameBasic,
					},
				},
				want:    2,
				wantErr: false,
			},
			{
				name: "Count filtered permissions",
				filter: &shared.PermissionsListFilter{
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
				got, err := store.CountPermissions(ctx, tt.filter)
				if (err != nil) != tt.wantErr {
					t.Errorf("CountPermissions() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("CountPermissions() = %v, want %v", got, tt.want)
				}
			})
		}
		return test.ErrEndTest
	})
}

func TestDeleteRolePermissions(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
		// Create a role and permission to test deletion
		rbacStore := stores.NewDbRBACStore(dbxx)
		role, err := rbacStore.CreateRole(ctx, &shared.CreateRoleDto{
			Name: "role_for_permissions",
		})
		if err != nil {
			t.Fatalf("failed to create test role: %v", err)
		}

		permission, err := rbacStore.CreatePermission(
			ctx,
			"test_permission",
			nil,
		)
		if err != nil {
			t.Fatalf("failed to create test permission: %v", err)
		}

		// Create role-permission association
		err = rbacStore.CreateRolePermissions(ctx, role.ID, permission.ID)
		if err != nil {
			t.Fatalf("failed to create role permissions: %v", err)
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
				name: "delete existing role permissions",
				args: args{
					ctx: ctx,

					id: role.ID,
				},
				wantErr: false,
			},
			{
				name: "delete non-existent role permissions",
				args: args{
					ctx: ctx,

					id: uuid.New(),
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := rbacStore.DeleteRolePermissions(tt.args.ctx, tt.args.id, permission.ID)
				if (err != nil) != tt.wantErr {
					t.Errorf("DeleteRolePermissions() error = %v, wantErr %v", err, tt.wantErr)
				}

				if tt.name == "delete existing role permissions" {
					// Verify the role permissions were deleted
					permissions, err := rbacStore.LoadRolePermissions(tt.args.ctx, tt.args.id)
					if err != nil {
						t.Errorf("Failed to check deleted role permissions: %v", err)
						return
					}
					if len(permissions[0]) != 0 {
						t.Errorf("Role permissions still exist after deletion")
					}
				}
			})
		}
		return errors.New("rollback")
	})
}

func TestFindOrCreatePermission(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
		rbacStore := stores.NewDbRBACStore(dbxx)
		type args struct {
			ctx            context.Context
			permissionName string
		}
		tests := []struct {
			name    string
			args    args
			want    string
			wantErr bool
		}{
			{
				name: "create new permission",
				args: args{
					ctx: ctx,

					permissionName: "test_permission",
				},
				want:    "test_permission",
				wantErr: false,
			},
			{
				name: "find existing permission",
				args: args{
					ctx: ctx,

					permissionName: "test_permission",
				},
				want:    "test_permission",
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := rbacStore.FindOrCreatePermission(tt.args.ctx, tt.args.permissionName)
				if (err != nil) != tt.wantErr {
					t.Errorf("FindOrCreatePermission() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got.Name != tt.want {
					t.Errorf("FindOrCreatePermission() = %v, want %v", got.Name, tt.want)
				}
			})
		}
		return errors.New("rollback")
	})
}

func TestCreatePermission(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
		rbacStore := stores.NewDbRBACStore(dbxx)
		description := "test description"

		type args struct {
			ctx         context.Context
			name        string
			description *string
		}
		tests := []struct {
			name    string
			args    args
			want    string
			wantErr bool
		}{
			{
				name: "create permission with name only",
				args: args{
					ctx: ctx,

					name:        "test_permission",
					description: nil,
				},
				want:    "test_permission",
				wantErr: false,
			},
			{
				name: "create permission with description",
				args: args{
					ctx: ctx,

					name:        "test_permission_2",
					description: &description,
				},
				want:    "test_permission_2",
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := rbacStore.CreatePermission(tt.args.ctx, tt.args.name, tt.args.description)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreatePermission() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got.Name != tt.want {
					t.Errorf("CreatePermission() = %v, want %v", got.Name, tt.want)
				}
				if tt.args.description != nil {
					if *got.Description != *tt.args.description {
						t.Errorf("CreatePermission() description = %v, want %v", *got.Description, *tt.args.description)
					}
				}
			})
		}
		return errors.New("rollback")
	})
}
func TestFindPermissionById(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
		rbacStore := stores.NewDbRBACStore(dbxx)
		// Create a test permission to find
		permission, err := rbacStore.CreatePermission(ctx, "test_permission", nil)
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
			want    *models.Permission
			wantErr bool
		}{
			{
				name: "find existing permission",
				args: args{
					ctx: ctx,

					id: permission.ID,
				},
				want:    permission,
				wantErr: false,
			},
			{
				name: "find non-existent permission",
				args: args{
					ctx: ctx,

					id: uuid.New(),
				},
				want:    nil,
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := rbacStore.FindPermissionById(tt.args.ctx, tt.args.id)
				if (err != nil) != tt.wantErr {
					t.Errorf("FindPermissionById() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.want == nil && got != nil {
					t.Errorf("FindPermissionById() = %v, want nil", got)
				} else if tt.want != nil && got == nil {
					t.Errorf("FindPermissionById() = nil, want %v", tt.want)
				} else if tt.want != nil && got != nil && got.ID != tt.want.ID {
					t.Errorf("FindPermissionById() = %v, want %v", got.ID, tt.want.ID)
				}
			})
		}
		return errors.New("rollback")
	})
}

func TestUpdatePermission(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
		// Create initial permission to update
		rbacStore := stores.NewDbRBACStore(dbxx)
		permission, err := rbacStore.CreatePermission(ctx, "initial_permission", nil)
		if err != nil {
			t.Fatalf("failed to create initial permission: %v", err)
		}

		description := "updated description"

		type args struct {
			ctx     context.Context
			id      uuid.UUID
			roledto *shared.UpdatePermissionDto
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "update existing permission",
				args: args{
					ctx: ctx,

					id: permission.ID,
					roledto: &shared.UpdatePermissionDto{
						Name:        "updated_permission",
						Description: &description,
					},
				},
				wantErr: false,
			},
			{
				name: "update non-existent permission",
				args: args{
					ctx: ctx,

					id: uuid.New(),
					roledto: &shared.UpdatePermissionDto{
						Name: "test_permission",
					},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := rbacStore.UpdatePermission(tt.args.ctx, tt.args.id, tt.args.roledto)
				if (err != nil) != tt.wantErr {
					t.Errorf("UpdatePermission() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if tt.name == "update existing permission" {
					// Verify the update
					updatedPermission, err := repository.Permission.GetOne(ctx,
						dbxx,
						&map[string]any{
							"id": map[string]any{
								"_eq": tt.args.id.String(),
							},
						})
					if err != nil {
						t.Errorf("Failed to get updated permission: %v", err)
						return
					}
					if updatedPermission.Name != tt.args.roledto.Name {
						t.Errorf("Permission name = %v, want %v", updatedPermission.Name, tt.args.roledto.Name)
					}
					if *updatedPermission.Description != *tt.args.roledto.Description {
						t.Errorf("Permission description = %v, want %v", *updatedPermission.Description, *tt.args.roledto.Description)
					}
				}
			})
		}
		return errors.New("rollback")
	})
}
func TestCreateRolePermissions(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
		rbacStore := stores.NewDbRBACStore(dbxx)
		role, err := rbacStore.FindOrCreateRole(ctx, "basic")
		if err != nil {
			t.Fatalf("failed to find or create role: %v", err)
		}
		if role == nil {
			t.Fatalf("role should not be nil")
		}
		permission, err := rbacStore.FindOrCreatePermission(ctx, "basic")
		if err != nil {
			t.Fatalf("failed to find or create permission: %v", err)
		}

		type args struct {
			ctx           context.Context
			roleId        uuid.UUID
			permissionIds []uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "create role permission",
				args: args{
					ctx:           ctx,
					roleId:        role.ID,
					permissionIds: []uuid.UUID{permission.ID},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := rbacStore.CreateRolePermissions(tt.args.ctx, tt.args.roleId, tt.args.permissionIds...); (err != nil) != tt.wantErr {
					t.Errorf("CreateRolePermissions() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
		return errors.New("rollback")
	})
}

func TestEnsureRoleAndPermissions(t *testing.T) {
	test.Short(t)
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, dbxx database.Dbx) {
		rbacStore := stores.NewDbRBACStore(dbxx)
		type args struct {
			ctx             context.Context
			roleName        string
			permissionNames []string
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "ensure role and permission",
				args: args{
					ctx: ctx,

					roleName:        "test_role",
					permissionNames: []string{"test_permission"},
				},
				wantErr: false,
			},
			{
				name: "ensure role with multiple permissions",
				args: args{
					ctx:             ctx,
					roleName:        "test_role_2",
					permissionNames: []string{"perm_1", "perm_2", "perm_3"},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := rbacStore.EnsureRoleAndPermissions(tt.args.ctx, tt.args.roleName, tt.args.permissionNames...); (err != nil) != tt.wantErr {
					t.Errorf("EnsureRoleAndPermissions() error = %v, wantErr %v", err, tt.wantErr)
				}

				// Verify role was created
				role, err := rbacStore.FindRoleByName(tt.args.ctx, tt.args.roleName)
				if err != nil {
					t.Errorf("Failed to find created role: %v", err)
				}
				if role.Name != tt.args.roleName {
					t.Errorf("Role name = %v, want %v", role.Name, tt.args.roleName)
				}

				// Verify permissions were created and assigned
				perms, err := rbacStore.LoadRolePermissions(tt.args.ctx, role.ID)
				if err != nil {
					t.Errorf("Failed to load role permissions: %v", err)
				}
				if len(perms[0]) != len(tt.args.permissionNames) {
					t.Errorf("Got %v permissions, want %v", len(perms[0]), len(tt.args.permissionNames))
				}
			})
		}
	})
}
