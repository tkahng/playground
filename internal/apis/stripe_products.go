package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
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

func (api *Api) StripeProductsWithPrices(ctx context.Context, inputt *StripeProductsWithPricesInput) (*shared.PaginatedOutput[*shared.StripeProductWithPrices], error) {
	db := api.app.Db()
	input := &shared.StripeProductListParams{
		PaginatedInput: inputt.PaginatedInput,
		StripeProductListFilter: shared.StripeProductListFilter{
			Active: shared.Active,
		},
		SortParams: inputt.SortParams,
	}
	users, err := repository.ListProducts(ctx, db, input)
	if err != nil {
		return nil, err
	}

	err = users.LoadStripeProductProductStripePrices(ctx, db,
		models.SelectWhere.StripePrices.Active.EQ(true),
	)
	if err != nil {
		return nil, err
	}

	count, err := repository.CountProducts(ctx, db, &input.StripeProductListFilter)
	if err != nil {
		return nil, err
	}
	prods := mapper.Map(users, func(user *models.StripeProduct) *shared.StripeProductWithPrices {
		return &shared.StripeProductWithPrices{
			Product: shared.ModelToProduct(user),
			Prices:  mapper.Map(user.R.ProductStripePrices, shared.ModelToPrice),
		}

	})

	return &shared.PaginatedOutput[*shared.StripeProductWithPrices]{
		Body: shared.PaginatedResponse[*shared.StripeProductWithPrices]{
			Data: prods,
			Meta: shared.GenerateMeta(input.PaginatedInput, count),
		},
	}, nil
}
