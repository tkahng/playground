package apis

import (
	"context"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/mapper"
)

type PlanFeature struct {
	ID              uuid.UUID `json:"id"`
	StripeProductID string    `json:"stripe_product_id"`
	DailyAiTokens   int64     `json:"daily_ai_tokens"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func fromModelPlanFeature(pf *models.PlanFeatures) *PlanFeature {
	if pf == nil {
		return nil
	}
	return &PlanFeature{
		ID:              pf.ID,
		StripeProductID: pf.StripeProductID,
		DailyAiTokens:   pf.DailyAiTokens,
		CreatedAt:       pf.CreatedAt,
		UpdatedAt:       pf.UpdatedAt,
	}
}

type PlanFeaturesUpsertBody struct {
	DailyAiTokens int64 `json:"daily_ai_tokens" minimum:"0" required:"true"`
}

type PlanFeaturesUpsertInput struct {
	ProductID string                 `path:"product-id" required:"true"`
	Body      PlanFeaturesUpsertBody `json:"body"`
}

func (api *Api) AdminPlanFeaturesList(ctx context.Context, _ *struct{}) (*struct {
	Body []*PlanFeature
}, error) {
	rows, err := api.App().Adapter().PlanFeatures().List(ctx)
	if err != nil {
		return nil, err
	}
	return &struct {
		Body []*PlanFeature
	}{
		Body: mapper.Map(rows, fromModelPlanFeature),
	}, nil
}

func (api *Api) AdminPlanFeaturesUpsert(ctx context.Context, input *PlanFeaturesUpsertInput) (*struct {
	Body *PlanFeature
}, error) {
	product, err := api.App().Adapter().Product().FindProductById(ctx, input.ProductID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, huma.Error404NotFound("product not found")
	}

	pf, err := api.App().Adapter().PlanFeatures().Upsert(ctx, &models.PlanFeatures{
		StripeProductID: product.ID,
		DailyAiTokens:   input.Body.DailyAiTokens,
	})
	if err != nil {
		return nil, err
	}
	return &struct {
		Body *PlanFeature
	}{
		Body: fromModelPlanFeature(pf),
	}, nil
}

func (api *Api) AdminPlanFeaturesGet(ctx context.Context, input *struct {
	ProductID string `path:"product-id" required:"true"`
}) (*struct {
	Body *PlanFeature
}, error) {
	pf, err := api.App().Adapter().PlanFeatures().FindByProductID(ctx, input.ProductID)
	if err != nil {
		return nil, err
	}
	if pf == nil {
		return nil, huma.Error404NotFound("plan features not found for product")
	}
	return &struct {
		Body *PlanFeature
	}{
		Body: fromModelPlanFeature(pf),
	}, nil
}

