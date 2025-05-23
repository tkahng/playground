package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/shared"
)

type StripePaymentPayload struct {
	// StripeCustomerID string `json:"stripe_customer_id"`
	PriceID string `json:"price_id"`
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
	customer := contextstore.GetContextCurrentCustomer(ctx)
	if customer == nil {
		return nil, huma.Error403Forbidden("No customer found")
	}
	if input.Body.PriceID == "" {
		return nil, huma.Error400BadRequest("Price ID is required")
	}
	url, err := a.app.Payment().CreateCheckoutSession(ctx, customer.ID, input.Body.PriceID)
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
}
type StripeBillingPortalInput struct {
	// HxRequestHeaders
	Body StripeBillingPortalBody
}

func (a *Api) StripeBillingPortal(ctx context.Context, input *struct{}) (*StripeUrlOutput, error) {
	customer := contextstore.GetContextCurrentCustomer(ctx)
	if customer == nil {
		return nil, huma.Error403Forbidden("No customer found")
	}
	url, err := a.app.Payment().CreateBillingPortalSession(ctx, customer.ID)
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
			Subscription: shared.FromModelSubscription(&cs.Subscription),
			Price: &shared.StripePricesWithProduct{
				Price:   shared.FromModelPrice(&cs.Price),
				Product: shared.FromModelProduct(&cs.Product),
			},
		},
	}, nil
}
