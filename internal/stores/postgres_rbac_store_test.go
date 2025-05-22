package stores_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/test"
)

func TestListPermissions(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(tx database.Dbx) error {
		err := queries.EnsureRoleAndPermissions(
			ctx,
			tx,
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
				store := stores.NewPostgresRBACStore(tx)
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
		err := queries.EnsureRoleAndPermissions(
			ctx,
			tx,
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
				store := stores.NewPostgresRBACStore(tx)
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
		err := queries.EnsureRoleAndPermissions(
			ctx,
			tx,
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
				store := stores.NewPostgresRBACStore(tx)
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
		err := queries.EnsureRoleAndPermissions(
			ctx,
			tx,
			shared.PermissionNameAdmin,
			shared.PermissionNameAdmin,
			shared.PermissionNameBasic,
		)

		if err != nil {
			t.Fatalf("failed to ensure role and permissions: %v", err)
		}
		err = queries.EnsureRoleAndPermissions(
			ctx,
			tx,
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
				store := stores.NewPostgresRBACStore(tx)
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
				store := stores.NewPostgresRBACStore(dbxx)
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
		// Create test user
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "test@test.com",
		})
		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}

		// Create test role and permission
		role, err := queries.CreateRole(ctx, dbxx, &queries.CreateRoleDto{
			Name: "test_role",
		})
		if err != nil {
			t.Fatalf("failed to create test role: %v", err)
		}

		perm, err := queries.CreatePermission(ctx, dbxx, &queries.CreatePermissionDto{
			Name: "test_permission",
		})
		if err != nil {
			t.Fatalf("failed to create test permission: %v", err)
		}

		// Assign role to user and permission to role
		err = queries.CreateUserRoles(ctx, dbxx, user.ID, role.ID)
		if err != nil {
			t.Fatalf("failed to assign role to user: %v", err)
		}

		err = queries.CreateRolePermissions(ctx, dbxx, role.ID, perm.ID)
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
				store := stores.NewPostgresRBACStore(dbxx)
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
		// Create test user
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "test@test.com",
		})
		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}

		// Create test role and permission
		role, err := queries.CreateRole(ctx, dbxx, &queries.CreateRoleDto{
			Name: "test_role",
		})
		if err != nil {
			t.Fatalf("failed to create test role: %v", err)
		}

		perm, err := queries.CreatePermission(ctx, dbxx, &queries.CreatePermissionDto{
			Name: "test_permission",
		})
		if err != nil {
			t.Fatalf("failed to create test permission: %v", err)
		}

		// Assign role to user and permission to role
		err = queries.CreateUserRoles(ctx, dbxx, user.ID, role.ID)
		if err != nil {
			t.Fatalf("failed to assign role to user: %v", err)
		}

		err = queries.CreateRolePermissions(ctx, dbxx, role.ID, perm.ID)
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
				store := stores.NewPostgresRBACStore(dbxx)
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
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "test@test.com",
		})
		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}

		// Create test role and permission that will be assigned
		role, err := queries.CreateRole(ctx, dbxx, &queries.CreateRoleDto{
			Name: "test_role",
		})
		if err != nil {
			t.Fatalf("failed to create test role: %v", err)
		}

		assignedPerm, err := queries.CreatePermission(ctx, dbxx, &queries.CreatePermissionDto{
			Name: "assigned_permission",
		})
		if err != nil {
			t.Fatalf("failed to create assigned permission: %v", err)
		}

		// Create an unassigned permission
		unassignedPerm, err := queries.CreatePermission(ctx, dbxx, &queries.CreatePermissionDto{
			Name: "unassigned_permission",
		})
		if err != nil {
			t.Fatalf("failed to create unassigned permission: %v", err)
		}

		// Assign role to user and permission to role
		err = queries.CreateUserRoles(ctx, dbxx, user.ID, role.ID)
		if err != nil {
			t.Fatalf("failed to assign role to user: %v", err)
		}

		err = queries.CreateRolePermissions(ctx, dbxx, role.ID, assignedPerm.ID)
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
				store := stores.NewPostgresRBACStore(dbxx)
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
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "test@test.com",
		})
		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}

		// Create test role and permission that will be assigned
		role, err := queries.CreateRole(ctx, dbxx, &queries.CreateRoleDto{
			Name: "test_role",
		})
		if err != nil {
			t.Fatalf("failed to create test role: %v", err)
		}

		assignedPerm, err := queries.CreatePermission(ctx, dbxx, &queries.CreatePermissionDto{
			Name: "assigned_permission",
		})
		if err != nil {
			t.Fatalf("failed to create assigned permission: %v", err)
		}

		// Create an unassigned permission
		_, err = queries.CreatePermission(ctx, dbxx, &queries.CreatePermissionDto{
			Name: "unassigned_permission",
		})
		if err != nil {
			t.Fatalf("failed to create unassigned permission: %v", err)
		}

		// Assign role to user and permission to role
		err = queries.CreateUserRoles(ctx, dbxx, user.ID, role.ID)
		if err != nil {
			t.Fatalf("failed to assign role to user: %v", err)
		}

		err = queries.CreateRolePermissions(ctx, dbxx, role.ID, assignedPerm.ID)
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
				store := stores.NewPostgresRBACStore(dbxx)
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
		paymentStore := stores.NewPostgresPaymentStore(dbxx)
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
		userStore := stores.NewPostgresUserStore(dbxx)
		rbacStore := stores.NewPostgresRBACStore(dbxx)
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
