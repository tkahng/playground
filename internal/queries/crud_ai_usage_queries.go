package queries

import (
	"context"
	"strings"

	"github.com/tkahng/authgo/internal/crud/repository"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
)

// CountTasks implements AdminCrudActions.
func CountAiUsages(ctx context.Context, db Queryer, filter *shared.AiUsageListFilter) (int64, error) {
	where := ListAiUsagesFilterFunc(filter)
	return repository.AiUsage.Count(ctx, db, where)
}

func ListAiUsages(ctx context.Context, db Queryer, input *shared.AiUsageListParams) ([]*models.AiUsage, error) {
	filter := input.AiUsageListFilter
	pageInput := &input.PaginatedInput

	limit, offset := PaginateRepo(pageInput)
	order := ListAiUsagesOrderByFunc(input)
	where := ListAiUsagesFilterFunc(&filter)
	data, err := repository.AiUsage.Get(
		ctx,
		db,
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
	// if slices.Contains(models.AiUsages.Columns().Names(), input.SortBy) {
	order[input.SortBy] = strings.ToUpper(input.SortOrder)
	// }
	return &order
}

func ListAiUsagesFilterFunc(filter *shared.AiUsageListFilter) *map[string]any {
	if filter == nil {
		return nil
	}
	where := make(map[string]any)
	// if filter.Q != "" {

	// }

	if len(filter.UserID) > 0 {
		where["user_id"] = map[string]any{
			"_eq": filter.UserID,
		}
	}

	return &where
}
