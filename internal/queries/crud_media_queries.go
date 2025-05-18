package queries

import (
	"context"
	"slices"
	"strings"

	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

var (
	MediaColumns = []string{
		"id",
		"user_id",
		"disk",
		"directory",
		"filename",
		"original_filename",
		"extension",
		"mime_type",
		"size",
		"created_at",
		"updated_at",
	}
)

func ListMedia(ctx context.Context, dbx database.Dbx, input *shared.MediaListParams) ([]*models.Medium, error) {
	filter := input.MediaListFilter
	pageInput := &input.PaginatedInput

	limit, offset := database.PaginateRepo(pageInput)
	where := ListMediaFilterFunc(&filter)
	order := ListMediaOrderByFunc(input)
	data, err := crudrepo.Media.Get(
		ctx,
		dbx,
		where,
		order,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ListMediaFilterFunc(mediaListFilter *shared.MediaListFilter) *map[string]any {
	if mediaListFilter == nil {
		return nil
	}
	where := make(map[string]any)
	if mediaListFilter.UserID != "" {
		where["user_id"] = map[string]any{
			"_eq": mediaListFilter.UserID,
		}
	}
	if mediaListFilter.Q != "" {
		where["_or"] = []map[string]any{
			{
				"disk": map[string]any{
					"_ilike": "%" + mediaListFilter.Q + "%",
				},
			},
			{
				"directory": map[string]any{
					"_ilike": "%" + mediaListFilter.Q + "%",
				},
			},
			{
				"filename": map[string]any{
					"_ilike": "%" + mediaListFilter.Q + "%",
				},
			},
			{
				"original_filename": map[string]any{
					"_ilike": "%" + mediaListFilter.Q + "%",
				},
			},
		}
	}

	return &where
}

func ListMediaOrderByFunc(input *shared.MediaListParams) *map[string]string {
	order := make(map[string]string)
	if input == nil || input.SortBy == "" {
		order["created_at"] = "DESC"
		order["id"] = "DESC"
		return &order
	}
	if slices.Contains(MediaColumns, input.SortBy) {
		order[input.SortBy] = strings.ToUpper(input.SortOrder)
	}
	return &order
}

func CountMedia(ctx context.Context, db database.Dbx, input *shared.MediaListFilter) (int64, error) {
	where := ListMediaFilterFunc(input)
	c, err := crudrepo.Media.Count(
		ctx,
		db,
		where,
	)
	return c, err
}
