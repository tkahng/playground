package stores

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
)

type PlayersFilter struct {
	repository.PaginatedInput
	repository.SortParams
	Ids          []uuid.UUID `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Q            string      `query:"q,omitempty" required:"false"`
	Emails       []string    `query:"emails,omitempty" required:"false" minimum:"1" maximum:"100" format:"email"`
	DisplayNames []string    `query:"display_names,omitempty" required:"false" minimum:"1" maximum:"100" format:"email"`
}

func (s *DBGamingStore) filterPlayerWhere(filter *PlayersFilter) *map[string]any {
	if filter == nil {
		return nil
	}
	where := map[string]any{}
	if filter.Ids != nil {
		where["id"] = map[string]any{
			"_in": filter.Ids,
		}
	}
	if filter.Q != "" {
		where["_or"] = []map[string]any{
			{
				"_and": []map[string]any{
					{
						"display_name": map[string]any{"_ilike": "%" + filter.Q + "%"},
					},
				},
			},
			{
				"_and": []map[string]any{
					{
						"email": map[string]any{"_ilike": "%" + filter.Q + "%"},
					},
				},
			},
		}
	}
	if filter.Emails != nil {
		where["email"] = map[string]any{
			"_in": filter.Emails,
		}
	}
	if filter.DisplayNames != nil {
		where["display_name"] = map[string]any{
			"_in": filter.DisplayNames,
		}
	}
	return &where
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
	limit, offset := filter.Pagination()
	sortBy, sortOrder := filter.Sort()
	var sort *map[string]string
	if sortBy != "" {
		sort = &map[string]string{
			sortBy: sortOrder,
		}
	}
	where := s.filterPlayerWhere(filter)
	return repository.Player.GetWithOptions(
		ctx,
		s.db,
		repository.WithLimit(limit),
		repository.WithOffset(offset),
		repository.WithWhere(where),
		repository.WithOrder(sort),
	)
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
	where := s.filterPlayerWhere(filter)
	return repository.Player.Delete(ctx, s.db, where)
}

func (s *DBGamingStore) CountPlayers(ctx context.Context, filter *PlayersFilter) (int64, error) {
	where := s.filterPlayerWhere(filter)
	return repository.Player.Count(ctx, s.db, where)
}
