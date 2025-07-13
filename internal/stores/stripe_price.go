package stores

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/repository"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/types"
	"github.com/tkahng/playground/internal/tools/utils"
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
	dbx := s.db
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
	_, err := database.ExecWithBuilder(ctx, dbx, q.PlaceholderFormat(squirrel.Dollar))
	return err
}

func listPriceOrderByMap(input *StripePriceFilter) *map[string]string {
	if input == nil {
		return nil
	}
	if input.SortBy == "" {
		return nil
	}
	return &map[string]string{
		input.SortBy: input.SortOrder,
	}
}

type StripePriceFilter struct {
	PaginatedInput
	SortParams
	Q          string                                        `query:"q,omitempty" required:"false"`
	Ids        []string                                      `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	Active     types.OptionalParam[bool]                     `query:"active,omitempty" required:"false"`
	ProductIds []string                                      `query:"product_ids,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	Type       types.OptionalParam[models.StripePricingType] `query:"type,omitempty" required:"false" enum:"recurring,one_time"`
}

func listPriceFilterFuncMap(filter *StripePriceFilter) *map[string]any {
	if filter == nil {
		return nil
	}
	param := map[string]any{}

	if filter.Active.IsSet {
		param[models.StripePriceTable.Active] = map[string]any{
			"_eq": filter.Active.Value,
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
	if filter.Type.IsSet {
		param[models.StripePriceTable.Type] = map[string]any{
			"_eq": filter.Type.Value,
		}
	}

	return &param
}

// ListPrices implements PaymentStore.
func (s *DbPriceStore) ListPrices(ctx context.Context, input *StripePriceFilter) ([]*models.StripePrice, error) {
	dbx := s.db

	limit, offset := input.LimitOffset()
	param := listPriceFilterFuncMap(input)
	sort := listPriceOrderByMap(input)
	data, err := repository.StripePrice.Get(
		ctx,
		dbx,
		param,
		sort,
		&limit,
		&offset,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *DbPriceStore) CountPrices(ctx context.Context, filter *StripePriceFilter) (int64, error) {
	filermap := listPriceFilterFuncMap(filter)
	data, err := repository.StripePrice.Count(ctx, s.db, filermap)
	if err != nil {
		return 0, err
	}
	return data, nil
}

func SelectStripePriceColumns(qs squirrel.SelectBuilder, prefix string) squirrel.SelectBuilder {
	qs = qs.Column(models.StripePriceTablePrefix.ID + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripePriceTable.ID))).
		Column(models.StripePriceTablePrefix.ProductID + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripePriceTable.ProductID))).
		Column(models.StripePriceTablePrefix.LookupKey + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripePriceTable.LookupKey))).
		Column(models.StripePriceTablePrefix.Active + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripePriceTable.Active))).
		Column(models.StripePriceTablePrefix.UnitAmount + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripePriceTable.UnitAmount))).
		Column(models.StripePriceTablePrefix.Currency + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripePriceTable.Currency))).
		Column(models.StripePriceTablePrefix.Type + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripePriceTable.Type))).
		Column(models.StripePriceTablePrefix.Interval + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripePriceTable.Interval))).
		Column(models.StripePriceTablePrefix.IntervalCount + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripePriceTable.IntervalCount))).
		Column(models.StripePriceTablePrefix.TrialPeriodDays + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripePriceTable.TrialPeriodDays))).
		Column(models.StripePriceTablePrefix.Metadata + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripePriceTable.Metadata))).
		Column(models.StripePriceTablePrefix.CreatedAt + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripePriceTable.CreatedAt))).
		Column(models.StripePriceTablePrefix.UpdatedAt + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripePriceTable.UpdatedAt)))
	return qs
}
func (s *DbPriceStore) FindPrice(ctx context.Context, filter *StripePriceFilter) (*models.StripePrice, error) {
	param := listPriceFilterFuncMap(filter)
	data, err := repository.StripePrice.GetOne(
		ctx,
		s.db,
		param,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *DbPriceStore) LoadPricesByProductIds(ctx context.Context, productIds ...string) ([][]*models.StripePrice, error) {

	prices, err := repository.StripePrice.Get(
		ctx,
		s.db,
		&map[string]any{
			models.StripePriceTable.ProductID: map[string]any{
				"_in": productIds,
			},
		},
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

func (s *DbPriceStore) LoadPricesByIds(ctx context.Context, priceIds ...string) ([]*models.StripePrice, error) {
	if len(priceIds) == 0 {
		return nil, nil
	}
	prices, err := repository.StripePrice.Get(
		ctx,
		s.db,
		&map[string]any{
			models.StripePriceTable.ID: map[string]any{
				"_in": priceIds,
			},
		},
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return mapper.MapToPointer(prices, priceIds, func(t *models.StripePrice) string {
		if t == nil {
			return ""
		}
		return t.ID
	}), nil
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

var _ DbPriceStoreInterface = (*DbPriceStore)(nil)

type DbPriceStoreInterface interface {
	LoadPricesByIds(ctx context.Context, priceIds ...string) ([]*models.StripePrice, error)
	LoadPricesByProductIds(ctx context.Context, productIds ...string) ([][]*models.StripePrice, error)
	UpsertPrice(ctx context.Context, price *models.StripePrice) error
	ListPrices(ctx context.Context, input *StripePriceFilter) ([]*models.StripePrice, error)
	CountPrices(ctx context.Context, filter *StripePriceFilter) (int64, error)
	FindPrice(ctx context.Context, filter *StripePriceFilter) (*models.StripePrice, error)
	UpsertPriceFromStripe(ctx context.Context, price *stripe.Price) error
}
