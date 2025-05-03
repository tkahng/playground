package queries

import (
	"context"
	"slices"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/scan"
	"github.com/stephenafamo/scan/pgxscan"
	"github.com/tkahng/authgo/internal/crud/crudModels"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

var (
	StripeProductColumnNames = []string{
		"id",
		"active",
		"name",
		"description",
		"image",
		"metadata",
		"created_at",
		"updated_at",
	}
	StripePriceColumnNames = []string{
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
	}
	StripeCustomerColumnNames = []string{
		"id",
		"stripe_id",
		"billing_address",
		"payment_method",
		"created_at",
		"updated_at",
	}
	StripeSubscriptionColumnNames = models.StripeSubscriptions.Columns().Names()
	MetadataIndexName             = "metadata.index"
)

func ListProducts(ctx context.Context, db Queryer, input *shared.StripeProductListParams) ([]*crudModels.StripeProduct, error) {

	q := squirrel.Select("stripe_products.*").
		From("stripe_products")
	filter := input.StripeProductListFilter
	pageInput := &input.PaginatedInput

	q = Paginate(q, pageInput)
	q = ListProductFilterFuncQuery(q, &filter)
	data, err := ExecQuery[*crudModels.StripeProduct](ctx, db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return data, nil
}

const (
	GetProductRolesQuery = `
	SELECT rp.product_id as key,
        COALESCE(
                json_agg(
                        jsonb_build_object(
                                'id',
                                p.id,
                                'name',
                                p.name,
                                'description',
                                p.description,
                                'created_at',
                                p.created_at,
                                'updated_at',
                                p.updated_at
                        )
                ) FILTER (
                        WHERE p.id IS NOT NULL
                ),
                '[]'
        ) AS data
FROM public.product_roles rp
        LEFT JOIN public.roles p ON p.id = rp.role_id
        WHERE rp.product_id	 = ANY (
                $1::uuid []
        )
GROUP BY rp.product_id;`
)

func LoadProductRoles(ctx context.Context, db Queryer, productIds ...string) ([]shared.JoinedResult[*crudModels.Role, string], error) {
	data, err := pgxscan.All(
		ctx,
		db,
		scan.StructMapper[shared.JoinedResult[*crudModels.Role, string]](),
		GetProductRolesQuery,
		productIds,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

const (
	GetProductPricesQuery = `
	SELECT sp.product_id as key,
        COALESCE(
                json_agg(
                        jsonb_build_object(
                                'id',
                                sp.id,
                                'product_id',
                                sp.product_id,
                                'lookup_key',
                                sp.lookup_key,
                                'active',
                                sp.active,
                                'unit_amount',
                                sp.unit_amount,
                                'currency',
                                sp.currency,
                                'type',
                                sp.type,
                                'interval',
                                sp.interval,
                                'interval_count',
                                sp.interval_count,
                                'trial_period_days',
                                sp.trial_period_days,
                                'created_at',
                                sp.created_at,
                                'updated_at',
                                sp.updated_at
                        )
                ),
                '[]'
        ) AS data
FROM public.stripe_prices sp
WHERE sp.product_id = ANY ($1::text [])
GROUP BY sp.product_id;`
)

func LoadProductPrices(ctx context.Context, db Queryer, productIds ...string) ([]*shared.JoinedResult[*crudModels.StripePrice, string], error) {
	data, err := pgxscan.All(
		ctx,
		db,
		scan.StructMapper[shared.JoinedResult[*crudModels.StripePrice, string]](),
		GetProductPricesQuery,
		productIds,
	)
	if err != nil {
		return nil, err
	}
	// return data, nil
	return mapper.MapTo(data, productIds, func(a shared.JoinedResult[*crudModels.StripePrice, string]) string {
		return a.Key
	}), nil
}

func ListProductOrderByFunc(ctx context.Context, q *psql.ViewQuery[*models.StripeProduct, models.StripeProductSlice], input *shared.StripeProductListParams) {
	if q == nil {
		return
	}
	if input == nil {
		q.Apply(
			sm.OrderBy("metadata->'index'").Asc(),
		)
		return
	}
	if input.SortParams.SortBy == "" {
		q.Apply(
			sm.OrderBy("metadata->'index'").Asc(),
		)
		return
	}
	if input.SortBy == MetadataIndexName {
		q.Apply(
			sm.OrderBy("metadata->'index'").Asc(),
		)
		return
	} else if slices.Contains(StripeProductColumnNames, input.SortBy) {
		if input.SortParams.SortOrder == "desc" {
			q.Apply(
				sm.OrderBy(input.SortBy).Desc(),
				sm.OrderBy(models.StripeProductColumns.ID).Desc(),
			)
			return
		} else if input.SortParams.SortOrder == "asc" {
			q.Apply(
				sm.OrderBy(input.SortBy).Asc(),
				sm.OrderBy(models.StripeProductColumns.ID).Asc(),
			)
			return
		}
	}
}
func ListProductOrderByFunc2(q squirrel.SelectBuilder, input *shared.StripeProductListParams) squirrel.SelectBuilder {
	if input == nil {
		return q
	}
	if input.SortParams.SortBy == "" {
		q = q.OrderBy("metadata->'index'" + " " + strings.ToUpper(input.SortOrder))
	}
	if input.SortBy == MetadataIndexName {
		q = q.OrderBy("metadata->'index'" + " " + strings.ToUpper(input.SortOrder))
	} else if slices.Contains(StripeProductColumnNames, input.SortBy) {
		q = q.OrderBy(input.SortBy + " " + strings.ToUpper(input.SortOrder))
	}
	return q
}

func ListProductFilterFuncQuery(q squirrel.SelectBuilder, filter *shared.StripeProductListFilter) squirrel.SelectBuilder {
	if filter == nil {
		return q
	}
	if filter.Active != "" {
		if filter.Active == shared.Active {
			q = q.Where("active = ?", true)
		}
		if filter.Active == shared.Inactive {
			q = q.Where("active = ?", false)
		}
	}
	if len(filter.Ids) > 0 {
		q = q.Where("id in (?)", filter.Ids)
	}
	return q
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
func CountProducts(ctx context.Context, db Queryer, filter *shared.StripeProductListFilter) (int64, error) {
	q := squirrel.Select("COUNT(stripe_products.*)").
		From("stripe_products")

	q = ListProductFilterFuncQuery(q, filter)
	data, err := ExecQuery[CountOutput](ctx, db, q.PlaceholderFormat(squirrel.Dollar))

	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil
	}

	return data[0].Count, nil
}

func ListPrices(ctx context.Context, db Queryer, input *shared.StripePriceListParams) (models.StripePriceSlice, error) {

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
				models.SelectWhere.StripePrices.Active.EQ(true),
			)
		}
		if filter.Active == shared.Inactive {
			q.Apply(
				models.SelectWhere.StripePrices.Active.EQ(false),
			)
		}
	}
	if len(filter.Ids) > 0 {
		q.Apply(
			models.SelectWhere.StripePrices.ID.In(filter.Ids...),
		)
	}
	if len(filter.ProductIds) > 0 {
		q.Apply(
			models.SelectWhere.StripePrices.ProductID.In(filter.ProductIds...),
		)
	}
}

func CountPrices(ctx context.Context, db Queryer, filter *shared.StripePriceListFilter) (int64, error) {
	q := models.StripePrices.Query()
	ListPriceFilterFunc(ctx, q, filter)
	data, err := q.Count(ctx, db)
	if err != nil {
		return 0, err
	}
	return data, nil
}

func ListCustomers(ctx context.Context, db Queryer, input *shared.StripeCustomerListParams) (models.StripeCustomerSlice, error) {

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

func CountCustomers(ctx context.Context, db Queryer, filter *shared.StripeCustomerListFilter) (int64, error) {
	q := models.StripeCustomers.Query()
	ListCustomerFilterFunc(ctx, q, filter)
	data, err := q.Count(ctx, db)
	if err != nil {
		return 0, err
	}
	return data, nil
}

func ListSubscriptions(ctx context.Context, db Queryer, input *shared.StripeSubscriptionListParams) (models.StripeSubscriptionSlice, error) {

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
		q.Apply(
			models.SelectWhere.StripeSubscriptions.ID.In(filter.Ids...),
		)
	}
	if len(filter.Status) > 0 {
		statuses := mapper.Map(filter.Status, shared.ToModelsStripeSubscriptionStatus)
		if len(statuses) > 0 {
			q.Apply(
				models.SelectWhere.StripeSubscriptions.Status.In(statuses...),
			)
		}
	}
	if filter.UserID != "" {
		userID, err := uuid.Parse(filter.UserID)
		if err != nil {
			return
		}
		q.Apply(
			models.SelectWhere.StripeSubscriptions.UserID.EQ(userID),
		)
	}
}

func CountSubscriptions(ctx context.Context, db Queryer, filter *shared.StripeSubscriptionListFilter) (int64, error) {
	q := models.StripeSubscriptions.Query()
	ListSubscriptionFilterFunc(ctx, q, filter)
	data, err := q.Count(ctx, db)
	if err != nil {
		return 0, err
	}
	return data, nil
}
