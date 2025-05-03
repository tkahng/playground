package shared

import (
	"time"

	crudModels "github.com/tkahng/authgo/internal/db/models"
	ty "github.com/tkahng/authgo/internal/types"
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

// ToModelsStripePricingPlanInterval converts a StripePricingPlanInterval to models.StripePricingPlanInterval

// ToStripePricingPlanInterval converts a models.StripePricingPlanInterval to StripePricingPlanInterval

type Price struct {
	ID              string                     `db:"id,pk" json:"id"`
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
}

func FromCrudPrice(price *crudModels.StripePrice) *Price {
	var interval *StripePricingPlanInterval
	if price.Interval != nil {
		interval = ty.Pointer(StripePricingPlanInterval(*price.Interval))
	}
	return &Price{
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
	}
}

type StripePricesWithProduct struct {
	*Price
	Product *Product `db:"product" json:"product,omitempty" required:"false"`
}

type StripePriceListFilter struct {
	Q          string       `query:"q,omitempty" required:"false"`
	Ids        []string     `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	Active     ActiveStatus `query:"active,omitempty" required:"false" enum:"active,inactive"`
	ProductIds []string     `query:"product_ids,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
}
type StripePriceListParams struct {
	PaginatedInput
	StripePriceListFilter
	SortParams
}
