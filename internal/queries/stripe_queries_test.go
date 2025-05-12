package queries_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/test"
)

func TestFindCustomerByStripeId(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		stripeId := "cus_test123"
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "test@example.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		err = queries.UpsertCustomerStripeId(ctx, dbxx, user.ID, stripeId)
		if err != nil {
			t.Fatalf("failed to upsert customer stripe id: %v", err)
		}
		customer, err := queries.FindCustomerByStripeId(ctx, dbxx, stripeId)
		if err != nil {
			t.Fatalf("failed to find customer by stripe id: %v", err)
		}
		if customer == nil {
			t.Fatalf("expected customer to be found, got nil")
		}
		if customer.StripeID != stripeId {
			t.Fatalf("expected stripe id %s, got %s", stripeId, customer.StripeID)
		}
		type args struct {
			ctx      context.Context
			dbx      db.Dbx
			stripeId string
		}
		tests := []struct {
			name    string
			args    args
			want    *models.StripeCustomer
			wantErr bool
		}{
			{
				name: "valid stripe id",
				args: args{
					ctx:      ctx,
					dbx:      dbxx,
					stripeId: stripeId,
				},
				want: &models.StripeCustomer{
					StripeID: stripeId,
					ID:       user.ID,
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.FindCustomerByStripeId(tt.args.ctx, tt.args.dbx, tt.args.stripeId)
				if (err != nil) != tt.wantErr {
					t.Errorf("FindCustomerByStripeId() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got.ID, tt.want.ID) {
					t.Errorf("FindCustomerByStripeId() got = %v, want %v", got.ID, tt.want.ID)
				}
				if !reflect.DeepEqual(got.StripeID, tt.want.StripeID) {
					t.Errorf("FindCustomerByStripeId() got = %v, want %v", got.StripeID, tt.want.StripeID)
				}
			})
		}
		return test.EndTestErr
	})
}
func TestFindCustomerByUserId(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		stripeId := "cus_test123"
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "test@example.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		err = queries.UpsertCustomerStripeId(ctx, dbxx, user.ID, stripeId)
		if err != nil {
			t.Fatalf("failed to upsert customer stripe id: %v", err)
		}

		customer, err := queries.FindCustomerByUserId(ctx, dbxx, user.ID)
		if err != nil {
			t.Fatalf("failed to find customer by user id: %v", err)
		}
		if customer == nil {
			t.Fatalf("expected customer to be found, got nil")
		}

		type args struct {
			ctx    context.Context
			dbx    db.Dbx
			userId uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			want    *models.StripeCustomer
			wantErr bool
		}{
			{
				name: "valid user id",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					userId: user.ID,
				},
				want: &models.StripeCustomer{
					ID:       user.ID,
					StripeID: stripeId,
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.FindCustomerByUserId(tt.args.ctx, tt.args.dbx, tt.args.userId)
				if (err != nil) != tt.wantErr {
					t.Errorf("FindCustomerByUserId() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got.ID, tt.want.ID) {
					t.Errorf("FindCustomerByUserId() got = %v, want %v", got.ID, tt.want.ID)
				}
				if !reflect.DeepEqual(got.StripeID, tt.want.StripeID) {
					t.Errorf("FindCustomerByUserId() got = %v, want %v", got.StripeID, tt.want.StripeID)
				}
			})
		}
		return test.EndTestErr
	})
}
func TestFindProductByStripeId(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		stripeId := "prod_test123"
		product := &models.StripeProduct{
			ID:          stripeId,
			Active:      true,
			Name:        "Test Product",
			Description: nil,
			Image:       nil,
			Metadata:    map[string]string{"key": "value"},
		}

		err := queries.UpsertProduct(ctx, dbxx, product)
		if err != nil {
			t.Fatalf("failed to upsert product: %v", err)
		}

		type args struct {
			ctx      context.Context
			dbx      db.Dbx
			stripeId string
		}
		tests := []struct {
			name    string
			args    args
			want    *models.StripeProduct
			wantErr bool
		}{
			{
				name: "valid stripe id",
				args: args{
					ctx:      ctx,
					dbx:      dbxx,
					stripeId: stripeId,
				},
				want: &models.StripeProduct{
					ID:     stripeId,
					Active: true,
					Name:   "Test Product",
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.FindProductByStripeId(tt.args.ctx, tt.args.dbx, tt.args.stripeId)
				if (err != nil) != tt.wantErr {
					t.Errorf("FindProductByStripeId() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got.ID, tt.want.ID) {
					t.Errorf("FindProductByStripeId() got = %v, want %v", got.ID, tt.want.ID)
				}
				if !reflect.DeepEqual(got.Active, tt.want.Active) {
					t.Errorf("FindProductByStripeId() got = %v, want %v", got.Active, tt.want.Active)
				}
				if !reflect.DeepEqual(got.Name, tt.want.Name) {
					t.Errorf("FindProductByStripeId() got = %v, want %v", got.Name, tt.want.Name)
				}
			})
		}
		return test.EndTestErr
	})
}
func TestUpsertCustomerStripeId(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "test@example.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		type args struct {
			ctx              context.Context
			dbx              db.Dbx
			userId           uuid.UUID
			stripeCustomerId string
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "insert new customer",
				args: args{
					ctx:              ctx,
					dbx:              dbxx,
					userId:           user.ID,
					stripeCustomerId: "cus_test123",
				},
				wantErr: false,
			},
			{
				name: "update existing customer",
				args: args{
					ctx:              ctx,
					dbx:              dbxx,
					userId:           user.ID,
					stripeCustomerId: "cus_test456",
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := queries.UpsertCustomerStripeId(tt.args.ctx, tt.args.dbx, tt.args.userId, tt.args.stripeCustomerId)
				if (err != nil) != tt.wantErr {
					t.Errorf("UpsertCustomerStripeId() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				// Verify customer was created/updated
				customer, err := queries.FindCustomerByUserId(tt.args.ctx, tt.args.dbx, tt.args.userId)
				if err != nil {
					t.Errorf("Failed to verify customer: %v", err)
					return
				}
				if customer.StripeID != tt.args.stripeCustomerId {
					t.Errorf("Customer stripe ID = %v, want %v", customer.StripeID, tt.args.stripeCustomerId)
				}
			})
		}
		return test.EndTestErr
	})
}
