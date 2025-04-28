package shared

import (
	"time"

	"github.com/tkahng/authgo/internal/db/models"
	ty "github.com/tkahng/authgo/internal/types"
)

type StripePricingType string

const (
	StripePricingTypeOneTime   StripePricingType = "one_time"
	StripePricingTypeRecurring StripePricingType = "recurring"
)

// ToModelsStripePricingType converts a StripePricingType to models.StripePricingType
func ToModelsStripePricingType(pt StripePricingType) models.StripePricingType {
	switch pt {
	case StripePricingTypeOneTime:
		return models.StripePricingTypeOneTime
	case StripePricingTypeRecurring:
		return models.StripePricingTypeRecurring
	default:
		return models.StripePricingTypeOneTime
	}
}

// ToStripePricingType converts a models.StripePricingType to StripePricingType
func ToStripePricingType(pt models.StripePricingType) StripePricingType {
	switch pt {
	case models.StripePricingTypeOneTime:
		return StripePricingTypeOneTime
	case models.StripePricingTypeRecurring:
		return StripePricingTypeRecurring
	default:
		return StripePricingTypeOneTime
	}
}

type StripePricingPlanInterval string

const (
	StripePricingPlanIntervalDay   StripePricingPlanInterval = "day"
	StripePricingPlanIntervalWeek  StripePricingPlanInterval = "week"
	StripePricingPlanIntervalMonth StripePricingPlanInterval = "month"
	StripePricingPlanIntervalYear  StripePricingPlanInterval = "year"
)

// ToModelsStripePricingPlanInterval converts a StripePricingPlanInterval to models.StripePricingPlanInterval
func ToModelsStripePricingPlanInterval(pt StripePricingPlanInterval) models.StripePricingPlanInterval {
	switch pt {
	case StripePricingPlanIntervalDay:
		return models.StripePricingPlanIntervalDay
	case StripePricingPlanIntervalWeek:
		return models.StripePricingPlanIntervalWeek
	case StripePricingPlanIntervalMonth:
		return models.StripePricingPlanIntervalMonth
	case StripePricingPlanIntervalYear:
		return models.StripePricingPlanIntervalYear
	default:
		return models.StripePricingPlanIntervalMonth
	}
}

// ToStripePricingPlanInterval converts a models.StripePricingPlanInterval to StripePricingPlanInterval
func ToStripePricingPlanInterval(pt *models.StripePricingPlanInterval) *StripePricingPlanInterval {
	if pt == nil {
		return nil
	}
	switch *pt {
	case models.StripePricingPlanIntervalDay:
		return ty.Pointer(StripePricingPlanIntervalDay)
	case models.StripePricingPlanIntervalWeek:
		return ty.Pointer(StripePricingPlanIntervalWeek)
	case models.StripePricingPlanIntervalMonth:
		return ty.Pointer(StripePricingPlanIntervalMonth)
	case models.StripePricingPlanIntervalYear:
		return ty.Pointer(StripePricingPlanIntervalYear)
	default:
		return nil
	}
}

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

func ModelToPrice(price *models.StripePrice) *Price {
	return &Price{
		ID:              price.ID,
		ProductID:       price.ProductID,
		LookupKey:       price.LookupKey.Ptr(),
		Active:          price.Active,
		UnitAmount:      price.UnitAmount.Ptr(),
		Currency:        price.Currency,
		Type:            ToStripePricingType(price.Type),
		Interval:        ToStripePricingPlanInterval(price.Interval.Ptr()),
		IntervalCount:   price.IntervalCount.Ptr(),
		TrialPeriodDays: price.TrialPeriodDays.Ptr(),
		Metadata:        price.Metadata.Val,
		CreatedAt:       price.CreatedAt,
		UpdatedAt:       price.UpdatedAt,
	}
}

type StripePricesWithProduct struct {
	*Price
	Product *Product `db:"product" json:"product,omitempty" required:"false"`
}
