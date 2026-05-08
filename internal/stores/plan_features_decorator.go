package stores

import (
	"context"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
)

type PlanFeaturesStoreDecorator struct {
	Delegate              *DbPlanFeaturesStore
	FindByProductIDFunc   func(ctx context.Context, productID string) (*models.PlanFeatures, error)
	UpsertFunc            func(ctx context.Context, pf *models.PlanFeatures) (*models.PlanFeatures, error)
	InsertIfMissingFunc   func(ctx context.Context, pf *models.PlanFeatures) error
	ListFunc              func(ctx context.Context) ([]*models.PlanFeatures, error)
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

func (s *PlanFeaturesStoreDecorator) List(ctx context.Context) ([]*models.PlanFeatures, error) {
	if s.ListFunc != nil {
		return s.ListFunc(ctx)
	}
	return s.Delegate.List(ctx)
}
