package apis

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
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

type Price struct {
	ID              string                            `db:"id,pk" json:"id"`
	ProductID       string                            `db:"product_id" json:"product_id"`
	LookupKey       *string                           `db:"lookup_key" json:"lookup_key"`
	Active          bool                              `db:"active" json:"active"`
	UnitAmount      *int64                            `db:"unit_amount" json:"unit_amount"`
	Currency        string                            `db:"currency" json:"currency"`
	Type            models.StripePricingType          `db:"type" json:"type" required:"true" enum:"one_time,recurring"`
	Interval        *models.StripePricingPlanInterval `db:"interval" json:"interval,omitempty" enum:"day,week,month,year"`
	IntervalCount   *int64                            `db:"interval_count" json:"interval_count"`
	TrialPeriodDays *int64                            `db:"trial_period_days" json:"trial_period_days"`
	Metadata        map[string]string                 `db:"metadata" json:"metadata"`
	CreatedAt       time.Time                         `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time                         `db:"updated_at" json:"updated_at"`
}

func ModelToPrice(price *models.StripePrice) *Price {
	return &Price{
		ID:              price.ID,
		ProductID:       price.ProductID,
		LookupKey:       price.LookupKey.Ptr(),
		Active:          price.Active,
		UnitAmount:      price.UnitAmount.Ptr(),
		Currency:        price.Currency,
		Type:            price.Type,
		Interval:        price.Interval.Ptr(),
		IntervalCount:   price.IntervalCount.Ptr(),
		TrialPeriodDays: price.TrialPeriodDays.Ptr(),
		Metadata:        price.Metadata.Val,
		CreatedAt:       price.CreatedAt,
		UpdatedAt:       price.UpdatedAt,
	}
}

type Product struct {
	ID          string            `db:"id,pk" json:"id"`
	Active      bool              `db:"active" json:"active"`
	Name        string            `db:"name" json:"name"`
	Description *string           `db:"description" json:"description"`
	Image       *string           `db:"image" json:"image"`
	Metadata    map[string]string `db:"metadata" json:"metadata"`
	CreatedAt   time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `db:"updated_at" json:"updated_at"`
}

func ModelToProduct(product *models.StripeProduct) *Product {
	return &Product{
		ID:          product.ID,
		Active:      product.Active,
		Name:        product.Name,
		Description: product.Description.Ptr(),
		Image:       product.Image.Ptr(),
		Metadata:    product.Metadata.Val,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}

type StripePricesWithProduct struct {
	*Price
	Product *Product `db:"product" json:"product,omitempty" required:"false"`
}
type StripeProductWithPrices struct {
	*Product
	Prices []*Price `db:"prices" json:"prices,omitempty" required:"false"`
}

type StripeProductsWithPricesInput struct {
	shared.PaginatedInput
	shared.SortParams
}

func (api *Api) StripeProductsWithPrices(ctx context.Context, inputt *StripeProductsWithPricesInput) (*shared.PaginatedOutput[*StripeProductWithPrices], error) {
	db := api.app.Db()
	input := &shared.StripeProductListParams{
		PaginatedInput: inputt.PaginatedInput,
		StripeProductListFilter: shared.StripeProductListFilter{
			Active: shared.Active,
		},
		SortParams: inputt.SortParams,
	}
	users, err := repository.ListProducts(ctx, db, input)
	if err != nil {
		return nil, err
	}

	err = users.LoadStripeProductProductStripePrices(ctx, db,
		models.SelectWhere.StripePrices.Active.EQ(true),
	)
	if err != nil {
		return nil, err
	}

	count, err := repository.CountProducts(ctx, db, &input.StripeProductListFilter)
	if err != nil {
		return nil, err
	}
	prods := mapper.Map(users, func(user *models.StripeProduct) *StripeProductWithPrices {
		return &StripeProductWithPrices{
			Product: ModelToProduct(user),
			Prices:  mapper.Map(user.R.ProductStripePrices, ModelToPrice),
		}

	})

	return &shared.PaginatedOutput[*StripeProductWithPrices]{
		Body: shared.PaginatedResponse[*StripeProductWithPrices]{
			Data: prods,
			Meta: shared.Meta{
				Page:    input.Page,
				PerPage: input.PerPage,
				Total:   count,
			},
		},
	}, nil
}
