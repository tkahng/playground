package repository

import (
	"context"

	"github.com/aarondl/opt/omit"
	"github.com/google/uuid"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/tkahng/authgo/internal/db/models"
)

func FindCustomerByUserId(ctx context.Context, dbx bob.Executor, userId uuid.UUID) (*models.StripeCustomer, error) {
	data, err := models.StripeCustomers.Query(
		models.SelectWhere.StripeCustomers.ID.EQ(userId),
	).One(ctx, dbx)
	return OptionalRow(data, err)
}

func UpsertCustomer(ctx context.Context, dbx bob.Executor, userId uuid.UUID, stripeCustomerId string) error {
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
			im.SetCol("description").To(
				psql.Raw("EXCLUDED.description"),
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
