package queries

import (
	"context"
	"slices"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/tkahng/authgo/internal/crud/crudrepo"
	"github.com/tkahng/authgo/internal/crud/models"
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
	StripeSubscriptionColumnNames = []string{
		"id",
		"status",
		"user_id",
		"metadata",
		"stripe_id",
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
	}
	MetadataIndexName = "metadata.index"
)

func ListProducts(ctx context.Context, db Queryer, input *shared.StripeProductListParams) ([]*models.StripeProduct, error) {

	q := squirrel.Select("stripe_products.*").
		From("stripe_products")
	filter := input.StripeProductListFilter
	pageInput := &input.PaginatedInput

	q = Paginate(q, pageInput)
	q = ListProductFilterFuncQuery(q, &filter)
	data, err := QueryWithBuilder[*models.StripeProduct](ctx, db, q.PlaceholderFormat(squirrel.Dollar))
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

func LoadProductRoles(ctx context.Context, db Queryer, productIds ...string) ([][]*models.Role, error) {
	data, err := QueryAll[shared.JoinedResult[*models.Role, string]](
		ctx,
		db,
		GetProductRolesQuery,
		productIds,
	)
	if err != nil {
		return nil, err
	}
	return mapper.Map(mapper.MapTo(data, productIds, func(a shared.JoinedResult[*models.Role, string]) string {
		return a.Key
	}), func(a *shared.JoinedResult[*models.Role, string]) []*models.Role {
		if a == nil {
			return nil
		}
		return a.Data
	}), nil
}

func LoadeProductPrices(ctx context.Context, db Queryer, where *map[string]any, productIds ...string) ([][]*models.StripePrice, error) {
	if where == nil {
		where = &map[string]any{}
	}
	(*where)["product_id"] = map[string]any{
		"_in": productIds,
	}
	prices, err := crudrepo.StripePrice.Get(
		ctx,
		db,
		where,
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return mapper.MapToManyPointer(prices, productIds, func(t *models.StripePrice) string {
		return t.ProductID
	}), nil
}

func ListProductOrderByQuery(q squirrel.SelectBuilder, input *shared.StripeProductListParams) squirrel.SelectBuilder {
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

// CountUsers implements AdminCrudActions.
func CountProducts(ctx context.Context, db Queryer, filter *shared.StripeProductListFilter) (int64, error) {
	q := squirrel.Select("COUNT(stripe_products.*)").
		From("stripe_products")

	q = ListProductFilterFuncQuery(q, filter)
	data, err := QueryWithBuilder[CountOutput](ctx, db, q.PlaceholderFormat(squirrel.Dollar))

	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil
	}

	return data[0].Count, nil
}

func ListPrices(ctx context.Context, db Queryer, input *shared.StripePriceListParams) ([]*models.StripePrice, error) {

	filter := input.StripePriceListFilter
	pageInput := &input.PaginatedInput

	limit, offset := PaginateRepo(pageInput)
	param := ListPriceFilterFuncMap(&filter)
	sort := ListPriceOrderByMap(input)

	data, err := crudrepo.StripePrice.Get(
		ctx,
		db,
		param,
		sort,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ListPriceOrderByMap(input *shared.StripePriceListParams) *map[string]string {
	if input == nil {
		return nil
	}
	if input.SortParams.SortBy == "" {
		return nil
	}
	return &map[string]string{
		input.SortParams.SortBy: input.SortParams.SortOrder,
	}
}

func ListPriceFilterFuncMap(filter *shared.StripePriceListFilter) *map[string]any {
	if filter == nil {
		return nil
	}
	param := map[string]any{}
	if filter.Q != "" {
		param["_or"] = []map[string]any{
			{
				"name": map[string]any{
					"_ilike": "%" + filter.Q + "%",
				},
			},
			{
				"description": map[string]any{
					"_ilike": "%" + filter.Q + "%",
				},
			},
		}
	}

	if filter.Active != "" {
		if filter.Active == shared.Active {
			// q.Apply(
			// 	models.SelectWhere.StripePrices.Active.EQ(true),
			// )
			param["active"] = map[string]any{
				"_eq": true,
			}
		}
		if filter.Active == shared.Inactive {
			// q.Apply(
			// 	models.SelectWhere.StripePrices.Active.EQ(false),
			// )
			param["active"] = map[string]any{
				"_eq": false,
			}
		}
	}
	if len(filter.Ids) > 0 {
		// q.Apply(
		// 	models.SelectWhere.StripePrices.ID.In(filter.Ids...),
		// )
		param["id"] = map[string]any{
			"_in": filter.Ids,
		}
	}
	if len(filter.ProductIds) > 0 {
		// q.Apply(
		// 	models.SelectWhere.StripePrices.ProductID.In(filter.ProductIds...),
		// )
		param["product_id"] = map[string]any{
			"_in": filter.ProductIds,
		}
	}

	return &param
}

func CountPrices(ctx context.Context, db Queryer, filter *shared.StripePriceListFilter) (int64, error) {
	filermap := ListPriceFilterFuncMap(filter)
	data, err := crudrepo.StripePrice.Count(ctx, db, filermap)
	if err != nil {
		return 0, err
	}
	return data, nil
}

func ListCustomers(ctx context.Context, db Queryer, input *shared.StripeCustomerListParams) ([]*models.StripeCustomer, error) {

	filter := input.StripeCustomerListFilter
	pageInput := &input.PaginatedInput

	limit, offset := PaginateRepo(pageInput)
	where := ListCustomerFilterFunc(&filter)
	order := StripeCustomerOrderByFunc(input)
	data, err := crudrepo.StripeCustomer.Get(
		ctx,
		db,
		where,
		order,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func StripeCustomerOrderByFunc(input *shared.StripeCustomerListParams) *map[string]string {
	if input == nil {
		return nil
	}
	order := make(map[string]string)
	if slices.Contains(StripeCustomerColumnNames, input.SortBy) {
		order[input.SortBy] = strings.ToUpper(input.SortOrder)
	}
	return &order
}

func ListCustomerFilterFunc(filter *shared.StripeCustomerListFilter) *map[string]any {
	if filter == nil {
		return nil
	}
	where := map[string]any{}
	if len(filter.Ids) > 0 {
		where["id"] = map[string]any{
			"_in": filter.Ids,
		}
	}
	return &where
}

func CountCustomers(ctx context.Context, db Queryer, filter *shared.StripeCustomerListFilter) (int64, error) {
	where := ListCustomerFilterFunc(filter)
	data, err := crudrepo.StripeCustomer.Count(ctx, db, where)
	if err != nil {
		return 0, err
	}
	return data, nil
}

func ListSubscriptions(ctx context.Context, db Queryer, input *shared.StripeSubscriptionListParams) ([]*models.StripeSubscription, error) {

	filter := input.StripeSubscriptionListFilter
	pageInput := &input.PaginatedInput

	limit, offset := PaginateRepo(pageInput)
	where := ListSubscriptionFilterFunc(&filter)
	order := StripeSubscriptionOrderByFunc(input)
	data, err := crudrepo.StripeSubscription.Get(
		ctx,
		db,
		where,
		order,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func StripeSubscriptionOrderByFunc(input *shared.StripeSubscriptionListParams) *map[string]string {
	if input == nil {
		return nil
	}
	order := make(map[string]string)
	if slices.Contains(StripeSubscriptionColumnNames, input.SortBy) {
		order[input.SortBy] = strings.ToUpper(input.SortOrder)
	}
	return &order
}

func ListSubscriptionFilterFunc(filter *shared.StripeSubscriptionListFilter) *map[string]any {
	if filter == nil {
		return nil
	}
	where := map[string]any{}
	if len(filter.Ids) > 0 {
		where["id"] = map[string]any{
			"_in": filter.Ids,
		}
	}
	if len(filter.Status) > 0 {
		statuses := mapper.Map(filter.Status, func(s shared.StripeSubscriptionStatus) string { return string(s) })
		where["status"] = map[string]any{
			"_in": statuses,
		}
	}
	if filter.UserID != "" {
		where["user_id"] = map[string]any{
			"_eq": filter.UserID,
		}
	}
	return &where
}

func CountSubscriptions(ctx context.Context, db Queryer, filter *shared.StripeSubscriptionListFilter) (int64, error) {
	where := ListSubscriptionFilterFunc(filter)
	data, err := crudrepo.StripeSubscription.Count(ctx, db, where)
	if err != nil {
		return 0, err
	}
	return data, nil
}
