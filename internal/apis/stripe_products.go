package apis

import (
	"context"
	"net/http"
	"time"

	"github.com/aarondl/opt/null"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stephenafamo/bob/types"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/dataloader"
	"github.com/tkahng/authgo/internal/tools/utils"
)

func (api *Api) StripeProductsWithPricesOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "stripe-products-with-prices",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "stripe-products-with-prices",
		Description: "stripe-products-with-prices",
		Tags:        []string{"Payment", "Stripe", "Products"},
		Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
	}
}

type Product struct {
	ID          string                        `db:"id,pk" json:"id"`
	Active      bool                          `db:"active" json:"active"`
	Name        string                        `db:"name" json:"name"`
	Description null.Val[string]              `db:"description" json:"description"`
	Image       null.Val[string]              `db:"image" json:"image"`
	Metadata    types.JSON[map[string]string] `db:"metadata" json:"metadata"`
	CreatedAt   time.Time                     `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time                     `db:"updated_at" json:"updated_at"`
}
type Price struct {
	ID              string                                     `db:"id,pk" json:"id"`
	ProductID       string                                     `db:"product_id" json:"product_id"`
	LookupKey       null.Val[string]                           `db:"lookup_key" json:"lookup_key"`
	Active          bool                                       `db:"active" json:"active"`
	UnitAmount      null.Val[int64]                            `db:"unit_amount" json:"unit_amount"`
	Currency        string                                     `db:"currency" json:"currency"`
	Type            models.StripePricingType                   `db:"type" json:"type" enum:"one_time,recurring"`
	Interval        null.Val[models.StripePricingPlanInterval] `db:"interval" json:"interval" enum:"day,week,month,year"`
	IntervalCount   null.Val[int64]                            `db:"interval_count" json:"interval_count"`
	TrialPeriodDays null.Val[int64]                            `db:"trial_period_days" json:"trial_period_days"`
	Metadata        types.JSON[map[string]string]              `db:"metadata" json:"metadata"`
	CreatedAt       time.Time                                  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time                                  `db:"updated_at" json:"updated_at"`
}
type StripeProductWithPrice struct {
	Product *Product `db:"product" json:"product"`
	Prices  []*Price `db:"prices" json:"prices"`
}
type StripeProductsOutput struct {
	Body PaginatedOutput[StripeProductWithPrice] `json:"body"`
}

func (api *Api) StripeProductsWithPrices(ctx context.Context, input *shared.StripeProductListParams) (*PaginatedOutput[*StripeProductWithPrice], error) {
	db := api.app.Db()
	utils.PrettyPrintJSON(input)
	users, err := repository.ListProducts(ctx, db, input)
	if err != nil {
		return nil, err
	}
	count, err := repository.CountProducts(ctx, db, &input.StripeProductListFilter)
	if err != nil {
		return nil, err
	}

	ids := dataloader.Map(users, func(user *models.StripeProduct) string {
		return user.ID
	})
	m := make(map[string][]*Price)
	claims, err := repository.PricesByProductIds(ctx, db, ids)
	if err != nil {
		return nil, err
	}
	for _, claim := range claims {
		m[claim.ProductID] = append(m[claim.ProductID], &Price{
			ID:              claim.ID,
			ProductID:       claim.ProductID,
			LookupKey:       claim.LookupKey,
			Active:          claim.Active,
			UnitAmount:      claim.UnitAmount,
			Currency:        claim.Currency,
			Type:            claim.Type,
			Interval:        claim.Interval,
			IntervalCount:   claim.IntervalCount,
			TrialPeriodDays: claim.TrialPeriodDays,
			Metadata:        claim.Metadata,
			CreatedAt:       claim.CreatedAt,
			UpdatedAt:       claim.UpdatedAt,
		})
	}
	info := dataloader.Map(users, func(user *models.StripeProduct) *StripeProductWithPrice {
		claims := m[user.ID]
		return &StripeProductWithPrice{
			Product: &Product{
				ID:          user.ID,
				Active:      user.Active,
				Name:        user.Name,
				Description: user.Description,
				Image:       user.Image,
				Metadata:    user.Metadata,
				CreatedAt:   user.CreatedAt,
				UpdatedAt:   user.UpdatedAt,
			},
			Prices: claims,
		}
	})

	return &PaginatedOutput[*StripeProductWithPrice]{
		Body: shared.PaginatedResponse[*StripeProductWithPrice]{
			Data: info,
			Meta: shared.Meta{
				Page:    input.Page,
				PerPage: input.PerPage,
				Total:   int(count),
			},
		},
	}, nil
}
