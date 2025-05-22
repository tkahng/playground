package apis

import (
	"context"

	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

type StripeProductsWithPricesInput struct {
	shared.PaginatedInput
	shared.SortParams
}

func (api *Api) StripeProductsWithPrices(ctx context.Context, inputt *StripeProductsWithPricesInput) (*shared.PaginatedOutput[*shared.StripeProductWithData], error) {
	input := &shared.StripeProductListParams{
		PaginatedInput: inputt.PaginatedInput,
		StripeProductListFilter: shared.StripeProductListFilter{
			Active: shared.Active,
		},
		SortParams: inputt.SortParams,
	}
	products, err := api.app.Payment().Store().ListProducts(ctx, input)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, u := range products {
		ids = append(ids, u.ID)
	}
	prices, err := api.app.Payment().Store().LoadProductPrices(ctx, &map[string]any{
		"product_id": map[string]any{
			"_in": ids,
		},
	}, ids...)
	if err != nil {
		return nil, err
	}
	for i, products := range products {
		price := prices[i]
		if len(price) > 0 {
			products.Prices = price
		}
	}

	count, err := api.app.Payment().Store().CountProducts(ctx, &input.StripeProductListFilter)
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
