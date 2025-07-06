package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/tools/types"
)

type DbConstraintStore struct {
	db database.Dbx
}

func (p *DbConstraintStore) WithTx(tx database.Dbx) *DbConstraintStore {
	return &DbConstraintStore{
		db: tx,
	}
}

// FindLatestActiveSubscriptionByTeamId implements services.ConstaintCheckerStore.
func (p *DbConstraintStore) FindLatestActiveSubscriptionByTeamId(ctx context.Context, teamId uuid.UUID) (*models.StripeSubscription, error) {
	subscriptions, err := repository.StripeSubscription.Get(
		ctx,
		p.db,
		&map[string]any{
			models.StripeSubscriptionTable.StripeCustomer: map[string]any{
				models.StripeCustomerTable.TeamID: map[string]any{
					"_eq": teamId,
				},
			},
			models.StripeSubscriptionTable.Status: map[string]any{"_in": []string{
				string(models.StripeSubscriptionStatusActive),
				string(models.StripeSubscriptionStatusTrialing),
			}},
		},
		&map[string]string{
			models.StripeSubscriptionTable.CreatedAt: "desc",
		},
		types.Pointer(1),
		nil,
	)
	if err != nil {
		return nil, err
	}
	if len(subscriptions) == 0 {
		return nil, nil
	}
	return subscriptions[0], nil
}

// FindUserByEmail implements services.ConstaintCheckerStore.
func (p *DbConstraintStore) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return repository.User.GetOne(
		ctx,
		p.db,
		&map[string]any{
			"email": map[string]any{
				"_eq": email,
			},
		},
	)
}

// FindLatestActiveSubscriptionByUserId implements services.ConstaintCheckerStore.
func (p *DbConstraintStore) FindLatestActiveSubscriptionByUserId(ctx context.Context, userId uuid.UUID) (*models.StripeSubscription, error) {
	subscriptions, err := repository.StripeSubscription.Get(
		ctx,
		p.db,
		&map[string]any{
			models.StripeSubscriptionTable.StripeCustomer: map[string]any{
				models.StripeCustomerTable.UserID: map[string]any{
					"_eq": userId,
				},
			},
			models.StripeSubscriptionTable.Status: map[string]any{
				"_in": []string{
					string(models.StripeSubscriptionStatusActive),
					string(models.StripeSubscriptionStatusTrialing),
				},
			},
		},
		&map[string]string{
			models.StripeSubscriptionTable.CreatedAt: "desc",
		},
		types.Pointer(1),
		nil,
	)
	if err != nil {
		return nil, err
	}
	if len(subscriptions) == 0 {
		return nil, nil
	}
	return subscriptions[0], nil
}

// FindUserByID implements services.ConstaintCheckerStore.
func (p *DbConstraintStore) FindUserById(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	return repository.User.GetOne(
		ctx,
		p.db,
		&map[string]any{
			models.UserTable.ID: map[string]any{
				"_eq": userId,
			},
		},
	)
}

func NewDbConstraintStore(db database.Dbx) *DbConstraintStore {
	return &DbConstraintStore{
		db: db,
	}
}
