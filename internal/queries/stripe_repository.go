package queries

import (
	"context"
	"errors"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/google/uuid"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/types"
	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/authgo/internal/db/models"
)

func FindCustomerByStripeId(ctx context.Context, dbx bob.Executor, stripeId string) (*models.StripeCustomer, error) {
	data, err := models.StripeCustomers.Query(
		models.SelectWhere.StripeCustomers.StripeID.EQ(stripeId),
	).One(ctx, dbx)
	return OptionalRow(data, err)
}

func FindCustomerByUserId(ctx context.Context, dbx bob.Executor, userId uuid.UUID) (*models.StripeCustomer, error) {
	data, err := models.StripeCustomers.Query(
		models.SelectWhere.StripeCustomers.ID.EQ(userId),
	).One(ctx, dbx)
	return OptionalRow(data, err)
}

func FindProductByStripeId(ctx context.Context, dbx bob.Executor, stripeId string) (*models.StripeProduct, error) {
	data, err := models.StripeProducts.Query(
		models.SelectWhere.StripeProducts.ID.EQ(stripeId),
	).One(ctx, dbx)
	return OptionalRow(data, err)
}

func UpsertCustomerStripeId(ctx context.Context, dbx bob.Executor, userId uuid.UUID, stripeCustomerId string) error {
	_, err := models.StripeCustomers.Insert(
		&models.StripeCustomerSetter{
			ID:       omit.From(userId),
			StripeID: omit.From(stripeCustomerId),
		},
		im.OnConflict("id").DoUpdate(
			im.SetCol("stripe_id").To(
				psql.Raw("EXCLUDED.stripe_id"),
			),
		),
	).Exec(ctx, dbx)
	return err
}

func UpsertProduct(ctx context.Context, dbx bob.Executor, product *models.StripeProductSetter) error {
	_, err := models.StripeProducts.Insert(
		product,
		im.OnConflict("id").DoUpdate(
			im.SetCol("active").To(
				psql.Raw("EXCLUDED.active"),
			),
			im.SetCol("name").To(
				psql.Raw("EXCLUDED.name"),
			),
			im.SetCol("description").To(
				psql.Raw("EXCLUDED.description"),
			),
			im.SetCol("image").To(
				psql.Raw("EXCLUDED.image"),
			),
			im.SetCol("metadata").To(
				psql.Raw("EXCLUDED.metadata"),
			),
		),
	).Exec(ctx, dbx)
	return err
}

func UpsertProductFromStripe(ctx context.Context, dbx bob.Executor, product *stripe.Product) error {
	if product == nil {
		return nil
	}
	var image *string
	if len(product.Images) > 0 {
		image = &product.Images[0]
	}
	param := &models.StripeProductSetter{
		ID:          omit.From(product.ID),
		Active:      omit.From(product.Active),
		Name:        omit.From(product.Name),
		Description: omitnull.From(product.Description),
		Image:       omitnull.FromPtr(image),
		Metadata:    omit.From(types.NewJSON(product.Metadata)),
	}
	if len(product.Images) > 0 {
		param.Image = omitnull.From(product.Images[0])
	}
	return UpsertProduct(ctx, dbx, param)
}

func UpsertPrice(ctx context.Context, dbx bob.Executor, price *models.StripePriceSetter) error {
	_, err := models.StripePrices.Insert(
		price,
		im.OnConflict("id").DoUpdate(
			im.SetCol("product_id").To(
				psql.Raw("EXCLUDED.product_id"),
			),
			im.SetCol("lookup_key").To(
				psql.Raw("EXCLUDED.lookup_key"),
			),
			im.SetCol("active").To(
				psql.Raw("EXCLUDED.active"),
			),
			im.SetCol("unit_amount").To(
				psql.Raw("EXCLUDED.unit_amount"),
			),
			im.SetCol("currency").To(
				psql.Raw("EXCLUDED.currency"),
			),
			im.SetCol("type").To(
				psql.Raw("EXCLUDED.type"),
			),
			im.SetCol("interval").To(
				psql.Raw("EXCLUDED.interval"),
			),
			im.SetCol("interval_count").To(
				psql.Raw("EXCLUDED.interval_count"),
			),
			im.SetCol("trial_period_days").To(
				psql.Raw("EXCLUDED.trial_period_days"),
			),
			im.SetCol("metadata").To(
				psql.Raw("EXCLUDED.metadata"),
			),
		),
	).Exec(ctx, dbx)
	return err
}

func UpsertPriceFromStripe(ctx context.Context, dbx bob.Executor, price *stripe.Price) error {
	if price == nil {
		return nil
	}
	param := &models.StripePriceSetter{
		ID:         omit.From(price.ID),
		ProductID:  omit.From(price.Product.ID),
		Active:     omit.From(price.Active),
		LookupKey:  omitnull.From(price.LookupKey),
		UnitAmount: omitnull.From(price.UnitAmount),
		Currency:   omit.From(string(price.Currency)),
		Type:       omit.From(PriceTypeConvert(price.Type)),
		Metadata:   omit.From(types.NewJSON(price.Metadata)),
	}
	if price.Recurring != nil {
		param.Interval = omitnull.From(PriceIntervalConvert(price.Recurring.Interval))
		param.IntervalCount = omitnull.From(price.Recurring.IntervalCount)
		param.TrialPeriodDays = omitnull.From(price.Recurring.TrialPeriodDays)
	}
	return UpsertPrice(ctx, dbx, param)
}

func PriceIntervalConvert(priceRecurringInterval stripe.PriceRecurringInterval) models.StripePricingPlanInterval {
	switch priceRecurringInterval {
	case stripe.PriceRecurringIntervalMonth:
		return models.StripePricingPlanIntervalMonth
	case stripe.PriceRecurringIntervalYear:
		return models.StripePricingPlanIntervalYear
	case stripe.PriceRecurringIntervalWeek:
		return models.StripePricingPlanIntervalWeek
	case stripe.PriceRecurringIntervalDay:
		return models.StripePricingPlanIntervalDay
	default:
		return models.StripePricingPlanIntervalMonth
	}
}

func PriceTypeConvert(priceType stripe.PriceType) models.StripePricingType {
	switch priceType {
	case stripe.PriceTypeOneTime:
		return models.StripePricingTypeOneTime
	case stripe.PriceTypeRecurring:
		return models.StripePricingTypeRecurring
	default:
		return models.StripePricingTypeRecurring
	}
}

func UpsertSubscription(ctx context.Context, dbx bob.Executor, subscription *models.StripeSubscriptionSetter) error {
	_, err := models.StripeSubscriptions.Insert(
		subscription,
		im.OnConflict("id").DoUpdate(
			im.SetCol("user_id").To(
				psql.Raw("EXCLUDED.user_id"),
			),
			im.SetCol("status").To(
				psql.Raw("EXCLUDED.status"),
			),
			im.SetCol("metadata").To(
				psql.Raw("EXCLUDED.metadata"),
			),
			im.SetCol("price_id").To(
				psql.Raw("EXCLUDED.price_id"),
			),
			im.SetCol("quantity").To(
				psql.Raw("EXCLUDED.quantity"),
			),
			im.SetCol("cancel_at_period_end").To(
				psql.Raw("EXCLUDED.cancel_at_period_end"),
			),
			im.SetCol("created").To(
				psql.Raw("EXCLUDED.created"),
			),
			im.SetCol("current_period_start").To(
				psql.Raw("EXCLUDED.current_period_start"),
			),
			im.SetCol("current_period_end").To(
				psql.Raw("EXCLUDED.current_period_end"),
			),
			im.SetCol("ended_at").To(
				psql.Raw("EXCLUDED.ended_at"),
			),
			im.SetCol("cancel_at").To(
				psql.Raw("EXCLUDED.cancel_at"),
			),
			im.SetCol("canceled_at").To(
				psql.Raw("EXCLUDED.canceled_at"),
			),
			im.SetCol("trial_start").To(
				psql.Raw("EXCLUDED.trial_start"),
			),
			im.SetCol("trial_end").To(
				psql.Raw("EXCLUDED.trial_end"),
			),
		),
	).Exec(ctx, dbx)
	return err
}

func UpsertSubscriptionFromStripe(ctx context.Context, exec bob.Executor, sub *stripe.Subscription, userId uuid.UUID) error {
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
	status := StripeSubscriptionStatusConvert(sub.Status)
	err := UpsertSubscription(ctx, exec, &models.StripeSubscriptionSetter{
		ID:                 omit.From(sub.ID),
		UserID:             omit.From(userId),
		Status:             omit.From(status),
		Metadata:           omit.From(types.NewJSON(sub.Metadata)),
		PriceID:            omit.From(item.Price.ID),
		Quantity:           omit.From(item.Quantity),
		CancelAtPeriodEnd:  omit.From(sub.CancelAtPeriodEnd),
		Created:            omit.From(Int64ToISODate(sub.Created)),
		CurrentPeriodStart: omit.From(Int64ToISODate(item.CurrentPeriodStart)),
		CurrentPeriodEnd:   omit.From(Int64ToISODate(item.CurrentPeriodEnd)),
		EndedAt:            omitnull.From(Int64ToISODate(sub.EndedAt)),
		CancelAt:           omitnull.From(Int64ToISODate(sub.CancelAt)),
		CanceledAt:         omitnull.From(Int64ToISODate(sub.CanceledAt)),
		TrialStart:         omitnull.From(Int64ToISODate(sub.TrialStart)),
		TrialEnd:           omitnull.From(Int64ToISODate(sub.TrialEnd)),
	})
	return err
}

func Int64ToISODate(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

func StripeSubscriptionStatusConvert(status stripe.SubscriptionStatus) models.StripeSubscriptionStatus {
	switch status {
	case stripe.SubscriptionStatusActive:
		return models.StripeSubscriptionStatusActive
	case stripe.SubscriptionStatusCanceled:
		return models.StripeSubscriptionStatusCanceled
	case stripe.SubscriptionStatusPastDue:
		return models.StripeSubscriptionStatusPastDue
	case stripe.SubscriptionStatusTrialing:
		return models.StripeSubscriptionStatusTrialing
	case stripe.SubscriptionStatusUnpaid:
		return models.StripeSubscriptionStatusUnpaid
	case stripe.SubscriptionStatusIncomplete:
		return models.StripeSubscriptionStatusIncomplete
	case stripe.SubscriptionStatusIncompleteExpired:
		return models.StripeSubscriptionStatusIncompleteExpired
	case stripe.SubscriptionStatusPaused:
		return models.StripeSubscriptionStatusPaused
	}
	return models.StripeSubscriptionStatusActive
}

func FindSubscriptionById(ctx context.Context, dbx bob.Executor, stripeId string) (*models.StripeSubscription, error) {
	data, err := models.StripeSubscriptions.Query(
		models.SelectWhere.StripeSubscriptions.ID.EQ(stripeId),
	).One(ctx, dbx)
	return OptionalRow(data, err)
}

func FindSubscriptionWithPriceById(ctx context.Context, dbx bob.Executor, stripeId string) (*models.StripeSubscription, error) {
	data, err := models.StripeSubscriptions.Query(
		models.SelectWhere.StripeSubscriptions.ID.EQ(stripeId),
		models.PreloadStripeSubscriptionPriceStripePrice(
			models.PreloadStripePriceProductStripeProduct(),
		),
	).One(ctx, dbx)
	return OptionalRow(data, err)
}

func FindLatestActiveSubscriptionByUserId(ctx context.Context, dbx bob.Executor, userId uuid.UUID) (*models.StripeSubscription, error) {
	data, err := models.StripeSubscriptions.Query(
		models.SelectWhere.StripeSubscriptions.UserID.EQ(userId),
		models.SelectWhere.StripeSubscriptions.Status.In(
			models.StripeSubscriptionStatusActive,
			models.StripeSubscriptionStatusTrialing,
		),
		sm.OrderBy(models.StripeSubscriptionColumns.Created).Desc(),
	).One(ctx, dbx)
	return OptionalRow(data, err)
}

func FindLatestActiveSubscriptionWithPriceByUserId(ctx context.Context, dbx bob.Executor, userId uuid.UUID) (*models.StripeSubscription, error) {
	data, err := models.StripeSubscriptions.Query(
		models.SelectWhere.StripeSubscriptions.UserID.EQ(userId),
		models.SelectWhere.StripeSubscriptions.Status.In(
			models.StripeSubscriptionStatusActive,
			models.StripeSubscriptionStatusTrialing,
		),
		sm.OrderBy(models.StripeSubscriptionColumns.Created).Desc(),
		models.PreloadStripeSubscriptionPriceStripePrice(
			models.PreloadStripePriceProductStripeProduct(),
		),
	).One(ctx, dbx)
	return OptionalRow(data, err)
}

func IsFirstSubscription(ctx context.Context, dbx bob.Executor, userId uuid.UUID) (bool, error) {
	data, err := models.StripeSubscriptions.Query(
		models.SelectWhere.StripeSubscriptions.UserID.EQ(userId),
	).Exists(ctx, dbx)
	return !data, err
	// return OptionalRow(data, err)
}

func FindValidPriceById(ctx context.Context, dbx bob.Executor, priceId string) (*models.StripePrice, error) {
	data, err := models.StripePrices.Query(
		models.SelectWhere.StripePrices.ID.EQ(priceId),
		models.SelectWhere.StripePrices.Type.EQ(models.StripePricingTypeRecurring),
	).One(ctx, dbx)
	return data, err
}
