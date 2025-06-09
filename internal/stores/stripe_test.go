package stores_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/test"
	"github.com/tkahng/authgo/internal/tools/types"
)

func TestStripeStore_CreateCustomer(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
		userStore := stores.NewDbUserStore(dbxx)
		teamStore := stores.NewDbTeamStore(dbxx)
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
				store := stores.NewDbStripeStore(tt.fields.db)
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

func TestStripeStore_ProductAndPrice(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
		store := stores.NewDbStripeStore(dbxx)

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
		found, err := store.FindProductById(ctx, "prod_123")
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
		validPrice, err := store.FindActivePriceById(ctx, "price_123")
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

func TestStripeStore_UpsertProductAndPriceFromStripe(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
		store := stores.NewDbStripeStore(dbxx)
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
		found, err := store.FindProductById(ctx, stripeProduct.ID)
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

func TestStripeStore_FindCustomer(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
		userStore := stores.NewDbUserStore(dbxx)
		user, err := userStore.CreateUser(ctx, &models.User{Email: "findcustomer@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		store := stores.NewDbStripeStore(dbxx)
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

func TestStripeStore_FindSubscriptionsWithPriceProductByIds(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
		store := stores.NewDbStripeStore(dbxx)
		userStore := stores.NewDbUserStore(dbxx)
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
		withPriceList, err := store.FindSubscriptionsWithPriceProductByIds(ctx, "sub_1")
		if err != nil {
			t.Fatalf("FindSubscriptionWithPriceProductById() error = %v", err)
		}
		if len(withPriceList) == 0 {
			t.Fatalf("FindSubscriptionWithPriceProductById() = %v, want at least 1", withPriceList)
		}
		withPrice := withPriceList[0]
		if withPrice == nil || withPrice.ID != "sub_1" {
			t.Errorf("FindSubscriptionWithPriceById() = %v, err = %v", withPrice, err)
		}
		if withPrice.Price == nil || withPrice.Price.ID != price.ID {
			t.Errorf("FindSubscriptionWithPriceById() Price = %v, want %v", withPrice.Price, price.ID)
		}
		if withPrice.Price.Product == nil || withPrice.Price.Product.ID != product.ID {
			t.Errorf("FindSubscriptionWithPriceById() Product = %v, want %v", withPrice.Price.Product, product.ID)
		}
		return errors.New("rollback")
	})
}

func TestStripeStore_FindActiveSubscriptionsByTeamIds(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
		store := stores.NewDbStripeStore(dbxx)
		userStore := stores.NewDbUserStore(dbxx)
		teamStore := stores.NewDbTeamStore(dbxx)
		user, err := userStore.CreateUser(ctx, &models.User{Email: "sub@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}

		team, err := teamStore.CreateTeam(ctx, "test", "test")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}

		_, err = teamStore.CreateTeamMember(
			ctx,
			team.ID,
			user.ID,
			models.TeamMemberRoleOwner,
			true,
		)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
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
			CustomerType: models.StripeCustomerTypeTeam,
			TeamID:       types.Pointer(team.ID),
			// UserID:       types.Pointer(user.ID),
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
		teamSubs, err := store.FindActiveSubscriptionsByTeamIds(ctx, team.ID)
		if err != nil {
			t.Fatalf("FindActiveSubscriptionsByTeamIds() error = %v", err)
		}
		if len(teamSubs) == 0 {
			t.Fatalf("FindActiveSubscriptionsByTeamIds() = %v, want at least 1", teamSubs)
		}

		withPrice := teamSubs[0]
		err = store.LoadSubscriptionsPriceProduct(ctx, withPrice)
		if err != nil {
			t.Fatalf("LoadSubscriptionstripe_pricesriceProduct() error = %v", err)
		}
		if withPrice == nil || withPrice.ID != "sub_1" {
			t.Errorf("FindSubscriptionWithPriceById() = %v, err = %v", withPrice, err)
		}
		if withPrice.Price == nil || withPrice.Price.ID != price.ID {
			t.Errorf("FindSubscriptionWithPriceById() Price = %v, want %v", withPrice.Price, price.ID)
		}
		if withPrice.Price.Product == nil || withPrice.Price.Product.ID != product.ID {
			t.Errorf("FindSubscriptionWithPriceById() Product = %v, want %v", withPrice.Price.Product, product.ID)
		}
		return errors.New("rollback")
	})
}
func TestStripeStore_FindActiveSubscriptionsByCustomerIds(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
		store := stores.NewDbStripeStore(dbxx)
		userStore := stores.NewDbUserStore(dbxx)
		teamStore := stores.NewDbTeamStore(dbxx)
		user, err := userStore.CreateUser(ctx, &models.User{Email: "sub@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}

		team, err := teamStore.CreateTeam(ctx, "test", "test")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}

		_, err = teamStore.CreateTeamMember(
			ctx,
			team.ID,
			user.ID,
			models.TeamMemberRoleOwner,
			true,
		)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
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
			CustomerType: models.StripeCustomerTypeTeam,
			TeamID:       types.Pointer(team.ID),
			// UserID:       types.Pointer(user.ID),
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
		customerSubs, err := store.FindActiveSubscriptionsByCustomerIds(ctx, customer.ID)
		if err != nil {
			t.Fatalf("FindActiveSubscriptionsByCustomerIds() error = %v", err)
		}
		if len(customerSubs) == 0 {
			t.Fatalf("FindActiveSubscriptionsByCustomerIds() = %v, want at least 1", customerSubs)
		}

		withPrice := customerSubs[0]
		err = store.LoadSubscriptionsPriceProduct(ctx, withPrice)
		if err != nil {
			t.Fatalf("LoadSubscriptionstripe_pricesriceProduct() error = %v", err)
		}
		if withPrice == nil || withPrice.ID != "sub_1" {
			t.Errorf("FindSubscriptionWithPriceById() = %v, err = %v", withPrice, err)
		}
		if withPrice.Price == nil || withPrice.Price.ID != price.ID {
			t.Errorf("FindSubscriptionWithPriceById() Price = %v, want %v", withPrice.Price, price.ID)
		}
		if withPrice.Price.Product == nil || withPrice.Price.Product.ID != product.ID {
			t.Errorf("FindSubscriptionWithPriceById() Product = %v, want %v", withPrice.Price.Product, product.ID)
		}
		return errors.New("rollback")
	})
}
func TestStripeStore_FindActiveSubscriptionsByUserIds(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
		store := stores.NewDbStripeStore(dbxx)
		userStore := stores.NewDbUserStore(dbxx)
		teamStore := stores.NewDbTeamStore(dbxx)
		user, err := userStore.CreateUser(ctx, &models.User{Email: "sub@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}

		team, err := teamStore.CreateTeam(ctx, "test", "test")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}

		_, err = teamStore.CreateTeamMember(
			ctx,
			team.ID,
			user.ID,
			models.TeamMemberRoleOwner,
			true,
		)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
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
			// TeamID:       types.Pointer(team.ID),
			UserID: types.Pointer(user.ID),
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
		customerSubs, err := store.FindActiveSubscriptionsByUserIds(ctx, user.ID)
		if err != nil {
			t.Fatalf("FindActiveSubscriptionsByUserIds() error = %v", err)
		}
		if len(customerSubs) == 0 {
			t.Fatalf("FindActiveSubscriptionsByUserIds() = %v, want at least 1", customerSubs)
		}

		withPrice := customerSubs[0]
		if withPrice == nil || withPrice.ID != "sub_1" {
			t.Errorf("FindSubscriptionWithPriceById() = %v, err = %v", withPrice, err)
		}
		err = store.LoadSubscriptionsPriceProduct(ctx, withPrice)
		if err != nil {
			t.Fatalf("LoadSubscriptionstripe_pricesriceProduct() error = %v", err)
		}
		if withPrice == nil || withPrice.ID != "sub_1" {
			t.Errorf("FindSubscriptionWithPriceById() = %v, err = %v", withPrice, err)
		}
		if withPrice.Price == nil || withPrice.Price.ID != price.ID {
			t.Errorf("FindSubscriptionWithPriceById() Price = %v, want %v", withPrice.Price, price.ID)
		}
		if withPrice.Price.Product == nil || withPrice.Price.Product.ID != product.ID {
			t.Errorf("FindSubscriptionWithPriceById() Product = %v, want %v", withPrice.Price.Product, product.ID)
		}
		return errors.New("rollback")
	})
}

func TestStripeStore_UpsertSubscriptionFromStripe(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
		store := stores.NewDbStripeStore(dbxx)
		userStore := stores.NewDbUserStore(dbxx)
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
func TestSelectStripePriceColumns(t *testing.T) {
	type args struct {
		tablePrefix string
		prefix      string
	}
	tests := []struct {
		name   string
		args   args
		expect []string
	}{
		{
			name: "no prefix",
			args: args{
				tablePrefix: "",
				prefix:      "",
			},
			expect: []string{
				`id AS "id"`,
				`product_id AS "product_id"`,
				`lookup_key AS "lookup_key"`,
				`active AS "active"`,
				`unit_amount AS "unit_amount"`,
				`currency AS "currency"`,
				`type AS "type"`,
				`interval AS "interval"`,
				`interval_count AS "interval_count"`,
				`trial_period_days AS "trial_period_days"`,
				`metadata AS "metadata"`,
				`created_at AS "created_at"`,
				`updated_at AS "updated_at"`,
			},
		},
		{
			name: "with tablePrefix and prefix",
			args: args{
				tablePrefix: "stripe_prices",
				prefix:      "price",
			},
			expect: []string{
				`stripe_prices.id AS "price.id"`,
				`stripe_prices.product_id AS "price.product_id"`,
				`stripe_prices.lookup_key AS "price.lookup_key"`,
				`stripe_prices.active AS "price.active"`,
				`stripe_prices.unit_amount AS "price.unit_amount"`,
				`stripe_prices.currency AS "price.currency"`,
				`stripe_prices.type AS "price.type"`,
				`stripe_prices.interval AS "price.interval"`,
				`stripe_prices.interval_count AS "price.interval_count"`,
				`stripe_prices.trial_period_days AS "price.trial_period_days"`,
				`stripe_prices.metadata AS "price.metadata"`,
				`stripe_prices.created_at AS "price.created_at"`,
				`stripe_prices.updated_at AS "price.updated_at"`,
			},
		},
		{
			name: "with tablePrefix only and double prefix",
			args: args{
				tablePrefix: "stripe_prices",
				prefix:      "some.price",
			},
			expect: []string{
				`stripe_prices.id AS "some.price.id"`,
				`stripe_prices.product_id AS "some.price.product_id"`,
				`stripe_prices.lookup_key AS "some.price.lookup_key"`,
				`stripe_prices.active AS "some.price.active"`,
				`stripe_prices.unit_amount AS "some.price.unit_amount"`,
				`stripe_prices.currency AS "some.price.currency"`,
				`stripe_prices.type AS "some.price.type"`,
				`stripe_prices.interval AS "some.price.interval"`,
				`stripe_prices.interval_count AS "some.price.interval_count"`,
				`stripe_prices.trial_period_days AS "some.price.trial_period_days"`,
				`stripe_prices.metadata AS "some.price.metadata"`,
				`stripe_prices.created_at AS "some.price.created_at"`,
				`stripe_prices.updated_at AS "some.price.updated_at"`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qs := squirrel.Select()
			qs = stores.SelectStripePriceColumns(qs, tt.args.prefix)
			sql, _, err := qs.ToSql()
			if err != nil {
				t.Fatalf("ToSql() error = %v", err)
			}
			for _, col := range tt.expect {
				if !containsSQLColumn(sql, col) {
					t.Errorf("Expected column %q in SQL: %s", col, sql)
				}
			}
		})
	}
}

func TestSelectStripeProductColumns(t *testing.T) {
	type args struct {
		tablePrefix string
		prefix      string
	}
	tests := []struct {
		name   string
		args   args
		expect []string
	}{
		{
			name: "no prefix",
			args: args{
				tablePrefix: "",
				prefix:      "",
			},
			expect: []string{
				`id AS "id"`,
				`name AS "name"`,
				`description AS "description"`,
				`active AS "active"`,
				`image AS "image"`,
				`metadata AS "metadata"`,
				`created_at AS "created_at"`,
				`updated_at AS "updated_at"`,
			},
		},
		{
			name: "with tablePrefix and prefix",
			args: args{
				tablePrefix: "p",
				prefix:      "product",
			},
			expect: []string{
				`stripe_products.id AS "product.id"`,
				`stripe_products.name AS "product.name"`,
				`stripe_products.description AS "product.description"`,
				`stripe_products.active AS "product.active"`,
				`stripe_products.image AS "product.image"`,
				`stripe_products.metadata AS "product.metadata"`,
				`stripe_products.created_at AS "product.created_at"`,
				`stripe_products.updated_at AS "product.updated_at"`,
			},
		},
		{
			name: "with tablePrefix only and double prefix",
			args: args{
				tablePrefix: "p",
				prefix:      "some.product",
			},
			expect: []string{
				`stripe_products.id AS "some.product.id"`,
				`stripe_products.name AS "some.product.name"`,
				`stripe_products.description AS "some.product.description"`,
				`stripe_products.active AS "some.product.active"`,
				`stripe_products.image AS "some.product.image"`,
				`stripe_products.metadata AS "some.product.metadata"`,
				`stripe_products.created_at AS "some.product.created_at"`,
				`stripe_products.updated_at AS "some.product.updated_at"`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qs := squirrel.Select()
			qs = stores.SelectStripeProductColumns(qs, tt.args.prefix)
			sql, _, err := qs.ToSql()
			if err != nil {
				t.Fatalf("ToSql() error = %v", err)
			}
			for _, col := range tt.expect {
				if !containsSQLColumn(sql, col) {
					t.Errorf("Expected column %q in SQL: %s", col, sql)
				}
			}
		})
	}
}

func TestSelectStripeSubscriptionColumns(t *testing.T) {
	type args struct {
		tablePrefix string
		prefix      string
	}
	tests := []struct {
		name   string
		args   args
		expect []string
	}{
		{
			name: "no prefix",
			args: args{
				tablePrefix: "",
				prefix:      "",
			},
			expect: []string{
				`stripe_subscriptions.id AS "id"`,
				`stripe_subscriptions.stripe_customer_id AS "stripe_customer_id"`,
				`stripe_subscriptions.status AS "status"`,
				`stripe_subscriptions.metadata AS "metadata"`,
				`stripe_subscriptions.item_id AS "item_id"`,
				`stripe_subscriptions.price_id AS "price_id"`,
				`stripe_subscriptions.quantity AS "quantity"`,
				`stripe_subscriptions.cancel_at_period_end AS "cancel_at_period_end"`,
				`stripe_subscriptions.created AS "created"`,
				`stripe_subscriptions.current_period_start AS "current_period_start"`,
				`stripe_subscriptions.current_period_end AS "current_period_end"`,
				`stripe_subscriptions.ended_at AS "ended_at"`,
				`stripe_subscriptions.cancel_at AS "cancel_at"`,
				`stripe_subscriptions.canceled_at AS "canceled_at"`,
				`stripe_subscriptions.trial_start AS "trial_start"`,
				`stripe_subscriptions.trial_end AS "trial_end"`,
				`stripe_subscriptions.created_at AS "created_at"`,
				`stripe_subscriptions.updated_at AS "updated_at"`,
			},
		},
		{
			name: "with tablePrefix and prefix",
			args: args{
				tablePrefix: "ss",
				prefix:      "subscription",
			},
			expect: []string{
				`stripe_subscriptions.id AS "subscription.id"`,
				`stripe_subscriptions.stripe_customer_id AS "subscription.stripe_customer_id"`,
				`stripe_subscriptions.status AS "subscription.status"`,
				`stripe_subscriptions.metadata AS "subscription.metadata"`,
				`stripe_subscriptions.item_id AS "subscription.item_id"`,
				`stripe_subscriptions.price_id AS "subscription.price_id"`,
				`stripe_subscriptions.quantity AS "subscription.quantity"`,
				`stripe_subscriptions.cancel_at_period_end AS "subscription.cancel_at_period_end"`,
				`stripe_subscriptions.created AS "subscription.created"`,
				`stripe_subscriptions.current_period_start AS "subscription.current_period_start"`,
				`stripe_subscriptions.current_period_end AS "subscription.current_period_end"`,
				`stripe_subscriptions.ended_at AS "subscription.ended_at"`,
				`stripe_subscriptions.cancel_at AS "subscription.cancel_at"`,
				`stripe_subscriptions.canceled_at AS "subscription.canceled_at"`,
				`stripe_subscriptions.trial_start AS "subscription.trial_start"`,
				`stripe_subscriptions.trial_end AS "subscription.trial_end"`,
				`stripe_subscriptions.created_at AS "subscription.created_at"`,
				`stripe_subscriptions.updated_at AS "subscription.updated_at"`,
			},
		},
		{
			name: "with tablePrefix only and double prefix",
			args: args{
				tablePrefix: "ss",
				prefix:      "some.subscription",
			},
			expect: []string{
				`stripe_subscriptions.id AS "some.subscription.id"`,
				`stripe_subscriptions.stripe_customer_id AS "some.subscription.stripe_customer_id"`,
				`stripe_subscriptions.status AS "some.subscription.status"`,
				`stripe_subscriptions.metadata AS "some.subscription.metadata"`,
				`stripe_subscriptions.item_id AS "some.subscription.item_id"`,
				`stripe_subscriptions.price_id AS "some.subscription.price_id"`,
				`stripe_subscriptions.quantity AS "some.subscription.quantity"`,
				`stripe_subscriptions.cancel_at_period_end AS "some.subscription.cancel_at_period_end"`,
				`stripe_subscriptions.created AS "some.subscription.created"`,
				`stripe_subscriptions.current_period_start AS "some.subscription.current_period_start"`,
				`stripe_subscriptions.current_period_end AS "some.subscription.current_period_end"`,
				`stripe_subscriptions.ended_at AS "some.subscription.ended_at"`,
				`stripe_subscriptions.cancel_at AS "some.subscription.cancel_at"`,
				`stripe_subscriptions.canceled_at AS "some.subscription.canceled_at"`,
				`stripe_subscriptions.trial_start AS "some.subscription.trial_start"`,
				`stripe_subscriptions.trial_end AS "some.subscription.trial_end"`,
				`stripe_subscriptions.created_at AS "some.subscription.created_at"`,
				`stripe_subscriptions.updated_at AS "some.subscription.updated_at"`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qs := squirrel.Select()
			qs = stores.SelectStripeSubscriptionColumns(qs, tt.args.prefix)
			sql, _, err := qs.ToSql()
			if err != nil {
				t.Fatalf("ToSql() error = %v", err)
			}
			for _, col := range tt.expect {
				if !containsSQLColumn(sql, col) {
					t.Errorf("Expected column %q in SQL: %s", col, sql)
				}
			}
		})
	}
}

// containsSQLColumn checks if the column string is present in the SELECT SQL.
func containsSQLColumn(sql, col string) bool {
	return strings.Contains(sql, col)
}
