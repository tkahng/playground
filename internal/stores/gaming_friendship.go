package stores

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
)

type FriendshipFilter struct {
	repository.PaginatedInput
	repository.SortParams
	Ids                          []uuid.UUID               `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	RequestingPlayerIds          []uuid.UUID               `query:"requesting_player_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	InvitedPlayerIds             []uuid.UUID               `query:"invited_player_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	RequestingOrInvitedPlayerIds []uuid.UUID               `query:"requesting_or_invited_player_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Statuses                     []models.FriendshipStatus `query:"statuses,omitempty" required:"false" minimum:"1" maximum:"100" enum:"pending,accepted,declined"`
}

func (s *DBGamingStore) friendshipFilterWhere(filter *FriendshipFilter) *map[string]any {
	if filter == nil {
		return nil
	}
	var setStatus bool
	where := map[string]any{}
	if filter.Ids != nil {
		where["id"] = map[string]any{
			"_in": filter.Ids,
		}
	}
	if len(filter.Statuses) > 0 {
		setStatus = true
		where["status"] = map[string]any{
			"_in": filter.Statuses,
		}
	}
	if filter.RequestingPlayerIds != nil {
		where["requesting_player_id"] = map[string]any{
			"_in": filter.RequestingPlayerIds,
		}
	}
	if filter.InvitedPlayerIds != nil {
		where["invited_player_id"] = map[string]any{
			"_in": filter.InvitedPlayerIds,
		}
	}
	if filter.RequestingOrInvitedPlayerIds != nil {
		var keys []string = []string{
			"requesting_player_id",
			"invited_player_id",
		}
		vars := []map[string]any{}
		for _, v := range keys {
			var arg map[string]any = map[string]any{}
			if setStatus {
				arg["status"] = map[string]any{
					"_in": filter.Statuses,
				}
			}
			arg[v] = map[string]any{
				"_in": filter.RequestingOrInvitedPlayerIds,
			}
			vars = append(vars, arg)
			if setStatus {

			}
		}

		where["_or"] = vars
	}

	return &where
}

func (s *DBGamingStore) FindFriendship(ctx context.Context, filter *FriendshipFilter) (*models.Frindship, error) {
	if filter == nil {
		filter = &FriendshipFilter{
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
	res, err := s.FindFriendships(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}
	return res[0], nil
}

func (s *DBGamingStore) FindFriendships(ctx context.Context, filter *FriendshipFilter) ([]*models.Frindship, error) {
	limit, offset := filter.Pagination()
	sortBy, sortOrder := filter.Sort()
	var sort *map[string]string
	if sortBy != "" {
		sort = &map[string]string{
			sortBy: sortOrder,
		}
	}
	where := s.friendshipFilterWhere(filter)
	return repository.Frindship.GetWithOptions(
		ctx,
		s.db,
		repository.WithLimit(limit),
		repository.WithOffset(offset),
		repository.WithWhere(where),
		repository.WithOrder(sort),
	)
}

func (s *DBGamingStore) CreateFriendship(ctx context.Context, friendship *models.Frindship) (*models.Frindship, error) {
	if friendship == nil {
		return nil, errors.New("player is nil")
	}
	if friendship.Status == "" {
		friendship.Status = models.FriendshipStatusPending
	}
	return repository.Frindship.PostOne(ctx, s.db, friendship)
}

func (s *DBGamingStore) UpdateFrindship(ctx context.Context, player *models.Frindship) (*models.Frindship, error) {
	return repository.Frindship.PutOne(ctx, s.db, player)
}

func (s *DBGamingStore) DeleteFrindships(ctx context.Context, filter *FriendshipFilter) (int64, error) {
	where := s.friendshipFilterWhere(filter)
	return repository.Frindship.Delete(ctx, s.db, where)
}

func (s *DBGamingStore) CountFriendships(ctx context.Context, filter *FriendshipFilter) (int64, error) {
	where := s.friendshipFilterWhere(filter)
	return repository.Frindship.Count(ctx, s.db, where)
}
