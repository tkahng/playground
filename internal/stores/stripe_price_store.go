package stores

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/types"
)

type DbPriceStore struct {
	db database.Dbx
}

func NewDbPriceStore(db database.Dbx) *DbPriceStore {
	return &DbPriceStore{
		db: db,
	}
}

func (s *DbPriceStore) WithTx(tx database.Dbx) *DbPriceStore {
	return &DbPriceStore{
		db: tx,
	}
}

func (s *DbPriceStore) UpsertPrice(ctx context.Context, price *models.StripePrice) error {
	var dbx database.Dbx = s.db
	q := squirrel.Insert("stripe_prices").Columns("id", "product_id", "lookup_key", "active", "unit_amount", "currency", "type", "interval", "interval_count", "trial_period_days", "metadata").Values(price.ID, price.ProductID, price.LookupKey, price.Active, price.UnitAmount, price.Currency, price.Type, price.Interval, price.IntervalCount, price.TrialPeriodDays, price.Metadata).Suffix(`
		ON CONFLICT(id) DO UPDATE SET 
			product_id = EXCLUDED.product_id,
			lookup_key = EXCLUDED.lookup_key,
			active = EXCLUDED.active,
			unit_amount = EXCLUDED.unit_amount,
			currency = EXCLUDED.currency,
			type = EXCLUDED.type,
			interval = EXCLUDED.interval,
			interval_count = EXCLUDED.interval_count,
			trial_period_days = EXCLUDED.trial_period_days,
			metadata = EXCLUDED.metadata
		`)
	return database.ExecWithBuilder(ctx, dbx, q.PlaceholderFormat(squirrel.Dollar))
}

func listPriceOrderByMap(input *shared.StripePriceListParams) *map[string]string {
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

func listPriceFilterFuncMap(filter *shared.StripePriceListFilter) *map[string]any {
	if filter == nil {
		return nil
	}
	param := map[string]any{}

	if filter.Active != "" {
		if filter.Active == shared.Active {
			param[models.StripePriceTable.Active] = map[string]any{
				"_eq": true,
			}
		}
		if filter.Active == shared.Inactive {
			param[models.StripePriceTable.Active] = map[string]any{
				"_eq": false,
			}
		}
	}
	if len(filter.Ids) > 0 {
		param[models.StripePriceTable.ID] = map[string]any{
			"_in": filter.Ids,
		}
	}
	if len(filter.ProductIds) > 0 {
		param[models.StripePriceTable.ProductID] = map[string]any{
			"_in": filter.ProductIds,
		}
	}

	return &param
}

// ListPrices implements PaymentStore.
func (s *DbPriceStore) ListPrices(ctx context.Context, input *shared.StripePriceListParams) ([]*models.StripePrice, error) {
	var dbx database.Dbx = s.db
	filter := input.StripePriceListFilter
	pageInput := &input.PaginatedInput
	limit, offset := database.PaginateRepo(pageInput)
	param := listPriceFilterFuncMap(&filter)
	sort := listPriceOrderByMap(input)
	data, err := repository.StripePrice.Get(
		ctx,
		dbx,
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

func (s *DbPriceStore) CountPrices(ctx context.Context, filter *shared.StripePriceListFilter) (int64, error) {
	filermap := listPriceFilterFuncMap(filter)
	data, err := repository.StripePrice.Count(ctx, s.db, filermap)
	if err != nil {
		return 0, err
	}
	return data, nil
}

func SelectStripePriceColumns(qs squirrel.SelectBuilder, prefix string) squirrel.SelectBuilder {
	qs = qs.Column(models.StripePriceTablePrefix.ID + " AS " + Quote(WithPrefix(prefix, models.StripePriceTable.ID))).
		Column(models.StripePriceTablePrefix.ProductID + " AS " + Quote(WithPrefix(prefix, models.StripePriceTable.ProductID))).
		Column(models.StripePriceTablePrefix.LookupKey + " AS " + Quote(WithPrefix(prefix, models.StripePriceTable.LookupKey))).
		Column(models.StripePriceTablePrefix.Active + " AS " + Quote(WithPrefix(prefix, models.StripePriceTable.Active))).
		Column(models.StripePriceTablePrefix.UnitAmount + " AS " + Quote(WithPrefix(prefix, models.StripePriceTable.UnitAmount))).
		Column(models.StripePriceTablePrefix.Currency + " AS " + Quote(WithPrefix(prefix, models.StripePriceTable.Currency))).
		Column(models.StripePriceTablePrefix.Type + " AS " + Quote(WithPrefix(prefix, models.StripePriceTable.Type))).
		Column(models.StripePriceTablePrefix.Interval + " AS " + Quote(WithPrefix(prefix, models.StripePriceTable.Interval))).
		Column(models.StripePriceTablePrefix.IntervalCount + " AS " + Quote(WithPrefix(prefix, models.StripePriceTable.IntervalCount))).
		Column(models.StripePriceTablePrefix.TrialPeriodDays + " AS " + Quote(WithPrefix(prefix, models.StripePriceTable.TrialPeriodDays))).
		Column(models.StripePriceTablePrefix.Metadata + " AS " + Quote(WithPrefix(prefix, models.StripePriceTable.Metadata))).
		Column(models.StripePriceTablePrefix.CreatedAt + " AS " + Quote(WithPrefix(prefix, models.StripePriceTable.CreatedAt))).
		Column(models.StripePriceTablePrefix.UpdatedAt + " AS " + Quote(WithPrefix(prefix, models.StripePriceTable.UpdatedAt)))
	return qs
}

// FindActivePriceById implements PaymentStore.
func (s *DbPriceStore) FindActivePriceById(ctx context.Context, priceId string) (*models.StripePrice, error) {
	data, err := repository.StripePrice.GetOne(
		ctx,
		s.db,
		&map[string]any{
			models.StripePriceTable.ID: map[string]any{
				"_eq": priceId,
			},
			models.StripePriceTable.Type: map[string]any{
				"_eq": string(models.StripePricingTypeRecurring),
			},
		},
	)
	return data, err
}

func (s *DbPriceStore) UpsertPriceFromStripe(ctx context.Context, price *stripe.Price) error {
	if price == nil {
		return nil
	}
	val := &models.StripePrice{
		ID:         price.ID,
		ProductID:  price.Product.ID,
		Active:     price.Active,
		LookupKey:  &price.LookupKey,
		UnitAmount: &price.UnitAmount,
		Currency:   string(price.Currency),
		Type:       models.StripePricingType(price.Type),
		Metadata:   price.Metadata,
	}
	if price.Recurring != nil {
		recur := price.Recurring
		val.Interval = types.Pointer(models.StripePricingPlanInterval(recur.Interval))
		val.IntervalCount = types.Pointer(recur.IntervalCount)
		val.TrialPeriodDays = types.Pointer(recur.TrialPeriodDays)
	}
	return s.UpsertPrice(ctx, val)
}
