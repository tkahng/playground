package stores

import (
	"context"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
)

type PlanFeaturesFilter struct {
	PaginatedInput
	SortParams
}

type PlanFeaturesStoreInterface interface {
	FindByProductID(ctx context.Context, productID string) (*models.PlanFeatures, error)
	Upsert(ctx context.Context, pf *models.PlanFeatures) (*models.PlanFeatures, error)
	// InsertIfMissing inserts a row only when no row exists for that product yet.
	// Used during startup sync so manually configured limits are never overwritten.
	InsertIfMissing(ctx context.Context, pf *models.PlanFeatures) error
	List(ctx context.Context, filter *PlanFeaturesFilter) ([]*models.PlanFeatures, error)
	CountPlanFeatures(ctx context.Context, filter *PlanFeaturesFilter) (int64, error)
	WithTx(db database.Dbx) *DbPlanFeaturesStore
}

type DbPlanFeaturesStore struct {
	db database.Dbx
}

var _ PlanFeaturesStoreInterface = (*DbPlanFeaturesStore)(nil)

func NewDbPlanFeaturesStore(db database.Dbx) *DbPlanFeaturesStore {
	return &DbPlanFeaturesStore{db: db}
}

func (s *DbPlanFeaturesStore) WithTx(db database.Dbx) *DbPlanFeaturesStore {
	return &DbPlanFeaturesStore{db: db}
}

func (s *DbPlanFeaturesStore) FindByProductID(ctx context.Context, productID string) (*models.PlanFeatures, error) {
	data, err := repository.PlanFeatures.GetOne(ctx, s.db, &map[string]any{
		"stripe_product_id": map[string]any{"_eq": productID},
	})
	return database.OptionalRow(data, err)
}

func (s *DbPlanFeaturesStore) InsertIfMissing(ctx context.Context, pf *models.PlanFeatures) error {
	q := squirrel.Insert("billing.plan_features").
		Columns("stripe_product_id", "daily_ai_tokens").
		Values(pf.StripeProductID, pf.DailyAiTokens).
		Suffix("ON CONFLICT (stripe_product_id) DO NOTHING").
		PlaceholderFormat(squirrel.Dollar)
	_, err := database.ExecWithBuilder(ctx, s.db, q)
	return err
}

func (s *DbPlanFeaturesStore) List(ctx context.Context, filter *PlanFeaturesFilter) ([]*models.PlanFeatures, error) {
	q := squirrel.Select("billing.plan_features.*").From("billing.plan_features")
	if filter != nil {
		q = queryPagination(q, filter)
		if filter.SortBy != "" && filter.SortOrder != "" {
			q = q.OrderBy("billing.plan_features." + filter.SortBy + " " + strings.ToUpper(filter.SortOrder))
		}
	}
	return database.QueryWithBuilder[*models.PlanFeatures](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
}

func (s *DbPlanFeaturesStore) CountPlanFeatures(ctx context.Context, _ *PlanFeaturesFilter) (int64, error) {
	q := squirrel.Select("COUNT(billing.plan_features.*)").From("billing.plan_features")
	data, err := database.QueryWithBuilder[database.CountOutput](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil
	}
	return data[0].Count, nil
}

func (s *DbPlanFeaturesStore) Upsert(ctx context.Context, pf *models.PlanFeatures) (*models.PlanFeatures, error) {
	q := squirrel.Insert("billing.plan_features").
		Columns("stripe_product_id", "daily_ai_tokens").
		Values(pf.StripeProductID, pf.DailyAiTokens).
		Suffix(`ON CONFLICT (stripe_product_id) DO UPDATE SET
			daily_ai_tokens = EXCLUDED.daily_ai_tokens,
			updated_at      = clock_timestamp()
		RETURNING *`).
		PlaceholderFormat(squirrel.Dollar)

	rows, err := database.QueryWithBuilder[models.PlanFeatures](ctx, s.db, q)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}
