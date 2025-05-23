package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/tools/types"
)

type PostgresConstraintStore struct {
	db database.Dbx
}

// FindLatestActiveSubscriptionByTeamId implements services.ConstaintCheckerStore.
func (p *PostgresConstraintStore) FindLatestActiveSubscriptionByTeamId(ctx context.Context, teamId uuid.UUID) (*models.StripeSubscription, error) {
	subscriptions, err := crudrepo.StripeSubscription.Get(
		ctx,
		p.db,
		&map[string]any{
			"stripe_customer": map[string]any{
				"team_id": map[string]any{
					"_eq": teamId.String(),
				},
			},
			"status": map[string]any{"_in": []string{
				string(models.StripeSubscriptionStatusActive),
				string(models.StripeSubscriptionStatusTrialing),
			}},
		},
		&map[string]string{
			"created_at": "desc",
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
func (p *PostgresConstraintStore) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return crudrepo.User.GetOne(
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
func (p *PostgresConstraintStore) FindLatestActiveSubscriptionByUserId(ctx context.Context, userId uuid.UUID) (*models.StripeSubscription, error) {
	subscriptions, err := crudrepo.StripeSubscription.Get(
		ctx,
		p.db,
		&map[string]any{
			"stripe_customer": map[string]any{
				"user_id": map[string]any{
					"_eq": userId.String(),
				},
			},
			"status": map[string]any{"_in": []string{
				string(models.StripeSubscriptionStatusActive),
				string(models.StripeSubscriptionStatusTrialing),
			}},
		},
		&map[string]string{
			"created_at": "desc",
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

// FindUserById implements services.ConstaintCheckerStore.
func (p *PostgresConstraintStore) FindUserById(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	return crudrepo.User.GetOne(
		ctx,
		p.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": userId.String(),
			},
		},
	)
}

func NewPostgresConstraintStore(db database.Dbx) *PostgresConstraintStore {
	return &PostgresConstraintStore{
		db: db,
	}
}

var _ services.ConstaintCheckerStore = (*PostgresConstraintStore)(nil)
