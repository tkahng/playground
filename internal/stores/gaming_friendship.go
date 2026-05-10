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
)

type GamingFriendshipStore interface {
	FindFriendship(ctx context.Context, filter *FriendshipFilter) (*models.Friendship, error)
	FindFriendships(ctx context.Context, filter *FriendshipFilter) ([]*models.Friendship, error)
	CreateFriendship(ctx context.Context, friendship *models.Friendship) (*models.Friendship, error)
	UpdateFriendship(ctx context.Context, player *models.Friendship) (*models.Friendship, error)
	DeleteFriendships(ctx context.Context, filter *FriendshipFilter) (int64, error)
	CountFriendships(ctx context.Context, filter *FriendshipFilter) (int64, error)
}

type FriendshipFilter struct {
	repository.PaginatedInput
	repository.SortParams
	Ids                          []uuid.UUID               `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	RequestingPlayerIds          []uuid.UUID               `query:"requesting_player_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	InvitedPlayerIds             []uuid.UUID               `query:"invited_player_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	RequestingOrInvitedPlayerIds []uuid.UUID               `query:"requesting_or_invited_player_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	// PlayerPair matches friendships in either direction between exactly two players:
	// (requesting=A AND invited=B) OR (requesting=B AND invited=A)
	PlayerPair                   *[2]uuid.UUID             `query:"-"`
	Statuses                     []models.FriendshipStatus `query:"statuses,omitempty" required:"false" minimum:"1" maximum:"100" enum:"pending,accepted,declined,blocked"`
	CreatedAfter                 *time.Time                `query:"-"`
}

func (s *DBGamingStore) friendshipFilterSelect(sq squirrel.SelectBuilder, filter *FriendshipFilter) squirrel.SelectBuilder {
	if filter == nil {
		return sq
	}
	if filter.Ids != nil {
		sq = sq.Where(squirrel.Eq{"gaming.friendships.id": filter.Ids})
	}
	if len(filter.Statuses) > 0 {
		sq = sq.Where(squirrel.Eq{"gaming.friendships.status": filter.Statuses})
	}
	if filter.RequestingPlayerIds != nil {
		sq = sq.Where(squirrel.Eq{"gaming.friendships.requesting_player_id": filter.RequestingPlayerIds})
	}
	if filter.InvitedPlayerIds != nil {
		sq = sq.Where(squirrel.Eq{"gaming.friendships.invited_player_id": filter.InvitedPlayerIds})
	}
	if filter.RequestingOrInvitedPlayerIds != nil {
		sq = sq.Where(
			squirrel.Or{
				squirrel.Eq{"gaming.friendships.requesting_player_id": filter.RequestingOrInvitedPlayerIds},
				squirrel.Eq{"gaming.friendships.invited_player_id": filter.RequestingOrInvitedPlayerIds},
			},
		)
	}
	if filter.PlayerPair != nil {
		a, b := filter.PlayerPair[0], filter.PlayerPair[1]
		sq = sq.Where(squirrel.Or{
			squirrel.And{
				squirrel.Eq{"gaming.friendships.requesting_player_id": a},
				squirrel.Eq{"gaming.friendships.invited_player_id": b},
			},
			squirrel.And{
				squirrel.Eq{"gaming.friendships.requesting_player_id": b},
				squirrel.Eq{"gaming.friendships.invited_player_id": a},
			},
		})
	}
	if filter.CreatedAfter != nil {
		sq = sq.Where(squirrel.GtOrEq{"gaming.friendships.created_at": *filter.CreatedAfter})
	}

	return sq
}
func (s *DBGamingStore) friendshipFilterDelete(sq squirrel.DeleteBuilder, filter *FriendshipFilter) squirrel.DeleteBuilder {
	if filter == nil {
		return sq
	}
	if filter.Ids != nil {
		sq = sq.Where(squirrel.Eq{"gaming.friendships.id": filter.Ids})
	}
	if len(filter.Statuses) > 0 {
		sq = sq.Where(squirrel.Eq{"gaming.friendships.status": filter.Statuses})
	}
	if filter.RequestingPlayerIds != nil {
		sq = sq.Where(squirrel.Eq{"gaming.friendships.requesting_player_id": filter.RequestingPlayerIds})
	}
	if filter.InvitedPlayerIds != nil {
		sq = sq.Where(squirrel.Eq{"gaming.friendships.invited_player_id": filter.InvitedPlayerIds})
	}
	if filter.RequestingOrInvitedPlayerIds != nil {
		sq = sq.Where(
			squirrel.Or{
				squirrel.Eq{"gaming.friendships.requesting_player_id": filter.RequestingOrInvitedPlayerIds},
				squirrel.Eq{"gaming.friendships.invited_player_id": filter.RequestingOrInvitedPlayerIds},
			},
		)
	}

	return sq
}

func (s *DBGamingStore) friendshipSortSelect(qs squirrel.SelectBuilder, filter Sortable) squirrel.SelectBuilder {
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
	if slices.Contains(repository.FriendshipBuilder.FieldNames(), sortBy) {
		qs = qs.OrderBy(sortBy + " " + strings.ToUpper(sortOrder))
	} else {
		slog.Warn("sort by field not found in repository columns", "sortBy", sortBy, "sortOrder", sortOrder)
	}
	return qs
}

func (s *DBGamingStore) FindFriendship(ctx context.Context, filter *FriendshipFilter) (*models.Friendship, error) {
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

func (s *DBGamingStore) FindFriendships(ctx context.Context, filter *FriendshipFilter) ([]*models.Friendship, error) {
	q := squirrel.Select(repository.FriendshipBuilder.ColumnNames()...).From("gaming.friendships")
	q = s.friendshipFilterSelect(q, filter)
	q = s.friendshipSortSelect(q, filter)
	q = queryPagination(q, filter)

	data, err := database.PgxQueryRowsToStruct[models.Friendship](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *DBGamingStore) CreateFriendship(ctx context.Context, friendship *models.Friendship) (*models.Friendship, error) {
	if friendship == nil {
		return nil, errors.New("player is nil")
	}
	if friendship.Status == "" {
		friendship.Status = models.FriendshipStatusPending
	}
	return repository.Friendship.PostOne(ctx, s.db, friendship)
}

func (s *DBGamingStore) UpdateFriendship(ctx context.Context, player *models.Friendship) (*models.Friendship, error) {
	return repository.Friendship.PutOne(ctx, s.db, player)
}

func (s *DBGamingStore) DeleteFriendships(ctx context.Context, filter *FriendshipFilter) (int64, error) {
	qs := squirrel.Delete(repository.FriendshipBuilder.TableName())
	qs = s.friendshipFilterDelete(qs, filter)
	return database.ExecWithBuilder(ctx, s.db, qs.PlaceholderFormat(squirrel.Dollar))
	// return database.PgxQuerySingleScalar[int64](ctx, s.db, qs.PlaceholderFormat(squirrel.Dollar))
}

func (s *DBGamingStore) CountFriendships(ctx context.Context, filter *FriendshipFilter) (int64, error) {
	q := squirrel.Select("COUNT(*)").From(repository.FriendshipBuilder.TableName())
	q = s.friendshipFilterSelect(q, filter)
	data, err := database.QueryWithBuilder[database.CountOutput](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil
	}

	return data[0].Count, nil
}
