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

type SubscriptionWithData struct {
	*Subscription
	Price            *StripePricesWithProduct `json:"price,omitempty" required:"false"`
	SubscriptionUser *shared.User       `json:"user,omitempty" required:"false"`
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
	subs := mapper.Map(subscriptions, func(sub *models.StripeSubscription) *SubscriptionWithData {
		ss := &SubscriptionWithData{
			Subscription: ModelToSubscription(sub),
		}
		if sub.R.User != nil {
			ss.SubscriptionUser = shared.ToUser(sub.R.User)
		}
		if sub.R.PriceStripePrice != nil {
			ss.Price = &StripePricesWithProduct{
				Price: &Price{
					ID:              sub.R.PriceStripePrice.ID,
					Active:          sub.R.PriceStripePrice.Active,
					UnitAmount:      sub.R.PriceStripePrice.UnitAmount.Ptr(),
					Currency:        sub.R.PriceStripePrice.Currency,
					Type:            sub.R.PriceStripePrice.Type,
					Interval:        sub.R.PriceStripePrice.Interval.Ptr(),
					IntervalCount:   sub.R.PriceStripePrice.IntervalCount.Ptr(),
					TrialPeriodDays: sub.R.PriceStripePrice.TrialPeriodDays.Ptr(),
					ProductID:       sub.R.PriceStripePrice.ProductID,
					Metadata:        sub.R.PriceStripePrice.Metadata.Val,
					CreatedAt:       sub.R.PriceStripePrice.CreatedAt,
					UpdatedAt:       sub.R.PriceStripePrice.UpdatedAt,
					LookupKey:       sub.R.PriceStripePrice.LookupKey.Ptr(),
				},
			}
			if sub.R.PriceStripePrice.R.ProductStripeProduct != nil {
				ss.Price.Product = &Product{
					ID:          sub.R.PriceStripePrice.R.ProductStripeProduct.ID,
					Active:      sub.R.PriceStripePrice.R.ProductStripeProduct.Active,
					Name:        sub.R.PriceStripePrice.R.ProductStripeProduct.Name,
					Description: sub.R.PriceStripePrice.R.ProductStripeProduct.Description.Ptr(),
					Image:       sub.R.PriceStripePrice.R.ProductStripeProduct.Image.Ptr(),
					Metadata:    sub.R.PriceStripePrice.R.ProductStripeProduct.Metadata.Val,
					CreatedAt:   sub.R.PriceStripePrice.R.ProductStripeProduct.CreatedAt,
					UpdatedAt:   sub.R.PriceStripePrice.R.ProductStripeProduct.UpdatedAt,
				}
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
