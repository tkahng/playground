package stores_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/test"
	"github.com/tkahng/authgo/internal/tools/types"
)

func TestPostgresStripeStore_CreateCustomer(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		userStore := stores.NewPostgresUserStore(dbxx)
		teamStore := stores.NewPostgresTeamStore(dbxx)
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			return err
		}
		user2, err := userStore.CreateUser(ctx, &models.User{
			Email: "user2@gmail.com",
		})
		if err != nil {
			return err
		}
		team, err := teamStore.CreateTeam(ctx, "test", "test")
		if err != nil {
			return err
		}

		type fields struct {
			db database.Dbx
		}
		type args struct {
			ctx      context.Context
			customer *models.StripeCustomer
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			want    *models.StripeCustomer
			wantErr bool
		}{
			{
				name: "create user customer",
				fields: fields{
					db: dbxx,
				},
				args: args{
					ctx: ctx,
					customer: &models.StripeCustomer{
						ID:           "cus_123",
						UserID:       types.Pointer(user.ID),
						Email:        user.Email,
						CustomerType: models.StripeCustomerTypeUser,
					},
				},
				want: &models.StripeCustomer{
					ID:           "cus_123",
					UserID:       types.Pointer(user.ID),
					Email:        user.Email,
					CustomerType: models.StripeCustomerTypeUser,
				},
				wantErr: false,
			},
			{
				name: "create customer with invalid user",
				fields: fields{
					db: dbxx,
				},
				args: args{
					ctx: ctx,
					customer: &models.StripeCustomer{
						ID:           "cus_456",
						UserID:       nil,
						Email:        "",
						CustomerType: models.StripeCustomerTypeUser,
					},
				},
				want:    nil,
				wantErr: true,
			},
			{
				name: "create team customer",
				fields: fields{
					db: dbxx,
				},
				args: args{
					ctx: ctx,
					customer: &models.StripeCustomer{
						ID:           "cus_789",
						TeamID:       types.Pointer(team.ID),
						Email:        "",
						CustomerType: models.StripeCustomerTypeTeam,
					},
				},
				want: &models.StripeCustomer{
					ID:           "cus_789",
					TeamID:       types.Pointer(team.ID),
					Email:        "",
					CustomerType: models.StripeCustomerTypeTeam,
				},
				wantErr: false,
			},
			{
				name: "create customer with user without type",
				fields: fields{
					db: dbxx,
				},
				args: args{
					ctx: ctx,
					customer: &models.StripeCustomer{
						ID:     "cus_101",
						UserID: types.Pointer(user2.ID),
						Email:  user2.Email,
					},
				},
				want: &models.StripeCustomer{
					ID:           "cus_101",
					UserID:       types.Pointer(user2.ID),
					Email:        user2.Email,
					CustomerType: models.StripeCustomerTypeUser,
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				store := stores.NewPostgresStripeStore(tt.fields.db)
				got, err := store.CreateCustomer(tt.args.ctx, tt.args.customer)
				if (err != nil) != tt.wantErr {
					t.Errorf("PostgresStripeStore.CreateCustomer() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != nil && tt.want != nil {
					if got.ID != tt.want.ID {
						t.Errorf("PostgresStripeStore.CreateCustomer() got = %v, want %v", got.ID, tt.want.ID)
					}
					if got.UserID != nil && tt.want.UserID != nil {
						if *got.UserID != *tt.want.UserID {
							t.Errorf("PostgresStripeStore.CreateCustomer() got = %v, want %v", *got.UserID, *tt.want.UserID)
						}
					}
					if got.TeamID != nil && tt.want.TeamID != nil {
						if *got.TeamID != *tt.want.TeamID {
							t.Errorf("PostgresStripeStore.CreateCustomer() got = %v, want %v", *got.TeamID, *tt.want.TeamID)
						}
					}

					if got.CustomerType != tt.want.CustomerType {
						t.Errorf("PostgresStripeStore.CreateCustomer() got.CustomerType = %v, want %v", got.CustomerType, tt.want.CustomerType)
					}
				}
			})
		}
		return errors.New("rollback")
	})
}
