package queries_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/test"
	"github.com/tkahng/authgo/internal/tools/types"
)

func TestFindCustomerByStripeId(t *testing.T) {
	test.Short(t)
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
	test.Short(t)
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
	test.Short(t)
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
	test.Short(t)
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
	test.Short(t)
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
	test.Short(t)
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
	test.Short(t)
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
	test.Short(t)
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
				if (err != nil) != tt.wantErr {
					t.Errorf("UpsertPriceFromStripe() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if tt.args.price != nil {
					// Verify price was created/updated
					price, err := crudrepo.StripePrice.GetOne(
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

func TestUpsertSubscription(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "test@example.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		if user == nil {
			t.Fatalf("expected user to be created, got nil")
		}
		userId := user.ID
		now := time.Now()

		// Create test product and price first
		product := &models.StripeProduct{
			ID:       "prod_test123",
			Active:   true,
			Name:     "Test Product",
			Metadata: map[string]string{"key": "value"},
		}
		err = queries.UpsertProduct(ctx, dbxx, product)
		if err != nil {
			t.Fatalf("failed to create test product: %v", err)
		}

		price := &models.StripePrice{
			ID:        "price_test123",
			ProductID: product.ID,
			Active:    true,
			Currency:  "usd",
			Type:      models.StripePricingTypeRecurring,
			Metadata:  map[string]string{"key": "value"},
		}
		err = queries.UpsertPrice(ctx, dbxx, price)
		if err != nil {
			t.Fatalf("failed to create test price: %v", err)
		}

		type args struct {
			ctx          context.Context
			dbx          db.Dbx
			subscription *models.StripeSubscription
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "insert new subscription",
				args: args{
					ctx: ctx,
					dbx: dbxx,
					subscription: &models.StripeSubscription{
						ID:                 "sub_test123",
						UserID:             nil,
						Status:             models.StripeSubscriptionStatusActive,
						PriceID:            price.ID,
						Quantity:           1,
						CancelAtPeriodEnd:  false,
						Created:            now,
						CurrentPeriodStart: now,
						CurrentPeriodEnd:   now.Add(30 * 24 * time.Hour),
						Metadata:           map[string]string{"key": "value"},
					},
				},
				wantErr: false,
			},
			{
				name: "update existing subscription",
				args: args{
					ctx: ctx,
					dbx: dbxx,
					subscription: &models.StripeSubscription{
						ID:                 "sub_test123",
						UserID:             types.Pointer(userId),
						Status:             models.StripeSubscriptionStatusCanceled,
						PriceID:            price.ID,
						Quantity:           2,
						CancelAtPeriodEnd:  true,
						Created:            now,
						CurrentPeriodStart: now,
						CurrentPeriodEnd:   now.Add(30 * 24 * time.Hour),
						CanceledAt:         &now,
						Metadata:           map[string]string{"key": "updated"},
					},
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := queries.UpsertSubscription(tt.args.ctx, tt.args.dbx, tt.args.subscription)
				if (err != nil) != tt.wantErr {
					t.Errorf("UpsertSubscription() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				// Verify subscription was created/updated
				sub, err := queries.FindSubscriptionById(tt.args.ctx, tt.args.dbx, tt.args.subscription.ID)
				if err != nil {
					t.Errorf("Failed to verify subscription: %v", err)
					return
				}
				if sub.ID != tt.args.subscription.ID {
					t.Errorf("Subscription ID = %v, want %v", sub.ID, tt.args.subscription.ID)
				}
				if sub.Status != tt.args.subscription.Status {
					t.Errorf("Subscription status = %v, want %v", sub.Status, tt.args.subscription.Status)
				}
				if sub.PriceID != tt.args.subscription.PriceID {
					t.Errorf("Subscription price_id = %v, want %v", sub.PriceID, tt.args.subscription.PriceID)
				}
				if sub.Quantity != tt.args.subscription.Quantity {
					t.Errorf("Subscription quantity = %v, want %v", sub.Quantity, tt.args.subscription.Quantity)
				}
				if sub.CancelAtPeriodEnd != tt.args.subscription.CancelAtPeriodEnd {
					t.Errorf("Subscription cancel_at_period_end = %v, want %v", sub.CancelAtPeriodEnd, tt.args.subscription.CancelAtPeriodEnd)
				}
			})
		}
		return test.EndTestErr
	})
}

func TestUpsertSubscriptionFromStripe(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "test@example.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		// Create test product and price first
		product := &models.StripeProduct{
			ID:       "prod_test123",
			Active:   true,
			Name:     "Test Product",
			Metadata: map[string]string{"key": "value"},
		}
		err = queries.UpsertProduct(ctx, dbxx, product)
		if err != nil {
			t.Fatalf("failed to create test product: %v", err)
		}

		price := &models.StripePrice{
			ID:        "price_test123",
			ProductID: product.ID,
			Active:    true,
			Currency:  "usd",
			Type:      models.StripePricingTypeRecurring,
			Metadata:  map[string]string{"key": "value"},
		}
		err = queries.UpsertPrice(ctx, dbxx, price)
		if err != nil {
			t.Fatalf("failed to create test price: %v", err)
		}

		timestamp := time.Now().Unix()

		type args struct {
			ctx    context.Context
			dbx    db.Dbx
			sub    *stripe.Subscription
			userId uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "nil subscription",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					sub:    nil,
					userId: user.ID,
				},
				wantErr: false,
			},
			{
				name: "subscription without items",
				args: args{
					ctx: ctx,
					dbx: dbxx,
					sub: &stripe.Subscription{
						ID:     "sub_test123",
						Status: "active",
						Items:  &stripe.SubscriptionItemList{},
					},
					userId: user.ID,
				},
				wantErr: true,
			},
			{
				name: "valid subscription",
				args: args{
					ctx: ctx,
					dbx: dbxx,
					sub: &stripe.Subscription{
						ID:                "sub_test456",
						Status:            "active",
						Created:           timestamp,
						CancelAtPeriodEnd: false,
						EndedAt:           timestamp,
						CancelAt:          timestamp,
						CanceledAt:        timestamp,
						TrialStart:        timestamp,
						TrialEnd:          timestamp,
						Metadata:          map[string]string{"key": "value"},
						Items: &stripe.SubscriptionItemList{
							Data: []*stripe.SubscriptionItem{
								{
									Price:              &stripe.Price{ID: price.ID},
									Quantity:           1,
									CurrentPeriodStart: timestamp,
									CurrentPeriodEnd:   timestamp,
								},
							},
						},
					},
					userId: user.ID,
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := queries.UpsertSubscriptionFromStripe(tt.args.ctx, tt.args.dbx, tt.args.sub, tt.args.userId)
				if (err != nil) != tt.wantErr {
					t.Errorf("UpsertSubscriptionFromStripe() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if tt.args.sub != nil && !tt.wantErr {
					// Verify subscription was created/updated
					sub, err := queries.FindSubscriptionById(tt.args.ctx, tt.args.dbx, tt.args.sub.ID)
					if err != nil {
						t.Errorf("Failed to verify subscription: %v", err)
						return
					}
					if sub.ID != tt.args.sub.ID {
						t.Errorf("Subscription ID = %v, want %v", sub.ID, tt.args.sub.ID)
					}
					if sub.Status != models.StripeSubscriptionStatus(tt.args.sub.Status) {
						t.Errorf("Subscription status = %v, want %v", sub.Status, tt.args.sub.Status)
					}
					if len(tt.args.sub.Items.Data) > 0 {
						if sub.PriceID != tt.args.sub.Items.Data[0].Price.ID {
							t.Errorf("Subscription price_id = %v, want %v", sub.PriceID, tt.args.sub.Items.Data[0].Price.ID)
						}
						if sub.Quantity != tt.args.sub.Items.Data[0].Quantity {
							t.Errorf("Subscription quantity = %v, want %v", sub.Quantity, tt.args.sub.Items.Data[0].Quantity)
						}
					}
				}
			})
		}
		return test.EndTestErr
	})
}
func TestFindSubscriptionById(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create test user
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "test@example.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		product := &models.StripeProduct{
			ID:       "prod_test123",
			Active:   true,
			Name:     "Test Product",
			Metadata: map[string]string{"key": "value"},
		}
		err = queries.UpsertProduct(ctx, dbxx, product)
		if err != nil {
			t.Fatalf("failed to create test product: %v", err)
		}
		price := &models.StripePrice{
			ID:        "price_test123",
			ProductID: product.ID,
			Active:    true,
			Currency:  "usd",
			Type:      models.StripePricingTypeRecurring,
			Metadata:  map[string]string{"key": "value"},
		}
		err = queries.UpsertPrice(ctx, dbxx, price)
		if err != nil {
			t.Fatalf("failed to create test price: %v", err)
		}

		// Create test subscription
		testSub := &models.StripeSubscription{
			ID:                 "sub_test123",
			UserID:             types.Pointer(user.ID),
			Status:             models.StripeSubscriptionStatusActive,
			PriceID:            price.ID,
			Quantity:           1,
			CancelAtPeriodEnd:  false,
			Created:            time.Now(),
			CurrentPeriodStart: time.Now(),
			CurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour),
			Metadata:           map[string]string{"key": "value"},
		}

		err = queries.UpsertSubscription(ctx, dbxx, testSub)
		if err != nil {
			t.Fatalf("failed to create test subscription: %v", err)
		}

		type args struct {
			ctx      context.Context
			dbx      db.Dbx
			stripeId string
		}
		tests := []struct {
			name    string
			args    args
			want    *models.StripeSubscription
			wantErr bool
		}{
			{
				name: "existing subscription",
				args: args{
					ctx:      ctx,
					dbx:      dbxx,
					stripeId: testSub.ID,
				},
				want:    testSub,
				wantErr: false,
			},
			{
				name: "non-existent subscription",
				args: args{
					ctx:      ctx,
					dbx:      dbxx,
					stripeId: "sub_nonexistent",
				},
				want:    nil,
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.FindSubscriptionById(tt.args.ctx, tt.args.dbx, tt.args.stripeId)
				if (err != nil) != tt.wantErr {
					t.Errorf("FindSubscriptionById() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.want == nil {
					if got != nil {
						t.Errorf("FindSubscriptionById() = %v, want nil", got)
					}
					return
				}
				if got == nil {
					t.Errorf("FindSubscriptionById() = nil, want %v", tt.want)
					return
				}
				if got.ID != tt.want.ID {
					t.Errorf("FindSubscriptionById() got ID = %v, want %v", got.ID, tt.want.ID)
				}
				if got.Status != tt.want.Status {
					t.Errorf("FindSubscriptionById() got Status = %v, want %v", got.Status, tt.want.Status)
				}
				if got.PriceID != tt.want.PriceID {
					t.Errorf("FindSubscriptionById() got PriceID = %v, want %v", got.PriceID, tt.want.PriceID)
				}
			})
		}
		return test.EndTestErr
	})
}
func TestFindSubscriptionWithPriceById(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create test user
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "test@example.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		product := &models.StripeProduct{
			ID:       "prod_test123",
			Active:   true,
			Name:     "Test Product",
			Metadata: map[string]string{"key": "value"},
		}
		err = queries.UpsertProduct(ctx, dbxx, product)
		if err != nil {
			t.Fatalf("failed to create test product: %v", err)
		}
		price := &models.StripePrice{
			ID:        "price_test123",
			ProductID: product.ID,
			Active:    true,
			Currency:  "usd",
			Type:      models.StripePricingTypeRecurring,
			Metadata:  map[string]string{"key": "value"},
		}
		err = queries.UpsertPrice(ctx, dbxx, price)
		if err != nil {
			t.Fatalf("failed to create test price: %v", err)
		}

		// Create test subscription
		testSub := &models.StripeSubscription{
			ID:                 "sub_test123",
			UserID:             types.Pointer(user.ID),
			Status:             models.StripeSubscriptionStatusActive,
			PriceID:            price.ID,
			Quantity:           1,
			CancelAtPeriodEnd:  false,
			Created:            time.Now(),
			CurrentPeriodStart: time.Now(),
			CurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour),
			Metadata:           map[string]string{"key": "value"},
		}

		err = queries.UpsertSubscription(ctx, dbxx, testSub)
		if err != nil {
			t.Fatalf("failed to create test subscription: %v", err)
		}

		type args struct {
			ctx      context.Context
			dbx      db.Dbx
			stripeId string
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
			check   func(*testing.T, *models.SubscriptionWithPrice, error)
		}{
			{
				name: "existing subscription",
				args: args{
					ctx:      ctx,
					dbx:      dbxx,
					stripeId: testSub.ID,
				},
				wantErr: false,
				check: func(t *testing.T, got *models.SubscriptionWithPrice, err error) {
					if err != nil {
						t.Errorf("unexpected error: %v", err)
						return
					}
					if got == nil {
						t.Error("expected subscription data, got nil")
						return
					}
					if got.Subscription.ID != testSub.ID {
						t.Errorf("expected subscription ID %v, got %v", testSub.ID, got.Subscription.ID)
					}
					if got.Price.ID != price.ID {
						t.Errorf("expected price ID %v, got %v", price.ID, got.Price.ID)
					}
					if got.Product.ID != product.ID {
						t.Errorf("expected product ID %v, got %v", product.ID, got.Product.ID)
					}
				},
			},
			{
				name: "non-existent subscription",
				args: args{
					ctx:      ctx,
					dbx:      dbxx,
					stripeId: "sub_nonexistent",
				},
				wantErr: false,
				check: func(t *testing.T, got *models.SubscriptionWithPrice, err error) {
					if err != nil {
						t.Errorf("unexpected error: %v", err)
						return
					}
					if got != nil {
						t.Errorf("expected nil for non-existent subscription, got %v", got)
					}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.FindSubscriptionWithPriceById(tt.args.ctx, tt.args.dbx, tt.args.stripeId)
				if (err != nil) != tt.wantErr {
					t.Errorf("FindSubscriptionWithPriceById() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				tt.check(t, got, err)
			})
		}
		return test.EndTestErr
	})
}
func TestFindLatestActiveSubscriptionByUserId(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create test user
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "test@example.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		// Create test product and price
		product := &models.StripeProduct{
			ID:       "prod_test123",
			Active:   true,
			Name:     "Test Product",
			Metadata: map[string]string{"key": "value"},
		}
		err = queries.UpsertProduct(ctx, dbxx, product)
		if err != nil {
			t.Fatalf("failed to create test product: %v", err)
		}

		price := &models.StripePrice{
			ID:        "price_test123",
			ProductID: product.ID,
			Active:    true,
			Currency:  "usd",
			Type:      models.StripePricingTypeRecurring,
			Metadata:  map[string]string{"key": "value"},
		}
		err = queries.UpsertPrice(ctx, dbxx, price)
		if err != nil {
			t.Fatalf("failed to create test price: %v", err)
		}

		// Create test subscriptions
		activeSub := &models.StripeSubscription{
			ID:                 "sub_active",
			UserID:             types.Pointer(user.ID),
			Status:             models.StripeSubscriptionStatusActive,
			PriceID:            price.ID,
			Quantity:           1,
			CancelAtPeriodEnd:  false,
			Created:            time.Now(),
			CurrentPeriodStart: time.Now(),
			CurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour),
			Metadata:           map[string]string{"key": "value"},
		}

		canceledSub := &models.StripeSubscription{
			ID:                 "sub_canceled",
			UserID:             types.Pointer(user.ID),
			Status:             models.StripeSubscriptionStatusCanceled,
			PriceID:            price.ID,
			Quantity:           1,
			CancelAtPeriodEnd:  true,
			Created:            time.Now(),
			CurrentPeriodStart: time.Now(),
			CurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour),
			Metadata:           map[string]string{"key": "value"},
		}

		// Insert subscriptions in different order
		for _, sub := range []*models.StripeSubscription{canceledSub, activeSub} {
			err = queries.UpsertSubscription(ctx, dbxx, sub)
			if err != nil {
				t.Fatalf("failed to create test subscription: %v", err)
			}
		}

		type args struct {
			ctx    context.Context
			dbx    db.Dbx
			userId uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			want    *models.StripeSubscription
			wantErr bool
		}{
			{
				name: "user with active subscription",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					userId: user.ID,
				},
				want:    activeSub, // Should get latest active subscription
				wantErr: false,
			},
			{
				name: "non-existent user",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					userId: uuid.New(),
				},
				want:    nil,
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.FindLatestActiveSubscriptionByUserId(tt.args.ctx, tt.args.dbx, tt.args.userId)
				if (err != nil) != tt.wantErr {
					t.Errorf("FindLatestActiveSubscriptionByUserId() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.want == nil {
					if got != nil {
						t.Errorf("FindLatestActiveSubscriptionByUserId() = %v, want nil", got)
					}
					return
				}
				if got == nil {
					t.Errorf("FindLatestActiveSubscriptionByUserId() = nil, want %v", tt.want)
					return
				}
				if got.ID != tt.want.ID {
					t.Errorf("FindLatestActiveSubscriptionByUserId() got ID = %v, want %v", got.ID, tt.want.ID)
				}
				if got.Status != tt.want.Status {
					t.Errorf("FindLatestActiveSubscriptionByUserId() got Status = %v, want %v", got.Status, tt.want.Status)
				}
				if got.UserID != tt.want.UserID {
					t.Errorf("FindLatestActiveSubscriptionByUserId() got UserID = %v, want %v", got.UserID, tt.want.UserID)
				}
			})
		}
		return test.EndTestErr
	})
}
func TestFindLatestActiveSubscriptionWithPriceByUserId(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create test user
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "test@example.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		// Create test product
		product := &models.StripeProduct{
			ID:       "prod_test123",
			Active:   true,
			Name:     "Test Product",
			Metadata: map[string]string{"key": "value"},
		}
		err = queries.UpsertProduct(ctx, dbxx, product)
		if err != nil {
			t.Fatalf("failed to create test product: %v", err)
		}

		// Create test price
		price := &models.StripePrice{
			ID:        "price_test123",
			ProductID: product.ID,
			Active:    true,
			Currency:  "usd",
			Type:      models.StripePricingTypeRecurring,
			Metadata:  map[string]string{"key": "value"},
		}
		err = queries.UpsertPrice(ctx, dbxx, price)
		if err != nil {
			t.Fatalf("failed to create test price: %v", err)
		}

		// Create active subscription
		activeSub := &models.StripeSubscription{
			ID:                 "sub_active",
			UserID:             types.Pointer(user.ID),
			Status:             models.StripeSubscriptionStatusActive,
			PriceID:            price.ID,
			Quantity:           1,
			CancelAtPeriodEnd:  false,
			Created:            time.Now(),
			CurrentPeriodStart: time.Now(),
			CurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour),
			Metadata:           map[string]string{"key": "value"},
		}
		err = queries.UpsertSubscription(ctx, dbxx, activeSub)
		if err != nil {
			t.Fatalf("failed to create active subscription: %v", err)
		}

		type args struct {
			ctx    context.Context
			dbx    db.Dbx
			userId uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
			check   func(*testing.T, *models.SubscriptionWithPrice, error)
		}{
			{
				name: "user with active subscription",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					userId: user.ID,
				},
				wantErr: false,
				check: func(t *testing.T, got *models.SubscriptionWithPrice, err error) {
					if err != nil {
						t.Errorf("unexpected error: %v", err)
						return
					}
					if got == nil {
						t.Error("expected subscription data, got nil")
						return
					}
					if got.Subscription.ID != activeSub.ID {
						t.Errorf("expected subscription ID %v, got %v", activeSub.ID, got.Subscription.ID)
					}
					if got.Price.ID != price.ID {
						t.Errorf("expected price ID %v, got %v", price.ID, got.Price.ID)
					}
					if got.Product.ID != product.ID {
						t.Errorf("expected product ID %v, got %v", product.ID, got.Product.ID)
					}
				},
			},
			{
				name: "non-existent user",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					userId: uuid.New(),
				},
				wantErr: false,
				check: func(t *testing.T, got *models.SubscriptionWithPrice, err error) {
					if err != nil {
						t.Errorf("unexpected error: %v", err)
						return
					}
					if got != nil {
						t.Errorf("expected nil for non-existent user, got %v", got)
					}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.FindLatestActiveSubscriptionWithPriceByUserId(tt.args.ctx, tt.args.dbx, tt.args.userId)
				if (err != nil) != tt.wantErr {
					t.Errorf("FindLatestActiveSubscriptionWithPriceByUserId() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				tt.check(t, got, err)
			})
		}
		return test.EndTestErr
	})
}
func TestIsFirstSubscription(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create test user
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "test@example.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		// Create test product and price
		product := &models.StripeProduct{
			ID:       "prod_test123",
			Active:   true,
			Name:     "Test Product",
			Metadata: map[string]string{"key": "value"},
		}
		err = queries.UpsertProduct(ctx, dbxx, product)
		if err != nil {
			t.Fatalf("failed to create test product: %v", err)
		}

		price := &models.StripePrice{
			ID:        "price_test123",
			ProductID: product.ID,
			Active:    true,
			Currency:  "usd",
			Type:      models.StripePricingTypeRecurring,
			Metadata:  map[string]string{"key": "value"},
		}
		err = queries.UpsertPrice(ctx, dbxx, price)
		if err != nil {
			t.Fatalf("failed to create test price: %v", err)
		}

		type args struct {
			ctx    context.Context
			dbx    db.Dbx
			userId uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			want    bool
			wantErr bool
			setup   func() error
		}{
			{
				name: "user without subscription",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					userId: user.ID,
				},
				want:    false,
				wantErr: false,
			},
			{
				name: "user with subscription",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					userId: user.ID,
				},
				want:    true,
				wantErr: false,
				setup: func() error {
					subscription := &models.StripeSubscription{
						ID:                 "sub_test123",
						UserID:             types.Pointer(user.ID),
						Status:             models.StripeSubscriptionStatusActive,
						PriceID:            price.ID,
						Quantity:           1,
						CancelAtPeriodEnd:  false,
						Created:            time.Now(),
						CurrentPeriodStart: time.Now(),
						CurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour),
						Metadata:           map[string]string{"key": "value"},
					}
					return queries.UpsertSubscription(ctx, dbxx, subscription)
				},
			},
			{
				name: "non-existent user",
				args: args{
					ctx:    ctx,
					dbx:    dbxx,
					userId: uuid.New(),
				},
				want:    false,
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.setup != nil {
					if err := tt.setup(); err != nil {
						t.Fatalf("test setup failed: %v", err)
					}
				}

				got, err := queries.IsFirstSubscription(tt.args.ctx, tt.args.dbx, tt.args.userId)
				if (err != nil) != tt.wantErr {
					t.Errorf("IsFirstSubscription() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("IsFirstSubscription() = %v, want %v", got, tt.want)
				}
			})
		}
		return test.EndTestErr
	})
}
func TestFindValidPriceById(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create test product first
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

		// Create recurring price
		recurringPrice := &models.StripePrice{
			ID:        "price_recurring",
			ProductID: product.ID,
			Active:    true,
			Currency:  "usd",
			Type:      models.StripePricingTypeRecurring,
			Metadata:  map[string]string{"key": "value"},
		}
		err = queries.UpsertPrice(ctx, dbxx, recurringPrice)
		if err != nil {
			t.Fatalf("failed to create recurring price: %v", err)
		}

		// Create one-time price
		oneTimePrice := &models.StripePrice{
			ID:        "price_onetime",
			ProductID: product.ID,
			Active:    true,
			Currency:  "usd",
			Type:      models.StripePricingTypeOneTime,
			Metadata:  map[string]string{"key": "value"},
		}
		err = queries.UpsertPrice(ctx, dbxx, oneTimePrice)
		if err != nil {
			t.Fatalf("failed to create one-time price: %v", err)
		}

		type args struct {
			ctx     context.Context
			dbx     db.Dbx
			priceId string
		}
		tests := []struct {
			name    string
			args    args
			want    *models.StripePrice
			wantErr bool
		}{
			{
				name: "valid recurring price",
				args: args{
					ctx:     ctx,
					dbx:     dbxx,
					priceId: recurringPrice.ID,
				},
				want:    recurringPrice,
				wantErr: false,
			},
			{
				name: "non-recurring price",
				args: args{
					ctx:     ctx,
					dbx:     dbxx,
					priceId: oneTimePrice.ID,
				},
				want:    nil,
				wantErr: false,
			},
			{
				name: "non-existent price",
				args: args{
					ctx:     ctx,
					dbx:     dbxx,
					priceId: "price_nonexistent",
				},
				want:    nil,
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.FindValidPriceById(tt.args.ctx, tt.args.dbx, tt.args.priceId)
				if (err != nil) != tt.wantErr {
					t.Errorf("FindValidPriceById() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.want == nil {
					if got != nil {
						t.Errorf("FindValidPriceById() = %v, want nil", got)
					}
					return
				}
				if got == nil {
					t.Errorf("FindValidPriceById() = nil, want %v", tt.want)
					return
				}
				if got.ID != tt.want.ID {
					t.Errorf("FindValidPriceById() got ID = %v, want %v", got.ID, tt.want.ID)
				}
				if got.Type != tt.want.Type {
					t.Errorf("FindValidPriceById() got Type = %v, want %v", got.Type, tt.want.Type)
				}
				if got.ProductID != tt.want.ProductID {
					t.Errorf("FindValidPriceById() got ProductID = %v, want %v", got.ProductID, tt.want.ProductID)
				}
			})
		}
		return test.EndTestErr
	})
}
