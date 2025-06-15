package apis

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
	"github.com/tkahng/authgo/internal/tools/utils"
)

type StripeWebhookInput struct {
	Signature string `header:"Stripe-Signature"`
	RawBody   []byte
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
		slog.ErrorContext(ctx, "⚠️  Webhook error while parsing basic request", slog.Any("error", err), slog.String("payload", string(payload)))
		return nil, huma.Error400BadRequest(err.Error())
	}
	cfg := api.app.Cfg()
	if cfg == nil {
		return nil, huma.Error400BadRequest("Missing config")
	}
	event, err := webhook.ConstructEvent(payload, input.Signature, cfg.Webhook)
	if err != nil {
		slog.ErrorContext(ctx, "⚠️  Webhook error while parsing basic request", slog.Any("error", err), slog.String("payload", string(payload)))
		return nil, huma.Error400BadRequest(err.Error())
	}
	payment := api.app.Payment()
	switch event.Type {
	case stripe.EventTypeProductCreated, stripe.EventTypeProductUpdated:
		product, err := utils.UnmarshalJSON[stripe.Product](event.Data.Raw)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to unmarshal product", err)
		}
		err = payment.UpsertProductFromStripe(ctx, &product)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to upsert product", err)
		}
		return nil, nil
	case stripe.EventTypePriceCreated, stripe.EventTypePriceUpdated:
		price, err := utils.UnmarshalJSON[stripe.Price](event.Data.Raw)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to unmarshal price", err)
		}
		err = payment.UpsertPriceFromStripe(ctx, &price)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to upsert price", err)
		}
		return nil, nil
	case stripe.EventTypeCheckoutSessionCompleted:
		session, err := utils.UnmarshalJSON[stripe.CheckoutSession](event.Data.Raw)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to unmarshal session", err)
		}
		err = payment.UpsertSubscriptionByIds(ctx, session.Customer.ID, session.Subscription.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to upsert checkout session complete", err)
		}
		return nil, nil
	case stripe.EventTypeCustomerSubscriptionCreated, stripe.EventTypeCustomerSubscriptionUpdated, stripe.EventTypeCustomerSubscriptionDeleted:
		subscription, err := utils.UnmarshalJSON[stripe.Subscription](event.Data.Raw)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to unmarshal subscription", err)
		}
		err = payment.UpsertSubscriptionByIds(ctx, subscription.Customer.ID, subscription.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to upsert subscription", err)
		}
		return nil, nil
	default:
		slog.WarnContext(ctx, "⚠️  Unhandled Stripe event type", slog.String("event_type", string(event.Type)))
		return nil, huma.Error400BadRequest("unhandled event type")
	}
	// return nil, huma.Error400BadRequest("unhandled event type")
}
