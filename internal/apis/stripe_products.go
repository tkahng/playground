package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/dataloader"
	"github.com/tkahng/authgo/internal/tools/utils"
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

type StripeProductWithPrices struct {
	*models.StripeProduct
	Prices []*models.StripePrice `db:"prices" json:"prices"`
}

func (api *Api) StripeProductsWithPrices(ctx context.Context, input *shared.StripeProductListParams) (*PaginatedOutput[*StripeProductWithPrices], error) {
	db := api.app.Db()
	utils.PrettyPrintJSON(input)
	users, err := repository.ListProducts(ctx, db, input)
	if err != nil {
		return nil, err
	}
	err = users.LoadStripeProductProductStripePrices(ctx, db)
	if err != nil {
		return nil, err
	}
	count, err := repository.CountProducts(ctx, db, &input.StripeProductListFilter)
	if err != nil {
		return nil, err
	}
	prods := dataloader.Map(users, func(user *models.StripeProduct) *StripeProductWithPrices {
		return &StripeProductWithPrices{
			StripeProduct: user,
			Prices:        user.R.ProductStripePrices,
		}

	})

	return &PaginatedOutput[*StripeProductWithPrices]{
		Body: shared.PaginatedResponse[*StripeProductWithPrices]{
			Data: prods,
			Meta: shared.Meta{
				Page:    input.Page,
				PerPage: input.PerPage,
				Total:   int(count),
			},
		},
	}, nil
}
