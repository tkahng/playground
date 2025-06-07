package stores_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/test"
)

func TestListPermissions(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(tx database.Dbx) error {
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
		return test.EndTestErr
	})
}

func TestCountPermissions(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(tx database.Dbx) error {
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
		return test.EndTestErr
	})
}

func TestListRoles(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(tx database.Dbx) error {
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
			input *shared.RolesListParams
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
					input: &shared.RolesListParams{
						RoleListFilter: shared.RoleListFilter{
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
		return test.EndTestErr
	})
}

func TestCountRoles(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(tx database.Dbx) error {
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
			filter  *shared.RoleListFilter
			want    int64
			wantErr bool
		}{
			{
				name: "Count all roles",
				filter: &shared.RoleListFilter{
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
				filter: &shared.RoleListFilter{
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
		return test.EndTestErr
	})
}

func TestFindPermissionsByIds(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
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

func TestListUserPermissionsSource(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
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
		role, err := rbacStore.CreateRole(ctx, &shared.CreateRoleDto{
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
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
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
		role, err := rbacstore.CreateRole(ctx, &shared.CreateRoleDto{
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
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
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
		role, err := rbacstore.CreateRole(ctx, &shared.CreateRoleDto{
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
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
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
		role, err := rbacstore.CreateRole(ctx, &shared.CreateRoleDto{
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

func TestCreateProductPermissions(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		// rbacStore := stores.NewPostgresRBACStore(dbxx)
		paymentStore := stores.NewDbPaymentStore(dbxx)
		permission, err := paymentStore.FindOrCreatePermission(ctx, "basic")
		if err != nil {
			t.Fatalf("failed to find or create permission: %v", err)
		}
		err = paymentStore.UpsertProduct(ctx, &models.StripeProduct{
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
			ctx           context.Context
			db            database.Dbx
			productId     string
			permissionIds []uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "create product permission",
				args: args{
					ctx:           ctx,
					db:            dbxx,
					productId:     "stripe-product-id",
					permissionIds: []uuid.UUID{permission.ID},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := paymentStore.CreateProductPermissions(tt.args.ctx, tt.args.productId, tt.args.permissionIds...); (err != nil) != tt.wantErr {
					t.Errorf("CreateProductPermissions() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
		return errors.New("rollback")
	})
}

func TestCreateUserRoles(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
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
		return test.EndTestErr
	})
}

// func TestLoadRolePermissions(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
// 		err := queries.EnsureRoleAndPermissions(ctx, dbxx, "basic", "basic")
// 		if err != nil {
// 			t.Fatalf("failed to ensure role and permissions: %v", err)
// 		}
// 		role, err := queries.FindOrCreateRole(ctx, dbxx, "basic")
// 		if err != nil {
// 			t.Fatalf("failed to find or create role: %v", err)
// 		}
// 		type args struct {
// 			ctx     context.Context
// 			db      database.Dbx
// 			roleIds []uuid.UUID
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			want    [][]*crudModels.Permission
// 			wantErr bool
// 		}{
// 			{
// 				name: "basic role",
// 				args: args{
// 					ctx:     ctx,
// 					db:      dbxx,
// 					roleIds: []uuid.UUID{role.ID},
// 				},
// 				want: [][]*crudModels.Permission{
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
// 				got, err := queries.LoadRolePermissions(tt.args.ctx, tt.args.db, tt.args.roleIds...)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("LoadRolePermissions() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if !reflect.DeepEqual(got[0][0].Name, tt.want[0][0].Name) {
// 					t.Errorf("LoadRolePermissions() = %v, want %v", got[0][0].Name, tt.want[0][0].Name)
// 				}
// 			})
// 		}
// 		return errors.New("rollback")
// 	})
// }
// func TestGetUserRoles(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
// 		err := queries.EnsureRoleAndPermissions(ctx, dbxx, "basic", "basic")
// 		if err != nil {
// 			return err
// 		}

// 		role, err := queries.FindOrCreateRole(ctx, dbxx, "basic")
// 		if err != nil {
// 			return err
// 		}
// 		user, err := queries.CreateUser(
// 			ctx,
// 			dbxx,
// 			&shared.AuthenticationInput{
// 				Email: "test@test.com",
// 			},
// 		)
// 		if err != nil {
// 			return err
// 		}

// 		err = queries.CreateUserRoles(ctx, dbxx, user.ID, role.ID)
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
// 			want    [][]*crudModels.Role
// 			wantErr bool
// 		}{
// 			{
// 				name: "get user roles",
// 				args: args{
// 					ctx:     ctx,
// 					db:      dbxx,
// 					userIds: []uuid.UUID{user.ID},
// 				},
// 				want: [][]*crudModels.Role{
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
// 				got, err := queries.GetUserRoles(tt.args.ctx, tt.args.db, tt.args.userIds...)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("GetUserRoles() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if !reflect.DeepEqual(got[0][0].Name, tt.want[0][0].Name) {
// 					t.Errorf("GetUserRoles() = %v, want %v", got[0][0].Name, tt.want[0][0].Name)
// 				}
// 			})
// 		}
// 		return errors.New("rollback")
// 	})
// }
// func TestGetUserPermissions(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
// 		permission, err := queries.FindOrCreatePermission(ctx, dbxx, "basic")
// 		if err != nil {
// 			t.Fatalf("failed to find or create permission: %v", err)
// 		}
// 		user, err := queries.CreateUser(
// 			ctx,
// 			dbxx,
// 			&shared.AuthenticationInput{
// 				Email: "test@test.com",
// 			},
// 		)
// 		if err != nil {
// 			return err
// 		}
// 		err = queries.CreateUserPermissions(ctx, dbxx, user.ID, permission.ID)
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
// 			want    [][]*crudModels.Permission
// 			wantErr bool
// 		}{
// 			{
// 				name: "get user permissions",
// 				args: args{
// 					ctx:     ctx,
// 					db:      dbxx,
// 					userIds: []uuid.UUID{user.ID},
// 				},
// 				want: [][]*crudModels.Permission{
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
// 				got, err := queries.GetUserPermissions(tt.args.ctx, tt.args.db, tt.args.userIds...)
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

// func TestCreateRolePermissions(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
// 		role, err := queries.FindOrCreateRole(ctx, dbxx, "basic")
// 		if err != nil {
// 			t.Fatalf("failed to find or create role: %v", err)
// 		}
// 		permission, err := queries.FindOrCreatePermission(ctx, dbxx, "basic")
// 		if err != nil {
// 			t.Fatalf("failed to find or create permission: %v", err)
// 		}

// 		type args struct {
// 			ctx           context.Context
// 			db            database.Dbx
// 			roleId        uuid.UUID
// 			permissionIds []uuid.UUID
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			wantErr bool
// 		}{
// 			{
// 				name: "create role permission",
// 				args: args{
// 					ctx:           ctx,
// 					db:            dbxx,
// 					roleId:        role.ID,
// 					permissionIds: []uuid.UUID{permission.ID},
// 				},
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				if err := queries.CreateRolePermissions(tt.args.ctx, tt.args.db, tt.args.roleId, tt.args.permissionIds...); (err != nil) != tt.wantErr {
// 					t.Errorf("CreateRolePermissions() error = %v, wantErr %v", err, tt.wantErr)
// 				}
// 			})
// 		}
// 		return errors.New("rollback")
// 	})
// }
// func TestCreateProductRoles(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
// 		role, err := queries.FindOrCreateRole(ctx, dbxx, "basic")
// 		if err != nil {
// 			t.Fatalf("failed to find or create role: %v", err)
// 		}
// 		err = queries.UpsertProduct(ctx, dbxx, &crudModels.StripeProduct{
// 			ID:          "stripe-product-id",
// 			Active:      true,
// 			Name:        "Test Product",
// 			Description: new(string),
// 			Image:       new(string),
// 			Metadata:    map[string]string{},
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to upsert product: %v", err)
// 		}
// 		type args struct {
// 			ctx       context.Context
// 			db        database.Dbx
// 			productId string
// 			roleIds   []uuid.UUID
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			wantErr bool
// 		}{
// 			{
// 				name: "create product role",
// 				args: args{
// 					ctx:       ctx,
// 					db:        dbxx,
// 					productId: "stripe-product-id",
// 					roleIds:   []uuid.UUID{role.ID},
// 				},
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				if err := queries.CreateProductRoles(tt.args.ctx, tt.args.db, tt.args.productId, tt.args.roleIds...); (err != nil) != tt.wantErr {
// 					t.Errorf("CreateProductRoles() error = %v, wantErr %v", err, tt.wantErr)
// 				}
// 			})
// 		}
// 		return errors.New("rollback")
// 	})
// }
// func TestEnsureRoleAndPermissions(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
// 		type args struct {
// 			ctx             context.Context
// 			db              database.Dbx
// 			roleName        string
// 			permissionNames []string
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			wantErr bool
// 		}{
// 			{
// 				name: "ensure role and permission",
// 				args: args{
// 					ctx:             ctx,
// 					db:              dbxx,
// 					roleName:        "test_role",
// 					permissionNames: []string{"test_permission"},
// 				},
// 				wantErr: false,
// 			},
// 			{
// 				name: "ensure role with multiple permissions",
// 				args: args{
// 					ctx:             ctx,
// 					db:              dbxx,
// 					roleName:        "test_role_2",
// 					permissionNames: []string{"perm_1", "perm_2", "perm_3"},
// 				},
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				if err := queries.EnsureRoleAndPermissions(tt.args.ctx, tt.args.db, tt.args.roleName, tt.args.permissionNames...); (err != nil) != tt.wantErr {
// 					t.Errorf("EnsureRoleAndPermissions() error = %v, wantErr %v", err, tt.wantErr)
// 				}

// 				// Verify role was created
// 				role, err := crudrepo.Role.GetOne(ctx, tt.args.db,
// 					&map[string]any{
// 						"name": map[string]any{
// 							"_eq": tt.args.roleName,
// 						},
// 					})
// 				if err != nil {
// 					t.Errorf("Failed to find created role: %v", err)
// 				}
// 				if role.Name != tt.args.roleName {
// 					t.Errorf("Role name = %v, want %v", role.Name, tt.args.roleName)
// 				}

// 				// Verify permissions were created and assigned
// 				perms, err := queries.LoadRolePermissions(tt.args.ctx, tt.args.db, role.ID)
// 				if err != nil {
// 					t.Errorf("Failed to load role permissions: %v", err)
// 				}
// 				if len(perms[0]) != len(tt.args.permissionNames) {
// 					t.Errorf("Got %v permissions, want %v", len(perms[0]), len(tt.args.permissionNames))
// 				}
// 			})
// 		}
// 		return errors.New("rollback")
// 	})
// }
// func TestFindOrCreateRole(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
// 		type args struct {
// 			ctx      context.Context
// 			db       database.Dbx
// 			roleName string
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			want    string
// 			wantErr bool
// 		}{
// 			{
// 				name: "create new role",
// 				args: args{
// 					ctx:      ctx,
// 					db:       dbxx,
// 					roleName: "test_role",
// 				},
// 				want:    "test_role",
// 				wantErr: false,
// 			},
// 			{
// 				name: "find existing role",
// 				args: args{
// 					ctx:      ctx,
// 					db:       dbxx,
// 					roleName: "test_role",
// 				},
// 				want:    "test_role",
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				got, err := queries.FindOrCreateRole(tt.args.ctx, tt.args.db, tt.args.roleName)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("FindOrCreateRole() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if got.Name != tt.want {
// 					t.Errorf("FindOrCreateRole() = %v, want %v", got.Name, tt.want)
// 				}
// 			})
// 		}
// 		return errors.New("rollback")
// 	})
// }
// func TestCreateRole(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
// 		type args struct {
// 			ctx  context.Context
// 			dbx  database.Dbx
// 			role *queries.CreateRoleDto
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			want    string
// 			wantErr bool
// 		}{
// 			{
// 				name: "create role with name only",
// 				args: args{
// 					ctx: ctx,
// 					dbx: dbxx,
// 					role: &queries.CreateRoleDto{
// 						Name: "test_role",
// 					},
// 				},
// 				want:    "test_role",
// 				wantErr: false,
// 			},
// 			{
// 				name: "create role with description",
// 				args: args{
// 					ctx: ctx,
// 					dbx: dbxx,
// 					role: &queries.CreateRoleDto{
// 						Name:        "test_role_2",
// 						Description: new(string),
// 					},
// 				},
// 				want:    "test_role_2",
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				got, err := queries.CreateRole(tt.args.ctx, tt.args.dbx, tt.args.role)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("CreateRole() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if got.Name != tt.want {
// 					t.Errorf("CreateRole() = %v, want %v", got.Name, tt.want)
// 				}
// 			})
// 		}
// 		return errors.New("rollback")
// 	})
// }
// func TestUpdateRole(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
// 		// Create initial role to update
// 		role, err := queries.CreateRole(ctx, dbxx, &queries.CreateRoleDto{
// 			Name: "initial_role",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create initial role: %v", err)
// 		}

// 		description := "updated description"

// 		type args struct {
// 			ctx     context.Context
// 			dbx     database.Dbx
// 			id      uuid.UUID
// 			roledto *queries.UpdateRoleDto
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			wantErr bool
// 		}{
// 			{
// 				name: "update existing role",
// 				args: args{
// 					ctx: ctx,
// 					dbx: dbxx,
// 					id:  role.ID,
// 					roledto: &queries.UpdateRoleDto{
// 						Name:        "updated_role",
// 						Description: &description,
// 					},
// 				},
// 				wantErr: false,
// 			},
// 			{
// 				name: "update non-existent role",
// 				args: args{
// 					ctx: ctx,
// 					dbx: dbxx,
// 					id:  uuid.New(),
// 					roledto: &queries.UpdateRoleDto{
// 						Name: "test_role",
// 					},
// 				},
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				err := queries.UpdateRole(tt.args.ctx, tt.args.dbx, tt.args.id, tt.args.roledto)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("UpdateRole() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}

// 				if tt.name == "update existing role" {
// 					// Verify the update
// 					updatedRole, err := crudrepo.Role.GetOne(ctx, tt.args.dbx,
// 						&map[string]any{
// 							"id": map[string]any{
// 								"_eq": tt.args.id.String(),
// 							},
// 						})
// 					if err != nil {
// 						t.Errorf("Failed to get updated role: %v", err)
// 						return
// 					}
// 					if updatedRole.Name != tt.args.roledto.Name {
// 						t.Errorf("Role name = %v, want %v", updatedRole.Name, tt.args.roledto.Name)
// 					}
// 					if *updatedRole.Description != *tt.args.roledto.Description {
// 						t.Errorf("Role description = %v, want %v", *updatedRole.Description, *tt.args.roledto.Description)
// 					}
// 				}
// 			})
// 		}
// 		return errors.New("rollback")
// 	})
// }
// func TestUpdatePermission(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
// 		// Create initial permission to update
// 		permission, err := queries.CreatePermission(ctx, dbxx, &queries.CreatePermissionDto{
// 			Name: "initial_permission",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create initial permission: %v", err)
// 		}

// 		description := "updated description"

// 		type args struct {
// 			ctx     context.Context
// 			dbx     database.Dbx
// 			id      uuid.UUID
// 			roledto *queries.UpdatePermissionDto
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			wantErr bool
// 		}{
// 			{
// 				name: "update existing permission",
// 				args: args{
// 					ctx: ctx,
// 					dbx: dbxx,
// 					id:  permission.ID,
// 					roledto: &queries.UpdatePermissionDto{
// 						Name:        "updated_permission",
// 						Description: &description,
// 					},
// 				},
// 				wantErr: false,
// 			},
// 			{
// 				name: "update non-existent permission",
// 				args: args{
// 					ctx: ctx,
// 					dbx: dbxx,
// 					id:  uuid.New(),
// 					roledto: &queries.UpdatePermissionDto{
// 						Name: "test_permission",
// 					},
// 				},
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				err := queries.UpdatePermission(tt.args.ctx, tt.args.dbx, tt.args.id, tt.args.roledto)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("UpdatePermission() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}

// 				if tt.name == "update existing permission" {
// 					// Verify the update
// 					updatedPermission, err := crudrepo.Permission.GetOne(ctx, tt.args.dbx,
// 						&map[string]any{
// 							"id": map[string]any{
// 								"_eq": tt.args.id.String(),
// 							},
// 						})
// 					if err != nil {
// 						t.Errorf("Failed to get updated permission: %v", err)
// 						return
// 					}
// 					if updatedPermission.Name != tt.args.roledto.Name {
// 						t.Errorf("Permission name = %v, want %v", updatedPermission.Name, tt.args.roledto.Name)
// 					}
// 					if *updatedPermission.Description != *tt.args.roledto.Description {
// 						t.Errorf("Permission description = %v, want %v", *updatedPermission.Description, *tt.args.roledto.Description)
// 					}
// 				}
// 			})
// 		}
// 		return errors.New("rollback")
// 	})
// }
// func TestDeleteRole(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
// 		// Create a role to delete
// 		role, err := queries.CreateRole(ctx, dbxx, &queries.CreateRoleDto{
// 			Name: "role_to_delete",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create test role: %v", err)
// 		}

// 		type args struct {
// 			ctx context.Context
// 			dbx database.Dbx
// 			id  uuid.UUID
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			wantErr bool
// 		}{
// 			{
// 				name: "delete existing role",
// 				args: args{
// 					ctx: ctx,
// 					dbx: dbxx,
// 					id:  role.ID,
// 				},
// 				wantErr: false,
// 			},
// 			{
// 				name: "delete non-existent role",
// 				args: args{
// 					ctx: ctx,
// 					dbx: dbxx,
// 					id:  uuid.New(),
// 				},
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				err := queries.DeleteRole(tt.args.ctx, tt.args.dbx, tt.args.id)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("DeleteRole() error = %v, wantErr %v", err, tt.wantErr)
// 				}

// 				if tt.name == "delete existing role" {
// 					// Verify the role was deleted
// 					deletedRole, err := crudrepo.Role.GetOne(ctx, tt.args.dbx,
// 						&map[string]any{
// 							"id": map[string]any{
// 								"_eq": tt.args.id.String(),
// 							},
// 						})
// 					if err != nil {
// 						t.Errorf("Failed to check deleted role: %v", err)
// 						return
// 					}
// 					if deletedRole != nil {
// 						t.Errorf("Role still exists after deletion")
// 					}
// 				}
// 			})
// 		}
// 		return errors.New("rollback")
// 	})
// }
// func TestDeleteRolePermissions(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
// 		// Create a role and permission to test deletion
// 		role, err := queries.CreateRole(ctx, dbxx, &queries.CreateRoleDto{
// 			Name: "role_for_permissions",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create test role: %v", err)
// 		}

// 		permission, err := queries.CreatePermission(ctx, dbxx, &queries.CreatePermissionDto{
// 			Name: "test_permission",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create test permission: %v", err)
// 		}

// 		// Create role-permission association
// 		err = queries.CreateRolePermissions(ctx, dbxx, role.ID, permission.ID)
// 		if err != nil {
// 			t.Fatalf("failed to create role permissions: %v", err)
// 		}

// 		type args struct {
// 			ctx context.Context
// 			dbx database.Dbx
// 			id  uuid.UUID
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			wantErr bool
// 		}{
// 			{
// 				name: "delete existing role permissions",
// 				args: args{
// 					ctx: ctx,
// 					dbx: dbxx,
// 					id:  role.ID,
// 				},
// 				wantErr: false,
// 			},
// 			{
// 				name: "delete non-existent role permissions",
// 				args: args{
// 					ctx: ctx,
// 					dbx: dbxx,
// 					id:  uuid.New(),
// 				},
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				err := queries.DeleteRolePermissions(tt.args.ctx, tt.args.dbx, tt.args.id)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("DeleteRolePermissions() error = %v, wantErr %v", err, tt.wantErr)
// 				}

// 				if tt.name == "delete existing role permissions" {
// 					// Verify the role permissions were deleted
// 					permissions, err := queries.LoadRolePermissions(tt.args.ctx, tt.args.dbx, tt.args.id)
// 					if err != nil {
// 						t.Errorf("Failed to check deleted role permissions: %v", err)
// 						return
// 					}
// 					if len(permissions[0]) != 0 {
// 						t.Errorf("Role permissions still exist after deletion")
// 					}
// 				}
// 			})
// 		}
// 		return errors.New("rollback")
// 	})
// }
// func TestFindOrCreatePermission(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
// 		type args struct {
// 			ctx            context.Context
// 			db             database.Dbx
// 			permissionName string
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			want    string
// 			wantErr bool
// 		}{
// 			{
// 				name: "create new permission",
// 				args: args{
// 					ctx:            ctx,
// 					db:             dbxx,
// 					permissionName: "test_permission",
// 				},
// 				want:    "test_permission",
// 				wantErr: false,
// 			},
// 			{
// 				name: "find existing permission",
// 				args: args{
// 					ctx:            ctx,
// 					db:             dbxx,
// 					permissionName: "test_permission",
// 				},
// 				want:    "test_permission",
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				got, err := queries.FindOrCreatePermission(tt.args.ctx, tt.args.db, tt.args.permissionName)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("FindOrCreatePermission() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if got.Name != tt.want {
// 					t.Errorf("FindOrCreatePermission() = %v, want %v", got.Name, tt.want)
// 				}
// 			})
// 		}
// 		return errors.New("rollback")
// 	})
// }
// func TestCreatePermission(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
// 		description := "test description"

// 		type args struct {
// 			ctx        context.Context
// 			dbx        database.Dbx
// 			permission *queries.CreatePermissionDto
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			want    string
// 			wantErr bool
// 		}{
// 			{
// 				name: "create permission with name only",
// 				args: args{
// 					ctx: ctx,
// 					dbx: dbxx,
// 					permission: &queries.CreatePermissionDto{
// 						Name: "test_permission",
// 					},
// 				},
// 				want:    "test_permission",
// 				wantErr: false,
// 			},
// 			{
// 				name: "create permission with description",
// 				args: args{
// 					ctx: ctx,
// 					dbx: dbxx,
// 					permission: &queries.CreatePermissionDto{
// 						Name:        "test_permission_2",
// 						Description: &description,
// 					},
// 				},
// 				want:    "test_permission_2",
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				got, err := queries.CreatePermission(tt.args.ctx, tt.args.dbx, tt.args.permission)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("CreatePermission() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if got.Name != tt.want {
// 					t.Errorf("CreatePermission() = %v, want %v", got.Name, tt.want)
// 				}
// 				if tt.args.permission.Description != nil {
// 					if *got.Description != *tt.args.permission.Description {
// 						t.Errorf("CreatePermission() description = %v, want %v", *got.Description, *tt.args.permission.Description)
// 					}
// 				}
// 			})
// 		}
// 		return errors.New("rollback")
// 	})
// }

// func TestDeletePermission(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
// 		// Create a permission to delete
// 		permission, err := queries.CreatePermission(ctx, dbxx, &queries.CreatePermissionDto{
// 			Name: "permission_to_delete",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create test permission: %v", err)
// 		}

// 		type args struct {
// 			ctx context.Context
// 			dbx database.Dbx
// 			id  uuid.UUID
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			wantErr bool
// 		}{
// 			{
// 				name: "delete existing permission",
// 				args: args{
// 					ctx: ctx,
// 					dbx: dbxx,
// 					id:  permission.ID,
// 				},
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				err := queries.DeletePermission(tt.args.ctx, tt.args.dbx, tt.args.id)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("DeletePermission() error = %v, wantErr %v", err, tt.wantErr)
// 				}

// 				if tt.name == "delete existing permission" {
// 					// Verify the permission was deleted
// 					deletedPermission, err := crudrepo.Permission.GetOne(ctx, tt.args.dbx,
// 						&map[string]any{
// 							"id": map[string]any{
// 								"_eq": tt.args.id.String(),
// 							},
// 						})
// 					if err != nil {
// 						t.Errorf("Failed to check deleted permission: %v", err)
// 						return
// 					}
// 					if deletedPermission != nil {
// 						t.Errorf("Permission still exists after deletion")
// 					}
// 				}
// 			})
// 		}
// 		return errors.New("rollback")
// 	})
// }
// func TestFindPermissionById(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
// 		// Create a test permission to find
// 		permission, err := queries.CreatePermission(ctx, dbxx, &queries.CreatePermissionDto{
// 			Name: "test_permission",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create test permission: %v", err)
// 		}

// 		type args struct {
// 			ctx context.Context
// 			dbx database.Dbx
// 			id  uuid.UUID
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			want    *crudModels.Permission
// 			wantErr bool
// 		}{
// 			{
// 				name: "find existing permission",
// 				args: args{
// 					ctx: ctx,
// 					dbx: dbxx,
// 					id:  permission.ID,
// 				},
// 				want:    permission,
// 				wantErr: false,
// 			},
// 			{
// 				name: "find non-existent permission",
// 				args: args{
// 					ctx: ctx,
// 					dbx: dbxx,
// 					id:  uuid.New(),
// 				},
// 				want:    nil,
// 				wantErr: false,
// 			},
// 		}

// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				got, err := queries.FindPermissionById(tt.args.ctx, tt.args.dbx, tt.args.id)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("FindPermissionById() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if tt.want == nil && got != nil {
// 					t.Errorf("FindPermissionById() = %v, want nil", got)
// 				} else if tt.want != nil && got == nil {
// 					t.Errorf("FindPermissionById() = nil, want %v", tt.want)
// 				} else if tt.want != nil && got != nil && got.ID != tt.want.ID {
// 					t.Errorf("FindPermissionById() = %v, want %v", got.ID, tt.want.ID)
// 				}
// 			})
// 		}
// 		return errors.New("rollback")
// 	})
// }
