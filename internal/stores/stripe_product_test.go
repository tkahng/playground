//go:build integration

package stores_test

import (
	"context"
	"testing"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/types"
)

func TestStripeStore_ProductAndPrice(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)

		// UpsertProduct
		product := &models.StripeProduct{
			ID:     "prod_123",
			Active: true,
			Name:   "Test Product",
			Metadata: map[string]string{
				"key1": "value1",
			},
		}
		err := adapter.Product().UpsertProduct(ctx, product)
		if err != nil {
			t.Fatalf("UpsertProduct() error = %v", err)
		}

		// FindProductByStripeId
		found, err := adapter.Product().FindProductById(ctx, "prod_123")
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
		err = adapter.Price().UpsertPrice(ctx, price)
		if err != nil {
			t.Fatalf("UpsertPrice() error = %v", err)
		}

		// FindValidPriceById
		validPrice, err := adapter.Price().FindPrice(ctx, &stores.StripePriceFilter{
			Ids: []string{price.ID},
			Active: types.OptionalParam[bool]{
				Value: true,
				IsSet: true,
			},
		})
		if err != nil {
			t.Fatalf("FindValidPriceById() error = %v", err)
		}
		if validPrice == nil || validPrice.ID != price.ID {
			t.Errorf("FindValidPriceById() = %v, want %v", validPrice, price.ID)
		}

		// ListProducts
		products, err := adapter.Product().ListProducts(ctx, &stores.StripeProductFilter{})
		if err != nil {
			t.Fatalf("ListProducts() error = %v", err)
		}
		if len(products) == 0 {
			t.Errorf("ListProducts() = %v, want at least 1", products)
		}

		// ListPrices
		prices, err := adapter.Price().ListPrices(ctx, &stores.StripePriceFilter{})
		if err != nil {
			t.Fatalf("ListPrices() error = %v", err)
		}
		if len(prices) == 0 {
			t.Errorf("ListPrices() = %v, want at least 1", prices)
		}
	})
}

func TestStripeStore_UpsertProductAndPrice(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		stripeProduct := &models.StripeProduct{
			ID:          "prod_stripe_1",
			Active:      true,
			Name:        "Stripe Product",
			Description: types.Pointer("Stripe Desc"),
			Metadata:    map[string]string{"foo": "bar"},
		}

		err := adapter.Product().UpsertProduct(ctx, stripeProduct)
		if err != nil {
			t.Fatalf("UpsertProductFromStripe() error = %v", err)
		}
		found, err := adapter.Product().FindProductById(ctx, stripeProduct.ID)
		if err != nil || found == nil || found.ID != stripeProduct.ID {
			t.Errorf("FindProductByStripeId() = %v, err = %v", found, err)
		}

		stripePrice := &models.StripePrice{
			ID:              "price_stripe_1",
			ProductID:       stripeProduct.ID,
			Active:          true,
			LookupKey:       types.Pointer("lookup_1"),
			UnitAmount:      types.Pointer(int64(5000)),
			Currency:        "usd",
			Type:            "recurring",
			Metadata:        map[string]string{"foo": "bar"},
			Interval:        types.Pointer(models.StripePricingPlanIntervalMonth),
			IntervalCount:   types.Pointer(int64(1)),
			TrialPeriodDays: types.Pointer(int64(14)),
		}
		err = adapter.Price().UpsertPrice(ctx, stripePrice)
		if err != nil {
			t.Fatalf("UpsertPrice() error = %v", err)
		}
	})
}
