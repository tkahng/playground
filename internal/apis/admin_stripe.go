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
