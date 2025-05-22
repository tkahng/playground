package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/shared"
)

func (api *Api) GetStripeSubscriptions(ctx context.Context, input *struct{}) (*struct {
	Body *shared.SubscriptionWithPrice `json:"body,omitempty" required:"false"`
}, error) {

	customer := contextstore.GetContextCurrentCustomer(ctx)
	if customer == nil {
		return nil, huma.Error403Forbidden("no customer found")
	}

	subscriptions, err := api.app.Payment().Store().FindLatestActiveSubscriptionWithPriceByCustomerId(ctx, customer.ID)
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
			Price:   shared.FromCrudPrice(&subscriptions.Price),
			Product: shared.FromCrudProduct(&subscriptions.Product),
		},
	}

	return output, nil

}
