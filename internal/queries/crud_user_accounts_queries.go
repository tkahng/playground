package queries

import (
	"context"
	"slices"

	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	crudmodels "github.com/tkahng/authgo/internal/crud/crudModels"
	"github.com/tkahng/authgo/internal/crud/crudrepo"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/types"
)

var (
	// UserColumnNames = models.Users.Columns().Names()
	UserAccountColumnNames = []string{"id", "user_id", "type", "provider", "provider_account_id", "created_at", "updated_at"}
	// UserAccountColumnNames = models.UserAccounts.Columns().Names()
)

// ListUserAccounts implements AdminCrudActions.
// ListUsers implements AdminCrudActions.
func ListUserAccounts2(ctx context.Context, db Queryer, input *shared.UserAccountListParams) (models.UserAccountSlice, error) {

	q := models.UserAccounts.Query()
	filter := input.UserAccountListFilter
	pageInput := &input.PaginatedInput

	ViewApplyPagination(q, pageInput)
	ListUserAccountsOrderByFunc(ctx, q, input)
	ListUserAccountFilterFunc(ctx, q, &filter)
	data, err := q.All(ctx, db)
	if err != nil {
		return nil, err
	}
	return data, nil
}
func ListUserAccounts(ctx context.Context, db Queryer, input *shared.UserAccountListParams) ([]*crudmodels.UserAccount, error) {
	where := UserAccountWhere(&input.UserAccountListFilter)
	sort := UserAccountOrderBy(&input.SortParams)
	data, err := crudrepo.UserAccount.Get(
		ctx,
		db,
		where,
		sort,
		types.Pointer(int(input.PaginatedInput.Page)),
		types.Pointer(int(input.PaginatedInput.PerPage)),
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ListUserAccountsOrderByFunc(ctx context.Context, q *psql.ViewQuery[*models.UserAccount, models.UserAccountSlice], input *shared.UserAccountListParams) {
	if q == nil {
		return
	}
	if input == nil || input.SortBy == "" {
		q.Apply(
			sm.OrderBy(models.UserAccountColumns.CreatedAt).Desc(),
			sm.OrderBy(models.UserAccountColumns.ID).Desc(),
		)
		return
	}
	if slices.Contains(UserAccountColumnNames, input.SortBy) {
		if input.SortParams.SortOrder == "desc" {
			q.Apply(
				sm.OrderBy(input.SortBy).Desc(),
				sm.OrderBy(models.UserAccountColumns.ID).Desc(),
			)
		} else if input.SortParams.SortOrder == "asc" {
			q.Apply(
				sm.OrderBy(input.SortBy).Asc(),
				sm.OrderBy(models.UserAccountColumns.ID).Asc(),
			)
		}
	}
}

func UserAccountOrderBy(params *shared.SortParams) *map[string]string {
	if params == nil {
		return nil
	}
	if slices.Contains(UserAccountColumnNames, params.SortBy) {
		return &map[string]string{
			params.SortBy: params.SortOrder,
		}
	}
	return nil
}

// func CreateUser(ctx context.Context, db Queryer, params *shared.AuthenticateUserParams) (*models.User, error) {
func UserAccountWhere(filter *shared.UserAccountListFilter) *map[string]any {
	where := make(map[string]any)
	if filter == nil {
		return &where
	}
	if len(filter.Providers) > 0 {
		var providers []string
		for _, p := range filter.Providers {
			providers = append(providers, p.String())
		}
		where["provider"] = map[string]any{
			"_in": providers,
		}
	}
	if len(filter.ProviderTypes) > 0 {
		var providerTypes []string
		for _, pt := range filter.ProviderTypes {
			providerTypes = append(providerTypes, pt.String())
		}
		where["type"] = map[string]any{
			"_in": providerTypes,
		}
	}
	if len(filter.Ids) > 0 {
		where["id"] = map[string]any{
			"_in": filter.Ids,
		}
	}
	if len(filter.UserIds) > 0 {
		where["user_id"] = map[string]any{
			"_in": filter.UserIds,
		}
	}
	return &where
}
func ListUserAccountFilterFunc(ctx context.Context, q *psql.ViewQuery[*models.UserAccount, models.UserAccountSlice], filter *shared.UserAccountListFilter) {
	if filter == nil {
		return
	}
	if len(filter.Providers) > 0 {
		var providers []models.Providers
		for _, p := range filter.Providers {
			providers = append(providers, shared.ToModelProvider(p))
		}
		q.Apply(
			models.SelectWhere.UserAccounts.Provider.In(providers...),
		)
	}
	if len(filter.ProviderTypes) > 0 {
		var providerTypes []models.ProviderTypes
		for _, pt := range filter.ProviderTypes {
			providerTypes = append(providerTypes, shared.ToModelProviderType(pt))
		}
		q.Apply(
			models.SelectWhere.UserAccounts.Type.In(providerTypes...),
		)
	}
	if len(filter.Ids) > 0 {
		var ids = ParseUUIDs(filter.Ids)
		q.Apply(
			models.SelectWhere.Users.ID.In(ids...),
		)
	}
	if len(filter.UserIds) > 0 {
		ids := ParseUUIDs(filter.UserIds)
		q.Apply(
			models.SelectWhere.UserAccounts.UserID.In(ids...),
		)
	}
}

// CountUsers implements AdminCrudActions.
func CountUserAccounts(ctx context.Context, db Queryer, filter *shared.UserAccountListFilter) (int64, error) {
	q := models.UserAccounts.Query()
	ListUserAccountFilterFunc(ctx, q, filter)
	data, err := q.Count(ctx, db)
	if err != nil {
		return 0, err
	}
	return data, nil
}
