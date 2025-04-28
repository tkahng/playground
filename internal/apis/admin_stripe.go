package apis

import (
	"context"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

func (api *Api) AdminStripeSubscriptionsOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-stripe-subscriptions",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Admin stripe subscriptions",
		Description: "List of stripe subscriptions",
		Tags:        []string{"Admin", "Subscription", "Stripe"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminStripeSubscriptions(ctx context.Context,
	input *shared.StripeSubscriptionListParams,
) (*shared.PaginatedOutput[*shared.SubscriptionWithData], error) {
	db := api.app.Db()
	subscriptions, err := repository.ListSubscriptions(ctx, db, input)
	if err != nil {
		return nil, err
	}
	if slices.Contains(input.Expand, "user") {
		err = subscriptions.LoadStripeSubscriptionUser(ctx, db)
		if err != nil {
			return nil, err
		}
	}
	if slices.Contains(input.Expand, "price") {
		if slices.Contains(input.Expand, "product") {
			err = subscriptions.LoadStripeSubscriptionPriceStripePrice(ctx, db,
				models.PreloadStripePriceProductStripeProduct(),
			)
			if err != nil {
				return nil, err
			}
		} else {
			err = subscriptions.LoadStripeSubscriptionPriceStripePrice(ctx, db)
			if err != nil {
				return nil, err
			}
		}
	}
	subs := mapper.Map(subscriptions, func(sub *models.StripeSubscription) *shared.SubscriptionWithData {
		ss := &shared.SubscriptionWithData{
			Subscription: shared.ModelToSubscription(sub),
		}
		if sub.R.User != nil {
			ss.SubscriptionUser = shared.ToUser(sub.R.User)
		}
		if sub.R.PriceStripePrice != nil {
			ss.Price = &shared.StripePricesWithProduct{
				Price: shared.ModelToPrice(sub.R.PriceStripePrice),
			}
			if sub.R.PriceStripePrice.R.ProductStripeProduct != nil {
				ss.Price.Product = shared.ModelToProduct(sub.R.PriceStripePrice.R.ProductStripeProduct)
			}
		}
		return ss
	})
	count, err := repository.CountSubscriptions(ctx, db, &input.StripeSubscriptionListFilter)
	if err != nil {
		return nil, err
	}
	return &shared.PaginatedOutput[*shared.SubscriptionWithData]{
		Body: shared.PaginatedResponse[*shared.SubscriptionWithData]{
			Data: subs,
			Meta: shared.GenerateMeta(input.PaginatedInput, count),
		},
	}, nil

}

func (api *Api) AdminStripeProductsOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-stripe-products",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Admin stripe products",
		Description: "List of stripe products",
		Tags:        []string{"Admin", "Product", "Stripe"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}
func (api *Api) AdminStripeProducts(ctx context.Context,
	input *shared.StripeProductListParams,
) (*shared.PaginatedOutput[*shared.StripeProductWithData], error) {
	db := api.app.Db()
	products, err := repository.ListProducts(ctx, db, input)
	if err != nil {
		return nil, err
	}
	if slices.Contains(input.Expand, "prices") {
		err = products.LoadStripeProductProductStripePrices(ctx, db)
		if err != nil {
			return nil, err
		}
	}
	if slices.Contains(input.Expand, "roles") {
		err = products.LoadStripeProductRoles(ctx, db)
		if err != nil {
			return nil, err
		}
	}
	prods := mapper.Map(products, func(p *models.StripeProduct) *shared.StripeProductWithData {
		spwd := &shared.StripeProductWithData{
			Product: shared.ModelToProduct(p),
		}
		if p.R.ProductStripePrices != nil {
			// If the product has prices, we map them to the shared model
			// and include them in the response.
			spwd.Prices = mapper.Map(p.R.ProductStripePrices, shared.ModelToPrice)
		}
		if p.R.Roles != nil {
			// If the product has prices and we are expanding prices,
			spwd.Roles = mapper.Map(p.R.Roles, shared.ToRole)
		}
		return spwd
	})
	count, err := repository.CountProducts(ctx, db, &input.StripeProductListFilter)
	if err != nil {
		return nil, err
	}
	return &shared.PaginatedOutput[*shared.StripeProductWithData]{
		Body: shared.PaginatedResponse[*shared.StripeProductWithData]{
			Data: prods,
			Meta: shared.GenerateMeta(input.PaginatedInput, count),
		},
	}, nil
}
