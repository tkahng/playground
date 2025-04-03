package repository

import (
	"context"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
)

func ListProducts(ctx context.Context, db bob.DB, input *shared.StripeProductListParams) (models.StripeProductSlice, error) {

	q := models.StripeProducts.Query()
	filter := input.StripeProductListFilter
	pageInput := &input.PaginatedInput

	ViewApplyPagination(q, pageInput)

	ListProductFilterFunc(ctx, q, &filter)
	data, err := q.All(ctx, db)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ListProductFilterFunc(ctx context.Context, q *psql.ViewQuery[*models.StripeProduct, models.StripeProductSlice], filter *shared.StripeProductListFilter) {
	if filter == nil {
		return
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

func PricesByProductIds(ctx context.Context, dbx bob.Executor, productIds []string) ([]*models.StripePrice, error) {
	data, err := models.StripePrices.Query(
		models.SelectWhere.StripePrices.ProductID.In(productIds...),
	).All(ctx, dbx)
	return data, err
}
