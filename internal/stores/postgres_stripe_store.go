package stores

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/types"
	"github.com/tkahng/authgo/internal/tools/utils"
)

var _ services.PaymentStripeStore = (*PostgresStripeStore)(nil)

func NewPostgresStripeStore(db database.Dbx) *PostgresStripeStore {
	return &PostgresStripeStore{
		db: db,
	}
}

type PostgresStripeStore struct {
	db database.Dbx
}

func (s *PostgresStripeStore) WithTx(tx database.Dbx) *PostgresStripeStore {
	return &PostgresStripeStore{
		db: tx,
	}
}

func (s *PostgresStripeStore) LoadPricesByIds(ctx context.Context, priceIds ...string) ([]*models.StripePrice, error) {
	if len(priceIds) == 0 {
		return nil, nil
	}
	prices, err := crudrepo.StripePrice.Get(
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

func (s *PostgresStripeStore) LoadPricesByProductIds(ctx context.Context, productIds ...string) ([][]*models.StripePrice, error) {

	prices, err := crudrepo.StripePrice.Get(
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

func (s *PostgresStripeStore) LoadProductsByIds(ctx context.Context, productIds ...string) ([]*models.StripeProduct, error) {
	if len(productIds) == 0 {
		return nil, nil
	}
	products, err := crudrepo.StripeProduct.Get(
		ctx,
		s.db,
		&map[string]any{
			models.StripeProductTable.ID: map[string]any{
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
	return mapper.MapToPointer(products, productIds, func(t *models.StripeProduct) string {
		if t == nil {
			return ""
		}
		return t.ID
	}), nil
}

func (s *PostgresStripeStore) LoadPricesWithProductByPriceIds(ctx context.Context, priceIds ...string) ([]*models.StripePrice, error) {
	if len(priceIds) == 0 {
		return nil, nil
	}
	prices, err := s.LoadPricesByIds(ctx, priceIds...)
	if err != nil {
		return nil, err
	}
	productIds := mapper.Map(prices, func(price *models.StripePrice) string {
		if price == nil || price.ProductID == "" {
			return ""
		}
		return price.ProductID
	})
	products, err := s.LoadProductsByIds(ctx, productIds...)
	if err != nil {
		return nil, err
	}
	for i, price := range prices {
		if price == nil {
			continue
		}
		product := products[i]
		if product == nil {
			continue
		}
		if product.ID != price.ProductID {
			continue
		}
		price.Product = product
	}
	return prices, nil
}

func (s *PostgresStripeStore) LoadSubscriptionsPriceProduct(ctx context.Context, subscriptions ...*models.StripeSubscription) error {
	if len(subscriptions) == 0 {
		return nil
	}
	priceIds := mapper.Map(subscriptions, func(sub *models.StripeSubscription) string {
		if sub == nil || sub.PriceID == "" {
			return ""
		}
		return sub.PriceID
	})
	prices, err := s.LoadPricesWithProductByPriceIds(ctx, priceIds...)
	if err != nil {
		return err
	}
	for i, sub := range subscriptions {
		if sub == nil {
			continue
		}
		price := prices[i]
		if price == nil {
			continue
		}
		if price.ID != sub.PriceID {
			continue
		}
		sub.Price = price
	}
	return nil
}

func (s *PostgresStripeStore) LoadSubscriptionsByIds(ctx context.Context, subscriptionIds ...string) ([]*models.StripeSubscription, error) {
	if len(subscriptionIds) == 0 {
		return nil, nil
	}
	where := map[string]any{
		models.StripeSubscriptionTable.ID: map[string]any{
			"_in": subscriptionIds,
		},
	}
	subscriptions, err := crudrepo.StripeSubscription.Get(
		ctx,
		s.db,
		&where,
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return mapper.MapToPointer(subscriptions, subscriptionIds, func(t *models.StripeSubscription) string {
		if t == nil {
			return ""
		}
		return t.ID
	}), nil
}

func (s *PostgresStripeStore) FindActiveSubscriptionsByCustomerIds(ctx context.Context, customerIds ...string) ([]*models.StripeSubscription, error) {
	if len(customerIds) == 0 {
		return nil, nil
	}
	qs := squirrel.Select()
	qs = SelectStripeSubscriptionColumns(qs, "")
	qs = qs.
		From("stripe_subscriptions").
		Where(squirrel.Or{
			squirrel.And{
				squirrel.Eq{
					"stripe_subscriptions.stripe_customer_id": customerIds,
				},
				squirrel.Eq{
					"stripe_subscriptions.status": models.StripeSubscriptionStatusActive,
				},
			},
			squirrel.And{
				squirrel.Eq{
					"stripe_subscriptions.stripe_customer_id": customerIds,
				},
				squirrel.Eq{
					"stripe_subscriptions.status": models.StripeSubscriptionStatusTrialing,
				},
				squirrel.Gt{
					"stripe_subscriptions.trial_end": time.Now().Format(time.RFC3339Nano),
				},
			},
		})
	subscriptions, err := database.QueryWithBuilder[*models.StripeSubscription](ctx, s.db, qs.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return mapper.MapToPointer(subscriptions, customerIds, func(s *models.StripeSubscription) string {

		if s == nil {
			return ""
		}
		return s.StripeCustomerID
	}), nil
}

func (s *PostgresStripeStore) FindActiveSubscriptionsByTeamIds(ctx context.Context, teamIds ...uuid.UUID) ([]*models.StripeSubscription, error) {
	if len(teamIds) == 0 {
		return nil, nil
	}
	qs := squirrel.Select()
	qs = SelectStripeSubscriptionColumns(qs, "")
	qs = SelectStripeCustomerColumns(qs, "stripe_customer")
	qs = qs.
		From("stripe_subscriptions").
		Join("stripe_customers ON stripe_subscriptions.stripe_customer_id = stripe_customers.id").
		Where(squirrel.Or{
			squirrel.And{
				squirrel.Eq{
					"stripe_customers.team_id": teamIds,
				},
				squirrel.Eq{
					"stripe_subscriptions.status": models.StripeSubscriptionStatusActive,
				},
			},
			squirrel.And{
				squirrel.Eq{
					"stripe_customers.team_id": teamIds,
				},
				squirrel.Eq{
					"stripe_subscriptions.status": models.StripeSubscriptionStatusTrialing,
				},
				squirrel.Gt{
					"stripe_subscriptions.trial_end": time.Now().Format(time.RFC3339Nano),
				},
			},
		})
	subscriptions, err := database.QueryWithBuilder[*models.StripeSubscription](ctx, s.db, qs.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return mapper.MapToPointer(subscriptions, teamIds, func(s *models.StripeSubscription) uuid.UUID {

		if s == nil || s.StripeCustomer == nil || s.StripeCustomer.TeamID == nil {
			return uuid.Nil
		}
		return *s.StripeCustomer.TeamID
	}), nil
}

func (s *PostgresStripeStore) FindActiveSubscriptionsByUserIds(ctx context.Context, userIds ...uuid.UUID) ([]*models.StripeSubscription, error) {
	if len(userIds) == 0 {
		return nil, nil
	}
	qs := squirrel.Select()
	qs = SelectStripeSubscriptionColumns(qs, "")
	qs = SelectStripeCustomerColumns(qs, "stripe_customer")
	qs = qs.
		From("stripe_subscriptions").
		Join("stripe_customers ON stripe_subscriptions.stripe_customer_id = stripe_customers.id").
		Where(squirrel.Or{
			squirrel.And{
				squirrel.Eq{
					"stripe_customers.user_id": userIds,
				},
				squirrel.Eq{
					"stripe_subscriptions.status": models.StripeSubscriptionStatusActive,
				},
			},
			squirrel.And{
				squirrel.Eq{
					"stripe_customers.user_id": userIds,
				},
				squirrel.Eq{
					"stripe_subscriptions.status": models.StripeSubscriptionStatusTrialing,
				},
				squirrel.Gt{
					"stripe_subscriptions.trial_end": time.Now().Format(time.RFC3339Nano),
				},
			},
		})
	subscriptions, err := database.QueryWithBuilder[*models.StripeSubscription](ctx, s.db, qs.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return mapper.MapToPointer(subscriptions, userIds, func(s *models.StripeSubscription) uuid.UUID {

		if s == nil || s.StripeCustomer == nil || s.StripeCustomer.UserID == nil {
			return uuid.Nil
		}
		return *s.StripeCustomer.UserID
	}), nil
}

func (s *PostgresStripeStore) FindSubscriptionsWithPriceProductByIds(ctx context.Context, subscriptionIds ...string) ([]*models.StripeSubscription, error) {
	qs := squirrel.Select()
	qs = SelectStripeSubscriptionColumns(qs, "")
	qs = SelectStripePriceColumns(qs, "price")
	qs = SelectStripeProductColumns(qs, "price.product")
	qs = qs.From(models.StripeSubscriptionTableName).
		Join(models.StripePriceTableName + " ON " + models.StripeSubscriptionTablePrefix.PriceID + " = " + models.StripePriceTablePrefix.ID).
		Join(models.StripeProductTableName + " ON " + models.StripePriceTablePrefix.ProductID + " = " + models.StripeProductTablePrefix.ID).
		Where(squirrel.Eq{models.StripeSubscriptionTablePrefix.ID: subscriptionIds})
	data, err := database.QueryWithBuilder[*models.StripeSubscription](ctx, s.db, qs.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}

	return mapper.MapToPointer(data, subscriptionIds, func(s *models.StripeSubscription) string {
		if s == nil {
			return ""
		}
		return s.ID
	}), nil
}

func (s *PostgresStripeStore) CreateCustomer(ctx context.Context, customer *models.StripeCustomer) (*models.StripeCustomer, error) {
	if customer == nil {
		return nil, errors.New("customer is nil")
	}
	if customer.UserID != nil {
		customer.CustomerType = models.StripeCustomerTypeUser
	} else if customer.TeamID != nil {
		customer.CustomerType = models.StripeCustomerTypeTeam
	} else {
		return nil, errors.New("customer type is not set")
	}
	return crudrepo.StripeCustomer.PostOne(
		ctx,
		s.db,
		customer,
	)
}

func (s *PostgresStripeStore) UpsertPriceFromStripe(ctx context.Context, price *stripe.Price) error {
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

func (s *PostgresStripeStore) UpsertPrice(ctx context.Context, price *models.StripePrice) error {
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
func (s *PostgresStripeStore) UpsertProductFromStripe(ctx context.Context, product *stripe.Product) error {
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

func (s *PostgresStripeStore) UpsertProduct(ctx context.Context, product *models.StripeProduct) error {
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

// FindCustomer implements PaymentStore.
func (s *PostgresStripeStore) FindCustomer(ctx context.Context, customer *models.StripeCustomer) (*models.StripeCustomer, error) {
	if customer == nil {
		return nil, nil
	}
	where := map[string]any{}
	if customer.ID != "" {
		where[models.StripeCustomerTable.ID] = map[string]any{
			"_eq": customer.ID,
		}
	}
	if customer.TeamID != nil {
		where[models.StripeCustomerTable.TeamID] = map[string]any{
			"_eq": customer.TeamID,
		}
	}
	if customer.UserID != nil {
		where[models.StripeCustomerTable.UserID] = map[string]any{
			"_eq": customer.UserID,
		}
	}
	data, err := crudrepo.StripeCustomer.GetOne(
		ctx,
		s.db,
		&where,
	)
	return database.OptionalRow(data, err)
}

// FindProductById implements PaymentStore.
func (s *PostgresStripeStore) FindProductById(ctx context.Context, productId string) (*models.StripeProduct, error) {
	data, err := crudrepo.StripeProduct.GetOne(
		ctx,
		s.db,
		&map[string]any{
			models.StripeProductTable.ID: map[string]any{
				"_eq": productId,
			},
		},
	)
	return database.OptionalRow(data, err)
}

func SelectStripeSubscriptionColumns(qs squirrel.SelectBuilder, prefix string) squirrel.SelectBuilder {
	qs = qs.
		Column(models.StripeSubscriptionTablePrefix.ID + " AS " + Quote(WithPrefix(prefix, models.StripeSubscriptionTable.ID))).
		Column(models.StripeSubscriptionTablePrefix.StripeCustomerID + " AS " + Quote(WithPrefix(prefix, models.StripeSubscriptionTable.StripeCustomerID))).
		Column(models.StripeSubscriptionTablePrefix.Status + " AS " + Quote(WithPrefix(prefix, models.StripeSubscriptionTable.Status))).
		Column(models.StripeSubscriptionTablePrefix.Metadata + " AS " + Quote(WithPrefix(prefix, models.StripeSubscriptionTable.Metadata))).
		Column(models.StripeSubscriptionTablePrefix.ItemID + " AS " + Quote(WithPrefix(prefix, models.StripeSubscriptionTable.ItemID))).
		Column(models.StripeSubscriptionTablePrefix.PriceID + " AS " + Quote(WithPrefix(prefix, models.StripeSubscriptionTable.PriceID))).
		Column(models.StripeSubscriptionTablePrefix.Quantity + " AS " + Quote(WithPrefix(prefix, models.StripeSubscriptionTable.Quantity))).
		Column(models.StripeSubscriptionTablePrefix.CancelAtPeriodEnd + " AS " + Quote(WithPrefix(prefix, models.StripeSubscriptionTable.CancelAtPeriodEnd))).
		Column(models.StripeSubscriptionTablePrefix.Created + " AS " + Quote(WithPrefix(prefix, models.StripeSubscriptionTable.Created))).
		Column(models.StripeSubscriptionTablePrefix.CurrentPeriodStart + " AS " + Quote(WithPrefix(prefix, models.StripeSubscriptionTable.CurrentPeriodStart))).
		Column(models.StripeSubscriptionTablePrefix.CurrentPeriodEnd + " AS " + Quote(WithPrefix(prefix, models.StripeSubscriptionTable.CurrentPeriodEnd))).
		Column(models.StripeSubscriptionTablePrefix.EndedAt + " AS " + Quote(WithPrefix(prefix, models.StripeSubscriptionTable.EndedAt))).
		Column(models.StripeSubscriptionTablePrefix.CancelAt + " AS " + Quote(WithPrefix(prefix, models.StripeSubscriptionTable.CancelAt))).
		Column(models.StripeSubscriptionTablePrefix.CanceledAt + " AS " + Quote(WithPrefix(prefix, models.StripeSubscriptionTable.CanceledAt))).
		Column(models.StripeSubscriptionTablePrefix.TrialStart + " AS " + Quote(WithPrefix(prefix, models.StripeSubscriptionTable.TrialStart))).
		Column(models.StripeSubscriptionTablePrefix.TrialEnd + " AS " + Quote(WithPrefix(prefix, models.StripeSubscriptionTable.TrialEnd))).
		Column(models.StripeSubscriptionTablePrefix.CreatedAt + " AS " + Quote(WithPrefix(prefix, models.StripeSubscriptionTable.CreatedAt))).
		Column(models.StripeSubscriptionTablePrefix.UpdatedAt + " AS " + Quote(WithPrefix(prefix, models.StripeSubscriptionTable.UpdatedAt)))
	return qs
}

func SelectStripeProductColumns(qs squirrel.SelectBuilder, prefix string) squirrel.SelectBuilder {
	qs = qs.Column(models.StripeProductTablePrefix.ID + " AS " + Quote(WithPrefix(prefix, models.StripeProductTable.ID))).
		Column(models.StripeProductTablePrefix.Name + " AS " + Quote(WithPrefix(prefix, models.StripeProductTable.Name))).
		Column(models.StripeProductTablePrefix.Description + " AS " + Quote(WithPrefix(prefix, models.StripeProductTable.Description))).
		Column(models.StripeProductTablePrefix.Active + " AS " + Quote(WithPrefix(prefix, models.StripeProductTable.Active))).
		Column(models.StripeProductTablePrefix.Image + " AS " + Quote(WithPrefix(prefix, models.StripeProductTable.Image))).
		Column(models.StripeProductTablePrefix.Metadata + " AS " + Quote(WithPrefix(prefix, models.StripeProductTable.Metadata))).
		Column(models.StripeProductTablePrefix.CreatedAt + " AS " + Quote(WithPrefix(prefix, models.StripeProductTable.CreatedAt))).
		Column(models.StripeProductTablePrefix.UpdatedAt + " AS " + Quote(WithPrefix(prefix, models.StripeProductTable.UpdatedAt)))

	return qs
}
func WithPrefix(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return fmt.Sprintf("%s.%s", prefix, name)
}

func Quote(name string) string {
	return fmt.Sprintf("\"%s\"", name)
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

func SelectStripeCustomerColumns(qs squirrel.SelectBuilder, prefix string) squirrel.SelectBuilder {
	qs = qs.Column(models.StripeCustomerTablePrefix.ID + " AS " + Quote(WithPrefix(prefix, models.StripeCustomerTable.ID))).
		Column(models.StripeCustomerTablePrefix.Email + " AS " + Quote(WithPrefix(prefix, models.StripeCustomerTable.Email))).
		Column(models.StripeCustomerTablePrefix.Name + " AS " + Quote(WithPrefix(prefix, models.StripeCustomerTable.Name))).
		Column(models.StripeCustomerTablePrefix.UserID + " AS " + Quote(WithPrefix(prefix, models.StripeCustomerTable.UserID))).
		Column(models.StripeCustomerTablePrefix.TeamID + " AS " + Quote(WithPrefix(prefix, models.StripeCustomerTable.TeamID))).
		Column(models.StripeCustomerTablePrefix.CustomerType + " AS " + Quote(WithPrefix(prefix, models.StripeCustomerTable.CustomerType))).
		Column(models.StripeCustomerTablePrefix.BillingAddress + " AS " + Quote(WithPrefix(prefix, models.StripeCustomerTable.BillingAddress))).
		Column(models.StripeCustomerTablePrefix.PaymentMethod + " AS " + Quote(WithPrefix(prefix, models.StripeCustomerTable.PaymentMethod))).
		Column(models.StripeCustomerTablePrefix.CreatedAt + " AS " + Quote(WithPrefix(prefix, models.StripeCustomerTable.CreatedAt))).
		Column(models.StripeCustomerTablePrefix.UpdatedAt + " AS " + Quote(WithPrefix(prefix, models.StripeCustomerTable.UpdatedAt)))
	return qs
}

const (
	getLatestActiveSubscriptionWithPriceByCustomerIdQuery = `
SELECT ss.id AS "subscription.id",
        ss.stripe_customer_id AS "subscription.stripe_customer_id",
        ss.status AS "subscription.status",
        ss.metadata AS "subscription.metadata",
		ss.item_id AS "subscription.item_id",
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
WHERE ss.stripe_customer_id = $1
        AND ss.status IN ('active', 'trialing')
ORDER BY ss.created_at DESC;
		`
)

func (s *PostgresStripeStore) FindActiveSubscriptionByCustomerId(ctx context.Context, customerId string) (*models.StripeSubscription, error) {
	data, err := s.FindActiveSubscriptionsByCustomerIds(ctx, customerId)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	subscription := data[0]
	if subscription == nil {
		return nil, nil
	}
	if subscription.Price == nil {
		return nil, fmt.Errorf("subscription %s has no price", subscription.ID)
	}
	if subscription.Price.Product == nil {
		return nil, fmt.Errorf("subscription %s has no product", subscription.ID)
	}
	return subscription, nil
}

// FindActivePriceById implements PaymentStore.
func (s *PostgresStripeStore) FindActivePriceById(ctx context.Context, priceId string) (*models.StripePrice, error) {
	data, err := crudrepo.StripePrice.GetOne(
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

// IsFirstSubscription implements PaymentStore.
func (s *PostgresStripeStore) IsFirstSubscription(ctx context.Context, customerID string) (bool, error) {
	data, err := crudrepo.StripeSubscription.Count(
		ctx,
		s.db,
		&map[string]any{
			models.StripeSubscriptionTable.StripeCustomerID: map[string]any{
				"_eq": customerID,
			},
		},
	)
	return data > 0, err
}

// ListPrices implements PaymentStore.
func (s *PostgresStripeStore) ListPrices(ctx context.Context, input *shared.StripePriceListParams) ([]*models.StripePrice, error) {
	var dbx database.Dbx = s.db
	filter := input.StripePriceListFilter
	pageInput := &input.PaginatedInput
	limit, offset := database.PaginateRepo(pageInput)
	param := listPriceFilterFuncMap(&filter)
	sort := listPriceOrderByMap(input)
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

type CountOutput struct {
	Count int64
}

func (s *PostgresStripeStore) CountProducts(ctx context.Context, filter *shared.StripeProductListFilter) (int64, error) {
	q := squirrel.Select("COUNT(stripe_products.*)").
		From("stripe_products")

	q = listProductFilterFuncQuery(q, filter)
	data, err := database.QueryWithBuilder[CountOutput](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))

	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil
	}

	return data[0].Count, nil
}

// UpsertSubscriptionFromStripe implements PaymentStore.
func (s *PostgresStripeStore) UpsertSubscriptionFromStripe(ctx context.Context, sub *stripe.Subscription) error {
	if sub == nil {
		return nil
	}
	var item *stripe.SubscriptionItem
	var customer *stripe.Customer = sub.Customer
	if len(sub.Items.Data) > 0 {
		item = sub.Items.Data[0]
	}
	if item == nil || item.Price == nil {
		return errors.New("price not found")
	}
	if customer == nil {
		return errors.New("customer not found")
	}
	status := models.StripeSubscriptionStatus(sub.Status)
	err := s.UpsertSubscription(
		ctx,
		&models.StripeSubscription{
			ID:                 sub.ID,
			StripeCustomerID:   customer.ID,
			Status:             models.StripeSubscriptionStatus(status),
			Metadata:           sub.Metadata,
			ItemID:             item.ID,
			PriceID:            item.Price.ID,
			Quantity:           item.Quantity,
			CancelAtPeriodEnd:  sub.CancelAtPeriodEnd,
			Created:            utils.Int64ToISODate(sub.Created),
			CurrentPeriodStart: utils.Int64ToISODate(item.CurrentPeriodStart),
			CurrentPeriodEnd:   utils.Int64ToISODate(item.CurrentPeriodEnd),
			EndedAt:            types.Pointer(utils.Int64ToISODate(sub.EndedAt)),
			CancelAt:           types.Pointer(utils.Int64ToISODate(sub.CancelAt)),
			CanceledAt:         types.Pointer(utils.Int64ToISODate(sub.CanceledAt)),
			TrialStart:         types.Pointer(utils.Int64ToISODate(sub.TrialStart)),
			TrialEnd:           types.Pointer(utils.Int64ToISODate(sub.TrialEnd)),
		},
	)
	return err
}

func (s *PostgresStripeStore) UpsertSubscription(ctx context.Context, sub *models.StripeSubscription) error {
	q := squirrel.Insert("stripe_subscriptions").
		Columns(
			"id",
			"stripe_customer_id",
			"status",
			"metadata",
			"item_id",
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
		).Values(
		sub.ID,
		sub.StripeCustomerID,
		sub.Status,
		sub.Metadata,
		sub.ItemID,
		sub.PriceID,
		sub.Quantity,
		sub.CancelAtPeriodEnd,
		sub.Created,
		sub.CurrentPeriodStart,
		sub.CurrentPeriodEnd,
		sub.EndedAt,
		sub.CancelAt,
		sub.CanceledAt,
		sub.TrialStart,
		sub.TrialEnd,
	).Suffix("ON CONFLICT (id) DO UPDATE SET " +
		"stripe_customer_id = EXCLUDED.stripe_customer_id," +
		"status = EXCLUDED.status," +
		"metadata = EXCLUDED.metadata," +
		"item_id = EXCLUDED.item_id," +
		"price_id = EXCLUDED.price_id," +
		"quantity = EXCLUDED.quantity," +
		"cancel_at_period_end = EXCLUDED.cancel_at_period_end," +
		"created = EXCLUDED.created," +
		"current_period_start = EXCLUDED.current_period_start," +
		"current_period_end = EXCLUDED.current_period_end," +
		"ended_at = EXCLUDED.ended_at," +
		"cancel_at = EXCLUDED.cancel_at," +
		"canceled_at = EXCLUDED.canceled_at," +
		"trial_start = EXCLUDED.trial_start," +
		"trial_end = EXCLUDED.trial_end")
	return database.ExecWithBuilder(ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
}

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

func (s *PostgresStripeStore) ListProducts(ctx context.Context, input *shared.StripeProductListParams) ([]*models.StripeProduct, error) {
	q := squirrel.Select("stripe_products.*").
		From("stripe_products")
	filter := input.StripeProductListFilter
	pageInput := &input.PaginatedInput

	q = database.Paginate(q, pageInput)
	q = listProductFilterFuncQuery(q, &filter)
	q = listProductOrderByQuery(q, input)
	data, err := database.QueryWithBuilder[*models.StripeProduct](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return data, nil
}

const (
	getProductRolesQuery = `
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
                $1::text []
        )
GROUP BY rp.product_id;`
)

func (s *PostgresStripeStore) LoadProductRoles(ctx context.Context, productIds ...string) ([][]*models.Role, error) {
	data, err := database.QueryAll[shared.JoinedResult[*models.Role, string]](
		ctx,
		s.db,
		getProductRolesQuery,
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

func listProductOrderByQuery(q squirrel.SelectBuilder, input *shared.StripeProductListParams) squirrel.SelectBuilder {
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

func listProductFilterFuncQuery(q squirrel.SelectBuilder, filter *shared.StripeProductListFilter) squirrel.SelectBuilder {
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

func (s *PostgresStripeStore) CountPrices(ctx context.Context, filter *shared.StripePriceListFilter) (int64, error) {
	filermap := listPriceFilterFuncMap(filter)
	data, err := crudrepo.StripePrice.Count(ctx, s.db, filermap)
	if err != nil {
		return 0, err
	}
	return data, nil
}

func (s *PostgresStripeStore) ListCustomers(ctx context.Context, input *shared.StripeCustomerListParams) ([]*models.StripeCustomer, error) {

	filter := input.StripeCustomerListFilter
	pageInput := &input.PaginatedInput

	limit, offset := database.PaginateRepo(pageInput)
	where := listCustomerFilterFunc(&filter)
	order := stripeCustomerOrderByFunc(input)
	data, err := crudrepo.StripeCustomer.Get(
		ctx,
		s.db,
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

func stripeCustomerOrderByFunc(input *shared.StripeCustomerListParams) *map[string]string {
	if input == nil {
		return nil
	}
	order := make(map[string]string)
	if slices.Contains(StripeCustomerColumnNames, input.SortBy) {
		order[input.SortBy] = strings.ToUpper(input.SortOrder)
	}
	return &order
}

func listCustomerFilterFunc(filter *shared.StripeCustomerListFilter) *map[string]any {
	if filter == nil {
		return nil
	}
	where := map[string]any{}
	if len(filter.Ids) > 0 {
		where[models.StripeCustomerTable.ID] = map[string]any{
			"_in": filter.Ids,
		}
	}
	return &where
}

func (s *PostgresStripeStore) CountCustomers(ctx context.Context, filter *shared.StripeCustomerListFilter) (int64, error) {
	where := listCustomerFilterFunc(filter)
	data, err := crudrepo.StripeCustomer.Count(ctx, s.db, where)
	if err != nil {
		return 0, err
	}
	return data, nil
}

func (s *PostgresStripeStore) ListSubscriptions(ctx context.Context, input *shared.StripeSubscriptionListParams) ([]*models.StripeSubscription, error) {

	filter := input.StripeSubscriptionListFilter
	pageInput := &input.PaginatedInput

	limit, offset := database.PaginateRepo(pageInput)
	where := listSubscriptionFilterFunc(&filter)
	order := listSubscriptionOrderByFunc(input)
	data, err := crudrepo.StripeSubscription.Get(
		ctx,
		s.db,
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

func listSubscriptionOrderByFunc(input *shared.StripeSubscriptionListParams) *map[string]string {
	if input == nil {
		return nil
	}
	order := make(map[string]string)
	if slices.Contains(StripeSubscriptionColumnNames, input.SortBy) {
		order[input.SortBy] = strings.ToUpper(input.SortOrder)
	}
	return &order
}

func listSubscriptionFilterFunc(filter *shared.StripeSubscriptionListFilter) *map[string]any {
	if filter == nil {
		return nil
	}
	where := map[string]any{}
	if len(filter.Ids) > 0 {
		where[models.StripeSubscriptionTable.ID] = map[string]any{
			"_in": filter.Ids,
		}
	}
	if len(filter.Status) > 0 {
		statuses := mapper.Map(filter.Status, func(s shared.StripeSubscriptionStatus) string { return string(s) })
		where[models.StripeSubscriptionTable.Status] = map[string]any{
			"_in": statuses,
		}
	}
	if len(filter.UserIDs) > 0 {
		where[models.StripeSubscriptionTable.StripeCustomer] = map[string]any{
			models.StripeCustomerTable.UserID: map[string]any{
				"_eq": filter.UserIDs,
			},
		}
	}
	return &where
}

func (s *PostgresStripeStore) CountSubscriptions(ctx context.Context, filter *shared.StripeSubscriptionListFilter) (int64, error) {
	where := listSubscriptionFilterFunc(filter)
	data, err := crudrepo.StripeSubscription.Count(ctx, s.db, where)
	if err != nil {
		return 0, err
	}
	return data, nil
}
