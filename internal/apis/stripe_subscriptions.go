package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/types"
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
	customer, err := api.app.Payment().Store().FindCustomer(ctx, &models.StripeCustomer{
		TeamID: types.Pointer(member.TeamID),
	})
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, huma.Error400BadRequest("No stripe customer id")
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
