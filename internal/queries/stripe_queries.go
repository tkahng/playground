package queries

import (
	"context"
	"errors"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/types"
)

func FindCustomerByStripeId(ctx context.Context, dbx Queryer, stripeId string) (*models.StripeCustomer, error) {
	data, err := repository.StripeCustomer.GetOne(
		ctx,
		dbx,
		&map[string]any{
			"stripe_id": map[string]any{
				"_eq": stripeId,
			},
		},
	)
	// data, err := crudModels.StripeCustomers.Query(
	// 	models.SelectWhere.StripeCustomers.StripeID.EQ(stripeId),
	// ).One(ctx, dbx)
	return OptionalRow(data, err)
}

func FindCustomerByUserId(ctx context.Context, dbx Queryer, userId uuid.UUID) (*models.StripeCustomer, error) {
	data, err := repository.StripeCustomer.GetOne(
		ctx,
		dbx,
		&map[string]any{
			"id": map[string]any{
				"_eq": userId.String(),
			},
		},
	)
	return OptionalRow(data, err)
}

func FindProductByStripeId(ctx context.Context, dbx Queryer, stripeId string) (*models.StripeProduct, error) {
	// data, err := models.StripeProducts.Query(
	// 	models.SelectWhere.StripeProducts.ID.EQ(stripeId),
	// ).One(ctx, dbx)
	data, err := repository.StripeProduct.GetOne(
		ctx,
		dbx,
		&map[string]any{
			"id": map[string]any{
				"_eq": stripeId,
			},
		},
	)
	return OptionalRow(data, err)
}

func UpsertCustomerStripeId(ctx context.Context, dbx Queryer, userId uuid.UUID, stripeCustomerId string) error {
	// _, err := models.StripeCustomers.Insert(
	// 	&models.StripeCustomerSetter{
	// 		ID:       omit.From(userId),
	// 		StripeID: omit.From(stripeCustomerId),
	// 	},
	// 	im.OnConflict("id").DoUpdate(
	// 		im.SetCol("stripe_id").To(
	// 			psql.Raw("EXCLUDED.stripe_id"),
	// 		),
	// 	),
	// ).Exec(ctx, dbx)
	q := squirrel.
		Insert("stripe_customers").
		Columns("id", "stripe_id").
		Values(userId, stripeCustomerId).
		Suffix(`ON CONFLICT (id) DO UPDATE SET stripe_id = EXCLUDED.stripe_id`)
	return ExecWithBuilder(ctx, dbx, q.PlaceholderFormat(squirrel.Dollar))
}

func UpsertProduct(ctx context.Context, dbx Queryer, product *models.StripeProduct) error {
	// _, err := models.StripeProducts.Insert(
	// 	product,
	// 	im.OnConflict("id").DoUpdate(
	// 		im.SetCol("active").To(
	// 			psql.Raw("EXCLUDED.active"),
	// 		),
	// 		im.SetCol("name").To(
	// 			psql.Raw("EXCLUDED.name"),
	// 		),
	// 		im.SetCol("description").To(
	// 			psql.Raw("EXCLUDED.description"),
	// 		),
	// 		im.SetCol("image").To(
	// 			psql.Raw("EXCLUDED.image"),
	// 		),
	// 		im.SetCol("metadata").To(
	// 			psql.Raw("EXCLUDED.metadata"),
	// 		),
	// 	),
	// ).Exec(ctx, dbx)
	q := squirrel.
		Insert("stripe_products").
		Columns("id", "active", "name", "description", "image", "metadata").
		Values(product.ID, product.Active, product.Name, product.Description, product.Image, product.Metadata).
		Suffix(`ON CONFLICT (id) DO UPDATE SET active = EXCLUDED.active, name = EXCLUDED.name, description = EXCLUDED.description, image = EXCLUDED.image, metadata = EXCLUDED.metadata`)
	return ExecWithBuilder(ctx, dbx, q.PlaceholderFormat(squirrel.Dollar))
}

func UpsertProductFromStripe(ctx context.Context, dbx Queryer, product *stripe.Product) error {
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
	return UpsertProduct(ctx, dbx, param)
}

func UpsertPrice(ctx context.Context, dbx Queryer, price *models.StripePrice) error {
	// _, err := models.StripePrices.Insert(
	// 	price,
	// 	im.OnConflict("id").DoUpdate(
	// 		im.SetCol("product_id").To(
	// 			psql.Raw("EXCLUDED.product_id"),
	// 		),
	// 		im.SetCol("lookup_key").To(
	// 			psql.Raw("EXCLUDED.lookup_key"),
	// 		),
	// 		im.SetCol("active").To(
	// 			psql.Raw("EXCLUDED.active"),
	// 		),
	// 		im.SetCol("unit_amount").To(
	// 			psql.Raw("EXCLUDED.unit_amount"),
	// 		),
	// 		im.SetCol("currency").To(
	// 			psql.Raw("EXCLUDED.currency"),
	// 		),
	// 		im.SetCol("type").To(
	// 			psql.Raw("EXCLUDED.type"),
	// 		),
	// 		im.SetCol("interval").To(
	// 			psql.Raw("EXCLUDED.interval"),
	// 		),
	// 		im.SetCol("interval_count").To(
	// 			psql.Raw("EXCLUDED.interval_count"),
	// 		),
	// 		im.SetCol("trial_period_days").To(
	// 			psql.Raw("EXCLUDED.trial_period_days"),
	// 		),
	// 		im.SetCol("metadata").To(
	// 			psql.Raw("EXCLUDED.metadata"),
	// 		),
	// 	),
	// ).Exec(ctx, dbx)
	q := squirrel.
		Insert("stripe_prices").
		Columns("id", "product_id", "lookup_key", "active", "unit_amount", "currency", "type", "interval", "interval_count", "trial_period_days", "metadata").
		Values(price.ID, price.ProductID, price.LookupKey, price.Active, price.UnitAmount, price.Currency, price.Type, price.Interval, price.IntervalCount, price.TrialPeriodDays, price.Metadata).
		Suffix(`ON CONFLICT(id) DO UPDATE SET product_id = EXCLUDED.product_id,
        lookup_key = EXCLUDED.lookup_key,
        active = EXCLUDED.active,
        unit_amount = EXCLUDED.unit_amount,
        currency = EXCLUDED.currency,
        type = EXCLUDED.type,
        interval = EXCLUDED.interval,
        interval_count = EXCLUDED.interval_count,
        trial_period_days = EXCLUDED.trial_period_days,
        metadata = EXCLUDED.metadata`)
	return ExecWithBuilder(ctx, dbx, q.PlaceholderFormat(squirrel.Dollar))
}

func UpsertPriceFromStripe(ctx context.Context, dbx Queryer, price *stripe.Price) error {
	if price == nil {
		return nil
	}
	// param := &models.StripePriceSetter{
	// 	ID:         omit.From(price.ID),
	// 	ProductID:  omit.From(price.Product.ID),
	// 	Active:     omit.From(price.Active),
	// 	LookupKey:  omitnull.From(price.LookupKey),
	// 	UnitAmount: omitnull.From(price.UnitAmount),
	// 	Currency:   omit.From(string(price.Currency)),
	// 	Type:       omit.From(PriceTypeConvert(price.Type)),
	// 	Metadata:   omit.From(types.NewJSON(price.Metadata)),
	// }
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
		*val.Interval = models.StripePricingPlanInterval(price.Recurring.Interval)
		*val.IntervalCount = price.Recurring.IntervalCount
		*val.TrialPeriodDays = price.Recurring.TrialPeriodDays
	}
	return UpsertPrice(ctx, dbx, val)
}

func UpsertSubscription(ctx context.Context, dbx Queryer, subscription *models.StripeSubscription) error {
	// _, err := models.StripeSubscriptions.Insert(
	// 	subscription,
	// 	im.OnConflict("id").DoUpdate(
	// im.SetCol("user_id").To(
	// 	psql.Raw("EXCLUDED.user_id"),
	// ),
	// im.SetCol("status").To(
	// 	psql.Raw("EXCLUDED.status"),
	// ),
	// im.SetCol("metadata").To(
	// 	psql.Raw("EXCLUDED.metadata"),
	// ),
	// im.SetCol("price_id").To(
	// 	psql.Raw("EXCLUDED.price_id"),
	// ),
	// im.SetCol("quantity").To(
	// 	psql.Raw("EXCLUDED.quantity"),
	// ),
	// im.SetCol("cancel_at_period_end").To(
	// 	psql.Raw("EXCLUDED.cancel_at_period_end"),
	// ),
	// im.SetCol("created").To(
	// 	psql.Raw("EXCLUDED.created"),
	// ),
	// im.SetCol("current_period_start").To(
	// 	psql.Raw("EXCLUDED.current_period_start"),
	// ),
	// im.SetCol("current_period_end").To(
	// 	psql.Raw("EXCLUDED.current_period_end"),
	// ),
	// im.SetCol("ended_at").To(
	// 	psql.Raw("EXCLUDED.ended_at"),
	// ),
	// im.SetCol("cancel_at").To(
	// 	psql.Raw("EXCLUDED.cancel_at"),
	// ),
	// im.SetCol("canceled_at").To(
	// 	psql.Raw("EXCLUDED.canceled_at"),
	// ),
	// im.SetCol("trial_start").To(
	// 	psql.Raw("EXCLUDED.trial_start"),
	// ),
	// im.SetCol("trial_end").To(
	// 	psql.Raw("EXCLUDED.trial_end"),
	// ),
	// 	),
	// ).Exec(ctx, dbx)
	q := squirrel.
		Insert("stripe_subscriptions").
		Columns(
			"id",
			"user_id",
			"status",
			"metadata",
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
		subscription.ID,
		subscription.UserID,
		subscription.Status,
		subscription.Metadata,
		subscription.PriceID,
		subscription.Quantity,
		subscription.CancelAtPeriodEnd,
		subscription.Created,
		subscription.CurrentPeriodStart,
		subscription.CurrentPeriodEnd,
		subscription.EndedAt,
		subscription.CancelAt,
		subscription.CanceledAt,
		subscription.TrialStart,
		subscription.TrialEnd,
	).Suffix(
		"ON CONFLICT (id) DO UPDATE SET " +
			"user_id = EXCLUDED.user_id," +
			"status = EXCLUDED.status," +
			"metadata = EXCLUDED.metadata," +
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
			"trial_end = EXCLUDED.trial_end",
	)

	return ExecWithBuilder(ctx, dbx, q.PlaceholderFormat(squirrel.Dollar))
}

func UpsertSubscriptionFromStripe(ctx context.Context, exec Queryer, sub *stripe.Subscription, userId uuid.UUID) error {
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
	err := UpsertSubscription(ctx, exec, &models.StripeSubscription{
		// ID:                 omit.From(sub.ID),
		// UserID:             omit.From(userId),
		// Status:             omit.From(status),
		// Metadata:           omit.From(types.NewJSON(sub.Metadata)),
		// PriceID:            omit.From(item.Price.ID),
		// Quantity:           omit.From(item.Quantity),
		// CancelAtPeriodEnd:  omit.From(sub.CancelAtPeriodEnd),
		// Created:            omit.From(Int64ToISODate(sub.Created)),
		// CurrentPeriodStart: omit.From(Int64ToISODate(item.CurrentPeriodStart)),
		// CurrentPeriodEnd:   omit.From(Int64ToISODate(item.CurrentPeriodEnd)),
		// EndedAt:            omitnull.From(Int64ToISODate(sub.EndedAt)),
		// CancelAt:           omitnull.From(Int64ToISODate(sub.CancelAt)),
		// CanceledAt:         omitnull.From(Int64ToISODate(sub.CanceledAt)),
		// TrialStart:         omitnull.From(Int64ToISODate(sub.TrialStart)),
		// TrialEnd:           omitnull.From(Int64ToISODate(sub.TrialEnd)),
		ID:                 sub.ID,
		UserID:             userId,
		Status:             models.StripeSubscriptionStatus(status),
		Metadata:           sub.Metadata,
		PriceID:            item.Price.ID,
		Quantity:           item.Quantity,
		CancelAtPeriodEnd:  sub.CancelAtPeriodEnd,
		Created:            Int64ToISODate(sub.Created),
		CurrentPeriodStart: Int64ToISODate(item.CurrentPeriodStart),
		CurrentPeriodEnd:   Int64ToISODate(item.CurrentPeriodEnd),
		EndedAt:            types.Pointer(Int64ToISODate(sub.EndedAt)),
		CancelAt:           types.Pointer(Int64ToISODate(sub.CancelAt)),
		CanceledAt:         types.Pointer(Int64ToISODate(sub.CanceledAt)),
		TrialStart:         types.Pointer(Int64ToISODate(sub.TrialStart)),
		TrialEnd:           types.Pointer(Int64ToISODate(sub.TrialEnd)),
	})
	return err
}

func Int64ToISODate(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

func FindSubscriptionById(ctx context.Context, dbx Queryer, stripeId string) (*models.StripeSubscription, error) {
	// data, err := models.StripeSubscriptions.Query(
	// 	models.SelectWhere.StripeSubscriptions.ID.EQ(stripeId),
	// ).One(ctx, dbx)
	data, err := repository.StripeSubscription.GetOne(
		ctx,
		dbx,
		&map[string]any{
			"id": map[string]any{
				"_eq": stripeId,
			},
		},
	)
	return OptionalRow(data, err)
}

const (
	GetSubscriptionWithPriceByIdQuery = `
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

func FindSubscriptionWithPriceById(ctx context.Context, dbx Queryer, stripeId string) (*models.SubscriptionWithPrice, error) {
	// data, err := models.StripeSubscriptions.Query(
	// 	models.SelectWhere.StripeSubscriptions.ID.EQ(stripeId),
	// 	models.PreloadStripeSubscriptionPriceStripePrice(
	// 		models.PreloadStripePriceProductStripeProduct(),
	// 	),
	// ).One(ctx, dbx)
	data, err := QueryAll[*models.SubscriptionWithPrice](ctx, dbx, GetSubscriptionWithPriceByIdQuery, stripeId)
	if err != nil {
		return nil, err
	}
	first := ReturnFirst(data)
	if first == nil {
		return nil, nil
	}
	return first, nil
}

func FindLatestActiveSubscriptionByUserId(ctx context.Context, dbx Queryer, userId uuid.UUID) (*models.StripeSubscription, error) {
	// data, err := models.StripeSubscriptions.Query(
	// 	models.SelectWhere.StripeSubscriptions.UserID.EQ(userId),
	// 	models.SelectWhere.StripeSubscriptions.Status.In(
	// 		models.StripeSubscriptionStatusActive,
	// 		models.StripeSubscriptionStatusTrialing,
	// 	),
	// 	sm.OrderBy(models.StripeSubscriptionColumns.Created).Desc(),
	// ).One(ctx, dbx)
	data, err := repository.StripeSubscription.Get(
		ctx,
		dbx,
		&map[string]any{
			"user_id": map[string]any{
				"_eq": userId.String(),
			},
			"status": map[string]any{
				"_in": []string{
					string(models.StripeSubscriptionStatusActive),
					string(models.StripeSubscriptionStatusTrialing),
				},
			},
		},
		&map[string]string{
			"created_at": "DESC",
		},
		types.Pointer(1),
		nil,
	)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return OptionalRow(data[0], err)
}

const (
	GetLatestActiveSubscriptionWithPriceByIdQuery = `
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
WHERE ss.user_id = $1
        AND ss.status IN ('active', 'trialing')
ORDER BY ss.updated_at DESC;
		`
)

func FindLatestActiveSubscriptionWithPriceByUserId(ctx context.Context, dbx Queryer, userId uuid.UUID) (*models.SubscriptionWithPrice, error) {
	data, err := QueryAll[models.SubscriptionWithPrice](ctx, dbx, GetLatestActiveSubscriptionWithPriceByIdQuery, userId)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return OptionalRow(&data[0], err)
}

func IsFirstSubscription(ctx context.Context, dbx Queryer, userId uuid.UUID) (bool, error) {
	// data, err := models.StripeSubscriptions.Query(
	// 	models.SelectWhere.StripeSubscriptions.UserID.EQ(userId),
	// ).Exists(ctx, dbx)
	data, err := repository.StripeSubscription.Count(
		ctx,
		dbx,
		&map[string]any{
			"user_id": map[string]any{
				"_eq": userId.String(),
			},
		},
	)
	return data > 0, err
	// return OptionalRow(data, err)
}

func FindValidPriceById(ctx context.Context, dbx Queryer, priceId string) (*models.StripePrice, error) {
	// data, err := models.StripePrices.Query(
	// 	models.SelectWhere.StripePrices.ID.EQ(priceId),
	// 	models.SelectWhere.StripePrices.Type.EQ(models.StripePricingTypeRecurring),
	// ).One(ctx, dbx)
	data, err := repository.StripePrice.GetOne(
		ctx,
		dbx,
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
