package queries

import (
	"context"
	"strings"

	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

// CountTasks implements AdminCrudActions.
func CountAiUsages(ctx context.Context, db database.Dbx, filter *shared.AiUsageListFilter) (int64, error) {
	where := ListAiUsagesFilterFunc(filter)
	return crudrepo.AiUsage.Count(ctx, db, where)
}

func ListAiUsages(ctx context.Context, dbx database.Dbx, input *shared.AiUsageListParams) ([]*models.AiUsage, error) {
	filter := input.AiUsageListFilter
	pageInput := &input.PaginatedInput

	limit, offset := database.PaginateRepo(pageInput)
	order := ListAiUsagesOrderByFunc(input)
	where := ListAiUsagesFilterFunc(&filter)
	data, err := crudrepo.AiUsage.Get(
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

func ListAiUsagesOrderByFunc(input *shared.AiUsageListParams) *map[string]string {

	if input == nil || input.SortBy == "" || input.SortOrder == "" {
		return nil
	}
	order := make(map[string]string)
	order[input.SortBy] = strings.ToUpper(input.SortOrder)
	return &order
}

func ListAiUsagesFilterFunc(filter *shared.AiUsageListFilter) *map[string]any {
	if filter == nil {
		return nil
	}
	where := make(map[string]any)
	if len(filter.UserID) > 0 {
		where["user_id"] = map[string]any{
			"_eq": filter.UserID,
		}
	}

	return &where
}
