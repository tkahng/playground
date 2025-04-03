package apis

import (
	"context"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/dataloader"
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

type SubscriptionWithData struct {
	*models.StripeSubscription
	Price            *StripePricesWithProduct `json:"price,omitempty" required:"false"`
	SubscriptionUser *models.User             `json:"user,omitempty" required:"false"`
}

func (api *Api) AdminStripeSubscriptions(ctx context.Context,
	input *shared.StripeSubscriptionListParams,
) (*PaginatedOutput[*SubscriptionWithData], error) {
	db := api.app.Db()
	subscriptions, err := repository.ListSubscriptions(ctx, db, input)
	if err != nil {
		return nil, err
	}
	if slices.Contains(input.Expand, "user") {
		subscriptions.LoadStripeSubscriptionUser(ctx, db)
	}
	if slices.Contains(input.Expand, "price") {
		if slices.Contains(input.Expand, "product") {
			subscriptions.LoadStripeSubscriptionPriceStripePrice(ctx, db,
				models.PreloadStripePriceProductStripeProduct(),
			)
		} else {
			subscriptions.LoadStripeSubscriptionPriceStripePrice(ctx, db)
		}
	}
	subs := dataloader.Map(subscriptions, func(sub *models.StripeSubscription) *SubscriptionWithData {
		ss := &SubscriptionWithData{
			StripeSubscription: sub,
		}
		if sub.R.User != nil {
			ss.SubscriptionUser = sub.R.User
		}
		if sub.R.PriceStripePrice != nil {
			ss.Price = &StripePricesWithProduct{
				StripePrice: sub.R.PriceStripePrice,
			}
			if sub.R.PriceStripePrice.R.ProductStripeProduct != nil {
				ss.Price.Product = sub.R.PriceStripePrice.R.ProductStripeProduct
			}
		}
		return ss
	})
	count, err := repository.CountSubscriptions(ctx, db, &input.StripeSubscriptionListFilter)
	if err != nil {
		return nil, err
	}
	return &PaginatedOutput[*SubscriptionWithData]{
		Body: shared.PaginatedResponse[*SubscriptionWithData]{
			Data: subs,
			Meta: shared.Meta{
				Page:    input.PaginatedInput.Page,
				PerPage: input.PaginatedInput.PerPage,
				Total:   int(count),
			},
		},
	}, nil

}
