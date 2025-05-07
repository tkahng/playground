package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/shared"
)

type StripePaymentPayload struct {
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

func (a *Api) StripeCheckoutSessionOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "create-checkout-session",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "create checkout session",
		Description: "create checkout session",
		Tags:        []string{"Payment", "Stripe", "Checkout Session"},
		Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (a *Api) StripeCheckoutSession(ctx context.Context, input *StripePaymentInput) (*StripeUrlOutput, error) {
	db := a.app.Db()
	info := core.GetContextUserInfo(ctx)
	if info == nil {
		return nil, huma.Error403Forbidden("Not authenticated")
	}
	user := &info.User

	// return sesh.URL, nil
	url, err := a.app.Payment().CreateCheckoutSession(ctx, db, user.ID, input.Body.PriceID)
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

type StripeBillingPortalInput struct {
	// HxRequestHeaders
	// Body StripePaymentPayload
}

func (a *Api) StripeBillingPortalOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "stripe-billing-portal",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "billing-portal",
		Description: "billing-portals",
		Tags:        []string{"Payment", "Billing Portal", "Stripe"},
		Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (a *Api) StripeBillingPortal(ctx context.Context, input *StripeBillingPortalInput) (*StripeUrlOutput, error) {
	db := a.app.Db()
	info := core.GetContextUserInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("not authorized")
	}
	url, err := a.app.Payment().CreateBillingPortalSession(ctx, db, info.User.ID)
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

func (a *Api) StripeCheckoutSessionGetOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "get-checkout-session",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "get checkout session",
		Description: "get checkout session",
		Tags:        []string{"Payment", "Stripe", "Checkout Session"},
		Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type StripeCheckoutSessionInput struct {
	CheckoutSessionID string `path:"checkoutSessionId"`
}

func (a *Api) StripeCheckoutSessionGet(ctx context.Context, input *StripeCheckoutSessionInput) (*CheckoutSessionOutput, error) {
	db := a.app.Db()
	payment := a.app.Payment()
	cs, err := payment.FindSubscriptionWithPriceBySessionId(ctx, db, input.CheckoutSessionID)
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
