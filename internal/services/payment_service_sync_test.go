//go:build integration

package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
)

func TestStripeService_UpsertPriceProductFromStripe(t *testing.T) {
	t.Run("syncs products and prices from stripe client", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			adpt := stores.NewDbAdapterDecorators(db)
			client := services.NewMockPaymentClient()
			svc := services.NewPaymentService(client, adpt)

			err := svc.UpsertPriceProductFromStripe(ctx)
			require.NoError(t, err)

			products, err := adpt.Product().ListProducts(ctx, &stores.StripeProductFilter{
				PaginatedInput: stores.PaginatedInput{PerPage: 100},
			})
			require.NoError(t, err)
			assert.Len(t, products, len(services.Products), "all mock products should be upserted")

			prices, err := adpt.Price().ListPrices(ctx, &stores.StripePriceFilter{
				PaginatedInput: stores.PaginatedInput{PerPage: 100},
			})
			require.NoError(t, err)
			assert.Len(t, prices, len(services.Prices), "all mock prices should be upserted")
		})
	})

	t.Run("is idempotent - second sync overwrites without error", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			adpt := stores.NewDbAdapterDecorators(db)
			client := services.NewMockPaymentClient()
			svc := services.NewPaymentService(client, adpt)

			require.NoError(t, svc.UpsertPriceProductFromStripe(ctx))
			require.NoError(t, svc.UpsertPriceProductFromStripe(ctx))

			products, err := adpt.Product().ListProducts(ctx, &stores.StripeProductFilter{
				PaginatedInput: stores.PaginatedInput{PerPage: 100},
			})
			require.NoError(t, err)
			assert.Len(t, products, len(services.Products))
		})
	})

	t.Run("returns error when FindAllProducts fails", func(t *testing.T) {
		adpt := stores.NewAdapterDecorators()
		client := services.NewMockPaymentClient()
		svc := services.NewPaymentService(client, adpt)

		client.FindAllProductsFunc = func() ([]*stripe.Product, error) {
			return nil, errors.New("stripe unavailable")
		}

		err := svc.UpsertPriceProductFromStripe(context.Background())
		assert.Error(t, err)
	})

	t.Run("returns error when FindAllPrices fails", func(t *testing.T) {
		adpt := stores.NewAdapterDecorators()
		client := services.NewMockPaymentClient()
		svc := services.NewPaymentService(client, adpt)

		client.FindAllProductsFunc = func() ([]*stripe.Product, error) {
			return []*stripe.Product{}, nil
		}
		client.FindAllPricesFunc = func() ([]*stripe.Price, error) {
			return nil, errors.New("stripe prices unavailable")
		}

		err := svc.UpsertPriceProductFromStripe(context.Background())
		assert.Error(t, err)
	})
}
