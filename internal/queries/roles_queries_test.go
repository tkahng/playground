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
