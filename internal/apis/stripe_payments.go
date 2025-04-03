package apis

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/tkahng/authgo/internal/types"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/repository"
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
	priceId := input.Body.PriceID
	db := a.app.Db()
	userr := core.GetContextUserClaims(ctx)
	if userr == nil || userr.User == nil {
		return nil, huma.Error403Forbidden("Not authenticated")
	}
	user := userr.User

	srv := a.app.Payment()
	dbcus, err := srv.FindOrCreateCustomerFromUser(ctx, db, user)
	if err != nil {
		return nil, err
	}
	val, err := repository.FindLatestActiveSubscriptionByUserId(ctx, db, user.ID)
	if err != nil {
		return nil, err
	}
	if val != nil {
		return nil, errors.New("user already has a valid subscription")
	}
	firstSub, err := repository.IsFirstSubscription(ctx, db, user.ID)
	if err != nil {
		return nil, err
	}
	var trialDays *int64
	if firstSub {
		trialDays = types.Pointer(int64(14))
	} else {
		trialDays = nil
	}
	valPrice, err := repository.FindValidPriceById(ctx, db, priceId)
	if err != nil {
		return nil, err
	}
	if valPrice == nil {
		return nil, errors.New("price is not valid")
	}
	sesh, err := srv.Client().CreateCheckoutSession(dbcus.StripeID, priceId, trialDays)
	if err != nil {
		return nil, err
	}
	// return sesh.URL, nil
	// url, err := a.app.Payment().CreateCheckoutSession(ctx, db, user.User, input.Body.PriceID)
	// if err != nil {
	// 	return nil, err
	// }
	return &StripeUrlOutput{
		Body: struct {
			Url string `json:"url"`
		}{
			Url: sesh.URL,
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
	user := core.GetContextUserClaims(ctx)
	if user == nil || user.User == nil {
		return nil, huma.Error401Unauthorized("not authorized")
	}
	srv := a.app.Payment()
	dbcus, err := srv.FindOrCreateCustomerFromUser(ctx, db, user.User)
	if err != nil {
		return nil, err
	}
	// verify user has a valid subscriptio
	sub, err := repository.FindLatestActiveSubscriptionByUserId(ctx, db, user.User.ID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, huma.Error400BadRequest("no subscription.  subscribe to access billing portal")
	}
	url, err := srv.Client().CreateBillingPortalSession(dbcus.StripeID)
	if err != nil {
		log.Println(err)
		return nil, huma.Error500InternalServerError("failed to create checkout session")
	}
	if url == nil {
		return nil, huma.Error500InternalServerError("failed to create checkout session")
	}
	// return , nil
	// url, err := a.app.Payment().CreateBillingPortalSession(ctx, db, user.User)
	// if err != nil {
	// 	return nil, err
	// }
	return &StripeUrlOutput{
		Body: struct {
			Url string `json:"url"`
		}{
			Url: url.URL,
		},
	}, nil

}
