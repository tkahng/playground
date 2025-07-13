package apis

import (
	"context"

	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mapper"
)

type StripeProductsWithPricesInput struct {
	PaginatedInput
	SortParams
}

func (api *Api) StripeProductsWithPrices(ctx context.Context, input *StripeProductsWithPricesInput) (*ApiPaginatedOutput[*StripeProduct], error) {

	filter := &stores.StripeProductFilter{}
	filter.Page = input.Page
	filter.PerPage = input.PerPage
	filter.Active.IsSet = true
	filter.Active.Value = true
	filter.SortBy = input.SortBy
	filter.SortOrder = input.SortOrder

	products, err := api.App().Adapter().Product().ListProducts(ctx, filter)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, u := range products {
		ids = append(ids, u.ID)
	}
	prices, err := api.App().Adapter().Price().LoadPricesByProductIds(ctx, ids...)
	if err != nil {
		return nil, err
	}
	for i, products := range products {
		price := prices[i]
		if len(price) > 0 {
			products.Prices = price
		}
	}

	count, err := api.App().Adapter().Product().CountProducts(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &ApiPaginatedOutput[*StripeProduct]{Body: ApiPaginatedResponse[*StripeProduct]{
		Data: mapper.Map(products, FromModelProduct),
		Meta: ApiGenerateMeta(&input.PaginatedInput, count),
	}}, nil
}
