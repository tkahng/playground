package queries_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/test"
	"github.com/tkahng/authgo/internal/tools/utils"
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
func TestUpsertProduct(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		description := "Test Description"
		image := "test-image.jpg"
		metadata := map[string]string{"key": "value"}

		type args struct {
			ctx     context.Context
			dbx     db.Dbx
			product *models.StripeProduct
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "insert new product",
				args: args{
					ctx: ctx,
					dbx: dbxx,
					product: &models.StripeProduct{
						ID:          "prod_test123",
						Active:      true,
						Name:        "Test Product",
						Description: &description,
						Image:       &image,
						Metadata:    metadata,
					},
				},
				wantErr: false,
			},
			{
				name: "update existing product",
				args: args{
					ctx: ctx,
					dbx: dbxx,
					product: &models.StripeProduct{
						ID:          "prod_test123",
						Active:      false,
						Name:        "Updated Product",
						Description: &description,
						Image:       &image,
						Metadata:    metadata,
					},
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := queries.UpsertProduct(tt.args.ctx, tt.args.dbx, tt.args.product)
				if (err != nil) != tt.wantErr {
					t.Errorf("UpsertProduct() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				// Verify product was created/updated
				product, err := queries.FindProductByStripeId(tt.args.ctx, tt.args.dbx, tt.args.product.ID)
				if err != nil {
					t.Errorf("Failed to verify product: %v", err)
					return
				}
				if product.Name != tt.args.product.Name {
					t.Errorf("Product name = %v, want %v", product.Name, tt.args.product.Name)
				}
				if product.Active != tt.args.product.Active {
					t.Errorf("Product active = %v, want %v", product.Active, tt.args.product.Active)
				}
			})
		}
		return test.EndTestErr
	})
}
func TestUpsertProductFromStripe(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		type args struct {
			ctx     context.Context
			dbx     db.Dbx
			product *stripe.Product
		}

		description := "Test Description"
		images := []string{"test-image.jpg"}
		metadata := map[string]string{"key": "value"}

		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "nil product",
				args: args{
					ctx:     ctx,
					dbx:     dbxx,
					product: nil,
				},
				wantErr: false,
			},
			{
				name: "valid product",
				args: args{
					ctx: ctx,
					dbx: dbxx,
					product: &stripe.Product{
						ID:          "prod_test123",
						Active:      true,
						Name:        "Test Product",
						Description: description,
						Images:      images,
						Metadata:    metadata,
					},
				},
				wantErr: false,
			},
			{
				name: "product without image",
				args: args{
					ctx: ctx,
					dbx: dbxx,
					product: &stripe.Product{
						ID:          "prod_test456",
						Active:      true,
						Name:        "Test Product No Image",
						Description: description,
						Images:      []string{},
						Metadata:    metadata,
					},
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := queries.UpsertProductFromStripe(tt.args.ctx, tt.args.dbx, tt.args.product)
				if (err != nil) != tt.wantErr {
					t.Errorf("UpsertProductFromStripe() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if tt.args.product != nil {
					// Verify product was created/updated
					product, err := queries.FindProductByStripeId(tt.args.ctx, tt.args.dbx, tt.args.product.ID)
					if err != nil {
						t.Errorf("Failed to verify product: %v", err)
						return
					}
					if product.Name != tt.args.product.Name {
						t.Errorf("Product name = %v, want %v", product.Name, tt.args.product.Name)
					}
					if product.Active != tt.args.product.Active {
						t.Errorf("Product active = %v, want %v", product.Active, tt.args.product.Active)
					}
					if *product.Description != tt.args.product.Description {
						t.Errorf("Product description = %v, want %v", *product.Description, tt.args.product.Description)
					}
					if len(tt.args.product.Images) > 0 {
						if *product.Image != tt.args.product.Images[0] {
							t.Errorf("Product image = %v, want %v", *product.Image, tt.args.product.Images[0])
						}
					}
				}
			})
		}
		return test.EndTestErr
	})
}

func TestUpsertPrice(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create a test product first
		product := &models.StripeProduct{
			ID:       "prod_test123",
			Active:   true,
			Name:     "Test Product",
			Metadata: map[string]string{"key": "value"},
		}
		err := queries.UpsertProduct(ctx, dbxx, product)
		if err != nil {
			t.Fatalf("failed to create test product: %v", err)
		}

		lookupKey := "test_key"
		unitAmount := int64(1000)
		interval := models.StripePricingPlanIntervalMonth
		intervalCount := int64(1)
		trialDays := int64(7)
		metadata := map[string]string{"key": "value"}

		type args struct {
			ctx   context.Context
			dbx   db.Dbx
			price *models.StripePrice
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "insert new price",
				args: args{
					ctx: ctx,
					dbx: dbxx,
					price: &models.StripePrice{
						ID:              "price_test123",
						ProductID:       product.ID,
						LookupKey:       &lookupKey,
						Active:          true,
						UnitAmount:      &unitAmount,
						Currency:        "usd",
						Type:            models.StripePricingTypeRecurring,
						Interval:        &interval,
						IntervalCount:   &intervalCount,
						TrialPeriodDays: &trialDays,
						Metadata:        metadata,
					},
				},
				wantErr: false,
			},
			{
				name: "update existing price",
				args: args{
					ctx: ctx,
					dbx: dbxx,
					price: &models.StripePrice{
						ID:              "price_test123",
						ProductID:       product.ID,
						LookupKey:       &lookupKey,
						Active:          false,
						UnitAmount:      &unitAmount,
						Currency:        "eur",
						Type:            models.StripePricingTypeRecurring,
						Interval:        &interval,
						IntervalCount:   &intervalCount,
						TrialPeriodDays: &trialDays,
						Metadata:        metadata,
					},
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := queries.UpsertPrice(tt.args.ctx, tt.args.dbx, tt.args.price)
				if (err != nil) != tt.wantErr {
					t.Errorf("UpsertPrice() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				// Verify price was created/updated
				price, err := queries.FindValidPriceById(tt.args.ctx, tt.args.dbx, tt.args.price.ID)
				if err != nil {
					t.Errorf("Failed to verify price: %v", err)
					return
				}
				if price.ProductID != tt.args.price.ProductID {
					t.Errorf("Price product_id = %v, want %v", price.ProductID, tt.args.price.ProductID)
				}
				if price.Active != tt.args.price.Active {
					t.Errorf("Price active = %v, want %v", price.Active, tt.args.price.Active)
				}
				if price.Currency != tt.args.price.Currency {
					t.Errorf("Price currency = %v, want %v", price.Currency, tt.args.price.Currency)
				}
			})
		}
		return test.EndTestErr
	})
}
func TestUpsertPriceFromStripe(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create a test product first
		meta := map[string]string{"key": "value"}
		product := &models.StripeProduct{
			ID:       "prod_test123",
			Active:   true,
			Name:     "Test Product",
			Metadata: meta,
		}
		err := queries.UpsertProduct(ctx, dbxx, product)
		if err != nil {
			t.Fatalf("failed to create test product: %v", err)
		}

		unitAmount := int64(1000)
		recurring := &stripe.PriceRecurring{
			Interval:        "month",
			IntervalCount:   1,
			TrialPeriodDays: 7,
		}

		type args struct {
			ctx   context.Context
			dbx   db.Dbx
			price *stripe.Price
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "nil price",
				args: args{
					ctx:   ctx,
					dbx:   dbxx,
					price: nil,
				},
				wantErr: false,
			},
			{
				name: "valid price with recurring",
				args: args{
					ctx: ctx,
					dbx: dbxx,
					price: &stripe.Price{
						ID:         "price_test123",
						Product:    &stripe.Product{ID: product.ID},
						Active:     true,
						LookupKey:  "test_key_1",
						UnitAmount: unitAmount,
						Currency:   "usd",
						Type:       "recurring",
						Recurring:  recurring,
						Metadata:   meta,
					},
				},
				wantErr: false,
			},
			{
				name: "price without recurring",
				args: args{
					ctx: ctx,
					dbx: dbxx,
					price: &stripe.Price{
						ID:         "price_test456",
						Product:    &stripe.Product{ID: product.ID},
						Active:     true,
						LookupKey:  "test_key_2",
						UnitAmount: unitAmount,
						Currency:   "usd",
						Type:       "one_time",
						Metadata:   meta,
					},
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := queries.UpsertPriceFromStripe(tt.args.ctx, tt.args.dbx, tt.args.price)
				utils.PrettyPrintJSON(err)
				if (err != nil) != tt.wantErr {
					t.Errorf("UpsertPriceFromStripe() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if tt.args.price != nil {
					// Verify price was created/updated
					price, err := repository.StripePrice.GetOne(
						ctx,
						dbxx,
						&map[string]any{
							"id": map[string]any{
								"eq": tt.args.price.ID,
							},
						},
					)
					if err != nil {
						t.Errorf("Failed to verify price: %v", err)
						return
					}
					if price == nil {
						t.Errorf("Price not found in database")
						return
					}
					if tt.args.price.Product != nil {
						if price.ProductID != tt.args.price.Product.ID {
							utils.PrettyPrintJSON(price)
							t.Errorf("Price product_id = %v, want %v", price.ProductID, tt.args.price.Product.ID)
						}
					}
					if price.Active != tt.args.price.Active {
						t.Errorf("Price active = %v, want %v", price.Active, tt.args.price.Active)
					}
					if *price.UnitAmount != tt.args.price.UnitAmount {
						t.Errorf("Price unit_amount = %v, want %v", *price.UnitAmount, tt.args.price.UnitAmount)
					}
					if price.Currency != string(tt.args.price.Currency) {
						t.Errorf("Price currency = %v, want %v", price.Currency, tt.args.price.Currency)
					}
					if tt.args.price.Recurring != nil {
						if *price.Interval != models.StripePricingPlanInterval(tt.args.price.Recurring.Interval) {
							t.Errorf("Price interval = %v, want %v", *price.Interval, tt.args.price.Recurring.Interval)
						}
					}
				}
			})
		}
		return test.EndTestErr
	})
}
