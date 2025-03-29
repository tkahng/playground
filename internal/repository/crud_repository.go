package repository

import (
	"context"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
)

func ListUserFilterFunc(ctx context.Context, q *psql.ViewQuery[*models.User, models.UserSlice], filter shared.UserListFilter) {
	if len(string(filter.Provider)) > 0 {
		q.Apply(
			models.SelectJoins.Users.InnerJoin.UserAccounts(ctx),
			models.SelectWhere.UserAccounts.Provider.EQ(filter.Provider),
			// models.SelectWhere.UserAccounts.Provider.EQ(filter.Provider.MustGet()),

		)
	}
}

// ListUsers implements AdminCrudActions.
func ListUsers(ctx context.Context, db bob.DB, input *shared.UserListParams) ([]*models.User, error) {

	q := models.Users.Query()
	filter := input.UserListFilter
	pageInput := &input.PaginatedInput

	ViewApplyPagination(q, pageInput)

	ListUserFilterFunc(ctx, q, filter)
	data, err := q.All(ctx, db)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CountUsers implements AdminCrudActions.
func CountUsers(ctx context.Context, db bob.DB, filter shared.UserListFilter) (int64, error) {
	q := models.Users.Query()
	ListUserFilterFunc(ctx, q, filter)
	data, err := q.Count(ctx, db)
	if err != nil {
		return 0, err
	}
	return data, nil
}

// ListUserAccounts implements AdminCrudActions.
func ListUserAccounts(ctx context.Context, db bob.DB, input *shared.PaginatedInput) ([]*models.UserAccount, error) {
	q := models.UserAccounts.Query()
	pageInput := input
	ViewApplyPagination(q, pageInput)
	data, err := q.All(ctx, db)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CountUserAccounts implements AdminCrudActions.
func CountUserAccounts(ctx context.Context, db bob.DB) (int64, error) {
	q := models.UserAccounts.Query()
	return CountExec(ctx, db, q)
}

// ListSessions implements AdminCrudActions.
func ListSessions(ctx context.Context, db bob.DB, input *shared.PaginatedInput) ([]*models.UserSession, error) {
	panic("unimplemented")
}

// CountSessions implements AdminCrudActions.
func CountSessions(ctx context.Context, db bob.DB) (int64, error) {
	q := models.UserSessions.Query()
	return CountExec(ctx, db, q)
}

// ListPermissions implements AdminCrudActions.
func ListPermissions(ctx context.Context, db bob.DB, input *shared.PaginatedInput) ([]models.Permission, error) {
	panic("unimplemented")
}

// CountPermissions implements AdminCrudActions.
func CountPermissions(ctx context.Context, db bob.DB) (int64, error) {
	q := models.Permissions.Query()
	return CountExec(ctx, db, q)
}

// ListRoles implements AdminCrudActions.
func ListRoles(ctx context.Context, db bob.DB, input *shared.PaginatedInput) ([]models.Role, error) {
	panic("unimplemented")
}

// CountRoles implements AdminCrudActions.
func CountRoles(ctx context.Context, db bob.DB) (int64, error) {
	q := models.Roles.Query()
	return CountExec(ctx, db, q)
}

// CountTokens implements AdminCrudActions.
func CountTokens(ctx context.Context, db bob.DB) (int64, error) {
	q := models.Tokens.Query()
	return CountExec(ctx, db, q)
}

// ListTokens implements AdminCrudActions.
func ListTokens(ctx context.Context, db bob.DB, input *shared.PaginatedInput) ([]*models.Token, error) {
	panic("unimplemented")
}
