package stores

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/types"
)

type GamingPlayerStore interface {
	FindPlayers(ctx context.Context, filter *PlayersFilter) ([]*models.Player, error)
	FindPlayer(ctx context.Context, filter *PlayersFilter) (*models.Player, error)
	FindHousePlayer(ctx context.Context) (*models.Player, error)
	CreatePlayer(ctx context.Context, player *models.Player) (*models.Player, error)
	UpdatePlayer(ctx context.Context, player *models.Player) (*models.Player, error)
	DeletePlayers(ctx context.Context, filter *PlayersFilter) (int64, error)
	CountPlayers(ctx context.Context, filter *PlayersFilter) (int64, error)
}

type PlayersFilter struct {
	repository.PaginatedInput
	repository.SortParams
	Ids          []uuid.UUID               `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Q            string                    `query:"q,omitempty" required:"false"`
	Emails       []string                  `query:"emails,omitempty" required:"false" minimum:"1" maximum:"100" format:"email"`
	DisplayNames []string                  `query:"display_names,omitempty" required:"false" minimum:"1" maximum:"100" format:"email"`
	UserIds      []uuid.UUID               `query:"user_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Registered   types.OptionalParam[bool] `query:"registered,omitempty" required:"false"`
	IsHouse      types.OptionalParam[bool] `query:"is_house,omitempty" required:"false"`
}

func (s *DBGamingStore) playerFilterSelect(qs squirrel.SelectBuilder, filter *PlayersFilter) squirrel.SelectBuilder {
	if filter == nil {
		// No filter: exclude house players by default.
		return qs.Where(squirrel.Eq{"gaming.players.is_house": false})
	}
	if filter.Ids != nil {
		qs = qs.Where(squirrel.Eq{"gaming.players.id": filter.Ids})
	}
	if filter.Q != "" {
		qs = qs.Where(squirrel.Or{
			squirrel.ILike{"gaming.players.display_name": "%" + filter.Q + "%"},
			squirrel.ILike{"gaming.players.email": "%" + filter.Q + "%"},
		})
	}
	if filter.Emails != nil {
		qs = qs.Where(squirrel.Eq{"gaming.players.email": filter.Emails})
	}
	if filter.DisplayNames != nil {
		qs = qs.Where(squirrel.Eq{"gaming.players.display_name": filter.DisplayNames})
	}
	if filter.UserIds != nil {
		qs = qs.Where(squirrel.Eq{"gaming.players.user_id": filter.UserIds})
	}
	if filter.Registered.IsSet {
		if filter.Registered.Value {
			qs = qs.Where(fmt.Sprintf("%s IS NOT NULL", "gaming.players.user_id"))
		} else {
			qs = qs.Where(fmt.Sprintf("%s IS NULL", "gaming.players.user_id"))
		}
	}
	if filter.IsHouse.IsSet {
		qs = qs.Where(squirrel.Eq{"gaming.players.is_house": filter.IsHouse.Value})
	} else if filter.Ids == nil {
		// Default: exclude the house player from all general queries.
		// ID-specific lookups (e.g., resolving participant players) bypass this.
		qs = qs.Where(squirrel.Eq{"gaming.players.is_house": false})
	}
	return qs
}
func (s *DBGamingStore) playerFilterDelete(qs squirrel.DeleteBuilder, filter *PlayersFilter) squirrel.DeleteBuilder {
	if filter == nil {
		return qs
	}
	if filter.Ids != nil {
		qs = qs.Where(squirrel.Eq{"gaming.players.id": filter.Ids})
	}
	if filter.Q != "" {
		qs = qs.Where(squirrel.Or{
			squirrel.ILike{"gaming.players.display_name": "%" + filter.Q + "%"},
			squirrel.ILike{"gaming.players.email": "%" + filter.Q + "%"},
		})
	}
	if filter.Emails != nil {
		qs = qs.Where(squirrel.Eq{"gaming.players.email": filter.Emails})
	}
	if filter.DisplayNames != nil {
		qs = qs.Where(squirrel.Eq{"gaming.players.display_name": filter.DisplayNames})
	}
	if filter.UserIds != nil {
		qs = qs.Where(squirrel.Eq{"gaming.players.user_id": filter.UserIds})
	}
	if filter.Registered.IsSet {
		if filter.Registered.Value {
			qs = qs.Where(fmt.Sprintf("%s IS NOT NULL", "gaming.players.user_id"))
		} else {
			qs = qs.Where(fmt.Sprintf("%s IS NULL", "gaming.players.user_id"))
		}
	}
	if filter.IsHouse.IsSet {
		qs = qs.Where(squirrel.Eq{"gaming.players.is_house": filter.IsHouse.Value})
	} else if filter.Ids == nil {
		// Prevent accidental bulk deletion of the house player.
		qs = qs.Where(squirrel.Eq{"gaming.players.is_house": false})
	}
	return qs
}

func (s *DBGamingStore) playerSortSelect(qs squirrel.SelectBuilder, filter Sortable) squirrel.SelectBuilder {
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
	if slices.Contains(repository.PlayerBuilder.FieldNames(), sortBy) {
		qs = qs.OrderBy(sortBy + " " + strings.ToUpper(sortOrder))
	} else {
		slog.Warn("sort by field not found in repository columns", "sortBy", sortBy, "sortOrder", sortOrder)
	}
	return qs
}

func (s *DBGamingStore) FindPlayers(ctx context.Context, filter *PlayersFilter) ([]*models.Player, error) {
	if filter == nil {
		filter = &PlayersFilter{
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
	q := squirrel.Select(repository.PlayerBuilder.ColumnNames()...).From("gaming.players")
	q = s.playerFilterSelect(q, filter)
	q = s.playerSortSelect(q, filter)
	q = queryPagination(q, filter)

	data, err := database.PgxQueryRowsToStruct[models.Player](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *DBGamingStore) FindPlayer(ctx context.Context, filter *PlayersFilter) (*models.Player, error) {
	if filter == nil {
		filter = &PlayersFilter{
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
	res, err := s.FindPlayers(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}
	return res[0], nil
}

func (s *DBGamingStore) FindHousePlayer(ctx context.Context) (*models.Player, error) {
	return s.FindPlayer(ctx, &PlayersFilter{
		IsHouse: types.OptionalParam[bool]{IsSet: true, Value: true},
	})
}

func (s *DBGamingStore) CreatePlayer(ctx context.Context, player *models.Player) (*models.Player, error) {
	if player == nil {
		return nil, errors.New("player is nil")
	}
	if player.Metadata == nil {
		player.Metadata = []byte("{}")
	}
	return repository.Player.PostOne(ctx, s.db, player)
}

func (s *DBGamingStore) UpdatePlayer(ctx context.Context, player *models.Player) (*models.Player, error) {
	return repository.Player.PutOne(ctx, s.db, player)
}

func (s *DBGamingStore) DeletePlayers(ctx context.Context, filter *PlayersFilter) (int64, error) {
	qs := squirrel.Delete(repository.PlayerBuilder.TableName())
	qs = s.playerFilterDelete(qs, filter)
	return database.ExecWithBuilder(ctx, s.db, qs.PlaceholderFormat(squirrel.Dollar))
}

func (s *DBGamingStore) CountPlayers(ctx context.Context, filter *PlayersFilter) (int64, error) {
	q := squirrel.Select("COUNT(*)").From(repository.PlayerBuilder.TableName())
	q = s.playerFilterSelect(q, filter)
	data, err := database.QueryWithBuilder[database.CountOutput](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil
	}

	return data[0].Count, nil
}
