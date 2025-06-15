package stores

import (
	"context"
	"log/slog"
	"slices"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/tools/utils"
)

type MediaStoreInterface interface {
	CreateMedia(ctx context.Context, media *models.Medium) (*models.Medium, error)
	FindMediaByID(ctx context.Context, mediaId uuid.UUID) (*models.Medium, error)
	UpdateMedia(ctx context.Context, media *models.Medium) (*models.Medium, error)
	FindMedia(ctx context.Context, filter *MediaListFilter) ([]*models.Medium, error)
	CountMedia(ctx context.Context, filter *MediaListFilter) (int64, error)
}

type DbMediaStore struct {
	dbx database.Dbx
}

func NewMediaStore(dbx database.Dbx) *DbMediaStore {
	return &DbMediaStore{
		dbx: dbx,
	}
}

func (s *DbMediaStore) UpdateMedia(ctx context.Context, media *models.Medium) (*models.Medium, error) {
	data, err := repository.Media.PutOne(
		ctx,
		s.dbx,
		media,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *DbMediaStore) CreateMedia(ctx context.Context, media *models.Medium) (*models.Medium, error) {
	data, err := repository.Media.PostOne(
		ctx,
		s.dbx,
		media,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *DbMediaStore) FindMediaByID(ctx context.Context, mediaId uuid.UUID) (*models.Medium, error) {
	data, err := repository.Media.GetOne(
		ctx,
		s.dbx,
		&map[string]any{
			"id": map[string]any{
				"_eq": mediaId,
			},
		},
	)
	return database.OptionalRow(data, err)
}

type MediaListFilter struct {
	PaginatedInput
	SortParams
	Q       string      `query:"q,omitempty" required:"false"`
	UserIds []uuid.UUID `query:"userId,omitempty" format:"uuid" required:"false"`
}

func (s *DbMediaStore) FindMedia(ctx context.Context, filter *MediaListFilter) ([]*models.Medium, error) {
	where := s.filter(filter)
	orderBy := s.sort(filter)

	limit, offset := pagination(filter)
	data, err := repository.Media.Get(
		ctx,
		s.dbx,
		where,
		orderBy,
		&limit,
		&offset,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *DbMediaStore) CountMedia(ctx context.Context, filter *MediaListFilter) (int64, error) {
	where := s.filter(filter)
	count, err := repository.Media.Count(
		ctx,
		s.dbx,
		where,
	)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *DbMediaStore) filter(filter *MediaListFilter) *map[string]any {
	if filter == nil {
		return nil
	}
	where := make(map[string]any)
	if len(filter.UserIds) > 0 {
		where["user_id"] = map[string]any{
			"_in": filter.UserIds,
		}
	}
	if filter.Q != "" {
		// where["_or"] = []map[string]any{
		// 	{
		// 		"disk": map[string]any{
		// 			"_ilike": "%" + filter.Q + "%",
		// 		},
		// 	},
		// 	{
		// 		"directory": map[string]any{
		// 			"_ilike": "%" + filter.Q + "%",
		// 		},
		// 	},
		// 	{
		// 		"filename": map[string]any{
		// 			"_ilike": "%" + filter.Q + "%",
		// 		},
		// 	},
		// 	{
		// 		"original_filename": map[string]any{
		// 			"_ilike": "%" + filter.Q + "%",
		// 		},
		// 	},
		// }
	}

	return &where
}

func (s *DbMediaStore) sort(filter Sortable) *map[string]string {
	if filter == nil {
		return nil // return nil if no filter is provided
	}

	sortBy, sortOrder := filter.Sort()
	if sortBy != "" && slices.Contains(repository.MediaBuilder.ColumnNames(), utils.Quote(sortBy)) {
		return &map[string]string{
			sortBy: sortOrder,
		}
	} else {
		slog.Info("sort by field not found in repository columns", "sortBy", sortBy, "sortOrder", sortOrder, "columns", repository.UserBuilder.ColumnNames())
	}

	return nil // default no sorting
}
