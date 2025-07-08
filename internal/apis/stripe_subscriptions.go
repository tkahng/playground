package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/contextstore"
)

func (api *Api) GetStripeSubscriptions(ctx context.Context, input *struct{}) (*struct {
	Body *StripeSubscription `json:"body,omitempty" required:"false"`
}, error) {

	customer := contextstore.GetContextCurrentCustomer(ctx)
	if customer == nil {
		return nil, huma.Error403Forbidden("no customer found")
	}

	subWithPriceProduct, err := api.App().Adapter().Subscription().FindActiveSubscriptionByCustomerId(ctx, customer.ID)
	if err != nil {
		return nil, err
	}

	output := &struct {
		Body *StripeSubscription `json:"body,omitempty" required:"false"`
	}{
		Body: FromModelSubscription(subWithPriceProduct),
	}

	return output, nil

}
func (api *Api) GetTeamStripeSubscriptions(ctx context.Context, input *struct {
	TeamID string `path:"team-id" json:"team_id" format:"uuid" required:"true"`
}) (*struct {
	Body *StripeSubscription `json:"body,omitempty" required:"false"`
}, error) {

	customer := contextstore.GetContextCurrentCustomer(ctx)
	if customer == nil {
		return nil, huma.Error403Forbidden("no customer found")
	}

	subWithPriceProduct, err := api.App().Adapter().Subscription().FindActiveSubscriptionByCustomerId(ctx, customer.ID)
	if err != nil {
		return nil, err
	}

	output := &struct {
		Body *StripeSubscription `json:"body,omitempty" required:"false"`
	}{
		Body: FromModelSubscription(subWithPriceProduct),
	}

	return output, nil

}
