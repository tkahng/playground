package stores_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
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

func TestPostgresStripeStore_ProductAndPrice(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		store := stores.NewPostgresStripeStore(dbxx)

		// UpsertProduct
		product := &models.StripeProduct{
			ID:     "prod_123",
			Active: true,
			Name:   "Test Product",
			Metadata: map[string]string{
				"key1": "value1",
			},
		}
		err := store.UpsertProduct(ctx, product)
		if err != nil {
			t.Fatalf("UpsertProduct() error = %v", err)
		}

		// FindProductByStripeId
		found, err := store.FindProductByStripeId(ctx, "prod_123")
		if err != nil {
			t.Fatalf("FindProductByStripeId() error = %v", err)
		}
		if found == nil || found.ID != product.ID {
			t.Errorf("FindProductByStripeId() = %v, want %v", found, product.ID)
		}

		// UpsertPrice
		price := &models.StripePrice{
			ID:         "price_123",
			ProductID:  product.ID,
			Active:     true,
			UnitAmount: types.Pointer(int64(1000)),
			Currency:   "usd",
			Type:       models.StripePricingTypeRecurring,
			Metadata: map[string]string{
				"key1": "value1",
			},
		}
		err = store.UpsertPrice(ctx, price)
		if err != nil {
			t.Fatalf("UpsertPrice() error = %v", err)
		}

		// FindValidPriceById
		validPrice, err := store.FindValidPriceById(ctx, "price_123")
		if err != nil {
			t.Fatalf("FindValidPriceById() error = %v", err)
		}
		if validPrice == nil || validPrice.ID != price.ID {
			t.Errorf("FindValidPriceById() = %v, want %v", validPrice, price.ID)
		}

		// ListProducts
		products, err := store.ListProducts(ctx, &shared.StripeProductListParams{})
		if err != nil {
			t.Fatalf("ListProducts() error = %v", err)
		}
		if len(products) == 0 {
			t.Errorf("ListProducts() = %v, want at least 1", products)
		}

		// ListPrices
		prices, err := store.ListPrices(ctx, &shared.StripePriceListParams{})
		if err != nil {
			t.Fatalf("ListPrices() error = %v", err)
		}
		if len(prices) == 0 {
			t.Errorf("ListPrices() = %v, want at least 1", prices)
		}

		return errors.New("rollback")
	})
}

func TestPostgresStripeStore_UpsertProductAndPriceFromStripe(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		store := stores.NewPostgresStripeStore(dbxx)
		stripeProduct := &stripe.Product{
			ID:          "prod_stripe_1",
			Active:      true,
			Name:        "Stripe Product",
			Description: "Stripe Desc",
			Images:      []string{"img1.jpg"},
			Metadata:    map[string]string{"foo": "bar"},
		}
		err := store.UpsertProductFromStripe(ctx, stripeProduct)
		if err != nil {
			t.Fatalf("UpsertProductFromStripe() error = %v", err)
		}
		found, err := store.FindProductByStripeId(ctx, stripeProduct.ID)
		if err != nil || found == nil || found.ID != stripeProduct.ID {
			t.Errorf("FindProductByStripeId() = %v, err = %v", found, err)
		}

		stripePrice := &stripe.Price{
			ID:         "price_stripe_1",
			Product:    &stripe.Product{ID: stripeProduct.ID},
			Active:     true,
			LookupKey:  "lookup_1",
			UnitAmount: 5000,
			Currency:   "usd",
			Type:       "recurring",
			Metadata:   map[string]string{"foo": "bar"},
			Recurring: &stripe.PriceRecurring{
				Interval:        "month",
				IntervalCount:   1,
				TrialPeriodDays: 14,
			},
		}
		err = store.UpsertPriceFromStripe(ctx, stripePrice)
		if err != nil {
			t.Fatalf("UpsertPriceFromStripe() error = %v", err)
		}
		return errors.New("rollback")
	})
}

func TestPostgresStripeStore_FindCustomer(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		userStore := stores.NewPostgresUserStore(dbxx)
		user, err := userStore.CreateUser(ctx, &models.User{Email: "findcustomer@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		store := stores.NewPostgresStripeStore(dbxx)
		customer := &models.StripeCustomer{
			ID:           "cus_find_1",
			UserID:       types.Pointer(user.ID),
			Email:        user.Email,
			CustomerType: models.StripeCustomerTypeUser,
		}
		_, err = store.CreateCustomer(ctx, customer)
		if err != nil {
			t.Fatalf("CreateCustomer() error = %v", err)
		}
		found, err := store.FindCustomer(ctx, &models.StripeCustomer{ID: "cus_find_1"})
		if err != nil || found == nil || found.ID != "cus_find_1" {
			t.Errorf("FindCustomer() = %v, err = %v", found, err)
		}
		return errors.New("rollback")
	})
}

func TestPostgresStripeStore_SubscriptionQueries(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		store := stores.NewPostgresStripeStore(dbxx)
		userStore := stores.NewPostgresUserStore(dbxx)
		user, err := userStore.CreateUser(ctx, &models.User{Email: "sub@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		// Insert product and price
		product := &models.StripeProduct{ID: "prod_sub_1", Active: true, Name: "Sub Product", Metadata: map[string]string{}}
		err = store.UpsertProduct(ctx, product)
		if err != nil {
			t.Fatalf("UpsertProduct() error = %v", err)
		}
		price := &models.StripePrice{
			ID:         "price_sub_1",
			ProductID:  product.ID,
			Active:     true,
			UnitAmount: types.Pointer(int64(2000)),
			Currency:   "usd",
			Type:       models.StripePricingTypeRecurring,
			Metadata:   map[string]string{},
		}
		err = store.UpsertPrice(ctx, price)
		if err != nil {
			t.Fatalf("UpsertPrice() error = %v", err)
		}
		// Insert customer
		customer := &models.StripeCustomer{
			ID:           "cus_sub_1",
			Email:        "sub@example.com",
			CustomerType: models.StripeCustomerTypeUser,
			UserID:       types.Pointer(user.ID),
		}
		_, err = store.CreateCustomer(ctx, customer)
		if err != nil {
			t.Fatalf("CreateCustomer() error = %v", err)
		}
		// Insert subscription
		sub := &models.StripeSubscription{
			ID:                 "sub_1",
			StripeCustomerID:   customer.ID,
			Status:             models.StripeSubscriptionStatusActive,
			Metadata:           map[string]string{},
			ItemID:             "item_1",
			PriceID:            price.ID,
			Quantity:           1,
			CancelAtPeriodEnd:  false,
			Created:            time.Now(),
			CurrentPeriodStart: time.Now(),
			CurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour),
		}
		err = store.UpsertSubscription(ctx, sub)
		if err != nil {
			t.Fatalf("UpsertSubscription() error = %v", err)
		}
		// FindSubscriptionWithPriceById
		withPrice, err := store.FindSubscriptionWithPriceById(ctx, "sub_1")
		if err != nil || withPrice == nil || withPrice.Subscription.ID != "sub_1" {
			t.Errorf("FindSubscriptionWithPriceById() = %v, err = %v", withPrice, err)
		}
		// FindLatestActiveSubscriptionWithPriceByCustomerId
		latest, err := store.FindLatestActiveSubscriptionWithPriceByCustomerId(ctx, customer.ID)
		if err != nil || latest == nil || latest.Subscription.ID != "sub_1" {
			t.Errorf("FindLatestActiveSubscriptionWithPriceByCustomerId() = %v, err = %v", latest, err)
		}
		// IsFirstSubscription
		isFirst, err := store.IsFirstSubscription(ctx, customer.ID)
		if err != nil {
			t.Errorf("IsFirstSubscription() error = %v", err)
		}
		if !isFirst {
			t.Errorf("IsFirstSubscription() = %v, want true", isFirst)
		}
		return errors.New("rollback")
	})
}

func TestPostgresStripeStore_UpsertSubscriptionFromStripe(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		store := stores.NewPostgresStripeStore(dbxx)
		userStore := stores.NewPostgresUserStore(dbxx)
		user, err := userStore.CreateUser(ctx, &models.User{Email: "sub@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		// Insert product and price
		product := &models.StripeProduct{ID: "prod_stripe_sub", Active: true, Name: "StripeSubProduct", Metadata: map[string]string{}}
		err = store.UpsertProduct(ctx, product)
		if err != nil {
			t.Fatalf("UpsertProduct() error = %v", err)
		}
		price := &models.StripePrice{
			ID:         "price_stripe_sub",
			ProductID:  product.ID,
			Active:     true,
			UnitAmount: types.Pointer(int64(3000)),
			Currency:   "usd",
			Type:       models.StripePricingTypeRecurring,
			Metadata:   map[string]string{},
		}
		err = store.UpsertPrice(ctx, price)
		if err != nil {
			t.Fatalf("UpsertPrice() error = %v", err)
		}
		// Insert customer
		customer := &models.StripeCustomer{
			ID:           "cus_stripe_sub",
			Email:        "stripe_sub@example.com",
			CustomerType: models.StripeCustomerTypeUser,
			UserID:       types.Pointer(user.ID),
		}
		_, err = store.CreateCustomer(ctx, customer)
		if err != nil {
			t.Fatalf("CreateCustomer() error = %v", err)
		}
		// UpsertSubscriptionFromStripe
		stripeSub := &stripe.Subscription{
			ID:       "sub_stripe_1",
			Customer: &stripe.Customer{ID: customer.ID},
			Status:   stripe.SubscriptionStatusActive,
			Metadata: map[string]string{},
			Items: &stripe.SubscriptionItemList{
				Data: []*stripe.SubscriptionItem{
					{
						ID:                 "item_stripe_1",
						Price:              &stripe.Price{ID: price.ID},
						Quantity:           1,
						CurrentPeriodStart: time.Now().Unix(),
						CurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour).Unix(),
					},
				},
			},
			CancelAtPeriodEnd: false,
			Created:           time.Now().Unix(),
		}
		err = store.UpsertSubscriptionFromStripe(ctx, stripeSub)
		if err != nil {
			t.Fatalf("UpsertSubscriptionFromStripe() error = %v", err)
		}
		return errors.New("rollback")
	})
}
