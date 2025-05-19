package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
)

func (api *Api) MyStripeSubscriptions(ctx context.Context, input *struct{}) (*struct {
	Body *shared.SubscriptionWithPrice `json:"body,omitempty" required:"false"`
}, error) {

	db := api.app.Db()
	user := contextstore.GetContextUserInfo(ctx)
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
			Price:   shared.FromCrudPrice(&subscriptions.Price),
			Product: shared.FromCrudProduct(&subscriptions.Product),
		},
	}

	return output, nil

}
