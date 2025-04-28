package shared

import (
	"time"

	"github.com/aarondl/opt/null"
	"github.com/google/uuid"
	"github.com/stephenafamo/bob/types"
	"github.com/tkahng/authgo/internal/db/models"
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

func ToStripeSubscriptionStatus(status models.StripeSubscriptionStatus) StripeSubscriptionStatus {
	switch status {
	case models.StripeSubscriptionStatusTrialing:
		return StripeSubscriptionStatusTrialing
	case models.StripeSubscriptionStatusActive:
		return StripeSubscriptionStatusActive
	case models.StripeSubscriptionStatusCanceled:
		return StripeSubscriptionStatusCanceled
	case models.StripeSubscriptionStatusIncomplete:
		return StripeSubscriptionStatusIncomplete
	case models.StripeSubscriptionStatusIncompleteExpired:
		return StripeSubscriptionStatusIncompleteExpired
	case models.StripeSubscriptionStatusPastDue:
		return StripeSubscriptionStatusPastDue
	case models.StripeSubscriptionStatusUnpaid:
		return StripeSubscriptionStatusUnpaid
	case models.StripeSubscriptionStatusPaused:
		return StripeSubscriptionStatusPaused
	default:
		return StripeSubscriptionStatusTrialing
	}
}

func ToModelsStripeSubscriptionStatus(status StripeSubscriptionStatus) models.StripeSubscriptionStatus {
	switch status {
	case StripeSubscriptionStatusTrialing:
		return models.StripeSubscriptionStatusTrialing
	case StripeSubscriptionStatusActive:
		return models.StripeSubscriptionStatusActive
	case StripeSubscriptionStatusCanceled:
		return models.StripeSubscriptionStatusCanceled
	case StripeSubscriptionStatusIncomplete:
		return models.StripeSubscriptionStatusIncomplete
	case StripeSubscriptionStatusIncompleteExpired:
		return models.StripeSubscriptionStatusIncompleteExpired
	case StripeSubscriptionStatusPastDue:
		return models.StripeSubscriptionStatusPastDue
	case StripeSubscriptionStatusUnpaid:
		return models.StripeSubscriptionStatusUnpaid
	case StripeSubscriptionStatusPaused:
		return models.StripeSubscriptionStatusPaused
	default:
		return models.StripeSubscriptionStatusTrialing
	}
}

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

func ModelToSubscription(md *models.StripeSubscription) *Subscription {
	return &Subscription{
		ID:                 md.ID,
		UserID:             md.UserID,
		Status:             ToStripeSubscriptionStatus(md.Status),
		Metadata:           md.Metadata.Val,
		PriceID:            md.PriceID,
		Quantity:           md.Quantity,
		CancelAtPeriodEnd:  md.CancelAtPeriodEnd,
		Created:            md.Created,
		CurrentPeriodStart: md.CurrentPeriodStart,
		CurrentPeriodEnd:   md.CurrentPeriodEnd,
		EndedAt:            md.EndedAt.Ptr(),
		CancelAt:           md.CancelAt.Ptr(),
		CanceledAt:         md.CanceledAt.Ptr(),
		TrialStart:         md.TrialStart.Ptr(),
		TrialEnd:           md.TrialEnd.Ptr(),
		CreatedAt:          md.CreatedAt,
		UpdatedAt:          md.UpdatedAt,
	}
}
func ToModelsSubscription(sub *Subscription) *models.StripeSubscription {
	if sub == nil {
		return nil
	}
	return &models.StripeSubscription{
		ID:                 sub.ID,
		UserID:             sub.UserID,
		Status:             ToModelsStripeSubscriptionStatus(sub.Status),
		Metadata:           types.JSON[map[string]string]{Val: sub.Metadata},
		PriceID:            sub.PriceID,
		Quantity:           sub.Quantity,
		CancelAtPeriodEnd:  sub.CancelAtPeriodEnd,
		Created:            sub.Created,
		CurrentPeriodStart: sub.CurrentPeriodStart,
		CurrentPeriodEnd:   sub.CurrentPeriodEnd,
		EndedAt:            null.FromPtr(sub.EndedAt),
		CancelAt:           null.FromPtr(sub.CancelAt),
		CanceledAt:         null.FromPtr(sub.CanceledAt),
		TrialStart:         null.FromPtr(sub.TrialStart),
		TrialEnd:           null.FromPtr(sub.TrialEnd),
		CreatedAt:          sub.CreatedAt,
		UpdatedAt:          sub.UpdatedAt,
	}
}

type SubscriptionWithData struct {
	*Subscription
	Price            *StripePricesWithProduct `json:"price,omitempty" required:"false"`
	SubscriptionUser *User                    `json:"user,omitempty" required:"false"`
}
