package queries_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/jaswdr/faker/v2"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/seeders"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/test"
	"github.com/tkahng/authgo/internal/tools/types"
	"github.com/tkahng/authgo/internal/tools/utils"
)

func TestListUsers(t *testing.T) {
	test.Short(t)
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
func TestCountUsers(t *testing.T) {
	test.Short(t)
ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		faker := faker.New().Internet()
		_, err := seeders.CreateUserWithAccountAndRole(ctx, dbxx, 5, models.ProvidersGoogle, "superuser", faker)
		if err != nil {
			t.Fatalf("failed to create users: %v", err)
		}
		_, err = seeders.CreateUserWithAccountAndRole(ctx, dbxx, 5, models.ProvidersCredentials, "basic", faker)
		if err != nil {
			t.Fatalf("failed to create users: %v", err)
		}
		_, err = seeders.CreateUserWithAccountAndRole(ctx, dbxx, 5, models.ProvidersGithub, "pro", faker)
		if err != nil {
			t.Fatalf("failed to create users: %v", err)
		}

		tests := []struct {
			name    string
			filter  *shared.UserListFilter
			want    int64
			wantErr bool
		}{
			{
				name: "count google users",
				filter: &shared.UserListFilter{
					Providers: []shared.Providers{shared.ProvidersGoogle},
				},
				want:    5,
				wantErr: false,
			},
			{
				name: "count credentials users",
				filter: &shared.UserListFilter{
					Providers: []shared.Providers{shared.ProvidersCredentials},
				},
				want:    5,
				wantErr: false,
			},
			{
				name: "count github users",
				filter: &shared.UserListFilter{
					Providers: []shared.Providers{shared.ProvidersGithub},
				},
				want:    5,
				wantErr: false,
			},
			{
				name:    "count all users",
				filter:  nil,
				want:    15,
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.CountUsers(ctx, dbxx, tt.filter)
				if (err != nil) != tt.wantErr {
					t.Errorf("CountUsers() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("CountUsers() = %v, want %v", got, tt.want)
				}
			})
		}
		return test.EndTestErr
	})
}
func TestDeleteUsers(t *testing.T) {
	test.Short(t)
ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		faker := faker.New().Internet()
		users, err := seeders.CreateUserWithAccountAndRole(ctx, dbxx, 1, models.ProvidersGoogle, "basic", faker)
		if err != nil {
			t.Fatalf("failed to create users: %v", err)
		}

		tests := []struct {
			name    string
			userId  uuid.UUID
			wantErr bool
		}{
			{
				name:    "delete existing user",
				userId:  users[0].ID,
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := queries.DeleteUsers(ctx, dbxx, tt.userId)
				if (err != nil) != tt.wantErr {
					t.Errorf("DeleteUsers() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
		return test.EndTestErr
	})
}
func TestUpdateUser(t *testing.T) {
	test.Short(t)
ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		faker := faker.New().Internet()
		users, err := seeders.CreateUserWithAccountAndRole(ctx, dbxx, 1, models.ProvidersGoogle, "basic", faker)
		if err != nil {
			t.Fatalf("failed to create users: %v", err)
		}

		tests := []struct {
			name    string
			userId  uuid.UUID
			input   *shared.UserMutationInput
			wantErr bool
		}{
			{
				name:   "update existing user",
				userId: users[0].ID,
				input: &shared.UserMutationInput{
					Email: "updated@example.com",
					Name:  types.Pointer("Updated Name"),
					Image: types.Pointer("updated-image.jpg"),
				},
				wantErr: false,
			},
			{
				name:   "update non-existent user",
				userId: uuid.New(),
				input: &shared.UserMutationInput{
					Email: "nonexistent@example.com",
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := queries.UpdateUser(ctx, dbxx, tt.userId, tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("UpdateUser() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
		return test.EndTestErr
	})
}
