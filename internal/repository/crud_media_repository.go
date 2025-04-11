package repository

import (
	"context"
	"log/slog"
	"slices"

	"github.com/google/uuid"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
)

func ListMedia(ctx context.Context, db bob.DB, input *shared.MediaListParams) (models.MediumSlice, error) {
	q := models.Media.Query()
	filter := input.MediaListFilter
	pageInput := &input.PaginatedInput

	ViewApplyPagination(q, pageInput)
	ListMediaOrderByFunc(ctx, q, input)
	ListMediaFilterFunc(ctx, q, &filter)
	data, err := q.All(ctx, db)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ListMediaFilterFunc(ctx context.Context, q *psql.ViewQuery[*models.Medium, models.MediumSlice], mediaListFilter *shared.MediaListFilter) {
	if q == nil {
		return
	}
	if mediaListFilter == nil {
		return
	}
	if mediaListFilter.UserID != "" {
		userId, err := uuid.Parse(mediaListFilter.UserID)
		if err != nil {
			slog.Error("failed to parse user id", "error", err)
			return
		}
		q.Apply(
			models.SelectWhere.Media.UserID.EQ(userId),
		)
	}
}

func ListMediaOrderByFunc(ctx context.Context, q *psql.ViewQuery[*models.Medium, models.MediumSlice], input *shared.MediaListParams) {
	if q == nil {
		return
	}
	if input == nil || input.SortBy == "" {
		q.Apply(
			sm.OrderBy(models.MediumColumns.CreatedAt).Desc(),
			sm.OrderBy(models.MediumColumns.ID).Desc(),
		)
		return
	}
	if slices.Contains(models.Media.Columns().Names(), input.SortBy) {
		if input.SortParams.SortOrder == "desc" {
			q.Apply(
				sm.OrderBy(input.SortBy).Desc(),
				sm.OrderBy(models.MediumColumns.ID).Desc(),
			)
		} else if input.SortParams.SortOrder == "asc" {
			q.Apply(
				sm.OrderBy(input.SortBy).Asc(),
				sm.OrderBy(models.MediumColumns.ID).Asc(),
			)
		}
	}
}

func CountMedia(ctx context.Context, db bob.DB, input *shared.MediaListFilter) (int64, error) {
	q := models.Media.Query()
	ListMediaFilterFunc(ctx, q, input)
	return CountExec(ctx, db, q)
}
