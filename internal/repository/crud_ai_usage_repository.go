package repository

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
)

// CountTasks implements AdminCrudActions.
func CountAiUsages(ctx context.Context, db Queryer, filter *shared.AiUsageListFilter) (int64, error) {
	q := models.AiUsages.Query()
	ListAiUsagesFilterFunc(ctx, q, filter)
	return CountExec(ctx, db, q)
}

func ListAiUsages(ctx context.Context, db Queryer, input *shared.AiUsageListParams) ([]*models.AiUsage, error) {
	q := models.AiUsages.Query()
	filter := input.AiUsageListFilter
	pageInput := &input.PaginatedInput

	ViewApplyPagination(q, pageInput)
	ListAiUsagesOrderByFunc(ctx, q, input)
	ListAiUsagesFilterFunc(ctx, q, &filter)
	data, err := q.All(ctx, db)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ListAiUsagesOrderByFunc(ctx context.Context, q *psql.ViewQuery[*models.AiUsage, models.AiUsageSlice], input *shared.AiUsageListParams) {
	if q == nil {
		return
	}
	if input == nil || input.SortBy == "" {
		q.Apply(
			sm.OrderBy(models.AiUsageColumns.CreatedAt).Desc(),
			sm.OrderBy(models.AiUsageColumns.ID).Desc(),
		)
		return
	}
	if slices.Contains(models.AiUsages.Columns().Names(), input.SortBy) {
		if input.SortParams.SortOrder == "desc" {
			q.Apply(
				sm.OrderBy(psql.Quote(input.SortBy)).Desc(),
				sm.OrderBy(models.AiUsageColumns.ID).Desc(),
			)
		} else if input.SortParams.SortOrder == "asc" || input.SortParams.SortOrder == "" {
			q.Apply(
				sm.OrderBy(psql.Quote(input.SortBy)).Asc(),
				sm.OrderBy(models.AiUsageColumns.ID).Asc(),
			)
		}
		return
	}

}

func ListAiUsagesFilterFunc(ctx context.Context, q *psql.ViewQuery[*models.AiUsage, models.AiUsageSlice], filter *shared.AiUsageListFilter) {
	if filter == nil {
		return
	}
	if filter.Q != "" {
		// q.Apply(
		// 	psql.WhereOr(models.SelectWhere.AiUsages.Name.ILike("%"+filter.Q+"%"),
		// 		models.SelectWhere.AiUsages.Description.ILike("%"+filter.Q+"%")),
		// )
	}

	if len(filter.UserID) > 0 {
		id, err := uuid.Parse(filter.UserID)
		if err != nil {
			return
		}
		q.Apply(
			models.SelectWhere.AiUsages.UserID.EQ(id),
		)
	}

}
