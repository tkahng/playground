package queries_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

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
