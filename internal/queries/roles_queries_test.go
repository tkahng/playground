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
	"github.com/tkahng/authgo/internal/test"
)

func TestLoadRolePermissions(t *testing.T) {
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
