package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/shared"
)

type StripePaymentPayload struct {
	// StripeCustomerID string `json:"stripe_customer_id"`
	PriceID string `json:"price_id" required:"true"`
}
type StripeTeamPaymentInput struct {
	TeamID string               `path:"team-id" required:"true"`
	Body   StripePaymentPayload `json:"body" required:"true"`
}
type StripeUserPaymentInput struct {
	Body StripePaymentPayload `json:"body" required:"true"`
}

type StripeUrlOutput struct {
	// HxResponseHeaders
	Body struct {
		Url string `json:"url"`
	} `json:"body"`
}

func (a *Api) bindCreateTeamCheckoutSession(stripeGroup huma.API) {
	huma.Register(
		stripeGroup,
		huma.Operation{
			OperationID: "create-team-checkout-session",
			Method:      http.MethodPost,
			Path:        "/teams/{team-id}/subscriptions/checkout-session",
			Summary:     "create checkout session",
			Description: "user create checkout session",
			Tags:        []string{"Subscriptions", "Checkout Session"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
			Middlewares: huma.Middlewares{
				a.middlewares.TeamInfoFromParam,
				a.middlewares.SelectCustomerFromTeam,
			},
		},
		a.CreateTeamCheckoutSession,
	)
}
func (api *Api) CreateTeamCheckoutSession(ctx context.Context, input *StripeTeamPaymentInput) (*StripeUrlOutput, error) {
	customer := contextstore.GetContextCurrentCustomer(ctx)
	if customer == nil {
		return nil, huma.Error403Forbidden("No customer found")
	}
	if input.Body.PriceID == "" {
		return nil, huma.Error400BadRequest("Price ID is required")
	}
	url, err := api.App().Payment().CreateCheckoutSession(ctx, customer.ID, input.Body.PriceID)
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
func (a *Api) bindCreateUserCheckoutSession(stripeGroup huma.API) {
	huma.Register(
		stripeGroup,
		huma.Operation{
			OperationID: "create-checkout-session",
			Method:      http.MethodPost,
			Path:        "/subscriptions/checkout-session",
			Summary:     "create checkout session",
			Description: "user create checkout session",
			Tags:        []string{"Subscriptions", "Checkout Session"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
			Middlewares: huma.Middlewares{
				a.middlewares.SelectCustomerFromUser,
			},
		},
		a.CreateUserCheckoutSession,
	)
}

func (api *Api) CreateUserCheckoutSession(ctx context.Context, input *StripeUserPaymentInput) (*StripeUrlOutput, error) {
	customer := contextstore.GetContextCurrentCustomer(ctx)
	if customer == nil {
		return nil, huma.Error403Forbidden("No customer found")
	}
	if input.Body.PriceID == "" {
		return nil, huma.Error400BadRequest("Price ID is required")
	}
	url, err := api.App().Payment().CreateCheckoutSession(ctx, customer.ID, input.Body.PriceID)
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
func (a *Api) bindStripeBillingPortal(stripeGroup huma.API) {
	huma.Register(
		stripeGroup,
		huma.Operation{
			OperationID: "stripe-billing-portal",
			Method:      http.MethodPost,
			Path:        "/subscriptions/billing-portals",
			Summary:     "create user billing-portal",
			Description: "billing-portals",
			Tags:        []string{"Subscriptions", "Billing Portal"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
			Middlewares: huma.Middlewares{
				a.middlewares.SelectCustomerFromUser,
			},
		},
		a.StripeBillingPortal,
	)
}
func (api *Api) StripeBillingPortal(ctx context.Context, input *struct{}) (*StripeUrlOutput, error) {
	customer := contextstore.GetContextCurrentCustomer(ctx)
	if customer == nil {
		return nil, huma.Error403Forbidden("No customer found")
	}
	url, err := api.App().Payment().CreateBillingPortalSession(ctx, customer.ID)
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

func (a *Api) bindStripeTeamBillingPortal(stripeGroup huma.API) {
	huma.Register(
		stripeGroup,
		huma.Operation{
			OperationID: "stripe-billing-portal-team",
			Method:      http.MethodPost,
			Path:        "/teams/{team-id}/subscriptions/billing-portals",
			Summary:     "create team billing-portal",
			Description: "billing-portals",
			Tags:        []string{"Subscriptions", "Billing Portal", "Team"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
			Middlewares: huma.Middlewares{
				a.middlewares.TeamInfoFromParam,
				a.middlewares.SelectCustomerFromTeam,
			},
		},
		a.StripeTeamBillingPortal,
	)
}
func (api *Api) StripeTeamBillingPortal(ctx context.Context, input *struct {
	TeamID string `path:"team-id" required:"true"`
}) (*StripeUrlOutput, error) {
	customer := contextstore.GetContextCurrentCustomer(ctx)
	if customer == nil {
		return nil, huma.Error403Forbidden("No customer found")
	}
	url, err := api.App().Payment().CreateBillingPortalSession(ctx, customer.ID)
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

func (a *Api) bindStripeCheckoutSessionGet(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-checkout-session",
			Method:      http.MethodGet,
			Path:        "/subscriptions/checkout-session/{checkoutSessionId}",
			Summary:     "get checkout session",
			Description: "get checkout session",
			Tags:        []string{"Subscriptions", "Checkout Session"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		a.StripeCheckoutSessionGet,
	)
}

type CheckoutSessionOutput struct {
	Body StripeSubscription
}

type StripeCheckoutSessionInput struct {
	CheckoutSessionID string `path:"checkoutSessionId"`
}

func (api *Api) StripeCheckoutSessionGet(ctx context.Context, input *StripeCheckoutSessionInput) (*CheckoutSessionOutput, error) {
	info := contextstore.GetContextUserInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	payment := api.App().Payment()
	cs, err := payment.FindSubscriptionWithPriceProductBySessionId(ctx, input.CheckoutSessionID)
	if err != nil {
		return nil, err
	}
	if cs == nil {
		return nil, huma.Error404NotFound("checkout session not found")
	}
	if cs.StripeCustomer != nil {
		if cs.StripeCustomer.TeamID != nil {
			teamInfo, err := api.App().Team().FindTeamInfo(ctx, *cs.StripeCustomer.TeamID, info.User.ID)
			if err != nil {
				return nil, err
			}
			if teamInfo == nil {
				return nil, huma.Error404NotFound("you are not a member of the team this checkout session is for")
			}
			cs.StripeCustomer.Team = &teamInfo.Team
		}
		if cs.StripeCustomer.UserID != nil {
			if *cs.StripeCustomer.UserID != info.User.ID {
				return nil, huma.Error403Forbidden("you are not the user this checkout session is for")
			}
			cs.StripeCustomer.User = &info.User
		}

	}
	return &CheckoutSessionOutput{
		Body: *fromModelSubscription(cs),
	}, nil
}
