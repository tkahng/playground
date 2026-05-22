//go:build integration

package stores_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/types"
)

func TestStripeStore_CreateCustomer(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)

		user, err := adapter.User().CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("got err %s", err.Error())
		}
		user2, err := adapter.User().CreateUser(ctx, &models.User{
			Email: "user2@gmail.com",
		})
		if err != nil {
			t.Fatalf("got err %s", err.Error())
		}
		team, err := adapter.TeamGroup().CreateTeam(ctx, "test", "test")
		if err != nil {
			t.Fatalf("got err %s", err.Error())
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
				got, err := adapter.Customer().CreateCustomer(tt.args.ctx, tt.args.customer)
				if (err != nil) != tt.wantErr {
					t.Errorf("PostgresStripeadapter.Customer().CreateCustomer() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != nil && tt.want != nil {
					if got.ID != tt.want.ID {
						t.Errorf("PostgresStripeadapter.Customer().CreateCustomer() got = %v, want %v", got.ID, tt.want.ID)
					}
					if got.UserID != nil && tt.want.UserID != nil {
						if *got.UserID != *tt.want.UserID {
							t.Errorf("PostgresStripeadapter.Customer().CreateCustomer() got = %v, want %v", *got.UserID, *tt.want.UserID)
						}
					}
					if got.TeamID != nil && tt.want.TeamID != nil {
						if *got.TeamID != *tt.want.TeamID {
							t.Errorf("PostgresStripeadapter.Customer().CreateCustomer() got = %v, want %v", *got.TeamID, *tt.want.TeamID)
						}
					}

					if got.CustomerType != tt.want.CustomerType {
						t.Errorf("PostgresStripeadapter.Customer().CreateCustomer() got.CustomerType = %v, want %v", got.CustomerType, tt.want.CustomerType)
					}
				}
			})
		}
	})
}

func TestStripeStore_FindCustomer(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		user, err := adapter.User().CreateUser(ctx, &models.User{Email: "findcustomer@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}

		customer := &models.StripeCustomer{
			ID:           "cus_find_1",
			UserID:       types.Pointer(user.ID),
			Email:        user.Email,
			CustomerType: models.StripeCustomerTypeUser,
		}
		_, err = adapter.Customer().CreateCustomer(ctx, customer)
		if err != nil {
			t.Fatalf("CreateCustomer() error = %v", err)
		}
		found, err := adapter.Customer().FindCustomer(ctx, &stores.StripeCustomerFilter{Ids: []string{"cus_find_1"}})
		if err != nil || found == nil || found.ID != "cus_find_1" {
			t.Errorf("FindCustomer() = %v, err = %v", found, err)
		}
	})
}

func TestDbCustomerStore_UpdateCustomer(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)
		user := stores.CreateUserWithOptions(t, adapter)
		customer, err := adapter.Customer().CreateCustomer(ctx, &models.StripeCustomer{
			ID:           "cus_1",
			UserID:       types.Pointer(user.User.ID),
			Email:        user.User.Email,
			Name:         types.Pointer("customer_name"),
			CustomerType: models.StripeCustomerTypeUser,
		})
		assert.NoError(t, err)
		assert.NotNil(t, customer)
		assert.Equal(t, "cus_1", customer.ID)
		customer.Email = "new@email.com"
		customer.Name = types.Pointer("new_customer_name")
		got, err := adapter.Customer().UpdateCustomer(ctx, customer)
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, customer.Email, got.Email)
		assert.Equal(t, *customer.Name, *got.Name)
	})
}
func TestDbCustomerStore_UpsertCustomer_Existing(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)
		user := stores.CreateUserWithOptions(t, adapter)
		fds := &models.StripeCustomer{
			ID:           "cus_1",
			UserID:       types.Pointer(user.User.ID),
			Email:        user.User.Email,
			Name:         types.Pointer("customer_name"),
			CustomerType: models.StripeCustomerTypeUser,
		}
		customer, err := adapter.Customer().CreateCustomer(ctx, fds)
		assert.NoError(t, err)
		assert.NotNil(t, customer)
		assert.Equal(t, "cus_1", customer.ID)
		customer.Email = "new@email.com"
		customer.Name = types.Pointer("new_customer_name")
		err = adapter.Customer().UpsertCustomer(ctx, customer)
		assert.NoError(t, err)
		got, err := adapter.Customer().FindCustomer(ctx, &stores.StripeCustomerFilter{
			Ids: []string{customer.ID},
		})
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, customer.Email, got.Email)
		assert.Equal(t, *customer.Name, *got.Name)
	})
}
func TestDbCustomerStore_UpsertCustomer_New(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)
		user := stores.CreateUserWithOptions(t, adapter)
		fds := &models.StripeCustomer{
			ID:           "cus_1",
			UserID:       types.Pointer(user.User.ID),
			Email:        user.User.Email,
			Name:         types.Pointer("customer_name"),
			CustomerType: models.StripeCustomerTypeUser,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		err := adapter.Customer().UpsertCustomer(ctx, fds)
		assert.NoError(t, err)
		got, err := adapter.Customer().FindCustomer(ctx, &stores.StripeCustomerFilter{
			Ids: []string{"cus_1"},
		})
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, fds.Email, got.Email)
		assert.Equal(t, *fds.Name, *got.Name)
	})
}
