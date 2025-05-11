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
	"github.com/tkahng/authgo/internal/tools/types"
)

func TestCreateUser(t *testing.T) {
	ctx, dbx := test.DbSetup()

	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		type args struct {
			ctx    context.Context
			db     db.Dbx
			params *shared.AuthenticationInput
		}
		tests := []struct {
			name    string
			args    args
			want    *crudModels.User
			wantErr bool
		}{
			{
				name: "create user",
				args: args{
					ctx: ctx,
					db:  dbxx,
					params: &shared.AuthenticationInput{
						Email: "tkahng@gmail.com",
						Name:  types.Pointer("tchunoo"),
					},
				},
				want: &crudModels.User{
					Email: "tkahng@gmail.com",
					Name:  types.Pointer("tchunoo"),
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.CreateUser(tt.args.ctx, tt.args.db, tt.args.params)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got.Email, tt.want.Email) {
					t.Errorf("CreateUser() = %v, want %v", got.Email, tt.want.Email)
				}
			})
		}
		return errors.New("test error")
	})

}

func TestCreateUserRoles(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create a user
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Errorf("failed to create user: %v", err)
			return err
		}
		role, err := queries.FindOrCreateRole(ctx, dbxx, "basic")
		if err != nil {
			t.Errorf("failed to create role: %v", err)
			return err
		}
		type args struct {
			ctx     context.Context
			db      db.Dbx
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
				if err := queries.CreateUserRoles(tt.args.ctx, tt.args.db, tt.args.userId, tt.args.roleIds...); (err != nil) != tt.wantErr {
					t.Errorf("createUserRoles() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
		return errors.New("test error")
	})
}

func TestCreateUserPermissions(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create a user
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Errorf("failed to create user: %v", err)
			return err
		}
		permission, err := queries.FindOrCreatePermission(ctx, dbxx, "basic")
		if err != nil {
			t.Errorf("failed to create role: %v", err)
			return err
		}
		type args struct {
			ctx           context.Context
			db            db.Dbx
			userId        uuid.UUID
			permissionIds []uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "create user permissions",
				args: args{
					ctx:           ctx,
					db:            dbxx,
					userId:        user.ID,
					permissionIds: []uuid.UUID{permission.ID},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := queries.CreateUserPermissions(tt.args.ctx, tt.args.db, tt.args.userId, tt.args.permissionIds...); (err != nil) != tt.wantErr {
					t.Errorf("CreateUserPermissions() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
		return errors.New("test error")
	})
}
