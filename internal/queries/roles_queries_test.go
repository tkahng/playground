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
