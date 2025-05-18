package queries

import (
	"context"
	"slices"

	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/types"
)

var (
	// UserColumnNames = models.Users.Columns().Names()
	UserAccountColumnNames = []string{"id", "user_id", "type", "provider", "provider_account_id", "created_at", "updated_at"}
	// UserAccountColumnNames = models.UserAccounts.Columns().Names()
)

// ListUserAccounts implements AdminCrudActions.
// ListUsers implements AdminCrudActions.

func ListUserAccounts(ctx context.Context, db database.Dbx, input *shared.UserAccountListParams) ([]*models.UserAccount, error) {
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

// func CreateUser(ctx context.Context, db db.Dbx, params *shared.AuthenticateUserParams) (*models.User, error) {
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

// CountUsers implements AdminCrudActions.
func CountUserAccounts(ctx context.Context, db database.Dbx, filter *shared.UserAccountListFilter) (int64, error) {
	where := UserAccountWhere(filter)
	data, err := crudrepo.UserAccount.Count(ctx, db, where)
	if err != nil {
		return 0, err
	}
	return data, nil
}
