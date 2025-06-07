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
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/types"
	"github.com/tkahng/authgo/internal/tools/utils"
)

type DbSubscriptionStore struct {
	db database.Dbx
}

func NewDbSubscriptionStore(db database.Dbx) *DbSubscriptionStore {
	return &DbSubscriptionStore{
		db: db,
	}
}
func (s *DbSubscriptionStore) WithTx(tx database.Dbx) *DbSubscriptionStore {
	return &DbSubscriptionStore{
		db: tx,
	}
}

func (s *DbSubscriptionStore) FindActiveSubscriptionsByCustomerIds(ctx context.Context, customerIds ...string) ([]*models.StripeSubscription, error) {
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

func (s *DbSubscriptionStore) FindActiveSubscriptionsByTeamIds(ctx context.Context, teamIds ...uuid.UUID) ([]*models.StripeSubscription, error) {
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

func (s *DbSubscriptionStore) FindActiveSubscriptionsByUserIds(ctx context.Context, userIds ...uuid.UUID) ([]*models.StripeSubscription, error) {
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

func (s *DbSubscriptionStore) FindSubscriptionsWithPriceProductByIds(ctx context.Context, subscriptionIds ...string) ([]*models.StripeSubscription, error) {
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

// UpsertSubscriptionFromStripe implements PaymentStore.
func (s *DbSubscriptionStore) UpsertSubscriptionFromStripe(ctx context.Context, sub *stripe.Subscription) error {
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

func (s *DbSubscriptionStore) UpsertSubscription(ctx context.Context, sub *models.StripeSubscription) error {
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

func (s *DbSubscriptionStore) FindActiveSubscriptionByCustomerId(ctx context.Context, customerId string) (*models.StripeSubscription, error) {
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

// IsFirstSubscription implements PaymentStore.
func (s *DbSubscriptionStore) IsFirstSubscription(ctx context.Context, customerID string) (bool, error) {
	data, err := repository.StripeSubscription.Count(
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

func (s *DbSubscriptionStore) ListSubscriptions(ctx context.Context, input *shared.StripeSubscriptionListParams) ([]*models.StripeSubscription, error) {

	filter := input.StripeSubscriptionListFilter
	pageInput := &input.PaginatedInput

	limit, offset := database.PaginateRepo(pageInput)
	where := listSubscriptionFilterFunc(&filter)
	order := listSubscriptionOrderByFunc(input)
	data, err := repository.StripeSubscription.Get(
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
	if slices.Contains(models.StripeSubscriptionTable.Columns, input.SortBy) {
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

func (s *DbSubscriptionStore) CountSubscriptions(ctx context.Context, filter *shared.StripeSubscriptionListFilter) (int64, error) {
	where := listSubscriptionFilterFunc(filter)
	data, err := repository.StripeSubscription.Count(ctx, s.db, where)
	if err != nil {
		return 0, err
	}
	return data, nil
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
