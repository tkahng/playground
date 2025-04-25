package apis

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
)

type Subscription struct {
	ID                 string                          `db:"id,pk" json:"id"`
	UserID             uuid.UUID                       `db:"user_id" json:"user_id"`
	Status             models.StripeSubscriptionStatus `db:"status" json:"status"`
	Metadata           map[string]string               `db:"metadata" json:"metadata"`
	PriceID            string                          `db:"price_id" json:"price_id"`
	Quantity           int64                           `db:"quantity" json:"quantity"`
	CancelAtPeriodEnd  bool                            `db:"cancel_at_period_end" json:"cancel_at_period_end"`
	Created            time.Time                       `db:"created" json:"created"`
	CurrentPeriodStart time.Time                       `db:"current_period_start" json:"current_period_start"`
	CurrentPeriodEnd   time.Time                       `db:"current_period_end" json:"current_period_end"`
	EndedAt            *time.Time                      `db:"ended_at" json:"ended_at"`
	CancelAt           *time.Time                      `db:"cancel_at" json:"cancel_at"`
	CanceledAt         *time.Time                      `db:"canceled_at" json:"canceled_at"`
	TrialStart         *time.Time                      `db:"trial_start" json:"trial_start"`
	TrialEnd           *time.Time                      `db:"trial_end" json:"trial_end"`
	CreatedAt          time.Time                       `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time                       `db:"updated_at" json:"updated_at"`
}

func ModelToSubscription(model *models.StripeSubscription) *Subscription {
	return &Subscription{
		ID:                 model.ID,
		UserID:             model.UserID,
		Status:             model.Status,
		Metadata:           model.Metadata.Val,
		PriceID:            model.PriceID,
		Quantity:           model.Quantity,
		CancelAtPeriodEnd:  model.CancelAtPeriodEnd,
		Created:            model.Created,
		CurrentPeriodStart: model.CurrentPeriodStart,
		CurrentPeriodEnd:   model.CurrentPeriodEnd,
		EndedAt:            model.EndedAt.Ptr(),
		CancelAt:           model.CancelAt.Ptr(),
		CanceledAt:         model.CanceledAt.Ptr(),
		TrialStart:         model.TrialStart.Ptr(),
		TrialEnd:           model.TrialEnd.Ptr(),
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
	}
}

type SubscriptionWithPrice struct {
	*Subscription
	Price *StripePricesWithProduct `json:"price,omitempty" required:"false"`
}

func (api *Api) MyStripeSubscriptionsOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "stripe-my-subscriptions",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "stripe-my-subscriptions",
		Description: "stripe-my-subscriptions",
		Tags:        []string{"Payment", "Stripe", "Subscriptions"},
		Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) MyStripeSubscriptions(ctx context.Context, input *struct{}) (*struct {
	Body *SubscriptionWithPrice `json:"body,omitempty" required:"false"`
}, error) {
	output := &struct {
		Body *SubscriptionWithPrice `json:"body,omitempty" required:"false"`
	}{}
	db := api.app.Db()
	user := core.GetContextUserClaims(ctx)
	if user == nil {
		return nil, huma.Error401Unauthorized("not authorized")
	}
	subscriptions, err := repository.FindLatestActiveSubscriptionWithPriceByUserId(ctx, db, user.User.ID)
	if err != nil {
		return nil, err
	}
	if subscriptions == nil {
		return output, nil
	}
	output.Body = &SubscriptionWithPrice{
		Subscription: ModelToSubscription(subscriptions),
	}
	var price *models.StripePrice
	var product *models.StripeProduct
	if subscriptions.R.PriceStripePrice != nil {
		price = subscriptions.R.PriceStripePrice
		if price.R.ProductStripeProduct != nil {
			product = price.R.ProductStripeProduct
			output.Body.Price = &StripePricesWithProduct{
				Price:   ModelToPrice(price),
				Product: ModelToProduct(product),
			}
		}
	}
	// subscriptions.LoadStripeSubscriptionPriceStripePrice(ctx, db, models.PreloadStripePriceProductStripeProduct())
	return output, nil

}
