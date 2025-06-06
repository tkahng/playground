package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/shared"
)

func (api *Api) GetStripeSubscriptions(ctx context.Context, input *struct{}) (*struct {
	Body *shared.Subscription `json:"body,omitempty" required:"false"`
}, error) {

	customer := contextstore.GetContextCurrentCustomer(ctx)
	if customer == nil {
		return nil, huma.Error403Forbidden("no customer found")
	}

	subWithPriceProduct, err := api.app.Payment().Store().FindLatestActiveSubscriptionWithPriceByCustomerId(ctx, customer.ID)
	if err != nil {
		return nil, err
	}
	output := &struct {
		Body *shared.Subscription `json:"body,omitempty" required:"false"`
	}{}

	if subWithPriceProduct == nil {
		return output, nil
	}

	return output, nil

}
