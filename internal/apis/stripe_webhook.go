package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/tools/utils"
)

type StripeWebhookInput struct {
	Signature string `header:"Stripe-Signature"`
	RawBody   []byte
}

func (api *Api) StripeWebhookOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "stripe-webhook",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "webhook",
		Description: "Webhook for stripe",
		Tags:        []string{"Payment", "Stripe", "Webhook"},
		Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
	}
}

func (api *Api) StripeWebhook(ctx context.Context, input *StripeWebhookInput) (*struct{}, error) {
	if input == nil {
		return nil, huma.Error400BadRequest("Missing input")
	}
	if input.Signature == "" {
		return nil, huma.Error400BadRequest("Missing signature header")
	}
	payload := input.RawBody

	event := stripe.Event{}

	if err := json.Unmarshal(payload, &event); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Webhook error while parsing basic request. %v\n", err.Error())

		return nil, huma.Error400BadRequest(err.Error())
	}
	cfg := api.app.Cfg()
	if cfg == nil {
		return nil, huma.Error400BadRequest("Missing config")
	}
	event, err := webhook.ConstructEvent(payload, input.Signature, cfg.StripeConfig.Webhook)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Webhook signature verification failed. %v\n", err)
		return nil, huma.Error400BadRequest(err.Error())
	}
	db := api.app.Db()
	payment := api.app.Payment()
	switch event.Type {
	case stripe.EventTypeProductCreated, stripe.EventTypeProductUpdated:
		product, err := utils.UnmarshalJSON[stripe.Product](event.Data.Raw)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to unmarshal product", err)
		}
		err = repository.UpsertProductFromStripe(ctx, db, &product)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to upsert product", err)
		}
		return nil, nil
	case stripe.EventTypePriceCreated, stripe.EventTypePriceUpdated:
		price, err := utils.UnmarshalJSON[stripe.Price](event.Data.Raw)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to unmarshal price", err)
		}
		err = repository.UpsertPriceFromStripe(ctx, db, &price)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to upsert price", err)
		}
		return nil, nil
	case stripe.EventTypeCheckoutSessionCompleted:
		session, err := utils.UnmarshalJSON[stripe.CheckoutSession](event.Data.Raw)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to unmarshal session", err)
		}
		err = payment.UpsertSubscriptionByIds(ctx, db, session.Customer.ID, session.Subscription.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to upsert checkout session complete", err)
		}
		return nil, nil
	case stripe.EventTypeCustomerSubscriptionCreated, stripe.EventTypeCustomerSubscriptionUpdated, stripe.EventTypeCustomerSubscriptionDeleted:
		subscription, err := utils.UnmarshalJSON[stripe.Subscription](event.Data.Raw)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to unmarshal subscription", err)
		}
		err = payment.UpsertSubscriptionByIds(ctx, db, subscription.Customer.ID, subscription.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to upsert subscription", err)
		}
		return nil, nil
	default:
		fmt.Fprintf(os.Stderr, "⚠️  Unhandled event type: %s\n", event.Type)
		return nil, huma.Error400BadRequest("unhandled event type")
	}
	// return nil, huma.Error400BadRequest("unhandled event type")
}
