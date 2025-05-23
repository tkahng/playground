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

func (api *Api) StripeProductsWithPrices(ctx context.Context, inputt *StripeProductsWithPricesInput) (*shared.PaginatedOutput[*shared.StripeProductWitPermission], error) {
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

	return &shared.PaginatedOutput[*shared.StripeProductWitPermission]{
		Body: shared.PaginatedResponse[*shared.StripeProductWitPermission]{
			Data: mapper.Map(products, func(p *models.StripeProduct) *shared.StripeProductWitPermission {
				return &shared.StripeProductWitPermission{
					Product:     shared.FromModelProduct(p),
					Permissions: mapper.Map(p.Permissions, shared.FromModelPermission),
					Prices:      mapper.Map(p.Prices, shared.FromModelPrice),
				}
			}),
			Meta: shared.GenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}
