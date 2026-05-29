package apis

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/tools/utils"
)

func (a *Api) bindStripeWebhook(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "stripe-webhook",
			Method:      http.MethodPost,
			Path:        "/stripe/webhook",
			Summary:     "webhook",
			Description: "Webhook for stripe",
			Tags:        []string{"Stripe", "Webhook"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
		},
		a.StripeWebhook,
	)
}

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

	cfg := api.App().Config()
	if cfg == nil {
		return nil, huma.Error500InternalServerError("server configuration unavailable")
	}
	event, err := webhook.ConstructEvent(payload, input.Signature, cfg.Webhook)
	if err != nil {
		// Signature failure → 400 so Stripe does not retry (it would fail again).
		slog.ErrorContext(ctx, "stripe webhook signature verification failed", slog.Any("error", err), slog.Int("payload_bytes", len(payload)))
		return nil, huma.Error400BadRequest("webhook signature verification failed")
	}
	payment := api.App().Payment()
	switch event.Type {
	case stripe.EventTypeProductCreated, stripe.EventTypeProductUpdated:
		product, err := utils.UnmarshalJSON[stripe.Product](event.Data.Raw)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to unmarshal product event")
		}
		if err = payment.UpsertProductFromStripe(ctx, &product); err != nil {
			slog.ErrorContext(ctx, "stripe webhook: upsert product failed", slog.Any("error", err))
			return nil, huma.Error500InternalServerError("internal error processing product event")
		}
		return nil, nil

	case stripe.EventTypePriceCreated, stripe.EventTypePriceUpdated:
		price, err := utils.UnmarshalJSON[stripe.Price](event.Data.Raw)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to unmarshal price event")
		}
		if err = payment.UpsertPriceFromStripe(ctx, &price); err != nil {
			slog.ErrorContext(ctx, "stripe webhook: upsert price failed", slog.Any("error", err))
			return nil, huma.Error500InternalServerError("internal error processing price event")
		}
		return nil, nil

	case stripe.EventTypeCheckoutSessionCompleted:
		cs, err := utils.UnmarshalJSON[stripe.CheckoutSession](event.Data.Raw)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to unmarshal checkout session event")
		}
		switch cs.Mode {
		case stripe.CheckoutSessionModePayment:
			purchaseType, ok := cs.Metadata["purchase_type"]
			if !ok || purchaseType != "points" {
				slog.WarnContext(ctx, "unhandled payment-mode checkout session", slog.String("session_id", cs.ID), slog.String("purchase_type", purchaseType))
				return nil, nil
			}
			userIDStr, ok := cs.Metadata["user_id"]
			if !ok {
				return nil, huma.Error400BadRequest("points purchase session missing user_id metadata")
			}
			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				return nil, huma.Error400BadRequest(fmt.Sprintf("invalid user_id in session metadata: %s", userIDStr))
			}
			pointsStr, ok := cs.Metadata["points_amount"]
			if !ok {
				return nil, huma.Error400BadRequest("points purchase session missing points_amount metadata")
			}
			pointsAmount, err := strconv.ParseInt(pointsStr, 10, 64)
			if err != nil || pointsAmount <= 0 {
				return nil, huma.Error400BadRequest(fmt.Sprintf("invalid points_amount in session metadata: %s", pointsStr))
			}
			txErr := api.App().Adapter().RunInTxCtx(ctx, func(txCtx context.Context) error {
				return services.FulfillPointsPurchase(txCtx, api.App().Adapter(), api.App().Ledger(), services.PointsPurchaseFulfillInput{
					UserID:          userID,
					PointsAmount:    pointsAmount,
					StripeSessionID: cs.ID,
				})
			})
			if txErr != nil {
				slog.ErrorContext(ctx, "stripe webhook: fulfill points purchase failed", slog.Any("error", txErr), slog.String("session_id", cs.ID))
				return nil, huma.Error500InternalServerError("internal error fulfilling points purchase")
			}
			return nil, nil

		case stripe.CheckoutSessionModeSubscription:
			if cs.Subscription == nil {
				return nil, huma.Error400BadRequest("subscription checkout session missing subscription field")
			}
			if err = payment.UpsertSubscriptionByIds(ctx, cs.Customer.ID, cs.Subscription.ID); err != nil {
				slog.ErrorContext(ctx, "stripe webhook: upsert subscription failed", slog.Any("error", err))
				return nil, huma.Error500InternalServerError("internal error processing subscription checkout")
			}
			return nil, nil

		default:
			slog.WarnContext(ctx, "unhandled checkout session mode", slog.String("session_id", cs.ID), slog.String("mode", string(cs.Mode)))
			return nil, nil
		}

	case stripe.EventTypeCustomerSubscriptionCreated, stripe.EventTypeCustomerSubscriptionUpdated, stripe.EventTypeCustomerSubscriptionDeleted:
		sub, err := utils.UnmarshalJSON[stripe.Subscription](event.Data.Raw)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to unmarshal subscription event")
		}
		if err = payment.UpsertSubscriptionByIds(ctx, sub.Customer.ID, sub.ID); err != nil {
			slog.ErrorContext(ctx, "stripe webhook: upsert subscription failed", slog.Any("error", err))
			return nil, huma.Error500InternalServerError("internal error processing subscription event")
		}
		return nil, nil

	default:
		slog.InfoContext(ctx, "stripe webhook: ignoring unhandled event type", slog.String("type", string(event.Type)))
		return nil, nil
	}
}
