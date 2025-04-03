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

type StripeProductWithPrice struct {
	Product *models.StripeProduct `db:"product" json:"product"`
	Prices  []*models.StripePrice `db:"prices" json:"prices"`
}
type StripeProductsOutput struct {
	Body PaginatedOutput[StripeProductWithPrice] `json:"body"`
}

func (api *Api) StripeProductsWithPrices(ctx context.Context, input *shared.PaginatedInput) (*PaginatedOutput[*StripeProductWithPrice], error) {
	db := api.app.Db()
	utils.PrettyPrintJSON(input)
	users, err := repository.ListProducts(ctx, db, input)
	if err != nil {
		return nil, err
	}
	count, err := repository.CountProducts(ctx, db, &struct{}{})
	if err != nil {
		return nil, err
	}

	ids := dataloader.Map(users, func(user *models.StripeProduct) string {
		return user.ID
	})
	m := make(map[string][]*models.StripePrice)
	claims, err := repository.PricesByProductIds(ctx, db, ids)
	if err != nil {
		return nil, err
	}
	for _, claim := range claims {
		m[claim.ProductID] = append(m[claim.ProductID], claim)
	}
	info := dataloader.Map(users, func(user *models.StripeProduct) *StripeProductWithPrice {
		claims := m[user.ID]
		return &StripeProductWithPrice{
			Product: user,
			Prices:  claims,
		}
	})

	return &PaginatedOutput[*StripeProductWithPrice]{
		Body: shared.PaginatedResponse[*StripeProductWithPrice]{
			Data: info,
			Meta: shared.Meta{
				Page:    input.Page,
				PerPage: input.PerPage,
				Total:   int(count),
			},
		},
	}, nil
}
