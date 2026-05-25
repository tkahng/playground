package stores_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/types"
)

func TestStripeStore_FindSubscriptionsWithPriceProductByIds(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		user, err := adapter.User().CreateUser(ctx, &models.User{Email: "sub@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		// Insert product and price
		product := &models.StripeProduct{ID: "prod_sub_1", Active: true, Name: "Sub Product", Metadata: map[string]string{}}
		err = adapter.Product().UpsertProduct(ctx, product)
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
		err = adapter.Price().UpsertPrice(ctx, price)
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
		_, err = adapter.Customer().CreateCustomer(ctx, customer)
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
		err = adapter.Subscription().UpsertSubscription(ctx, sub)
		if err != nil {
			t.Fatalf("UpsertSubscription() error = %v", err)
		}
		// FindSubscriptionWithPriceById
		withPriceList, err := adapter.Subscription().FindSubscriptionsWithPriceProductByIds(ctx, "sub_1")
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
		if withPrice.StripeCustomer == nil || withPrice.StripeCustomer.ID != customer.ID {
			t.Errorf("FindSubscriptionWithPriceById() StripeCustomer = %v, want %v", withPrice.StripeCustomer, customer.ID)
		}
	})
}

func TestStripeStore_FindActiveSubscriptionsByTeamIds(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)

		user, err := adapter.User().CreateUser(ctx, &models.User{Email: "sub@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}

		team, err := adapter.TeamGroup().CreateTeam(ctx, "test", "test")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}

		_, err = adapter.TeamMember().CreateTeamMember(
			ctx,
			&models.TeamMember{
				TeamID:           team.ID,
				UserID:           &user.ID,
				Role:             models.TeamMemberRoleOwner,
				Active:           true,
				HasBillingAccess: true,
			},
		)

		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		// Insert product and price
		product := &models.StripeProduct{ID: "prod_sub_1", Active: true, Name: "Sub Product", Metadata: map[string]string{}}
		err = adapter.Product().UpsertProduct(ctx, product)
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
		err = adapter.Price().UpsertPrice(ctx, price)
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
		_, err = adapter.Customer().CreateCustomer(ctx, customer)
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
		err = adapter.Subscription().UpsertSubscription(ctx, sub)
		if err != nil {
			t.Fatalf("UpsertSubscription() error = %v", err)
		}
		// FindSubscriptionWithPriceById
		teamSubs, err := adapter.Subscription().FindActiveSubscriptionsByTeamIds(ctx, team.ID)
		if err != nil {
			t.Fatalf("FindActiveSubscriptionsByTeamIds() error = %v", err)
		}
		if len(teamSubs) == 0 {
			t.Fatalf("FindActiveSubscriptionsByTeamIds() = %v, want at least 1", teamSubs)
		}

		withPrice := teamSubs[0]
		err = loadPricesWithProduct(t, ctx, withPrice, adapter)
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
	})
}

func loadPricesWithProduct(t *testing.T, ctx context.Context, withPrice *models.StripeSubscription, adapter stores.StorageAdapterInterface) error {
	t.Helper()
	if withPrice == nil {
		return nil
	}
	price, err := adapter.Price().FindPrice(ctx, &stores.StripePriceFilter{
		Ids: []string{withPrice.PriceID},
	})
	if err != nil {
		return err
	}
	if price == nil {
		return errors.New("price not found")
	}
	product, err := adapter.Product().FindProductById(ctx, price.ProductID)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New("product not found")
	}
	withPrice.Price = price
	withPrice.Price.Product = product
	return nil
}
func TestStripeStore_FindActiveSubscriptionsByCustomerIds(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		user, err := adapter.User().CreateUser(ctx, &models.User{Email: "sub@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}

		team, err := adapter.TeamGroup().CreateTeam(ctx, "test", "test")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}

		_, err = adapter.TeamMember().CreateTeamMember(
			ctx,
			&models.TeamMember{
				TeamID:           team.ID,
				UserID:           &user.ID,
				Role:             models.TeamMemberRoleOwner,
				HasBillingAccess: true,
				Active:           true,
			},
		)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		// Insert product and price
		product := &models.StripeProduct{ID: "prod_sub_1", Active: true, Name: "Sub Product", Metadata: map[string]string{}}
		err = adapter.Product().UpsertProduct(ctx, product)
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
		err = adapter.Price().UpsertPrice(ctx, price)
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
		_, err = adapter.Customer().CreateCustomer(ctx, customer)
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
		err = adapter.Subscription().UpsertSubscription(ctx, sub)
		if err != nil {
			t.Fatalf("UpsertSubscription() error = %v", err)
		}
		// FindSubscriptionWithPriceById
		customerSubs, err := adapter.Subscription().FindActiveSubscriptionsByCustomerIds(ctx, customer.ID)
		if err != nil {
			t.Fatalf("FindActiveSubscriptionsByCustomerIds() error = %v", err)
		}
		if len(customerSubs) == 0 {
			t.Fatalf("FindActiveSubscriptionsByCustomerIds() = %v, want at least 1", customerSubs)
		}

		withPrice := customerSubs[0]
		if withPrice == nil || withPrice.ID != "sub_1" {
			t.Errorf("FindSubscriptionWithPriceById() = %v, err = %v", withPrice, err)
		}
		if withPrice.Price == nil || withPrice.Price.ID != price.ID {
			t.Errorf("FindSubscriptionWithPriceById() Price = %v, want %v", withPrice.Price, price.ID)
		}
		if withPrice.Price.Product == nil || withPrice.Price.Product.ID != product.ID {
			t.Errorf("FindSubscriptionWithPriceById() Product = %v, want %v", withPrice.Price.Product, product.ID)
		}
	})
}
func TestStripeStore_FindActiveSubscriptionsByUserIds(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)

		user, err := adapter.User().CreateUser(ctx, &models.User{Email: "sub@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}

		team, err := adapter.TeamGroup().CreateTeam(ctx, "test", "test")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}

		_, err = adapter.TeamMember().CreateTeamMember(ctx, &models.TeamMember{
			TeamID:           team.ID,
			UserID:           &user.ID,
			Role:             models.TeamMemberRoleOwner,
			HasBillingAccess: true,
			Active:           true,
		})
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		// Insert product and price
		product := &models.StripeProduct{ID: "prod_sub_1", Active: true, Name: "Sub Product", Metadata: map[string]string{}}
		err = adapter.Product().UpsertProduct(ctx, product)
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
		err = adapter.Price().UpsertPrice(ctx, price)
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
		_, err = adapter.Customer().CreateCustomer(ctx, customer)
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
		err = adapter.Subscription().UpsertSubscription(ctx, sub)
		if err != nil {
			t.Fatalf("UpsertSubscription() error = %v", err)
		}
		customerSubs, err := adapter.Subscription().FindActiveSubscriptionsByUserIds(ctx, user.ID)
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
		err = loadPricesWithProduct(t, ctx, withPrice, adapter)
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
	})
}

func TestStripeStore_UpsertSubscriptionFromStripe(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)

		user, err := adapter.User().CreateUser(ctx, &models.User{Email: "sub@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		// Insert product and price
		product := &models.StripeProduct{ID: "prod_stripe_sub", Active: true, Name: "StripeSubProduct", Metadata: map[string]string{}}
		err = adapter.Product().UpsertProduct(ctx, product)
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
		err = adapter.Price().UpsertPrice(ctx, price)
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
		_, err = adapter.Customer().CreateCustomer(ctx, customer)
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
		err = adapter.Subscription().UpsertSubscriptionFromStripe(ctx, stripeSub)
		if err != nil {
			t.Fatalf("UpsertSubscriptionFromStripe() error = %v", err)
		}
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
				`billing.stripe_prices.id AS "id"`,
				`billing.stripe_prices.product_id AS "product_id"`,
				`billing.stripe_prices.lookup_key AS "lookup_key"`,
				`billing.stripe_prices.active AS "active"`,
				`billing.stripe_prices.unit_amount AS "unit_amount"`,
				`billing.stripe_prices.currency AS "currency"`,
				`billing.stripe_prices.type AS "type"`,
				`billing.stripe_prices.interval AS "interval"`,
				`billing.stripe_prices.interval_count AS "interval_count"`,
				`billing.stripe_prices.trial_period_days AS "trial_period_days"`,
				`billing.stripe_prices.metadata AS "metadata"`,
				`billing.stripe_prices.created_at AS "created_at"`,
				`billing.stripe_prices.updated_at AS "updated_at"`,
			},
		},
		{
			name: "with tablePrefix and prefix",
			args: args{
				tablePrefix: "stripe_prices",
				prefix:      "price",
			},
			expect: []string{
				`billing.stripe_prices.id AS "price.id"`,
				`billing.stripe_prices.product_id AS "price.product_id"`,
				`billing.stripe_prices.lookup_key AS "price.lookup_key"`,
				`billing.stripe_prices.active AS "price.active"`,
				`billing.stripe_prices.unit_amount AS "price.unit_amount"`,
				`billing.stripe_prices.currency AS "price.currency"`,
				`billing.stripe_prices.type AS "price.type"`,
				`billing.stripe_prices.interval AS "price.interval"`,
				`billing.stripe_prices.interval_count AS "price.interval_count"`,
				`billing.stripe_prices.trial_period_days AS "price.trial_period_days"`,
				`billing.stripe_prices.metadata AS "price.metadata"`,
				`billing.stripe_prices.created_at AS "price.created_at"`,
				`billing.stripe_prices.updated_at AS "price.updated_at"`,
			},
		},
		{
			name: "with tablePrefix only and double prefix",
			args: args{
				tablePrefix: "stripe_prices",
				prefix:      "some.price",
			},
			expect: []string{
				`billing.stripe_prices.id AS "some.price.id"`,
				`billing.stripe_prices.product_id AS "some.price.product_id"`,
				`billing.stripe_prices.lookup_key AS "some.price.lookup_key"`,
				`billing.stripe_prices.active AS "some.price.active"`,
				`billing.stripe_prices.unit_amount AS "some.price.unit_amount"`,
				`billing.stripe_prices.currency AS "some.price.currency"`,
				`billing.stripe_prices.type AS "some.price.type"`,
				`billing.stripe_prices.interval AS "some.price.interval"`,
				`billing.stripe_prices.interval_count AS "some.price.interval_count"`,
				`billing.stripe_prices.trial_period_days AS "some.price.trial_period_days"`,
				`billing.stripe_prices.metadata AS "some.price.metadata"`,
				`billing.stripe_prices.created_at AS "some.price.created_at"`,
				`billing.stripe_prices.updated_at AS "some.price.updated_at"`,
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
				`billing.stripe_products.id AS "id"`,
				`billing.stripe_products.name AS "name"`,
				`billing.stripe_products.description AS "description"`,
				`billing.stripe_products.active AS "active"`,
				`billing.stripe_products.image AS "image"`,
				`billing.stripe_products.metadata AS "metadata"`,
				`billing.stripe_products.created_at AS "created_at"`,
				`billing.stripe_products.updated_at AS "updated_at"`,
			},
		},
		{
			name: "with tablePrefix and prefix",
			args: args{
				tablePrefix: "p",
				prefix:      "product",
			},
			expect: []string{
				`billing.stripe_products.id AS "product.id"`,
				`billing.stripe_products.name AS "product.name"`,
				`billing.stripe_products.description AS "product.description"`,
				`billing.stripe_products.active AS "product.active"`,
				`billing.stripe_products.image AS "product.image"`,
				`billing.stripe_products.metadata AS "product.metadata"`,
				`billing.stripe_products.created_at AS "product.created_at"`,
				`billing.stripe_products.updated_at AS "product.updated_at"`,
			},
		},
		{
			name: "with tablePrefix only and double prefix",
			args: args{
				tablePrefix: "p",
				prefix:      "some.product",
			},
			expect: []string{
				`billing.stripe_products.id AS "some.product.id"`,
				`billing.stripe_products.name AS "some.product.name"`,
				`billing.stripe_products.description AS "some.product.description"`,
				`billing.stripe_products.active AS "some.product.active"`,
				`billing.stripe_products.image AS "some.product.image"`,
				`billing.stripe_products.metadata AS "some.product.metadata"`,
				`billing.stripe_products.created_at AS "some.product.created_at"`,
				`billing.stripe_products.updated_at AS "some.product.updated_at"`,
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
				`billing.stripe_subscriptions.id AS "id"`,
				`billing.stripe_subscriptions.stripe_customer_id AS "stripe_customer_id"`,
				`billing.stripe_subscriptions.status AS "status"`,
				`billing.stripe_subscriptions.metadata AS "metadata"`,
				`billing.stripe_subscriptions.item_id AS "item_id"`,
				`billing.stripe_subscriptions.price_id AS "price_id"`,
				`billing.stripe_subscriptions.quantity AS "quantity"`,
				`billing.stripe_subscriptions.cancel_at_period_end AS "cancel_at_period_end"`,
				`billing.stripe_subscriptions.created AS "created"`,
				`billing.stripe_subscriptions.current_period_start AS "current_period_start"`,
				`billing.stripe_subscriptions.current_period_end AS "current_period_end"`,
				`billing.stripe_subscriptions.ended_at AS "ended_at"`,
				`billing.stripe_subscriptions.cancel_at AS "cancel_at"`,
				`billing.stripe_subscriptions.canceled_at AS "canceled_at"`,
				`billing.stripe_subscriptions.trial_start AS "trial_start"`,
				`billing.stripe_subscriptions.trial_end AS "trial_end"`,
				`billing.stripe_subscriptions.created_at AS "created_at"`,
				`billing.stripe_subscriptions.updated_at AS "updated_at"`,
			},
		},
		{
			name: "with tablePrefix and prefix",
			args: args{
				tablePrefix: "ss",
				prefix:      "subscription",
			},
			expect: []string{
				`billing.stripe_subscriptions.id AS "subscription.id"`,
				`billing.stripe_subscriptions.stripe_customer_id AS "subscription.stripe_customer_id"`,
				`billing.stripe_subscriptions.status AS "subscription.status"`,
				`billing.stripe_subscriptions.metadata AS "subscription.metadata"`,
				`billing.stripe_subscriptions.item_id AS "subscription.item_id"`,
				`billing.stripe_subscriptions.price_id AS "subscription.price_id"`,
				`billing.stripe_subscriptions.quantity AS "subscription.quantity"`,
				`billing.stripe_subscriptions.cancel_at_period_end AS "subscription.cancel_at_period_end"`,
				`billing.stripe_subscriptions.created AS "subscription.created"`,
				`billing.stripe_subscriptions.current_period_start AS "subscription.current_period_start"`,
				`billing.stripe_subscriptions.current_period_end AS "subscription.current_period_end"`,
				`billing.stripe_subscriptions.ended_at AS "subscription.ended_at"`,
				`billing.stripe_subscriptions.cancel_at AS "subscription.cancel_at"`,
				`billing.stripe_subscriptions.canceled_at AS "subscription.canceled_at"`,
				`billing.stripe_subscriptions.trial_start AS "subscription.trial_start"`,
				`billing.stripe_subscriptions.trial_end AS "subscription.trial_end"`,
				`billing.stripe_subscriptions.created_at AS "subscription.created_at"`,
				`billing.stripe_subscriptions.updated_at AS "subscription.updated_at"`,
			},
		},
		{
			name: "with tablePrefix only and double prefix",
			args: args{
				tablePrefix: "ss",
				prefix:      "some.subscription",
			},
			expect: []string{
				`billing.stripe_subscriptions.id AS "some.subscription.id"`,
				`billing.stripe_subscriptions.stripe_customer_id AS "some.subscription.stripe_customer_id"`,
				`billing.stripe_subscriptions.status AS "some.subscription.status"`,
				`billing.stripe_subscriptions.metadata AS "some.subscription.metadata"`,
				`billing.stripe_subscriptions.item_id AS "some.subscription.item_id"`,
				`billing.stripe_subscriptions.price_id AS "some.subscription.price_id"`,
				`billing.stripe_subscriptions.quantity AS "some.subscription.quantity"`,
				`billing.stripe_subscriptions.cancel_at_period_end AS "some.subscription.cancel_at_period_end"`,
				`billing.stripe_subscriptions.created AS "some.subscription.created"`,
				`billing.stripe_subscriptions.current_period_start AS "some.subscription.current_period_start"`,
				`billing.stripe_subscriptions.current_period_end AS "some.subscription.current_period_end"`,
				`billing.stripe_subscriptions.ended_at AS "some.subscription.ended_at"`,
				`billing.stripe_subscriptions.cancel_at AS "some.subscription.cancel_at"`,
				`billing.stripe_subscriptions.canceled_at AS "some.subscription.canceled_at"`,
				`billing.stripe_subscriptions.trial_start AS "some.subscription.trial_start"`,
				`billing.stripe_subscriptions.trial_end AS "some.subscription.trial_end"`,
				`billing.stripe_subscriptions.created_at AS "some.subscription.created_at"`,
				`billing.stripe_subscriptions.updated_at AS "some.subscription.updated_at"`,
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
			joinedCols := strings.Join(tt.expect, ", ")
			// for _, col := range tt.expect {
			if !containsSQLColumn(sql, joinedCols) {
				t.Errorf("Expected column %q in SQL: %s", joinedCols, sql)
			}
			// }
		})
	}
}

// containsSQLColumn checks if the column string is present in the SELECT SQL.
func containsSQLColumn(sql, col string) bool {
	return strings.Contains(sql, col)
}
