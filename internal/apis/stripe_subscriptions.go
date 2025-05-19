package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/shared"
)

func (api *Api) MyStripeSubscriptions(ctx context.Context, input *struct{}) (*struct {
	Body *shared.SubscriptionWithPrice `json:"body,omitempty" required:"false"`
}, error) {

	user := contextstore.GetContextUserInfo(ctx)
	if user == nil {
		return nil, huma.Error401Unauthorized("not authorized")
	}
	member, err := api.app.Team().Store().FindLatestTeamMemberByUserID(ctx, user.User.ID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, huma.Error400BadRequest("No team selected")
	}
	subscriptions, err := api.app.Payment().Store().FindLatestActiveSubscriptionWithPriceByTeamId(ctx, member.TeamID)
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
