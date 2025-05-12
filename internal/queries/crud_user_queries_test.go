package queries_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/jaswdr/faker/v2"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/seeders"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/test"
	"github.com/tkahng/authgo/internal/tools/utils"
)

func TestListUsers(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		faker := faker.New().Internet()
		users1, err := seeders.CreateUserWithAccountAndRole(ctx, dbxx, 5, models.ProvidersGoogle, "superuser", faker)
		if err != nil {
			t.Fatalf("failed to create users: %v", err)
		}
		fmt.Println("users1", len(users1))
		_, err = seeders.CreateUserWithAccountAndRole(ctx, dbxx, 5, models.ProvidersCredentials, "basic", faker)
		if err != nil {
			t.Fatalf("failed to create users: %v", err)
		}
		_, err = seeders.CreateUserWithAccountAndRole(ctx, dbxx, 5, models.ProvidersGithub, "pro", faker)
		if err != nil {
			t.Fatalf("failed to create users: %v", err)
		}
		type args struct {
			ctx   context.Context
			db    db.Dbx
			input *shared.UserListParams
		}
		tests := []struct {
			name      string
			args      args
			wantCount int
			wantErr   bool
		}{
			{
				name: "query google users",
				args: args{
					ctx: ctx,
					db:  dbxx,
					input: &shared.UserListParams{
						UserListFilter: shared.UserListFilter{
							Providers: []shared.Providers{
								shared.ProvidersGoogle,
							},
						},
					},
				},
				wantCount: 5,
				wantErr:   false,
			},
			{
				name: "query credentials users",
				args: args{
					ctx: ctx,
					db:  dbxx,
					input: &shared.UserListParams{
						UserListFilter: shared.UserListFilter{
							Providers: []shared.Providers{
								shared.ProvidersCredentials,
							},
						},
					},
				},
				wantCount: 5,
				wantErr:   false,
			},
			{
				name: "query github users",
				args: args{
					ctx: ctx,
					db:  dbxx,
					input: &shared.UserListParams{
						UserListFilter: shared.UserListFilter{
							Providers: []shared.Providers{
								shared.ProvidersGithub,
							},
						},
					},
				},
				wantCount: 5,
				wantErr:   false,
			},
			{
				name: "query all users",
				args: args{
					ctx: ctx,
					db:  dbxx,
					input: &shared.UserListParams{
						PaginatedInput: shared.PaginatedInput{
							Page:    0,
							PerPage: 15,
						},
					},
				},
				wantCount: 15,
				wantErr:   false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.ListUsers(tt.args.ctx, tt.args.db, tt.args.input)
				utils.PrettyPrintJSON(got)
				if (err != nil) != tt.wantErr {
					t.Errorf("ListUsers() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(len(got), tt.wantCount) {
					t.Errorf("ListUsers() = %v, want %v", len(got), tt.wantCount)
				}
			})
		}
		return test.EndTestErr
	})
}
