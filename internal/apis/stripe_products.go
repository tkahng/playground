package apis

import (
	"context"

	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

type StripeProductsWithPricesInput struct {
	shared.PaginatedInput
	shared.SortParams
}

func (api *Api) StripeProductsWithPrices(ctx context.Context, input *StripeProductsWithPricesInput) (*ApiPaginatedOutput[*shared.StripeProduct], error) {

	filter := &stores.StripeProductFilter{}
	filter.Page = input.Page
	filter.PerPage = input.PerPage
	filter.Active.IsSet = true
	filter.Active.Value = true
	filter.SortBy = input.SortBy
	filter.SortOrder = input.SortOrder

	products, err := api.app.Adapter().Product().ListProducts(ctx, filter)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, u := range products {
		ids = append(ids, u.ID)
	}
	prices, err := api.app.Adapter().Price().LoadPricesByProductIds(ctx, ids...)
	if err != nil {
		return nil, err
	}
	for i, products := range products {
		price := prices[i]
		if len(price) > 0 {
			products.Prices = price
		}
	}

	count, err := api.app.Adapter().Product().CountProducts(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &ApiPaginatedOutput[*shared.StripeProduct]{Body: ApiPaginatedResponse[*shared.StripeProduct]{
		Data: mapper.Map(products, shared.FromModelProduct),
		Meta: GenerateMeta(&input.PaginatedInput, count),
	}}, nil
}
