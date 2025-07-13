package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/contextstore"
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
		Body: *FromModelSubscription(cs),
	}, nil
}
