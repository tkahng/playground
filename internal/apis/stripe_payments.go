package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/shared"
)

type StripePaymentPayload struct {
	StripeCustomerID string `json:"stripe_customer_id"`
	PriceID          string `json:"price_id"`
}
type StripePaymentInput struct {
	// HxRequestHeaders
	Body StripePaymentPayload
}

type StripeUrlOutput struct {
	// HxResponseHeaders
	Body struct {
		Url string `json:"url"`
	}
}

func (a *Api) StripeCheckoutSession(ctx context.Context, input *StripePaymentInput) (*StripeUrlOutput, error) {
	info := contextstore.GetContextUserInfo(ctx)
	if info == nil {
		return nil, huma.Error403Forbidden("Not authenticated")
	}

	// team := contextstore.GetContextSelectedTeam(ctx)
	// if team == nil {
	// 	return nil, huma.Error400BadRequest("No team selected")
	// }
	// if team.StripeCustomerID == nil {
	// 	return nil, huma.Error400BadRequest("No stripe customer id")
	// }

	// return sesh.URL, nil
	url, err := a.app.Payment().CreateCheckoutSession(ctx, input.Body.StripeCustomerID, input.Body.PriceID)
	if err != nil {
		return nil, err
	}
	return &StripeUrlOutput{
		Body: struct {
			Url string `json:"url"`
		}{
			Url: url,
		},
	}, nil

}

type StripeBillingPortalBody struct {
	StripeCustomerID string `json:"stripe_customer_id"`
}
type StripeBillingPortalInput struct {
	// HxRequestHeaders
	Body StripeBillingPortalBody
}

func (a *Api) StripeBillingPortal(ctx context.Context, input *StripeBillingPortalInput) (*StripeUrlOutput, error) {

	// team := contextstore.GetContextSelectedTeam(ctx)
	// if team == nil {
	// 	return nil, huma.Error401Unauthorized("not authorized")
	// }
	// if team.StripeCustomerID == nil {
	// 	return nil, huma.Error400BadRequest("No stripe customer id")
	// }
	url, err := a.app.Payment().CreateBillingPortalSession(ctx, input.Body.StripeCustomerID)
	if err != nil {
		return nil, err
	}
	return &StripeUrlOutput{
		Body: struct {
			Url string `json:"url"`
		}{
			Url: url,
		},
	}, nil

}

type CheckoutSession struct {
	ID      string          `json:"id"`
	Price   *shared.Price   `json:"price"`
	Product *shared.Product `json:"product"`
}

type CheckoutSessionOutput struct {
	Body shared.SubscriptionWithPrice
}

type StripeCheckoutSessionInput struct {
	CheckoutSessionID string `path:"checkoutSessionId"`
}

func (a *Api) StripeCheckoutSessionGet(ctx context.Context, input *StripeCheckoutSessionInput) (*CheckoutSessionOutput, error) {

	payment := a.app.Payment()
	cs, err := payment.FindSubscriptionWithPriceBySessionId(ctx, input.CheckoutSessionID)
	if err != nil {
		return nil, err
	}
	if cs == nil {
		return nil, huma.Error404NotFound("checkout session not found")
	}

	return &CheckoutSessionOutput{
		Body: shared.SubscriptionWithPrice{
			Subscription: shared.FromCrudSubscription(&cs.Subscription),
			Price: &shared.StripePricesWithProduct{
				Price:   shared.FromCrudPrice(&cs.Price),
				Product: shared.FromCrudProduct(&cs.Product),
			},
		},
	}, nil
}
