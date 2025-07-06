package apis

import (
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/tools/mapper"
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

type StripeSubscription struct {
	_                  struct{}                 `db:"stripe_subscriptions" json:"-"`
	ID                 string                   `db:"id" json:"id"`
	StripeCustomerID   string                   `db:"stripe_customer_id" json:"stripe_customer_id"`
	Status             StripeSubscriptionStatus `db:"status" json:"status"`
	Metadata           map[string]string        `db:"metadata" json:"metadata"`
	ItemID             string                   `db:"item_id" json:"item_id"`
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
	StripeCustomer     *StripeCustomer          `db:"stripe_customer" src:"stripe_customer_id" dest:"id" table:"stripe_customers" json:"stripe_customer,omitempty"`
	Price              *StripePrice             `db:"price" src:"price_id" dest:"id" table:"stripe_prices" json:"price,omitempty"`
}

type StripeCustomerType string

const (
	StripeCustomerTypeUser StripeCustomerType = "user"
	StripeCustomerTypeTeam StripeCustomerType = "team"
)

type StripeCustomer struct {
	_              struct{}              `db:"stripe_customers" json:"-"`
	ID             string                `db:"id" json:"id"`
	Email          string                `db:"email" json:"email"`
	Name           *string               `db:"name" json:"name,omitempty" required:"false"`
	UserID         *uuid.UUID            `db:"user_id" json:"user_id,omitempty" required:"false"`
	TeamID         *uuid.UUID            `db:"team_id" json:"team_id,omitempty" required:"false"`
	CustomerType   StripeCustomerType    `db:"customer_type" json:"customer_type" enum:"user,team"`
	BillingAddress *map[string]string    `db:"billing_address" json:"billing_address"`
	PaymentMethod  *map[string]string    `db:"payment_method" json:"payment_method"`
	CreatedAt      time.Time             `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time             `db:"updated_at" json:"updated_at"`
	Team           *Team                 `db:"team" src:"team_id" dest:"id" table:"teams" json:"team,omitempty"`
	User           *ApiUser              `db:"user" src:"user_id" dest:"id" table:"users" json:"user,omitempty"`
	Subscriptions  []*StripeSubscription `db:"subscriptions" src:"id" dest:"stripe_customer_id" table:"stripe_subscriptions" json:"subscriptions,omitempty"`
}

func FromModelCustomer(sub *models.StripeCustomer) *StripeCustomer {
	if sub == nil {
		return nil
	}
	return &StripeCustomer{
		ID:             sub.ID,
		Email:          sub.Email,
		Name:           sub.Name,
		UserID:         sub.UserID,
		TeamID:         sub.TeamID,
		CustomerType:   StripeCustomerType(sub.CustomerType),
		BillingAddress: sub.BillingAddress,
		PaymentMethod:  sub.PaymentMethod,
		CreatedAt:      sub.CreatedAt,
		UpdatedAt:      sub.UpdatedAt,
		Team:           FromTeamModel(sub.Team),
		User:           FromUserModel(sub.User),
		Subscriptions:  mapper.Map(sub.Subscriptions, FromModelSubscription),
	}
}

func FromModelSubscription(sub *models.StripeSubscription) *StripeSubscription {
	if sub == nil {
		return nil
	}
	return &StripeSubscription{
		ID:                 sub.ID,
		StripeCustomerID:   sub.StripeCustomerID,
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
		ItemID:             sub.ItemID,
		StripeCustomer:     FromModelCustomer(sub.StripeCustomer),
		Price:              FromModelPrice(sub.Price),
	}
}

type SubscriptionWithPrice struct {
	Price        StripePrice        `json:"price"`
	Subscription StripeSubscription `json:"subscription"`
	Product      StripeProduct      `json:"product"`
}

type StripeSubscriptionExpand struct {
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"user,price,product"`
}

type StripeSubscriptionGetParams struct {
	SubscriptionID string `path:"subscription-id" json:"subscription_id" required:"true"`
	StripeSubscriptionExpand
}
