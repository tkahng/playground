package seeders

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/repository"
	"github.com/tkahng/playground/internal/tools/types"
)

func CreateStripeProductPrices(ctx context.Context, dbx database.Dbx, count int) ([]*models.StripeProduct, error) {
	var products []models.StripeProduct
	for range count {
		uid := uuid.NewString()
		product := models.StripeProduct{
			ID:       uid,
			Name:     uid,
			Active:   true,
			Metadata: map[string]string{"key": "value"},
		}
		products = append(products, product)
	}
	res, err := repository.StripeProduct.Post(ctx, dbx, products)
	if err != nil {
		return nil, err
	}
	var prices []models.StripePrice
	for _, product := range res {
		price := models.StripePrice{
			ID:         uuid.NewString(),
			ProductID:  product.ID,
			UnitAmount: types.Pointer(int64(1000)),
			Currency:   "usd",
			Active:     true,
			Type:       models.StripePricingTypeRecurring,
			Interval:   types.Pointer(models.StripePricingPlanIntervalDay),
			Metadata:   map[string]string{"key": "value"},
		}
		prices = append(prices, price)
	}
	newPrices, err := repository.StripePrice.Post(ctx, dbx, prices)
	if err != nil {
		return nil, err
	}
	for _, prod := range res {
		for _, price := range newPrices {
			if prod.ID == price.ProductID {
				prod.Prices = append(prod.Prices, price)
			}
		}
	}

	return res, nil
}
