package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
)

func (api *Api) MyStripeSubscriptionsOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "stripe-my-subscriptions",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "stripe-my-subscriptions",
		Description: "stripe-my-subscriptions",
		Tags:        []string{"Payment", "Stripe", "Subscriptions"},
		Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) MyStripeSubscriptions(ctx context.Context, input *struct{}) (*struct {
	Body *shared.SubscriptionWithPrice `json:"body,omitempty" required:"false"`
}, error) {

	db := api.app.Db()
	user := core.GetContextUserInfo(ctx)
	if user == nil {
		return nil, huma.Error401Unauthorized("not authorized")
	}
	subscriptions, err := queries.FindLatestActiveSubscriptionWithPriceByUserId(ctx, db, user.User.ID)
	if err != nil {
		return nil, err
	}
	if subscriptions == nil {
		return nil, nil
	}
	output := &struct {
		Body *shared.SubscriptionWithPrice `json:"body,omitempty" required:"false"`
	}{}
	output.Body = &shared.SubscriptionWithPrice{
		Subscription: shared.FromCrudSubscription(&subscriptions.Subscription),
		Price: &shared.StripePricesWithProduct{
			Price:   shared.FromCrudModel(&subscriptions.Price),
			Product: shared.FromCrudProduct(&subscriptions.Product),
		},
	}

	// subscriptions.LoadStripeSubscriptionPriceStripePrice(ctx, db, models.PreloadStripePriceProductStripeProduct())
	return output, nil

}
