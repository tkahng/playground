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

// type StripeProduct

type stripeProductTable struct {
	Columns     []string
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
	Columns: []string{
		"id",
		"active",
		"name",
		"description",
		"image",
		"metadata",
		"created_at",
		"updated_at",
	},
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

var StripeProductTableName = "stripe_products"

var StripeProductTablePrefix = stripeProductTable{
	Columns: []string{
		StripeProductTableName + "." + "id",
		StripeProductTableName + "." + "active",
		StripeProductTableName + "." + "name",
		StripeProductTableName + "." + "description",
		StripeProductTableName + "." + "image",
		StripeProductTableName + "." + "metadata",
		StripeProductTableName + "." + "created_at",
		StripeProductTableName + "." + "updated_at",
	},
	ID:          StripeProductTableName + "." + "id",
	Active:      StripeProductTableName + "." + "active",
	Name:        StripeProductTableName + "." + "name",
	Description: StripeProductTableName + "." + "description",
	Image:       StripeProductTableName + "." + "image",
	Metadata:    StripeProductTableName + "." + "metadata",
	CreatedAt:   StripeProductTableName + "." + "created_at",
	UpdatedAt:   StripeProductTableName + "." + "updated_at",
	Prices:      StripeProductTableName + "." + "prices",
	Roles:       StripeProductTableName + "." + "roles",
	Permissions: StripeProductTableName + "." + "permissions",
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
	Columns         []string
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

var StripePriceTableName = "stripe_prices"

var StripePriceTable = stripePriceTable{
	Columns: []string{
		"id",
		"product_id",
		"lookup_key",
		"active",
		"unit_amount",
		"currency",
		"type",
		"interval",
		"interval_count",
		"trial_period_days",
		"metadata",
		"created_at",
		"updated_at",
	},
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
var StripePriceTablePrefix = stripePriceTable{
	Columns: []string{
		StripePriceTableName + "." + "id",
		StripePriceTableName + "." + "product_id",
		StripePriceTableName + "." + "lookup_key",
		StripePriceTableName + "." + "active",
		StripePriceTableName + "." + "unit_amount",
		StripePriceTableName + "." + "currency",
		StripePriceTableName + "." + "type",
		StripePriceTableName + "." + "interval",
		StripePriceTableName + "." + "interval_count",
		StripePriceTableName + "." + "trial_period_days",
		StripePriceTableName + "." + "metadata",
		StripePriceTableName + "." + "created_at",
		StripePriceTableName + "." + "updated_at",
	},
	ID:              StripePriceTableName + "." + "id",
	ProductID:       StripePriceTableName + "." + "product_id",
	LookupKey:       StripePriceTableName + "." + "lookup_key",
	Active:          StripePriceTableName + "." + "active",
	UnitAmount:      StripePriceTableName + "." + "unit_amount",
	Currency:        StripePriceTableName + "." + "currency",
	Type:            StripePriceTableName + "." + "type",
	Interval:        StripePriceTableName + "." + "interval",
	IntervalCount:   StripePriceTableName + "." + "interval_count",
	TrialPeriodDays: StripePriceTableName + "." + "trial_period_days",
	Metadata:        StripePriceTableName + "." + "metadata",
	CreatedAt:       StripePriceTableName + "." + "created_at",
	UpdatedAt:       StripePriceTableName + "." + "updated_at",
	Product:         StripePriceTableName + "." + "product",
	Subscriptions:   StripePriceTableName + "." + "subscriptions",
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

func (s StripeSubscriptionStatus) String() string {
	return string(s)
}

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
	Columns            []string
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

var StripeSubscriptionTableName = "stripe_subscriptions"
var StripeSubscriptionTable = stripeSubscriptionTable{
	Columns: []string{
		"id",
		"stripe_customer_id",
		"status",
		"metadata",
		"item_id",
		"price_id",
		"quantity",
		"cancel_at_period_end",
		"created",
		"current_period_start",
		"current_period_end",
		"ended_at",
		"cancel_at",
		"canceled_at",
		"trial_start",
		"trial_end",
		"created_at",
		"updated_at",
	},
	ID:                 "id", // primary key
	StripeCustomerID:   "stripe_customer_id",
	Status:             "status",
	Metadata:           "metadata",
	ItemID:             "item_id",  // item_id is the ID of the subscription item
	PriceID:            "price_id", // price_id is the ID of the price associated with the subscription
	Quantity:           "quantity",
	CancelAtPeriodEnd:  "cancel_at_period_end",
	Created:            "created",              // created is the timestamp when the subscription was created
	CurrentPeriodStart: "current_period_start", // current_period_start is the start of the current billing period
	CurrentPeriodEnd:   "current_period_end",   // current_period_end is the end of the current billing period
	EndedAt:            "ended_at",             // ended_at is the timestamp when the subscription ended, if applicable
	CancelAt:           "cancel_at",            // cancel_at is the timestamp when the subscription will be canceled, if applicable
	CanceledAt:         "canceled_at",          // canceled_at is the timestamp when the subscription was canceled, if applicable
	TrialStart:         "trial_start",          // trial_start is the timestamp when the trial period started, if applicable
	TrialEnd:           "trial_end",            // trial_end is the timestamp when the trial period ended, if applicable
	CreatedAt:          "created_at",           // created_at is the timestamp when the subscription was created in the database
	UpdatedAt:          "updated_at",           // updated_at is the timestamp when the subscription was last updated in the database
	StripeCustomer:     "stripe_customer",      // stripe_customer is the StripeCustomer associated with the subscription
	Price:              "price",                // price is the StripePrice associated with the subscription
}

var StripeSubscriptionTablePrefix = stripeSubscriptionTable{
	Columns: []string{
		StripeSubscriptionTableName + "." + "id",
		StripeSubscriptionTableName + "." + "stripe_customer_id",
		StripeSubscriptionTableName + "." + "status",
		StripeSubscriptionTableName + "." + "metadata",
		StripeSubscriptionTableName + "." + "item_id",
		StripeSubscriptionTableName + "." + "price_id",
		StripeSubscriptionTableName + "." + "quantity",
		StripeSubscriptionTableName + "." + "cancel_at_period_end",
		StripeSubscriptionTableName + "." + "created",
		StripeSubscriptionTableName + "." + "current_period_start",
		StripeSubscriptionTableName + "." + "current_period_end",
		StripeSubscriptionTableName + "." + "ended_at",
		StripeSubscriptionTableName + "." + "cancel_at",
		StripeSubscriptionTableName + "." + "canceled_at",
		StripeSubscriptionTableName + "." + "trial_start",
		StripeSubscriptionTableName + "." + "trial_end",
		StripeSubscriptionTableName + "." + "created_at",
		StripeSubscriptionTableName + "." + "updated_at",
	},
	ID:                 StripeSubscriptionTableName + "." + "id",
	StripeCustomerID:   StripeSubscriptionTableName + "." + "stripe_customer_id",
	Status:             StripeSubscriptionTableName + "." + "status",
	Metadata:           StripeSubscriptionTableName + "." + "metadata",
	ItemID:             StripeSubscriptionTableName + "." + "item_id",
	PriceID:            StripeSubscriptionTableName + "." + "price_id",
	Quantity:           StripeSubscriptionTableName + "." + "quantity",
	CancelAtPeriodEnd:  StripeSubscriptionTableName + "." + "cancel_at_period_end",
	Created:            StripeSubscriptionTableName + "." + "created",
	CurrentPeriodStart: StripeSubscriptionTableName + "." + "current_period_start",
	CurrentPeriodEnd:   StripeSubscriptionTableName + "." + "current_period_end",
	EndedAt:            StripeSubscriptionTableName + "." + "ended_at",
	CancelAt:           StripeSubscriptionTableName + "." + "cancel_at",
	CanceledAt:         StripeSubscriptionTableName + "." + "canceled_at",
	TrialStart:         StripeSubscriptionTableName + "." + "trial_start",
	TrialEnd:           StripeSubscriptionTableName + "." + "trial_end",
	CreatedAt:          StripeSubscriptionTableName + "." + "created_at",
	UpdatedAt:          StripeSubscriptionTableName + "." + "updated_at",
	StripeCustomer:     StripeSubscriptionTableName + "." + "stripe_customer",
	Price:              StripeSubscriptionTableName + "." + "price",
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
	// scannable
}

type stripeCustomerTable struct {
	Columns        []string
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

var StripeCustomerTableName = "stripe_customers"

var StripeCustomerTable = stripeCustomerTable{
	Columns: []string{
		"id",
		"email",
		"name",
		"user_id",
		"team_id",
		"customer_type",
		"billing_address",
		"payment_method",
		"created_at",
		"updated_at",
	},
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

var StripeCustomerTablePrefix = stripeCustomerTable{
	Columns: []string{
		StripeCustomerTableName + "." + "id",
		StripeCustomerTableName + "." + "email",
		StripeCustomerTableName + "." + "name",
		StripeCustomerTableName + "." + "user_id",
		StripeCustomerTableName + "." + "team_id",
		StripeCustomerTableName + "." + "customer_type",
		StripeCustomerTableName + "." + "billing_address",
		StripeCustomerTableName + "." + "payment_method",
		StripeCustomerTableName + "." + "created_at",
		StripeCustomerTableName + "." + "updated_at",
	},
	ID:             StripeCustomerTableName + "." + "id",
	Email:          StripeCustomerTableName + "." + "email",
	Name:           StripeCustomerTableName + "." + "name",
	UserID:         StripeCustomerTableName + "." + "user_id",
	TeamID:         StripeCustomerTableName + "." + "team_id",
	CustomerType:   StripeCustomerTableName + "." + "customer_type",
	BillingAddress: StripeCustomerTableName + "." + "billing_address",
	PaymentMethod:  StripeCustomerTableName + "." + "payment_method",
	CreatedAt:      StripeCustomerTableName + "." + "created_at",
	UpdatedAt:      StripeCustomerTableName + "." + "updated_at",
	Team:           StripeCustomerTableName + "." + "team",
	User:           StripeCustomerTableName + "." + "user",
	Subscriptions:  StripeCustomerTableName + "." + "subscriptions",
}

type SubscriptionWithPrice struct {
	Price        StripePrice        `json:"price"`
	Subscription StripeSubscription `json:"subscription"`
	Product      StripeProduct      `json:"product"`
}
