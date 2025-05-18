package queries_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
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
				got, err := queries.ListPermissions(tt.args.ctx, tt.args.db, tt.args.input)
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
				got, err := queries.CountPermissions(ctx, tx, tt.filter)
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
				got, err := queries.ListRoles(tt.args.ctx, tt.args.db, tt.args.input)
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
				got, err := queries.CountRoles(ctx, tx, tt.filter)
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
