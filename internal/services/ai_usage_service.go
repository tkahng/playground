package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/apierrors"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/gemini"
)

const (
	FreeTierDailyTokenLimit int64 = 10_000
	PaidTierDailyTokenLimit int64 = 100_000
)

type AiUsageService interface {
	// CheckQuota returns TooManyRequests if the team has exhausted its daily token allowance.
	CheckQuota(ctx context.Context, teamID uuid.UUID) error
	// RecordUsage persists a usage row after a successful AI call.
	RecordUsage(ctx context.Context, userID uuid.UUID, teamMemberID uuid.UUID, teamID uuid.UUID, usage gemini.Usage) error
	// GetDailyLimit returns the team's configured daily token limit based on its subscription.
	GetDailyLimit(ctx context.Context, teamID uuid.UUID) (int64, error)
}

type aiUsageService struct {
	adapter stores.StorageAdapterInterface
}

func NewAiUsageService(adapter stores.StorageAdapterInterface) AiUsageService {
	return &aiUsageService{adapter: adapter}
}

func (s *aiUsageService) GetDailyLimit(ctx context.Context, teamID uuid.UUID) (int64, error) {
	subs, err := s.adapter.Subscription().FindActiveSubscriptionsByTeamIds(ctx, teamID)
	if err != nil {
		return 0, err
	}

	// No active subscription → free tier
	if len(subs) == 0 || subs[0] == nil {
		return FreeTierDailyTokenLimit, nil
	}

	sub := subs[0]
	// Load the full subscription with price + product to get the product ID
	full, err := s.adapter.Subscription().FindSubscriptionsWithPriceProductByIds(ctx, sub.ID)
	if err != nil {
		return 0, err
	}
	if len(full) == 0 || full[0] == nil || full[0].Price == nil {
		return PaidTierDailyTokenLimit, nil
	}

	pf, err := s.adapter.PlanFeatures().FindByProductID(ctx, full[0].Price.ProductID)
	if err != nil {
		return 0, err
	}
	if pf == nil {
		// Paid subscription but no plan_features row configured yet
		return PaidTierDailyTokenLimit, nil
	}

	return pf.DailyAiTokens, nil
}

func (s *aiUsageService) CheckQuota(ctx context.Context, teamID uuid.UUID) error {
	limit, err := s.GetDailyLimit(ctx, teamID)
	if err != nil {
		return err
	}

	used, err := s.adapter.AiUsage().GetDailyTokensByTeam(ctx, teamID, time.Now().UTC())
	if err != nil {
		return err
	}

	if used >= limit {
		return apierrors.TooManyRequests("daily AI token limit reached")
	}
	return nil
}

func (s *aiUsageService) RecordUsage(ctx context.Context, userID uuid.UUID, teamMemberID uuid.UUID, teamID uuid.UUID, usage gemini.Usage) error {
	row := &models.AiUsage{
		UserID:           userID,
		TeamMemberID:     &teamMemberID,
		TeamID:           &teamID,
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
	}
	_, err := s.adapter.AiUsage().CreateAiUsage(ctx, row)
	return err
}
