package repository

import (
	"context"
	"slices"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
)

var (
	StripeProductColumnNames      = models.StripeProducts.Columns().Names()
	StripePriceColumnNames        = models.StripePrices.Columns().Names()
	StripeCustomerColumnNames     = models.StripeCustomers.Columns().Names()
	StripeSubscriptionColumnNames = models.StripeSubscriptions.Columns().Names()
	MetadataIndexName             = "metadata.index"
)

func ListProducts(ctx context.Context, db bob.DB, input *shared.StripeProductListParams) (models.StripeProductSlice, error) {

	q := models.StripeProducts.Query()
	filter := input.StripeProductListFilter
	pageInput := &input.PaginatedInput

	ViewApplyPagination(q, pageInput)

	ListProductFilterFunc(ctx, q, &filter)
	ListProductOrderByFunc(ctx, q, input)
	data, err := q.All(ctx, db)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ListProductOrderByFunc(ctx context.Context, q *psql.ViewQuery[*models.StripeProduct, models.StripeProductSlice], input *shared.StripeProductListParams) {
	if input == nil {
		return
	}
	if input.SortParams.SortBy == "" {
		return
	}
	if input.SortBy == MetadataIndexName {
		q.Apply(
			sm.OrderBy("metadata->'index'").Asc(),
		)
	} else if slices.Contains(StripeProductColumnNames, input.SortBy) {
		var order = sm.OrderBy(input.SortBy)
		if input.SortParams.SortOrder == "desc" {
			order = sm.OrderBy(input.SortBy).Desc()
		} else if input.SortParams.SortOrder == "asc" {
			order = sm.OrderBy(input.SortBy).Asc()
		}
		q.Apply(
			order,
		)
	}
}

func ListProductFilterFunc(ctx context.Context, q *psql.ViewQuery[*models.StripeProduct, models.StripeProductSlice], filter *shared.StripeProductListFilter) {
	if filter == nil {
		return
	}
	if filter.Active != "" {
		if filter.Active == shared.Active {
			q.Apply(
				models.SelectWhere.StripeProducts.Active.EQ(true),
			)
		}
		if filter.Active == shared.Inactive {
			q.Apply(
				models.SelectWhere.StripeProducts.Active.EQ(false),
			)
		}
	}
	if len(filter.Ids) > 0 {
		q.Apply(
			models.SelectWhere.StripeProducts.ID.In(filter.Ids...),
		)
	}
}

// CountUsers implements AdminCrudActions.
func CountProducts(ctx context.Context, db bob.DB, filter *shared.StripeProductListFilter) (int64, error) {
	q := models.StripeProducts.Query()
	ListProductFilterFunc(ctx, q, filter)
	data, err := q.Count(ctx, db)
	if err != nil {
		return 0, err
	}
	return data, nil
}

func ListPrices(ctx context.Context, db bob.DB, input *shared.StripePriceListParams) (models.StripePriceSlice, error) {

	q := models.StripePrices.Query()
	filter := input.StripePriceListFilter
	pageInput := &input.PaginatedInput

	ViewApplyPagination(q, pageInput)

	ListPriceFilterFunc(ctx, q, &filter)
	ListPriceOrderByFunc(ctx, q, input)
	data, err := q.All(ctx, db)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ListPriceOrderByFunc(ctx context.Context, q *psql.ViewQuery[*models.StripePrice, models.StripePriceSlice], input *shared.StripePriceListParams) {
	if input == nil {
		return
	}
	if input.SortParams.SortBy == "" {
		return
	}
	if slices.Contains(StripeCustomerColumnNames, input.SortBy) {
		var order = sm.OrderBy(input.SortBy)
		if input.SortParams.SortOrder == "desc" {
			order = sm.OrderBy(input.SortBy).Desc()
		} else if input.SortParams.SortOrder == "asc" {
			order = sm.OrderBy(input.SortBy).Asc()
		}
		q.Apply(
			order,
		)
	}
}

func ListPriceFilterFunc(ctx context.Context, q *psql.ViewQuery[*models.StripePrice, models.StripePriceSlice], filter *shared.StripePriceListFilter) {
	if filter == nil {
		return
	}
	if filter.Active != "" {
		if filter.Active == shared.Active {
			q.Apply(
				models.SelectWhere.StripeProducts.Active.EQ(true),
			)
		}
		if filter.Active == shared.Inactive {
			q.Apply(
				models.SelectWhere.StripeProducts.Active.EQ(false),
			)
		}
	}
	if len(filter.Ids) > 0 {
		q.Apply(
			models.SelectWhere.StripeProducts.ID.In(filter.Ids...),
		)
	}
}

func CountPrices(ctx context.Context, db bob.DB, filter *shared.StripePriceListFilter) (int64, error) {
	q := models.StripePrices.Query()
	ListPriceFilterFunc(ctx, q, filter)
	data, err := q.Count(ctx, db)
	if err != nil {
		return 0, err
	}
	return data, nil
}

func ListCustomers(ctx context.Context, db bob.DB, input *shared.StripeCustomerListParams) (models.StripeCustomerSlice, error) {

	q := models.StripeCustomers.Query()
	filter := input.StripeCustomerListFilter
	pageInput := &input.PaginatedInput

	ViewApplyPagination(q, pageInput)
	ListCustomerFilterFunc(ctx, q, &filter)
	StripeCustomerOrderByFunc(ctx, q, input)
	data, err := q.All(ctx, db)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func StripeCustomerOrderByFunc(ctx context.Context, q *psql.ViewQuery[*models.StripeCustomer, models.StripeCustomerSlice], input *shared.StripeCustomerListParams) {
	if input == nil {
		return
	}
	if input.SortParams.SortBy == "" {
		return
	}
	if slices.Contains(StripeCustomerColumnNames, input.SortBy) {
		var order = sm.OrderBy(input.SortBy)
		if input.SortParams.SortOrder == "desc" {
			order = sm.OrderBy(input.SortBy).Desc()
		} else if input.SortParams.SortOrder == "asc" {
			order = sm.OrderBy(input.SortBy).Asc()
		}
		q.Apply(
			order,
		)
	}
}

func ListCustomerFilterFunc(ctx context.Context, q *psql.ViewQuery[*models.StripeCustomer, models.StripeCustomerSlice], filter *shared.StripeCustomerListFilter) {
	if filter == nil {
		return
	}
	if len(filter.Ids) > 0 {
		ids := ParseUUIDs(filter.Ids)
		q.Apply(
			models.SelectWhere.StripeCustomers.ID.In(ids...),
		)
	}

}

func CountCustomers(ctx context.Context, db bob.DB, filter *shared.StripePriceListFilter) (int64, error) {
	q := models.StripePrices.Query()
	ListPriceFilterFunc(ctx, q, filter)
	data, err := q.Count(ctx, db)
	if err != nil {
		return 0, err
	}
	return data, nil
}

func ListSubscriptions(ctx context.Context, db bob.DB, input *shared.StripeSubscriptionListParams) (models.StripeSubscriptionSlice, error) {

	q := models.StripeSubscriptions.Query()
	filter := input.StripeSubscriptionListFilter
	pageInput := &input.PaginatedInput

	ViewApplyPagination(q, pageInput)
	ListSubscriptionFilterFunc(ctx, q, &filter)
	StripeSubscriptionOrderByFunc(ctx, q, input)
	data, err := q.All(ctx, db)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func StripeSubscriptionOrderByFunc(ctx context.Context, q *psql.ViewQuery[*models.StripeSubscription, models.StripeSubscriptionSlice], input *shared.StripeSubscriptionListParams) {
	if input == nil {
		return
	}
	if input.SortParams.SortBy == "" {
		return
	}
	if slices.Contains(StripeSubscriptionColumnNames, input.SortBy) {
		var order = sm.OrderBy(input.SortBy)
		if input.SortParams.SortOrder == "desc" {
			order = sm.OrderBy(input.SortBy).Desc()
		} else if input.SortParams.SortOrder == "asc" {
			order = sm.OrderBy(input.SortBy).Asc()
		}
		q.Apply(
			order,
		)
	}
}

func ListSubscriptionFilterFunc(ctx context.Context, q *psql.ViewQuery[*models.StripeSubscription, models.StripeSubscriptionSlice], filter *shared.StripeSubscriptionListFilter) {
	if filter == nil {
		return
	}
	if len(filter.Ids) > 0 {
		ids := ParseUUIDs(filter.Ids)
		q.Apply(
			models.SelectWhere.StripeCustomers.ID.In(ids...),
		)
	}

}

func CountSubscriptions(ctx context.Context, db bob.DB, filter *shared.StripeSubscriptionListFilter) (int64, error) {
	q := models.StripeSubscriptions.Query()
	ListSubscriptionFilterFunc(ctx, q, filter)
	data, err := q.Count(ctx, db)
	if err != nil {
		return 0, err
	}
	return data, nil
}
