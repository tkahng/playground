package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/crud/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

func (api *Api) StripeProductsWithPricesOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "stripe-products-with-prices",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "stripe-products-with-prices",
		Description: "stripe-products-with-prices",
		Tags:        []string{"Payment", "Stripe", "Products"},
		Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
	}
}

type StripeProductsWithPricesInput struct {
	shared.PaginatedInput
	shared.SortParams
}

func (api *Api) StripeProductsWithPrices(ctx context.Context, inputt *StripeProductsWithPricesInput) (*shared.PaginatedOutput[*shared.StripeProductWithData], error) {
	db := api.app.Db()
	input := &shared.StripeProductListParams{
		PaginatedInput: inputt.PaginatedInput,
		StripeProductListFilter: shared.StripeProductListFilter{
			Active: shared.Active,
		},
		SortParams: inputt.SortParams,
	}
	products, err := queries.ListProducts(ctx, db, input)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, u := range products {
		ids = append(ids, u.ID)
	}
	prices, err := queries.LoadeProductPrices(ctx, db, &map[string]any{
		"product_id": map[string]any{
			"_in": ids,
		},
	}, ids...)
	for i, products := range products {
		price := prices[i]
		if len(price) > 0 {
			products.Prices = price
		}
	}
	if err != nil {
		return nil, err
	}

	count, err := queries.CountProducts(ctx, db, &input.StripeProductListFilter)
	if err != nil {
		return nil, err
	}

	return &shared.PaginatedOutput[*shared.StripeProductWithData]{
		Body: shared.PaginatedResponse[*shared.StripeProductWithData]{
			Data: mapper.Map(products, func(p *models.StripeProduct) *shared.StripeProductWithData {
				return &shared.StripeProductWithData{
					Product: shared.FromCrudProduct(p),
					Roles:   mapper.Map(p.Roles, shared.FromCrudRole),
					Prices:  mapper.Map(p.Prices, shared.FromCrudPrice),
				}
			}),
			Meta: shared.GenerateMeta(input.PaginatedInput, count),
		},
	}, nil
}
