package stores

import (
	"context"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
)

type PlanFeaturesStoreDecorator struct {
	Delegate                *DbPlanFeaturesStore
	FindByProductIDFunc     func(ctx context.Context, productID string) (*models.PlanFeatures, error)
	UpsertFunc              func(ctx context.Context, pf *models.PlanFeatures) (*models.PlanFeatures, error)
	InsertIfMissingFunc     func(ctx context.Context, pf *models.PlanFeatures) error
	ListFunc                func(ctx context.Context, filter *PlanFeaturesFilter) ([]*models.PlanFeatures, error)
	CountPlanFeaturesFunc   func(ctx context.Context, filter *PlanFeaturesFilter) (int64, error)
}

var _ PlanFeaturesStoreInterface = (*PlanFeaturesStoreDecorator)(nil)

func NewPlanFeaturesStoreDecorator(db database.Dbx) *PlanFeaturesStoreDecorator {
	return &PlanFeaturesStoreDecorator{Delegate: NewDbPlanFeaturesStore(db)}
}

func (s *PlanFeaturesStoreDecorator) WithTx(db database.Dbx) *DbPlanFeaturesStore {
	return s.Delegate.WithTx(db)
}

func (s *PlanFeaturesStoreDecorator) Cleanup() {
	s.FindByProductIDFunc = nil
	s.UpsertFunc = nil
	s.InsertIfMissingFunc = nil
	s.ListFunc = nil
	s.CountPlanFeaturesFunc = nil
}

func (s *PlanFeaturesStoreDecorator) FindByProductID(ctx context.Context, productID string) (*models.PlanFeatures, error) {
	if s.FindByProductIDFunc != nil {
		return s.FindByProductIDFunc(ctx, productID)
	}
	return s.Delegate.FindByProductID(ctx, productID)
}

func (s *PlanFeaturesStoreDecorator) Upsert(ctx context.Context, pf *models.PlanFeatures) (*models.PlanFeatures, error) {
	if s.UpsertFunc != nil {
		return s.UpsertFunc(ctx, pf)
	}
	return s.Delegate.Upsert(ctx, pf)
}

func (s *PlanFeaturesStoreDecorator) InsertIfMissing(ctx context.Context, pf *models.PlanFeatures) error {
	if s.InsertIfMissingFunc != nil {
		return s.InsertIfMissingFunc(ctx, pf)
	}
	return s.Delegate.InsertIfMissing(ctx, pf)
}

func (s *PlanFeaturesStoreDecorator) List(ctx context.Context, filter *PlanFeaturesFilter) ([]*models.PlanFeatures, error) {
	if s.ListFunc != nil {
		return s.ListFunc(ctx, filter)
	}
	return s.Delegate.List(ctx, filter)
}

func (s *PlanFeaturesStoreDecorator) CountPlanFeatures(ctx context.Context, filter *PlanFeaturesFilter) (int64, error) {
	if s.CountPlanFeaturesFunc != nil {
		return s.CountPlanFeaturesFunc(ctx, filter)
	}
	return s.Delegate.CountPlanFeatures(ctx, filter)
}
