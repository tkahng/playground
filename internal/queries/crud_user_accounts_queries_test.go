package queries_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/jaswdr/faker/v2"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/seeders"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/test"
)

func TestListUserAccounts(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		faker := faker.New().Internet()
		_, err := seeders.CreateUserWithAccountAndRole(ctx, dbxx, 5, models.ProvidersGoogle, "basic", faker)
		if err != nil {
			t.Fatalf("failed to create users: %v", err)
		}
		_, err = seeders.CreateUserWithAccountAndRole(ctx, dbxx, 5, models.ProvidersCredentials, "admin", faker)
		if err != nil {
			t.Fatalf("failed to create users: %v", err)
		}

		type args struct {
			ctx   context.Context
			db    db.Dbx
			input *shared.UserAccountListParams
		}
		tests := []struct {
			name      string
			args      args
			want      []*models.UserAccount
			wantCount int
			wantErr   bool
		}{
			{
				name: "query google users",
				args: args{
					ctx: ctx,
					db:  dbxx,
					input: &shared.UserAccountListParams{
						UserAccountListFilter: shared.UserAccountListFilter{
							Providers: []shared.Providers{
								shared.ProvidersGoogle,
							},
						},
					},
				},
				want:      nil,
				wantCount: 5,
				wantErr:   false,
			},
			{
				name: "query credentials users",
				args: args{
					ctx: ctx,
					db:  dbxx,
					input: &shared.UserAccountListParams{
						UserAccountListFilter: shared.UserAccountListFilter{
							Providers: []shared.Providers{
								shared.ProvidersCredentials,
							},
						},
					},
				},
				want:      nil,
				wantCount: 5,
				wantErr:   false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.ListUserAccounts(tt.args.ctx, tt.args.db, tt.args.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("ListUserAccounts() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ListUserAccounts() = %v, want %v", got, tt.want)
				}
			})
		}
		return test.EndTestErr
	})
}
func TestCountUserAccounts(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		faker := faker.New().Internet()
		_, err := seeders.CreateUserWithAccountAndRole(ctx, dbxx, 5, models.ProvidersGoogle, "basic", faker)
		if err != nil {
			t.Fatalf("failed to create users: %v", err)
		}
		_, err = seeders.CreateUserWithAccountAndRole(ctx, dbxx, 5, models.ProvidersCredentials, "admin", faker)
		if err != nil {
			t.Fatalf("failed to create users: %v", err)
		}

		type args struct {
			ctx    context.Context
			db     db.Dbx
			filter *shared.UserAccountListFilter
		}
		tests := []struct {
			name    string
			args    args
			want    int64
			wantErr bool
		}{
			{
				name: "count google users",
				args: args{
					ctx: ctx,
					db:  dbxx,
					filter: &shared.UserAccountListFilter{
						Providers: []shared.Providers{
							shared.ProvidersGoogle,
						},
					},
				},
				want:    5,
				wantErr: false,
			},
			{
				name: "count credentials users",
				args: args{
					ctx: ctx,
					db:  dbxx,
					filter: &shared.UserAccountListFilter{
						Providers: []shared.Providers{
							shared.ProvidersCredentials,
						},
					},
				},
				want:    5,
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.CountUserAccounts(tt.args.ctx, tt.args.db, tt.args.filter)
				if (err != nil) != tt.wantErr {
					t.Errorf("CountUserAccounts() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("CountUserAccounts() = %v, want %v", got, tt.want)
				}
			})
		}
		return test.EndTestErr
	})
}
