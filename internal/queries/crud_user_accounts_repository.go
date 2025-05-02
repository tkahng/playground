package queries

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
)

var (
	// UserColumnNames = models.Users.Columns().Names()
	UserAccountColumnNames = models.UserAccounts.Columns().Names()
	// UserAccountColumnNames = models.UserAccounts.Columns().Names()
)

// ListUserAccounts implements AdminCrudActions.
// ListUsers implements AdminCrudActions.
func ListUserAccounts(ctx context.Context, db Queryer, input *shared.UserAccountListParams) (models.UserAccountSlice, error) {

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

// func CreateUser(ctx context.Context, db Queryer, params *shared.AuthenticateUserParams) (*models.User, error) {

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
	if filter.UserId != "" {
		id, err := uuid.Parse(filter.UserId)
		if err != nil {
			return
		}
		q.Apply(
			models.SelectWhere.UserAccounts.UserID.EQ(id),
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
