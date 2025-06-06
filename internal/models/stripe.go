package models

import (
	"time"

	"github.com/google/uuid"
)

type StripeProduct struct {
	_           struct{}          `db:"stripe_products" json:"-"`
	ID          string            `db:"id" json:"id"`
	Active      bool              `db:"active" json:"active"`
	Name        string            `db:"name" json:"name"`
	Description *string           `db:"description" json:"description"`
	Image       *string           `db:"image" json:"image"`
	Metadata    map[string]string `db:"metadata" json:"metadata"`
	CreatedAt   time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `db:"updated_at" json:"updated_at"`
	Prices      []*StripePrice    `db:"prices" src:"id" dest:"product_id" table:"stripe_prices" json:"prices,omitempty"`
	Roles       []*Role           `db:"roles" src:"id" dest:"product_id" table:"roles" through:"product_roles,role_id,id" json:"roles,omitempty"`
	Permissions []*Permission     `db:"permissions" src:"id" dest:"product_id" table:"permissions" through:"product_permissions,permission_id,id" json:"permissions,omitempty"`
}

type stripeProductTable struct {
	ID          string
	Active      string
	Name        string
	Description string
	Image       string
	Metadata    string
	CreatedAt   string
	UpdatedAt   string
	Prices      string
	Roles       string
	Permissions string
}

var StripeProductTable = stripeProductTable{
	ID:          "id",
	Active:      "active",
	Name:        "name",
	Description: "description",
	Image:       "image",
	Metadata:    "metadata",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
	Prices:      "prices",
	Roles:       "roles",
	Permissions: "permissions",
}

type StripePricingType string

const (
	StripePricingTypeOneTime   StripePricingType = "one_time"
	StripePricingTypeRecurring StripePricingType = "recurring"
)

// ToModelsStripePricingType converts a StripePricingType to models.StripePricingType

type StripePricingPlanInterval string

const (
	StripePricingPlanIntervalDay   StripePricingPlanInterval = "day"
	StripePricingPlanIntervalWeek  StripePricingPlanInterval = "week"
	StripePricingPlanIntervalMonth StripePricingPlanInterval = "month"
	StripePricingPlanIntervalYear  StripePricingPlanInterval = "year"
)

// ToModelsStripePricingPlanInterval converts a StripePricingPlanInterval to models.StripePricingPlanInterval

type StripePrice struct {
	_               struct{}                   `db:"stripe_prices" json:"-"`
	ID              string                     `db:"id" json:"id"`
	ProductID       string                     `db:"product_id" json:"product_id"`
	LookupKey       *string                    `db:"lookup_key" json:"lookup_key"`
	Active          bool                       `db:"active" json:"active"`
	UnitAmount      *int64                     `db:"unit_amount" json:"unit_amount"`
	Currency        string                     `db:"currency" json:"currency"`
	Type            StripePricingType          `db:"type" json:"type" required:"true" enum:"one_time,recurring"`
	Interval        *StripePricingPlanInterval `db:"interval" json:"interval,omitempty" enum:"day,week,month,year"`
	IntervalCount   *int64                     `db:"interval_count" json:"interval_count"`
	TrialPeriodDays *int64                     `db:"trial_period_days" json:"trial_period_days"`
	Metadata        map[string]string          `db:"metadata" json:"metadata"`
	CreatedAt       time.Time                  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time                  `db:"updated_at" json:"updated_at"`
	Product         *StripeProduct             `db:"product" src:"product_id" dest:"id" table:"stripe_products" json:"product,omitempty"`
	Subscriptions   []*StripeSubscription      `db:"subscriptions" src:"id" dest:"price_id" table:"stripe_subscriptions" json:"subscriptions,omitempty"`
}

type stripePriceTable struct {
	ID              string
	ProductID       string
	LookupKey       string
	Active          string
	UnitAmount      string
	Currency        string
	Type            string
	Interval        string
	IntervalCount   string
	TrialPeriodDays string
	Metadata        string
	CreatedAt       string
	UpdatedAt       string
	Product         string
	Subscriptions   string
}

var StripePriceTable = stripePriceTable{
	ID:              "id",
	ProductID:       "product_id",
	LookupKey:       "lookup_key",
	Active:          "active",
	UnitAmount:      "unit_amount",
	Currency:        "currency",
	Type:            "type",
	Interval:        "interval",
	IntervalCount:   "interval_count",
	TrialPeriodDays: "trial_period_days",
	Metadata:        "metadata",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
	Product:         "product",
	Subscriptions:   "subscriptions",
}

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

type stripeSubscriptionTable struct {
	ID                 string
	StripeCustomerID   string
	Status             string
	Metadata           string
	ItemID             string
	PriceID            string
	Quantity           string
	CancelAtPeriodEnd  string
	Created            string
	CurrentPeriodStart string
	CurrentPeriodEnd   string
	EndedAt            string
	CancelAt           string
	CanceledAt         string
	TrialStart         string
	TrialEnd           string
	CreatedAt          string
	UpdatedAt          string
	StripeCustomer     string
	Price              string
}

var StripeSubscriptionTable = stripeSubscriptionTable{
	ID:                 "id",
	StripeCustomerID:   "stripe_customer_id",
	Status:             "status",
	Metadata:           "metadata",
	ItemID:             "item_id",
	PriceID:            "price_id",
	Quantity:           "quantity",
	CancelAtPeriodEnd:  "cancel_at_period_end",
	Created:            "created",
	CurrentPeriodStart: "current_period_start",
	CurrentPeriodEnd:   "current_period_end",
	EndedAt:            "ended_at",
	CancelAt:           "cancel_at",
	CanceledAt:         "canceled_at",
	TrialStart:         "trial_start",
	TrialEnd:           "trial_end",
	CreatedAt:          "created_at",
	UpdatedAt:          "updated_at",
	StripeCustomer:     "stripe_customer",
	Price:              "price",
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
	User           *User                 `db:"user" src:"user_id" dest:"id" table:"users" json:"user,omitempty"`
	Subscriptions  []*StripeSubscription `db:"subscriptions" src:"id" dest:"stripe_customer_id" table:"stripe_subscriptions" json:"subscriptions,omitempty"`
}

type stripeCustomerTable struct {
	ID             string
	Email          string
	Name           string
	UserID         string
	TeamID         string
	CustomerType   string
	BillingAddress string
	PaymentMethod  string
	CreatedAt      string
	UpdatedAt      string
	Team           string
	User           string
	Subscriptions  string
}

var StripeCustomerTable = stripeCustomerTable{
	ID:             "id",
	Email:          "email",
	Name:           "name",
	UserID:         "user_id",
	TeamID:         "team_id",
	CustomerType:   "customer_type",
	BillingAddress: "billing_address",
	PaymentMethod:  "payment_method",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
	Team:           "team",
	User:           "user",
	Subscriptions:  "subscriptions",
}

type SubscriptionWithPrice struct {
	Price        StripePrice        `json:"price"`
	Subscription StripeSubscription `json:"subscription"`
	Product      StripeProduct      `json:"product"`
}
