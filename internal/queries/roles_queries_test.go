package queries_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db"
	crudModels "github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/test"
)

func TestLoadRolePermissions(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		err := queries.EnsureRoleAndPermissions(ctx, dbxx, "basic", "basic")
		if err != nil {
			t.Fatalf("failed to ensure role and permissions: %v", err)
		}
		role, err := queries.FindOrCreateRole(ctx, dbxx, "basic")
		if err != nil {
			t.Fatalf("failed to find or create role: %v", err)
		}
		type args struct {
			ctx     context.Context
			db      db.Dbx
			roleIds []uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			want    [][]*crudModels.Permission
			wantErr bool
		}{
			{
				name: "basic role",
				args: args{
					ctx:     ctx,
					db:      dbxx,
					roleIds: []uuid.UUID{role.ID},
				},
				want: [][]*crudModels.Permission{
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
				got, err := queries.LoadRolePermissions(tt.args.ctx, tt.args.db, tt.args.roleIds...)
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
func TestGetUserRoles(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		err := queries.EnsureRoleAndPermissions(ctx, dbxx, "basic", "basic")
		if err != nil {
			return err
		}

		role, err := queries.FindOrCreateRole(ctx, dbxx, "basic")
		if err != nil {
			return err
		}
		user, err := queries.CreateUser(
			ctx,
			dbxx,
			&shared.AuthenticationInput{
				Email: "test@test.com",
			},
		)
		if err != nil {
			return err
		}

		err = queries.CreateUserRoles(ctx, dbxx, user.ID, role.ID)
		if err != nil {
			return err
		}

		type args struct {
			ctx     context.Context
			db      db.Dbx
			userIds []uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			want    [][]*crudModels.Role
			wantErr bool
		}{
			{
				name: "get user roles",
				args: args{
					ctx:     ctx,
					db:      dbxx,
					userIds: []uuid.UUID{user.ID},
				},
				want: [][]*crudModels.Role{
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
				got, err := queries.GetUserRoles(tt.args.ctx, tt.args.db, tt.args.userIds...)
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
func TestGetUserPermissions(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		permission, err := queries.FindOrCreatePermission(ctx, dbxx, "basic")
		if err != nil {
			t.Fatalf("failed to find or create permission: %v", err)
		}
		user, err := queries.CreateUser(
			ctx,
			dbxx,
			&shared.AuthenticationInput{
				Email: "test@test.com",
			},
		)
		if err != nil {
			return err
		}
		err = queries.CreateUserPermissions(ctx, dbxx, user.ID, permission.ID)
		if err != nil {
			return err
		}
		type args struct {
			ctx     context.Context
			db      db.Dbx
			userIds []uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			want    [][]*crudModels.Permission
			wantErr bool
		}{
			{
				name: "get user permissions",
				args: args{
					ctx:     ctx,
					db:      dbxx,
					userIds: []uuid.UUID{user.ID},
				},
				want: [][]*crudModels.Permission{
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
				got, err := queries.GetUserPermissions(tt.args.ctx, tt.args.db, tt.args.userIds...)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetUserPermissions() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(len(got[0]), len(tt.want[0])) {
					t.Errorf("GetUserPermissions() = %v, want %v", len(got[0]), len(tt.want[0]))
				}
				if !reflect.DeepEqual(got[0][0].Name, tt.want[0][0].Name) {
					t.Errorf("GetUserPermissions() = %v, want %v", got[0][0].Name, tt.want[0][0].Name)
				}
			})
		}
		return errors.New("rollback")
	})
}

func TestCreateRolePermissions(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		role, err := queries.FindOrCreateRole(ctx, dbxx, "basic")
		if err != nil {
			t.Fatalf("failed to find or create role: %v", err)
		}
		permission, err := queries.FindOrCreatePermission(ctx, dbxx, "basic")
		if err != nil {
			t.Fatalf("failed to find or create permission: %v", err)
		}

		type args struct {
			ctx           context.Context
			db            db.Dbx
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
					db:            dbxx,
					roleId:        role.ID,
					permissionIds: []uuid.UUID{permission.ID},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := queries.CreateRolePermissions(tt.args.ctx, tt.args.db, tt.args.roleId, tt.args.permissionIds...); (err != nil) != tt.wantErr {
					t.Errorf("CreateRolePermissions() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
		return errors.New("rollback")
	})
}
func TestCreateProductRoles(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		role, err := queries.FindOrCreateRole(ctx, dbxx, "basic")
		if err != nil {
			t.Fatalf("failed to find or create role: %v", err)
		}
		err = queries.UpsertProduct(ctx, dbxx, &crudModels.StripeProduct{
			ID:          "stripe-product-id",
			Active:      true,
			Name:        "Test Product",
			Description: new(string),
			Image:       new(string),
			Metadata:    map[string]string{},
		})
		if err != nil {
			t.Fatalf("failed to upsert product: %v", err)
		}
		type args struct {
			ctx       context.Context
			db        db.Dbx
			productId string
			roleIds   []uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "create product role",
				args: args{
					ctx:       ctx,
					db:        dbxx,
					productId: "stripe-product-id",
					roleIds:   []uuid.UUID{role.ID},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := queries.CreateProductRoles(tt.args.ctx, tt.args.db, tt.args.productId, tt.args.roleIds...); (err != nil) != tt.wantErr {
					t.Errorf("CreateProductRoles() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
		return errors.New("rollback")
	})
}
func TestEnsureRoleAndPermissions(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		type args struct {
			ctx             context.Context
			db              db.Dbx
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
					ctx:             ctx,
					db:              dbxx,
					roleName:        "test_role",
					permissionNames: []string{"test_permission"},
				},
				wantErr: false,
			},
			{
				name: "ensure role with multiple permissions",
				args: args{
					ctx:             ctx,
					db:              dbxx,
					roleName:        "test_role_2",
					permissionNames: []string{"perm_1", "perm_2", "perm_3"},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := queries.EnsureRoleAndPermissions(tt.args.ctx, tt.args.db, tt.args.roleName, tt.args.permissionNames...); (err != nil) != tt.wantErr {
					t.Errorf("EnsureRoleAndPermissions() error = %v, wantErr %v", err, tt.wantErr)
				}

				// Verify role was created
				role, err := repository.Role.GetOne(ctx, tt.args.db,
					&map[string]any{
						"name": map[string]any{
							"_eq": tt.args.roleName,
						},
					})
				if err != nil {
					t.Errorf("Failed to find created role: %v", err)
				}
				if role.Name != tt.args.roleName {
					t.Errorf("Role name = %v, want %v", role.Name, tt.args.roleName)
				}

				// Verify permissions were created and assigned
				perms, err := queries.LoadRolePermissions(tt.args.ctx, tt.args.db, role.ID)
				if err != nil {
					t.Errorf("Failed to load role permissions: %v", err)
				}
				if len(perms[0]) != len(tt.args.permissionNames) {
					t.Errorf("Got %v permissions, want %v", len(perms[0]), len(tt.args.permissionNames))
				}
			})
		}
		return errors.New("rollback")
	})
}
func TestFindOrCreateRole(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		type args struct {
			ctx      context.Context
			db       db.Dbx
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
					ctx:      ctx,
					db:       dbxx,
					roleName: "test_role",
				},
				want:    "test_role",
				wantErr: false,
			},
			{
				name: "find existing role",
				args: args{
					ctx:      ctx,
					db:       dbxx,
					roleName: "test_role",
				},
				want:    "test_role",
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.FindOrCreateRole(tt.args.ctx, tt.args.db, tt.args.roleName)
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
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		type args struct {
			ctx  context.Context
			dbx  db.Dbx
			role *queries.CreateRoleDto
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
					dbx: dbxx,
					role: &queries.CreateRoleDto{
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
					dbx: dbxx,
					role: &queries.CreateRoleDto{
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
				got, err := queries.CreateRole(tt.args.ctx, tt.args.dbx, tt.args.role)
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
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create initial role to update
		role, err := queries.CreateRole(ctx, dbxx, &queries.CreateRoleDto{
			Name: "initial_role",
		})
		if err != nil {
			t.Fatalf("failed to create initial role: %v", err)
		}

		description := "updated description"

		type args struct {
			ctx     context.Context
			dbx     db.Dbx
			id      uuid.UUID
			roledto *queries.UpdateRoleDto
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
					dbx: dbxx,
					id:  role.ID,
					roledto: &queries.UpdateRoleDto{
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
					dbx: dbxx,
					id:  uuid.New(),
					roledto: &queries.UpdateRoleDto{
						Name: "test_role",
					},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := queries.UpdateRole(tt.args.ctx, tt.args.dbx, tt.args.id, tt.args.roledto)
				if (err != nil) != tt.wantErr {
					t.Errorf("UpdateRole() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if tt.name == "update existing role" {
					// Verify the update
					updatedRole, err := repository.Role.GetOne(ctx, tt.args.dbx,
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
func TestUpdatePermission(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create initial permission to update
		permission, err := queries.CreatePermission(ctx, dbxx, &queries.CreatePermissionDto{
			Name: "initial_permission",
		})
		if err != nil {
			t.Fatalf("failed to create initial permission: %v", err)
		}

		description := "updated description"

		type args struct {
			ctx     context.Context
			dbx     db.Dbx
			id      uuid.UUID
			roledto *queries.UpdatePermissionDto
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
					dbx: dbxx,
					id:  permission.ID,
					roledto: &queries.UpdatePermissionDto{
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
					dbx: dbxx,
					id:  uuid.New(),
					roledto: &queries.UpdatePermissionDto{
						Name: "test_permission",
					},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := queries.UpdatePermission(tt.args.ctx, tt.args.dbx, tt.args.id, tt.args.roledto)
				if (err != nil) != tt.wantErr {
					t.Errorf("UpdatePermission() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if tt.name == "update existing permission" {
					// Verify the update
					updatedPermission, err := repository.Permission.GetOne(ctx, tt.args.dbx,
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
func TestDeleteRole(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create a role to delete
		role, err := queries.CreateRole(ctx, dbxx, &queries.CreateRoleDto{
			Name: "role_to_delete",
		})
		if err != nil {
			t.Fatalf("failed to create test role: %v", err)
		}

		type args struct {
			ctx context.Context
			dbx db.Dbx
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
					dbx: dbxx,
					id:  role.ID,
				},
				wantErr: false,
			},
			{
				name: "delete non-existent role",
				args: args{
					ctx: ctx,
					dbx: dbxx,
					id:  uuid.New(),
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := queries.DeleteRole(tt.args.ctx, tt.args.dbx, tt.args.id)
				if (err != nil) != tt.wantErr {
					t.Errorf("DeleteRole() error = %v, wantErr %v", err, tt.wantErr)
				}

				if tt.name == "delete existing role" {
					// Verify the role was deleted
					deletedRole, err := repository.Role.GetOne(ctx, tt.args.dbx,
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
func TestDeleteRolePermissions(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create a role and permission to test deletion
		role, err := queries.CreateRole(ctx, dbxx, &queries.CreateRoleDto{
			Name: "role_for_permissions",
		})
		if err != nil {
			t.Fatalf("failed to create test role: %v", err)
		}

		permission, err := queries.CreatePermission(ctx, dbxx, &queries.CreatePermissionDto{
			Name: "test_permission",
		})
		if err != nil {
			t.Fatalf("failed to create test permission: %v", err)
		}

		// Create role-permission association
		err = queries.CreateRolePermissions(ctx, dbxx, role.ID, permission.ID)
		if err != nil {
			t.Fatalf("failed to create role permissions: %v", err)
		}

		type args struct {
			ctx context.Context
			dbx db.Dbx
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
					dbx: dbxx,
					id:  role.ID,
				},
				wantErr: false,
			},
			{
				name: "delete non-existent role permissions",
				args: args{
					ctx: ctx,
					dbx: dbxx,
					id:  uuid.New(),
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := queries.DeleteRolePermissions(tt.args.ctx, tt.args.dbx, tt.args.id)
				if (err != nil) != tt.wantErr {
					t.Errorf("DeleteRolePermissions() error = %v, wantErr %v", err, tt.wantErr)
				}

				if tt.name == "delete existing role permissions" {
					// Verify the role permissions were deleted
					permissions, err := queries.LoadRolePermissions(tt.args.ctx, tt.args.dbx, tt.args.id)
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
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		type args struct {
			ctx            context.Context
			db             db.Dbx
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
					ctx:            ctx,
					db:             dbxx,
					permissionName: "test_permission",
				},
				want:    "test_permission",
				wantErr: false,
			},
			{
				name: "find existing permission",
				args: args{
					ctx:            ctx,
					db:             dbxx,
					permissionName: "test_permission",
				},
				want:    "test_permission",
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.FindOrCreatePermission(tt.args.ctx, tt.args.db, tt.args.permissionName)
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
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		description := "test description"

		type args struct {
			ctx        context.Context
			dbx        db.Dbx
			permission *queries.CreatePermissionDto
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
					dbx: dbxx,
					permission: &queries.CreatePermissionDto{
						Name: "test_permission",
					},
				},
				want:    "test_permission",
				wantErr: false,
			},
			{
				name: "create permission with description",
				args: args{
					ctx: ctx,
					dbx: dbxx,
					permission: &queries.CreatePermissionDto{
						Name:        "test_permission_2",
						Description: &description,
					},
				},
				want:    "test_permission_2",
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.CreatePermission(tt.args.ctx, tt.args.dbx, tt.args.permission)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreatePermission() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got.Name != tt.want {
					t.Errorf("CreatePermission() = %v, want %v", got.Name, tt.want)
				}
				if tt.args.permission.Description != nil {
					if *got.Description != *tt.args.permission.Description {
						t.Errorf("CreatePermission() description = %v, want %v", *got.Description, *tt.args.permission.Description)
					}
				}
			})
		}
		return errors.New("rollback")
	})
}
func TestFindPermissionsByIds(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create test permissions
		perm1, err := queries.CreatePermission(ctx, dbxx, &queries.CreatePermissionDto{
			Name: "test_perm_1",
		})
		if err != nil {
			t.Fatalf("failed to create test permission 1: %v", err)
		}

		perm2, err := queries.CreatePermission(ctx, dbxx, &queries.CreatePermissionDto{
			Name: "test_perm_2",
		})
		if err != nil {
			t.Fatalf("failed to create test permission 2: %v", err)
		}

		type args struct {
			ctx    context.Context
			dbx    db.Dbx
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
				got, err := queries.FindPermissionsByIds(tt.args.ctx, tt.args.dbx, tt.args.params)
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
func TestDeletePermission(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create a permission to delete
		permission, err := queries.CreatePermission(ctx, dbxx, &queries.CreatePermissionDto{
			Name: "permission_to_delete",
		})
		if err != nil {
			t.Fatalf("failed to create test permission: %v", err)
		}

		type args struct {
			ctx context.Context
			dbx db.Dbx
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
					dbx: dbxx,
					id:  permission.ID,
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := queries.DeletePermission(tt.args.ctx, tt.args.dbx, tt.args.id)
				if (err != nil) != tt.wantErr {
					t.Errorf("DeletePermission() error = %v, wantErr %v", err, tt.wantErr)
				}

				if tt.name == "delete existing permission" {
					// Verify the permission was deleted
					deletedPermission, err := repository.Permission.GetOne(ctx, tt.args.dbx,
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
func TestFindPermissionById(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create a test permission to find
		permission, err := queries.CreatePermission(ctx, dbxx, &queries.CreatePermissionDto{
			Name: "test_permission",
		})
		if err != nil {
			t.Fatalf("failed to create test permission: %v", err)
		}

		type args struct {
			ctx context.Context
			dbx db.Dbx
			id  uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			want    *crudModels.Permission
			wantErr bool
		}{
			{
				name: "find existing permission",
				args: args{
					ctx: ctx,
					dbx: dbxx,
					id:  permission.ID,
				},
				want:    permission,
				wantErr: false,
			},
			{
				name: "find non-existent permission",
				args: args{
					ctx: ctx,
					dbx: dbxx,
					id:  uuid.New(),
				},
				want:    nil,
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.FindPermissionById(tt.args.ctx, tt.args.dbx, tt.args.id)
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
