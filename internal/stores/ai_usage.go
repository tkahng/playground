package stores

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
)

type AiUsageFilter struct {
	PaginatedInput
	TeamID *uuid.UUID
	Since  *time.Time
	Until  *time.Time
}

type AiUsageStoreInterface interface {
	CreateAiUsage(ctx context.Context, input *models.AiUsage) (*models.AiUsage, error)
	GetDailyTokensByTeamMember(ctx context.Context, teamMemberID uuid.UUID, date time.Time) (int64, error)
	GetDailyTokensByTeam(ctx context.Context, teamID uuid.UUID, date time.Time) (int64, error)
	ListAiUsages(ctx context.Context, filter *AiUsageFilter) ([]*models.AiUsage, error)
	CountAiUsages(ctx context.Context, filter *AiUsageFilter) (int64, error)
	WithTx(db database.Dbx) *DbAiUsageStore
}

type DbAiUsageStore struct {
	db database.Dbx
}

var _ AiUsageStoreInterface = (*DbAiUsageStore)(nil)

func NewDbAiUsageStore(db database.Dbx) *DbAiUsageStore {
	return &DbAiUsageStore{db: db}
}

func (s *DbAiUsageStore) WithTx(db database.Dbx) *DbAiUsageStore {
	return &DbAiUsageStore{db: db}
}

func (s *DbAiUsageStore) CreateAiUsage(ctx context.Context, input *models.AiUsage) (*models.AiUsage, error) {
	return repository.AiUsage.PostOne(ctx, s.db, input)
}

func (s *DbAiUsageStore) GetDailyTokensByTeamMember(ctx context.Context, teamMemberID uuid.UUID, date time.Time) (int64, error) {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24 * time.Hour)

	qs := squirrel.Select("COALESCE(SUM(total_tokens), 0)").
		From("app.ai_usages").
		Where(squirrel.Eq{"team_member_id": teamMemberID}).
		Where(squirrel.GtOrEq{"created_at": dayStart}).
		Where(squirrel.Lt{"created_at": dayEnd}).
		PlaceholderFormat(squirrel.Dollar)

	return database.PgxQuerySingleScalar[int64](ctx, s.db, qs)
}

func (s *DbAiUsageStore) whereFilter(filter *AiUsageFilter) squirrel.And {
	cond := squirrel.And{}
	if filter == nil {
		return cond
	}
	if filter.TeamID != nil {
		cond = append(cond, squirrel.Eq{"team_id": *filter.TeamID})
	}
	if filter.Since != nil {
		cond = append(cond, squirrel.GtOrEq{"created_at": *filter.Since})
	}
	if filter.Until != nil {
		cond = append(cond, squirrel.Lt{"created_at": *filter.Until})
	}
	return cond
}

func (s *DbAiUsageStore) ListAiUsages(ctx context.Context, filter *AiUsageFilter) ([]*models.AiUsage, error) {
	qs := squirrel.Select(repository.AiUsageBuilder.FieldNames()...).
		From("app.ai_usages").
		Where(s.whereFilter(filter)).
		OrderBy("created_at DESC").
		PlaceholderFormat(squirrel.Dollar)
	if filter != nil {
		limit, offset := filter.LimitOffset()
		qs = qs.Limit(uint64(limit)).Offset(uint64(offset))
	}
	return database.QueryWithBuilder[*models.AiUsage](ctx, s.db, qs)
}

func (s *DbAiUsageStore) CountAiUsages(ctx context.Context, filter *AiUsageFilter) (int64, error) {
	qs := squirrel.Select("COUNT(*)").
		From("app.ai_usages").
		Where(s.whereFilter(filter)).
		PlaceholderFormat(squirrel.Dollar)
	return database.PgxQuerySingleScalar[int64](ctx, s.db, qs)
}

func (s *DbAiUsageStore) GetDailyTokensByTeam(ctx context.Context, teamID uuid.UUID, date time.Time) (int64, error) {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24 * time.Hour)

	qs := squirrel.Select("COALESCE(SUM(total_tokens), 0)").
		From("app.ai_usages").
		Where(squirrel.Eq{"team_id": teamID}).
		Where(squirrel.GtOrEq{"created_at": dayStart}).
		Where(squirrel.Lt{"created_at": dayEnd}).
		PlaceholderFormat(squirrel.Dollar)

	return database.PgxQuerySingleScalar[int64](ctx, s.db, qs)
}
