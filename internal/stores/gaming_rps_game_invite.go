package stores

import (
	"context"
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

type RpsGameInviteFilter struct {
	repository.PaginatedInput
	repository.SortParams
	Ids       []uuid.UUID                    `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Tokens    []string                       `query:"statuses,omitempty" required:"false" minimum:"1" maximum:"100" enum:"pending,cancelled,completed"`
	ExpiresAt types.OptionalParam[time.Time] `query:"expires_at,omitempty" required:"false"`
}

func filterSelectRpsGameInvites(qs squirrel.SelectBuilder, filter *RpsGameInviteFilter) squirrel.SelectBuilder {
	if filter == nil {
		return qs
	}
	if len(filter.Ids) > 0 {
		qs = qs.Where(squirrel.Eq{"gaming.rps_game_invites.id": filter.Ids})
	}
	if len(filter.Tokens) > 0 {
		qs = qs.Where(squirrel.Eq{"gaming.rps_game_invites.token": filter.Tokens})
	}
	if filter.ExpiresAt.IsSet {
		qs = qs.Where(squirrel.Eq{"gaming.rps_game_invites.expires_at": filter.ExpiresAt.Value})
	}
	return qs
}

func filterDeleteRpsGameInvites(qs squirrel.DeleteBuilder, filter *RpsGameInviteFilter) squirrel.DeleteBuilder {
	if filter == nil {
		return qs
	}
	if len(filter.Ids) > 0 {
		qs = qs.Where(squirrel.Eq{"gaming.rps_game_invites.id": filter.Ids})
	}
	if len(filter.Tokens) > 0 {
		qs = qs.Where(squirrel.Eq{"gaming.rps_game_invites.token": filter.Tokens})
	}
	if filter.ExpiresAt.IsSet {
		qs = qs.Where(squirrel.Eq{"gaming.rps_game_invites.expires_at": filter.ExpiresAt.Value})
	}
	return qs
}

func rpGameInviteSortSelect(qs squirrel.SelectBuilder, filter Sortable) squirrel.SelectBuilder {
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

type RpsGameInviteStore interface {
	FindRpsGameInvite(ctx context.Context, filter *RpsGameInviteFilter) (*models.RpsGameInvite, error)
	FindRpsGameInvites(ctx context.Context, filter *RpsGameInviteFilter) ([]*models.RpsGameInvite, error)
	CreateRpsGameInvite(ctx context.Context, invite *models.RpsGameInvite) (*models.RpsGameInvite, error)
	UpdateRpsGameInvite(ctx context.Context, player *models.RpsGameInvite) (*models.RpsGameInvite, error)
	DeleteRpGameInvites(ctx context.Context, filter *RpsGameInviteFilter) (int64, error)
	CountRpsGameInvites(ctx context.Context, filter *RpsGameInviteFilter) (int64, error)
}

func (s *DBGamingStore) FindRpsGameInvite(ctx context.Context, filter *RpsGameInviteFilter) (*models.RpsGameInvite, error) {
	if filter == nil {
		filter = &RpsGameInviteFilter{
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
	res, err := s.FindRpsGameInvites(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}
	return res[0], nil
}

func (s *DBGamingStore) FindRpsGameInvites(ctx context.Context, filter *RpsGameInviteFilter) ([]*models.RpsGameInvite, error) {
	if filter == nil {
		filter = &RpsGameInviteFilter{
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
	q := squirrel.Select(repository.RpsGameInviteBuilder.ColumnNames()...).From(repository.RpsGameInviteBuilder.TableName())
	q = filterSelectRpsGameInvites(q, filter)
	q = rpGameInviteSortSelect(q, filter)
	q = queryPagination(q, filter)

	data, err := database.PgxQueryWithBuilder[models.RpsGameInvite](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *DBGamingStore) CreateRpsGameInvite(ctx context.Context, invite *models.RpsGameInvite) (*models.RpsGameInvite, error) {
	if invite.Metadata == nil {
		invite.Metadata = []byte("{}")
	}
	return repository.RpsGameInvite.PostOne(ctx, s.db, invite)
}

func (s *DBGamingStore) UpdateRpsGameInvite(ctx context.Context, invite *models.RpsGameInvite) (*models.RpsGameInvite, error) {
	return repository.RpsGameInvite.PutOne(ctx, s.db, invite)
}

func (s *DBGamingStore) DeleteRpGameInvites(ctx context.Context, filter *RpsGameInviteFilter) (int64, error) {
	qs := squirrel.Delete(repository.RpsGameInviteBuilder.TableName())
	qs = filterDeleteRpsGameInvites(qs, filter)
	return database.ExecWithBuilder(ctx, s.db, qs.PlaceholderFormat(squirrel.Dollar))
}

func (s *DBGamingStore) CountRpsGameInvites(ctx context.Context, filter *RpsGameInviteFilter) (int64, error) {
	q := squirrel.Select("COUNT(*)").From(repository.RpsGameInviteBuilder.TableName())
	q = filterSelectRpsGameInvites(q, filter)
	data, err := database.QueryWithBuilder[database.CountOutput](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil
	}

	return data[0].Count, nil
}
