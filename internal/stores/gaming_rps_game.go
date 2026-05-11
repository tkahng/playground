package stores

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/queries"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/types"
)

type RpsGameStore interface {
	CountRpsGames(ctx context.Context, filter *RpsGameFilter) (int64, error)
	FindRpsGames(ctx context.Context, filter *RpsGameFilter) ([]*models.RpsGame, error)
	FindRpsGame(ctx context.Context, filter *RpsGameFilter) (*models.RpsGame, error)
	FindRpsGameForUpdate(ctx context.Context, gameID uuid.UUID) (*models.RpsGame, error)
	FindExpiredPendingBetGames(ctx context.Context) ([]*models.RpsGame, error)
	FindPendingGamesExpiringWithin(ctx context.Context, within time.Duration) ([]*models.RpsGame, error)
	MarkRpsGameExpirySent(ctx context.Context, game *models.RpsGame) error
	CreateRpsGame(ctx context.Context, game *models.RpsGame) (*models.RpsGame, error)
	UpdateRpsGame(ctx context.Context, game *models.RpsGame) (*models.RpsGame, error)
}

type RpsGameFilter struct {
	repository.PaginatedInput
	repository.SortParams
	Ids            []uuid.UUID                    `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Statuses       []models.RpsGameStatus         `query:"statuses,omitempty" required:"false" minimum:"1" maximum:"100" enum:"pending,cancelled,completed"`
	CompletedAt    types.OptionalParam[time.Time] `query:"completed_at,omitempty" required:"false"`
	CompletedAtOp  queries.FilterOperator         `query:"completed_at_op,omitempty" required:"false" enum:"eq,gt,gte,lt,lte"`
	ExpiresAt      types.OptionalParam[time.Time] `query:"expires_at,omitempty" required:"false"`
	ExpiresAtOp    queries.FilterOperator         `query:"expires_at_op,omitempty" required:"false" enum:"eq,gt,gte,lt,lte"`
	ParticipantIds []uuid.UUID                    `query:"participant_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
}

func filterRpsGames(qs squirrel.SelectBuilder, filter *RpsGameFilter) squirrel.SelectBuilder {
	if filter == nil {
		return qs
	}
	if len(filter.Ids) > 0 {
		qs = qs.Where(squirrel.Eq{"gaming.rps_games.id": filter.Ids})
	}
	if len(filter.Statuses) > 0 {
		qs = qs.Where(squirrel.Eq{"gaming.rps_games.status": filter.Statuses})
	}
	if filter.CompletedAtOp != "" {
		qs = queries.ToSquirrelOp(qs, filter.CompletedAtOp, "gaming.rps_games.completed_at", filter.CompletedAt.Value)
	}
	if filter.ExpiresAtOp != "" {
		qs = queries.ToSquirrelOp(qs, filter.ExpiresAtOp, "gaming.rps_games.expires_at", filter.ExpiresAt.Value)
	}
	if len(filter.ParticipantIds) > 0 {
		//  WHERE g.id IN (
		//     SELECT rp.gameid
		//     FROM gaming.rpsparticipants rp
		//     WHERE rp.playerid = ANY($1)  -- $1 = array of UUIDs
		// );

		qs = qs.Where(`
			gaming.rps_games.id IN (
				SELECT rp.game_id
				FROM gaming.rps_participants rp
				WHERE rp.player_id = ANY(?)  -- $1 = array of UUIDs
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
	q := squirrel.Select(repository.RpsGameBuilder.ColumnNames()...).From("gaming.rps_games")
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
	q := squirrel.Select(repository.RpsGameBuilder.ColumnNames()...).From("gaming.rps_games")
	q = filterRpsGames(q, filter)
	q = rpsgamesSortSelect(q, filter)
	q = queryPagination(q, filter)

	data, err := database.PgxQueryRowsToStruct[models.RpsGame](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *DBGamingStore) CountRpsGames(ctx context.Context, filter *RpsGameFilter) (int64, error) {
	q := squirrel.Select("COUNT(*)").From("gaming.rps_games")
	q = filterRpsGames(q, filter)
	c, err := database.QueryWithBuilder[database.CountOutput](
		ctx,
		s.db,
		q.PlaceholderFormat(squirrel.Dollar),
	)
	if err != nil {
		return 0, err
	}
	if len(c) == 0 {
		return 0, nil
	}
	return c[0].Count, nil
}

// CreateRpsGame implements [GamingStore].
func (s *DBGamingStore) CreateRpsGame(ctx context.Context, game *models.RpsGame) (*models.RpsGame, error) {
	if game == nil {
		return nil, errors.New("game is nil")
	}
	if game.Metadata == nil {
		game.Metadata = []byte("{}")
	}
	return repository.RpsGame.PostOne(ctx, s.db, game)
}

// UpdateRpsGame implements [GamingStore].
func (s *DBGamingStore) UpdateRpsGame(ctx context.Context, game *models.RpsGame) (*models.RpsGame, error) {
	return repository.RpsGame.PutOne(ctx, s.db, game)
}

// FindRpsGameForUpdate fetches a game row and holds a row-level lock for the
// duration of the surrounding transaction. Call this inside RunInTxCtx to
// prevent concurrent double-settlement.
func (s *DBGamingStore) FindRpsGameForUpdate(ctx context.Context, gameID uuid.UUID) (*models.RpsGame, error) {
	cols := strings.Join(repository.RpsGameBuilder.ColumnNames(), ", ")
	query := fmt.Sprintf("SELECT %s FROM gaming.rps_games WHERE id = $1 FOR UPDATE", cols)
	data, err := database.QueryAll[*models.RpsGame](ctx, s.db, query, gameID)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return data[0], nil
}

// FindPendingGamesExpiringWithin returns pending games whose expiry falls
// within the window [now, now+within) that have not yet had a warning sent
// (metadata->>'expiry_warning_sent' is absent or false).
func (s *DBGamingStore) FindPendingGamesExpiringWithin(ctx context.Context, within time.Duration) ([]*models.RpsGame, error) {
	cols := strings.Join(repository.RpsGameBuilder.ColumnNames(), ", ")
	query := fmt.Sprintf(
		`SELECT %s FROM gaming.rps_games
		 WHERE status = $1
		   AND expires_at > clock_timestamp()
		   AND expires_at <= clock_timestamp() + $2::interval
		   AND COALESCE((metadata->>'expiry_warning_sent')::boolean, false) = false`,
		cols,
	)
	return database.QueryAll[*models.RpsGame](ctx, s.db, query,
		string(models.RpsGameStatusPending),
		within.String(),
	)
}

// MarkRpsGameExpirySent sets metadata->expiry_warning_sent = true on a game.
func (s *DBGamingStore) MarkRpsGameExpirySent(ctx context.Context, game *models.RpsGame) error {
	const query = `UPDATE gaming.rps_games SET metadata = metadata || '{"expiry_warning_sent":true}' WHERE id = $1`
	_, err := s.db.Exec(ctx, query, game.ID)
	return err
}

// FindExpiredPendingBetGames returns pending games with a host bet escrow whose
// expiry time has passed. These need their host escrow refunded and status set to cancelled.
func (s *DBGamingStore) FindExpiredPendingBetGames(ctx context.Context) ([]*models.RpsGame, error) {
	cols := strings.Join(repository.RpsGameBuilder.ColumnNames(), ", ")
	query := fmt.Sprintf(
		"SELECT %s FROM gaming.rps_games WHERE status = $1 AND expires_at < clock_timestamp() AND host_bet_transfer_id IS NOT NULL",
		cols,
	)
	return database.QueryAll[*models.RpsGame](ctx, s.db, query, string(models.RpsGameStatusPending))
}
