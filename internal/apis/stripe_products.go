package apis

import (
	"context"

	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

type StripeProductsWithPricesInput struct {
	shared.PaginatedInput
	shared.SortParams
}

func (api *Api) StripeProductsWithPrices(ctx context.Context, inputt *StripeProductsWithPricesInput) (*shared.PaginatedOutput[*shared.StripeProduct], error) {
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
	prices, err := api.app.Payment().Store().LoadPricesByProductIds(ctx, ids...)
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

	return &shared.PaginatedOutput[*shared.StripeProduct]{Body: shared.PaginatedResponse[*shared.StripeProduct]{
		Data: mapper.Map(products, shared.FromModelProduct),
		Meta: shared.GenerateMeta(&input.PaginatedInput, count),
	}}, nil
}
