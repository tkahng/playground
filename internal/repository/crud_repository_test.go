package repository_test

import (
	"context"
	"testing"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/db/seeders"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/test"
)

func TestListUsers(t *testing.T) {
	ctx, db, pl := test.DbSetup()
	t.Cleanup(func() {
		repository.TruncateModels(ctx, db)
		pl.Close()
	})
	_ = seeders.UserCredentialsFactory(ctx, db, 10)
	_ = seeders.UserOauthFactory(ctx, db, 10, models.ProvidersGoogle)
	_ = seeders.UserOauthFactory(ctx, db, 10, models.ProvidersGithub)
	type args struct {
		ctx   context.Context
		db    bob.DB
		input *shared.UserListParams
		count int64
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "only 10 users, default pagination",
			args: args{
				ctx: ctx,
				db:  db,
				input: &shared.UserListParams{
					PaginatedInput: shared.PaginatedInput{
						// PerPage: shared.From(10),
						// Page:    shared.From(1),
						PerPage: 10,
						Page:    1,
					},
				},
				count: 10,
			},
			want:    10,
			wantErr: false,
		},
		{
			name: "10 users, 5 per page, 2nd page",
			args: args{
				ctx: ctx,
				db:  db,
				input: &shared.UserListParams{
					PaginatedInput: shared.PaginatedInput{
						// PerPage: shared.From(5),
						// Page:    shared.From(2),
						PerPage: 5,
						Page:    2,
					},
					UserListFilter: shared.UserListFilter{
						// Provider: shared.From(models.ProvidersGoogle),
						Providers: []models.Providers{models.ProvidersGoogle},
					},
				},
				count: 15,
			},
			want:    5,
			wantErr: false,
		},
		{
			name: "10 users, 10 per page, 2rd page",
			args: args{
				ctx: ctx,
				db:  db,
				input: &shared.UserListParams{
					PaginatedInput: shared.PaginatedInput{
						// PerPage: shared.From(10),
						// Page:    shared.From(2),
						PerPage: 10,
						Page:    2,
					},
					UserListFilter: shared.UserListFilter{
						// Provider: shared.From(models.ProvidersCredentials),
						Providers: []models.Providers{
							models.ProvidersCredentials,
						},
					},
				},
				count: 10,
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "10 users, 10 per page, 2rd page",
			args: args{
				ctx: ctx,
				db:  db,
				input: &shared.UserListParams{
					PaginatedInput: shared.PaginatedInput{
						// PerPage: shared.From(10),
						// Page:    shared.From(2),
						PerPage: 10,
						Page:    2,
					},
					UserListFilter: shared.UserListFilter{
						// Provider: shared.From(models.ProvidersGithub),
						Providers: []models.Providers{
							models.ProvidersGithub,
						},
					},
				},
				count: 10,
			},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Cleanup(func() {
			// 	repository.TruncateModels(tt.args.ctx, tt.args.db)
			// })
			// f.NewUser().CreateMany(tt.args.ctx, tt.args.db, int(tt.args.count))
			got, err := repository.ListUsers(tt.args.ctx, tt.args.db, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !(len(got) == int(tt.want)) {
				t.Errorf("ListUsers() = %v, want %v", len(got), tt.want)
			}
		})
	}
}
