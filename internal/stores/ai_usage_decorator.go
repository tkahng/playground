package stores

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
)

type AiUsageStoreDecorator struct {
	Delegate                    *DbAiUsageStore
	CreateAiUsageFunc           func(ctx context.Context, input *models.AiUsage) (*models.AiUsage, error)
	GetDailyTokensByTeamMemberFunc func(ctx context.Context, teamMemberID uuid.UUID, date time.Time) (int64, error)
	GetDailyTokensByTeamFunc    func(ctx context.Context, teamID uuid.UUID, date time.Time) (int64, error)
}

var _ AiUsageStoreInterface = (*AiUsageStoreDecorator)(nil)

func NewAiUsageStoreDecorator(db database.Dbx) *AiUsageStoreDecorator {
	return &AiUsageStoreDecorator{
		Delegate: NewDbAiUsageStore(db),
	}
}

func (s *AiUsageStoreDecorator) WithTx(db database.Dbx) *DbAiUsageStore {
	return s.Delegate.WithTx(db)
}

func (s *AiUsageStoreDecorator) Cleanup() {
	s.CreateAiUsageFunc = nil
	s.GetDailyTokensByTeamMemberFunc = nil
	s.GetDailyTokensByTeamFunc = nil
}

func (s *AiUsageStoreDecorator) CreateAiUsage(ctx context.Context, input *models.AiUsage) (*models.AiUsage, error) {
	if s.CreateAiUsageFunc != nil {
		return s.CreateAiUsageFunc(ctx, input)
	}
	return s.Delegate.CreateAiUsage(ctx, input)
}

func (s *AiUsageStoreDecorator) GetDailyTokensByTeamMember(ctx context.Context, teamMemberID uuid.UUID, date time.Time) (int64, error) {
	if s.GetDailyTokensByTeamMemberFunc != nil {
		return s.GetDailyTokensByTeamMemberFunc(ctx, teamMemberID, date)
	}
	return s.Delegate.GetDailyTokensByTeamMember(ctx, teamMemberID, date)
}

func (s *AiUsageStoreDecorator) GetDailyTokensByTeam(ctx context.Context, teamID uuid.UUID, date time.Time) (int64, error) {
	if s.GetDailyTokensByTeamFunc != nil {
		return s.GetDailyTokensByTeamFunc(ctx, teamID, date)
	}
	return s.Delegate.GetDailyTokensByTeam(ctx, teamID, date)
}
