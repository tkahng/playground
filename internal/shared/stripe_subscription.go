package shared

import (
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crud/models"
)

// enum:"trialing,active,canceled,incomplete,incomplete_expired,past_due,unpaid,paused"
type StripeSubscriptionStatus string

const (
	StripeSubscriptionStatusTrialing          StripeSubscriptionStatus = "trialing"
	StripeSubscriptionStatusActive            StripeSubscriptionStatus = "active"
	StripeSubscriptionStatusCanceled          StripeSubscriptionStatus = "canceled"
	StripeSubscriptionStatusIncomplete        StripeSubscriptionStatus = "incomplete"
	StripeSubscriptionStatusIncompleteExpired StripeSubscriptionStatus = "incomplete_expired"
	StripeSubscriptionStatusPastDue           StripeSubscriptionStatus = "past_due"
	StripeSubscriptionStatusUnpaid            StripeSubscriptionStatus = "unpaid"
	StripeSubscriptionStatusPaused            StripeSubscriptionStatus = "paused"
)

type Subscription struct {
	ID                 string                   `db:"id,pk" json:"id"`
	UserID             uuid.UUID                `db:"user_id" json:"user_id"`
	Status             StripeSubscriptionStatus `db:"status" json:"status"`
	Metadata           map[string]string        `db:"metadata" json:"metadata"`
	PriceID            string                   `db:"price_id" json:"price_id"`
	Quantity           int64                    `db:"quantity" json:"quantity"`
	CancelAtPeriodEnd  bool                     `db:"cancel_at_period_end" json:"cancel_at_period_end"`
	Created            time.Time                `db:"created" json:"created"`
	CurrentPeriodStart time.Time                `db:"current_period_start" json:"current_period_start"`
	CurrentPeriodEnd   time.Time                `db:"current_period_end" json:"current_period_end"`
	EndedAt            *time.Time               `db:"ended_at" json:"ended_at"`
	CancelAt           *time.Time               `db:"cancel_at" json:"cancel_at"`
	CanceledAt         *time.Time               `db:"canceled_at" json:"canceled_at"`
	TrialStart         *time.Time               `db:"trial_start" json:"trial_start"`
	TrialEnd           *time.Time               `db:"trial_end" json:"trial_end"`
	CreatedAt          time.Time                `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time                `db:"updated_at" json:"updated_at"`
}
type SubscriptionWithPrice struct {
	*Subscription
	Price *StripePricesWithProduct `json:"price,omitempty" required:"false"`
}

func FromCrudToSubWithUserAndPrice(sub *models.SubscriptionWithPrice) *SubscriptionWithData {
	return &SubscriptionWithData{
		Subscription: FromCrudSubscription(&sub.Subscription),
		Price: &StripePricesWithProduct{
			Product: FromCrudProduct(&sub.Product),
			Price:   FromCrudPrice(&sub.Price),
		},
	}
}

func FromCrudSubscription(sub *models.StripeSubscription) *Subscription {
	return &Subscription{
		ID:                 sub.ID,
		UserID:             sub.UserID,
		Status:             StripeSubscriptionStatus(sub.Status),
		Metadata:           sub.Metadata,
		PriceID:            sub.PriceID,
		Quantity:           sub.Quantity,
		CancelAtPeriodEnd:  sub.CancelAtPeriodEnd,
		Created:            sub.Created,
		CurrentPeriodStart: sub.CurrentPeriodStart,
		CurrentPeriodEnd:   sub.CurrentPeriodEnd,
		EndedAt:            sub.EndedAt,
		CancelAt:           sub.CancelAt,
		CanceledAt:         sub.CanceledAt,
		TrialStart:         sub.TrialStart,
		TrialEnd:           sub.TrialEnd,
		CreatedAt:          sub.CreatedAt,
		UpdatedAt:          sub.UpdatedAt,
	}
}

type SubscriptionWithData struct {
	*Subscription
	Price            *StripePricesWithProduct `json:"price,omitempty" required:"false"`
	SubscriptionUser *User                    `json:"user,omitempty" required:"false"`
}

type StripeSubscriptionListFilter struct {
	Q      string                     `query:"q,omitempty" required:"false"`
	Ids    []string                   `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	UserID string                     `query:"user_id,omitempty" required:"false" format:"uuid"`
	Status []StripeSubscriptionStatus `query:"status,omitempty" required:"false" minimum:"1" maximum:"100" enum:"trialing,active,canceled,incomplete,incomplete_expired,past_due,unpaid,paused"`
}
type StripeSubscriptionListParams struct {
	PaginatedInput
	StripeSubscriptionListFilter
	SortParams
	StripeSubscriptionExpand
	// Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"user,price,product"`
}

type StripeSubscriptionExpand struct {
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"user,price,product"`
}

type StripeSubscriptionGetParams struct {
	SubscriptionID string `path:"subscription-id" json:"subscription_id" required:"true"`
	StripeSubscriptionExpand
}
