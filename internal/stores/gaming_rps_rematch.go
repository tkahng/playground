package stores

import (
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
)

type RpsRematchFilter struct {
	Ids                 []uuid.UUID
	OriginalGameIDs     []uuid.UUID
	RequestingPlayerIDs []uuid.UUID
	InvitedPlayerIDs    []uuid.UUID
	Statuses            []models.RpsRematchStatus
}

type RpsRematchStore interface {
	CreateRpsRematchRequest(ctx context.Context, req *models.RpsRematchRequest) (*models.RpsRematchRequest, error)
	FindRpsRematchRequest(ctx context.Context, filter *RpsRematchFilter) (*models.RpsRematchRequest, error)
	UpdateRpsRematchRequest(ctx context.Context, req *models.RpsRematchRequest) (*models.RpsRematchRequest, error)
	FindExpiredPendingRpsRematches(ctx context.Context) ([]*models.RpsRematchRequest, error)
}

func (s *DBGamingStore) rematchFilterSelect(qs squirrel.SelectBuilder, filter *RpsRematchFilter) squirrel.SelectBuilder {
	if filter == nil {
		return qs
	}
	if len(filter.Ids) > 0 {
		qs = qs.Where(squirrel.Eq{"gaming.rps_rematch_requests.id": filter.Ids})
	}
	if len(filter.OriginalGameIDs) > 0 {
		qs = qs.Where(squirrel.Eq{"gaming.rps_rematch_requests.original_game_id": filter.OriginalGameIDs})
	}
	if len(filter.RequestingPlayerIDs) > 0 {
		qs = qs.Where(squirrel.Eq{"gaming.rps_rematch_requests.requesting_player_id": filter.RequestingPlayerIDs})
	}
	if len(filter.InvitedPlayerIDs) > 0 {
		qs = qs.Where(squirrel.Eq{"gaming.rps_rematch_requests.invited_player_id": filter.InvitedPlayerIDs})
	}
	if len(filter.Statuses) > 0 {
		qs = qs.Where(squirrel.Eq{"gaming.rps_rematch_requests.status": filter.Statuses})
	}
	return qs
}

func (s *DBGamingStore) CreateRpsRematchRequest(ctx context.Context, req *models.RpsRematchRequest) (*models.RpsRematchRequest, error) {
	if req.Metadata == nil {
		req.Metadata = []byte("{}")
	}
	return repository.RpsRematchRequest.PostOne(ctx, s.db, req)
}

func (s *DBGamingStore) FindRpsRematchRequest(ctx context.Context, filter *RpsRematchFilter) (*models.RpsRematchRequest, error) {
	cols := strings.Join(repository.RpsRematchRequestBuilder.ColumnNames(), ", ")
	qs := squirrel.Select(cols).From("gaming.rps_rematch_requests")
	qs = s.rematchFilterSelect(qs, filter)
	qs = qs.Limit(1).PlaceholderFormat(squirrel.Dollar)
	rows, err := database.QueryWithBuilder[*models.RpsRematchRequest](ctx, s.db, qs)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

func (s *DBGamingStore) UpdateRpsRematchRequest(ctx context.Context, req *models.RpsRematchRequest) (*models.RpsRematchRequest, error) {
	return repository.RpsRematchRequest.PutOne(ctx, s.db, req)
}

func (s *DBGamingStore) FindExpiredPendingRpsRematches(ctx context.Context) ([]*models.RpsRematchRequest, error) {
	cols := strings.Join(repository.RpsRematchRequestBuilder.ColumnNames(), ", ")
	query := fmt.Sprintf(
		"SELECT %s FROM gaming.rps_rematch_requests WHERE status = $1 AND expires_at < clock_timestamp()",
		cols,
	)
	return database.QueryAll[*models.RpsRematchRequest](ctx, s.db, query, string(models.RpsRematchStatusPending))
}
