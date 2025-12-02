package stores

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/types"
)

type RpsParticipantStore interface {
	FindRpsParticipant(ctx context.Context, filter *RpsParticipantFilter) (*models.RpsParticipant, error)
	FindRpsParticipants(ctx context.Context, filter *RpsParticipantFilter) ([]*models.RpsParticipant, error)
	CountRpsParticipants(ctx context.Context, filter *RpsParticipantFilter) (int64, error)
	CreateRpsParticipant(ctx context.Context, participant *models.RpsParticipant) (*models.RpsParticipant, error)
	UpdateRpsParticipant(ctx context.Context, participant *models.RpsParticipant) (*models.RpsParticipant, error)
}
type RpsParticipantFilter struct {
	repository.PaginatedInput
	repository.SortParams
	Ids         []uuid.UUID                    `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	RpsGameIds  []uuid.UUID                    `query:"rps_game_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	PlayerIds   []uuid.UUID                    `query:"player_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Statuses    []models.RpsParticipantStatus  `query:"statuses,omitempty" required:"false" minimum:"1" maximum:"100" enum:"pending,declined,completed"`
	Moves       []models.RpsParticipantMove    `query:"moves,omitempty" required:"false" minimum:"1" maximum:"100" enum:"rock,paper,scissors"`
	Results     []models.RpsParticipantResult  `query:"results,omitempty" required:"false" minimum:"1" maximum:"100" enum:"win,lose,draw"`
	RespondedAt types.OptionalParam[time.Time] `query:"responded_at,omitempty" required:"false"`
	Responded   types.OptionalParam[bool]      `query:"responded,omitempty" required:"false"`
}

func rpsParticipantsSortSelect(qs squirrel.SelectBuilder, filter Sortable) squirrel.SelectBuilder {
	if filter == nil {
		return qs // return original query if no filter is provided
	}
	sortBy, sortOrder := filter.Sort()
	if sortBy == "" {
		return qs
	}
	if sortOrder == "" {
		sortOrder = "ASC"
	}
	// if sortBy is in the registered fieldnames, it is a scalar field. direct sort.
	if slices.Contains(repository.RpsGameBuilder.FieldNames(), sortBy) {
		qs = qs.OrderBy(sortBy + " " + strings.ToUpper(sortOrder))
	} else {
		slog.Warn("sort by field not found in repository columns", "sortBy", sortBy, "sortOrder", sortOrder)
	}
	return qs
}
func rpsParticipantsFilterSelect(q squirrel.SelectBuilder, filter *RpsParticipantFilter) squirrel.SelectBuilder {
	if filter == nil {
		return q // return original query if no filter is provided
	}
	if len(filter.Ids) > 0 {
		q = q.Where(squirrel.Eq{"gaming.rps_participants.id": filter.Ids})
	}
	if len(filter.RpsGameIds) > 0 {
		q = q.Where(squirrel.Eq{"gaming.rps_participants.rps_game_id": filter.RpsGameIds})
	}
	if len(filter.PlayerIds) > 0 {
		q = q.Where(squirrel.Eq{"gaming.rps_participants.player_id": filter.PlayerIds})
	}
	if len(filter.Statuses) > 0 {
		q = q.Where(squirrel.Eq{"gaming.rps_participants.status": filter.Statuses})
	}
	if len(filter.Moves) > 0 {
		q = q.Where(squirrel.Eq{"gaming.rps_participants.move": filter.Moves})
	}
	if len(filter.Results) > 0 {
		q = q.Where(squirrel.Eq{"gaming.rps_participants.result": filter.Results})
	}
	if filter.RespondedAt.IsSet {
		q = q.Where(squirrel.Eq{"gaming.rps_participants.responded_at": filter.RespondedAt.Value})
	}
	if filter.Responded.IsSet {
		if filter.Responded.Value {
			q = q.Where("gaming.rps_participants.responded_at IS NOT NULL")
		} else {
			q = q.Where("gaming.rps_participants.responded_at IS NULL")
		}
	}
	return q
}

func (s *DBGamingStore) FindRpsParticipants(ctx context.Context, filter *RpsParticipantFilter) ([]*models.RpsParticipant, error) {
	if filter == nil {
		filter = &RpsParticipantFilter{
			PaginatedInput: repository.PaginatedInput{
				Page:    0,
				PerPage: 10,
			},
			SortParams: repository.SortParams{
				SortBy:    "created_at",
				SortOrder: "desc",
			},
		}
	}
	q := squirrel.Select(repository.RpsParticipantBuilder.ColumnNames()...).From("gaming.rps_participants")
	q = rpsParticipantsFilterSelect(q, filter)
	q = rpsParticipantsSortSelect(q, filter)
	q = queryPagination(q, filter)
	data, err := database.PgxQueryRowsToStruct[models.RpsParticipant](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *DBGamingStore) FindRpsParticipant(ctx context.Context, filter *RpsParticipantFilter) (*models.RpsParticipant, error) {
	if filter == nil {
		filter = &RpsParticipantFilter{
			PaginatedInput: repository.PaginatedInput{
				Page:    0,
				PerPage: 1,
			},
			SortParams: repository.SortParams{
				SortBy:    "created_at",
				SortOrder: "desc",
			},
		}
	} else {
		filter.PerPage = 1
		filter.Page = 0
	}
	res, err := s.FindRpsParticipants(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}
	return res[0], nil
}

func (s *DBGamingStore) CountRpsParticipants(ctx context.Context, filter *RpsParticipantFilter) (int64, error) {
	q := squirrel.Select("COUNT(*)").From("gaming.rps_participants")
	q = rpsParticipantsFilterSelect(q, filter)
	return database.ExecWithBuilder(ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
}

// CreateParticipant
func (s *DBGamingStore) CreateRpsParticipant(ctx context.Context, rpsParticipant *models.RpsParticipant) (*models.RpsParticipant, error) {
	if rpsParticipant == nil {
		return nil, errors.New("rpsParticipant is nil")
	}
	if rpsParticipant.Metadata == nil {
		rpsParticipant.Metadata = []byte("{}")
	}
	return repository.RpsParticipant.PostOne(ctx, s.db, rpsParticipant)
}

// UpdateParticipant
func (s *DBGamingStore) UpdateRpsParticipant(ctx context.Context, rpsParticipant *models.RpsParticipant) (*models.RpsParticipant, error) {
	return repository.RpsParticipant.PutOne(ctx, s.db, rpsParticipant)
}
