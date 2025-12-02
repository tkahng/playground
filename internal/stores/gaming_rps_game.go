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

type RpsGameStore interface {
	FindRpsGames(ctx context.Context, filter *RpsGameFilter) ([]*models.RpsGame, error)
	FindRpsGame(ctx context.Context, filter *RpsGameFilter) (*models.RpsGame, error)
	CreateRpsGame(ctx context.Context, game *models.RpsGame) (*models.RpsGame, error)
	UpdateRpsGame(ctx context.Context, game *models.RpsGame) (*models.RpsGame, error)
	CreateGameWithRequest(ctx context.Context, input *GameCreateInput) (*models.RpsGame, error)
}

type RpsGameFilter struct {
	repository.PaginatedInput
	repository.SortParams
	IDs            []uuid.UUID                    `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Statuses       []models.RpsGameStatus         `query:"statuses,omitempty" required:"false" minimum:"1" maximum:"100" enum:"pending,cancelled,completed"`
	CompletedAt    types.OptionalParam[time.Time] `query:"completed_at,omitempty" required:"false"`
	ExpiresAt      types.OptionalParam[time.Time] `query:"expires_at,omitempty" required:"false"`
	ParticipantIds []uuid.UUID                    `query:"participant_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
}

func filterRpsGames(qs squirrel.SelectBuilder, filter *RpsGameFilter) squirrel.SelectBuilder {
	if filter == nil {
		return qs
	}
	if len(filter.IDs) > 0 {
		qs = qs.Where(squirrel.Eq{"g.id": filter.IDs})
	}
	if len(filter.Statuses) > 0 {
		qs = qs.Where(squirrel.Eq{"g.status": filter.Statuses})
	}
	if filter.CompletedAt.IsSet {
		qs = qs.Where(squirrel.Eq{"g.completed_at": filter.CompletedAt.Value})
	}
	if filter.ExpiresAt.IsSet {
		qs = qs.Where(squirrel.Eq{"g.expires_at": filter.ExpiresAt.Value})
	}
	if len(filter.ParticipantIds) > 0 {
		//  WHERE g.id IN (
		//     SELECT rp.gameid
		//     FROM gaming.rpsparticipants rp
		//     WHERE rp.playerid = ANY($1)  -- $1 = array of UUIDs
		// );

		qs = qs.Where(`
			g.id IN (
				SELECT rp.gameid
				FROM gaming.rpsparticipants rp
				WHERE rp.playerid = ANY(?)  -- $1 = array of UUIDs
			)
		`, filter.ParticipantIds)
	}
	return qs
}
func rpsgamesSortSelect(qs squirrel.SelectBuilder, filter Sortable) squirrel.SelectBuilder {
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

func (s *DBGamingStore) FindRpsGame(ctx context.Context, filter *RpsGameFilter) (*models.RpsGame, error) {
	q := squirrel.Select(repository.RpsGameBuilder.ColumnNames()...).From("gaming.rpsgames")
	q = filterRpsGames(q, filter)
	q = rpsgamesSortSelect(q, filter)

	data, err := database.PgxQueryRowsToStruct[models.RpsGame](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		return data[0], nil
	}
	return nil, nil
}
func (s *DBGamingStore) FindRpsGames(ctx context.Context, filter *RpsGameFilter) ([]*models.RpsGame, error) {
	q := squirrel.Select(repository.RpsGameBuilder.ColumnNames()...).From("gaming.rpsgames")
	q = filterRpsGames(q, filter)
	q = rpsgamesSortSelect(q, filter)
	q = queryPagination(q, filter)

	data, err := database.PgxQueryRowsToStruct[models.RpsGame](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return data, nil
}

type GameCreateInput struct {
	RequestingPlayerID   uuid.UUID
	InvitedPlayerID      uuid.UUID
	RequestingPlayerMove models.RpsParticipantMove
}

// CreateGame implements [GamingStore].
func (s *DBGamingStore) CreateGameWithRequest(ctx context.Context, input *GameCreateInput) (*models.RpsGame, error) {
	players, err := s.FindPlayers(ctx, &PlayersFilter{
		Ids: []uuid.UUID{input.RequestingPlayerID, input.InvitedPlayerID},
	})
	if err != nil {
		return nil, err
	}
	if len(players) != 2 {
		return nil, errors.New("did not find 2 players.")
	}
	var invitedPlayer, requestingPlayer *models.Player
	for _, player := range players {
		if player.ID == input.InvitedPlayerID {
			invitedPlayer = player
		}
		if player.ID == input.RequestingPlayerID {
			requestingPlayer = player
		}
	}
	if invitedPlayer == nil || requestingPlayer == nil {
		return nil, errors.New("did not find 2 players.")
	}
	return nil, nil
}

// CreateRpsGame implements [GamingStore].
func (s *DBGamingStore) CreateRpsGame(ctx context.Context, game *models.RpsGame) (*models.RpsGame, error) {
	return repository.RpsGame.PostOne(ctx, s.db, game)
}

// UpdateRpsGame implements [GamingStore].
func (s *DBGamingStore) UpdateRpsGame(ctx context.Context, game *models.RpsGame) (*models.RpsGame, error) {
	return repository.RpsGame.PutOne(ctx, s.db, game)
}
