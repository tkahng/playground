package payment

import (
	"context"

	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/types"
)

type PosrgresStripeStore struct {
	db database.Dbx
}

func NewPostgresPaymentStore(db database.Dbx) *PosrgresStripeStore {
	return &PosrgresStripeStore{db: db}
}

var _ PaymentStore = (*PosrgresStripeStore)(nil)

func (s *PosrgresStripeStore) UpsertPriceFromStripe(ctx context.Context, price *stripe.Price) error {
	if price == nil {
		return nil
	}
	val := &models.StripePrice{ID: price.ID, ProductID: price.Product.ID, Active: price.Active, LookupKey: &price.LookupKey, UnitAmount: &price.UnitAmount, Currency: string(price.Currency), Type: models.StripePricingType(price.Type), Metadata: price.Metadata}
	if price.Recurring != nil {
		recur := price.Recurring
		val.Interval = types.Pointer(models.StripePricingPlanInterval(recur.Interval))
		val.IntervalCount = types.Pointer(recur.IntervalCount)
		val.TrialPeriodDays = types.Pointer(recur.TrialPeriodDays)
	}
	return s.UpsertPrice(ctx, val)
}

func (s *PosrgresStripeStore) UpsertPrice(ctx context.Context, price *models.StripePrice) error {
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

// UpsertProductFromStripe implements PaymentStore.
func (s *PosrgresStripeStore) UpsertProductFromStripe(ctx context.Context, product *stripe.Product) error {
	if product == nil {
		return nil
	}
	var image *string
	if len(product.Images) > 0 {
		image = &product.Images[0]
	}
	param := &models.StripeProduct{
		ID:          product.ID,
		Active:      product.Active,
		Name:        product.Name,
		Description: &product.Description,
		Image:       image,
		Metadata:    product.Metadata,
	}
	return s.UpsertProduct(ctx, param)
}

func (s *PosrgresStripeStore) UpsertProduct(ctx context.Context, product *models.StripeProduct) error {
	var dbx database.Dbx = s.db
	q := squirrel.Insert("stripe_products").
		Columns(
			"id",
			"active",
			"name",
			"description",
			"image",
			"metadata",
		).
		Values(
			product.ID,
			product.Active,
			product.Name,
			product.Description,
			product.Image,
			product.Metadata,
		).Suffix(`ON CONFLICT (id) DO UPDATE SET 
						active = EXCLUDED.active, 
						name = EXCLUDED.name, 
						description = EXCLUDED.description, 
						image = EXCLUDED.image, 
						metadata = EXCLUDED.metadata
		`)
	return database.ExecWithBuilder(ctx, dbx, q.PlaceholderFormat(squirrel.Dollar))
}

// FindCustomerByStripeId implements PaymentStore.
func (s *PosrgresStripeStore) FindCustomerByStripeId(ctx context.Context, stripeId string) (*models.StripeCustomer, error) {
	data, err := crudrepo.StripeCustomer.GetOne(
		ctx,
		s.db,
		&map[string]any{"stripe_id": map[string]any{"_eq": stripeId}},
	)
	return database.OptionalRow(data, err)
}

// FindCustomerByUserId implements PaymentStore.
func (s *PosrgresStripeStore) FindCustomerByUserId(ctx context.Context, userId uuid.UUID) (*models.StripeCustomer, error) {
	data, err := crudrepo.StripeCustomer.GetOne(
		ctx,
		s.db,
		&map[string]any{"id": map[string]any{"_eq": userId.String()}},
	)
	return database.OptionalRow(data, err)
}

// FindLatestActiveSubscriptionByTeamId implements PaymentStore.
func (s *PosrgresStripeStore) FindLatestActiveSubscriptionByTeamId(ctx context.Context, teamId uuid.UUID) (*models.StripeSubscription, error) {
	data, err := crudrepo.StripeSubscription.Get(
		ctx,
		s.db,
		&map[string]any{
			"team_id": map[string]any{"_eq": teamId.String()},
			"status": map[string]any{"_in": []string{
				string(models.StripeSubscriptionStatusActive),
				string(models.StripeSubscriptionStatusTrialing),
			}},
		},
		&map[string]string{"created_at": "DESC"},
		types.Pointer(1),
		nil,
	)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return database.OptionalRow(data[0], err)
}

// FindProductByStripeId implements PaymentStore.
func (s *PosrgresStripeStore) FindProductByStripeId(ctx context.Context, productId string) (*models.StripeProduct, error) {
	data, err := crudrepo.StripeProduct.GetOne(
		ctx,
		s.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": productId,
			},
		},
	)
	return database.OptionalRow(data, err)
}

const (
	getSubscriptionWithPriceByIdQuery = `
SELECT ss.id AS "subscription.id",
        ss.user_id AS "subscription.user_id",
        ss.status AS "subscription.status",
        ss.metadata AS "subscription.metadata",
        ss.price_id AS "subscription.price_id",
        ss.quantity AS "subscription.quantity",
        ss.cancel_at_period_end AS "subscription.cancel_at_period_end",
        ss.created AS "subscription.created",
        ss.current_period_start AS "subscription.current_period_start",
        ss.current_period_end AS "subscription.current_period_end",
        ss.ended_at AS "subscription.ended_at",
        ss.cancel_at AS "subscription.cancel_at",
        ss.canceled_at AS "subscription.canceled_at",
        ss.trial_start AS "subscription.trial_start",
        ss.trial_end AS "subscription.trial_end",
        ss.created_at AS "subscription.created_at",
        ss.updated_at AS "subscription.updated_at",
        sp.id AS "price.id",
        sp.product_id AS "price.product_id",
        sp.lookup_key AS "price.lookup_key",
        sp.active AS "price.active",
        sp.unit_amount AS "price.unit_amount",
        sp.currency AS "price.currency",
        sp.type AS "price.type",
        sp.interval AS "price.interval",
        sp.interval_count AS "price.interval_count",
        sp.trial_period_days AS "price.trial_period_days",
        sp.metadata AS "price.metadata",
        sp.created_at AS "price.created_at",
        sp.updated_at AS "price.updated_at",
        p.id AS "product.id",
        p.name AS "product.name",
        p.description AS "product.description",
        p.active AS "product.active",
        p.image AS "product.image",
        p.metadata AS "product.metadata",
        p.created_at AS "product.created_at",
        p.updated_at AS "product.updated_at"
FROM public.stripe_subscriptions ss
        JOIN public.stripe_prices sp ON ss.price_id = sp.id
        JOIN public.stripe_products p ON sp.product_id = p.id
WHERE ss.id = $1
		`
)

// FindSubscriptionWithPriceById implements PaymentStore.
func (s *PosrgresStripeStore) FindSubscriptionWithPriceById(ctx context.Context, stripeId string) (*models.SubscriptionWithPrice, error) {
	data, err := database.QueryAll[*models.SubscriptionWithPrice](ctx, s.db, getSubscriptionWithPriceByIdQuery, stripeId)
	if err != nil {
		return nil, err
	}
	var first *models.SubscriptionWithPrice
	if len(data) > 0 {
		first = data[0]
	}
	return first, nil
}

// FindTeamById implements PaymentStore.
func (s *PosrgresStripeStore) FindTeamById(ctx context.Context, teamId uuid.UUID) (*models.Team, error) {
	return crudrepo.Team.GetOne(
		ctx,
		s.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": teamId.String(),
			},
		},
	)
}

// FindValidPriceById implements PaymentStore.
func (s *PosrgresStripeStore) FindValidPriceById(ctx context.Context, priceId string) (*models.StripePrice, error) {
	data, err := crudrepo.StripePrice.GetOne(
		ctx,
		s.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": priceId,
			},
			"type": map[string]any{
				"_eq": string(models.StripePricingTypeRecurring),
			},
		},
	)
	return data, err
}

// IsFirstSubscription implements PaymentStore.
func (s *PosrgresStripeStore) IsFirstSubscription(ctx context.Context, userId uuid.UUID) (bool, error) {
	data, err := crudrepo.StripeSubscription.Count(
		ctx,
		s.db,
		&map[string]any{
			"team_id": map[string]any{
				"_eq": userId.String(),
			},
		},
	)
	return data > 0, err
}

// ListPrices implements PaymentStore.
func (s *PosrgresStripeStore) ListPrices(ctx context.Context, input *shared.StripePriceListParams) ([]*models.StripePrice, error) {
	var dbx database.Dbx = s.db
	filter := input.StripePriceListFilter
	pageInput := &input.PaginatedInput
	limit, offset := database.PaginateRepo(pageInput)
	param := queries.ListPriceFilterFuncMap(&filter)
	sort := queries.ListPriceOrderByMap(input)
	data, err := crudrepo.StripePrice.Get(
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

// ListProducts implements PaymentStore.
func (s *PosrgresStripeStore) ListProducts(ctx context.Context, input *shared.StripeProductListParams) ([]*models.StripeProduct, error) {
	var dbx database.Dbx = s.db
	q := squirrel.Select("stripe_products.*").From("stripe_products")
	filter := input.StripeProductListFilter
	pageInput := &input.PaginatedInput
	q = database.Paginate(q, pageInput)
	q = queries.ListProductFilterFuncQuery(q, &filter)
	data, err := database.QueryWithBuilder[*models.StripeProduct](
		ctx,
		dbx,
		q.PlaceholderFormat(squirrel.Dollar),
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// UpsertCustomerStripeId implements PaymentStore.
func (s *PosrgresStripeStore) UpsertCustomerStripeId(ctx context.Context, userId uuid.UUID, stripeCustomerId string) error {
	var dbx database.Dbx = s.db
	q := squirrel.Insert("stripe_customers").Columns("id", "stripe_id").Values(userId, stripeCustomerId).Suffix(`ON CONFLICT (id) DO UPDATE SET stripe_id = EXCLUDED.stripe_id`)
	return database.ExecWithBuilder(ctx, dbx, q.PlaceholderFormat(squirrel.Dollar))
}

// UpsertSubscriptionFromStripe implements PaymentStore.
func (s *PosrgresStripeStore) UpsertSubscriptionFromStripe(ctx context.Context, sub *stripe.Subscription, userId uuid.UUID) error {
	if sub == nil {
		return nil
	}
	var item *stripe.SubscriptionItem
	if len(sub.Items.Data) > 0 {
		item = sub.Items.Data[0]
	}
	if item == nil || item.Price == nil {
		return errors.New("price not found")
	}
	status := models.StripeSubscriptionStatus(sub.Status)
	err := s.UpsertSubscription(
		ctx,
		&models.StripeSubscription{
			ID:                 sub.ID,
			TeamID:             userId,
			Status:             models.StripeSubscriptionStatus(status),
			Metadata:           sub.Metadata,
			PriceID:            item.Price.ID,
			Quantity:           item.Quantity,
			CancelAtPeriodEnd:  sub.CancelAtPeriodEnd,
			Created:            queries.Int64ToISODate(sub.Created),
			CurrentPeriodStart: queries.Int64ToISODate(item.CurrentPeriodStart),
			CurrentPeriodEnd:   queries.Int64ToISODate(item.CurrentPeriodEnd),
			EndedAt:            types.Pointer(queries.Int64ToISODate(sub.EndedAt)),
			CancelAt:           types.Pointer(queries.Int64ToISODate(sub.CancelAt)),
			CanceledAt:         types.Pointer(queries.Int64ToISODate(sub.CanceledAt)),
			TrialStart:         types.Pointer(queries.Int64ToISODate(sub.TrialStart)),
			TrialEnd:           types.Pointer(queries.Int64ToISODate(sub.TrialEnd)),
		},
	)
	return err
}

func (s *PosrgresStripeStore) UpsertSubscription(ctx context.Context, sub *models.StripeSubscription) error {
	err := queries.UpsertSubscription(
		ctx,
		s.db,
		sub,
	)
	return err
}
