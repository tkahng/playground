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

func (a *Api) StripePaymentOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "create-checkout-session",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "create checkout session",
		Description: "create checkout session",
		Tags:        []string{"Stripe"},
		Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (a *Api) CreateCheckoutSession(ctx context.Context, input *StripePaymentInput) (*StripeUrlOutput, error) {
	db := a.app.Db()
	user := core.GetContextUserClaims(ctx)
	if user == nil || user.User == nil {
		return nil, huma.Error403Forbidden("Not authenticated")
	}
	url, err := a.app.Payment().CreateCheckoutSession(ctx, db, user.User, input.Body.PriceID)
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
		Tags:        []string{"Stripe"},
		Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (a *Api) StripeBillingPortal(ctx context.Context, input *StripeBillingPortalInput) (*StripeUrlOutput, error) {
	db := a.app.Db()
	user := core.GetContextUserClaims(ctx)
	if user == nil || user.User == nil {
		return nil, huma.Error401Unauthorized("not authorized")
	}
	url, err := a.app.Payment().CreateBillingPortalSession(ctx, db, user.User)
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
