package apis

import (
	"time"

	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/tools/mapper"
	ty "github.com/tkahng/authgo/internal/tools/types"
)

type StripePricingType string

const (
	StripePricingTypeOneTime   StripePricingType = "one_time"
	StripePricingTypeRecurring StripePricingType = "recurring"
)

// ToModelsStripePricingType converts a StripePricingType to models.StripePricingType
// ToStripePricingType converts a models.StripePricingType to StripePricingType

type StripePricingPlanInterval string

const (
	StripePricingPlanIntervalDay   StripePricingPlanInterval = "day"
	StripePricingPlanIntervalWeek  StripePricingPlanInterval = "week"
	StripePricingPlanIntervalMonth StripePricingPlanInterval = "month"
	StripePricingPlanIntervalYear  StripePricingPlanInterval = "year"
)

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
	Subscriptions   []*Subscription            `db:"subscriptions" src:"id" dest:"price_id" table:"stripe_subscriptions" json:"subscriptions,omitempty"`
}

func FromModelPrice(price *models.StripePrice) *StripePrice {
	var interval *StripePricingPlanInterval
	if price.Interval != nil {
		interval = ty.Pointer(StripePricingPlanInterval(*price.Interval))
	}
	return &StripePrice{
		ID:              price.ID,
		ProductID:       price.ProductID,
		LookupKey:       price.LookupKey,
		Active:          price.Active,
		UnitAmount:      price.UnitAmount,
		Currency:        price.Currency,
		Type:            StripePricingType(price.Type),
		Interval:        interval,
		IntervalCount:   price.IntervalCount,
		TrialPeriodDays: price.TrialPeriodDays,
		Metadata:        price.Metadata,
		CreatedAt:       price.CreatedAt,
		UpdatedAt:       price.UpdatedAt,
		Product:         FromModelProduct(price.Product),
		Subscriptions:   mapper.Map(price.Subscriptions, FromModelSubscription),
	}
}
