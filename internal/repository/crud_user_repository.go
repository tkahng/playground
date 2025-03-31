package repository

import (
	"context"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
)

// func CreateUser(ctx context.Context, db bob.Executor, params *shared.AuthenticateUserParams) (*models.User, error) {

func ListUserFilterFunc(ctx context.Context, q *psql.ViewQuery[*models.User, models.UserSlice], filter shared.UserListFilter) {
	if len(filter.Providers) > 0 {
		q.Apply(
			models.SelectJoins.Users.InnerJoin.UserAccounts(ctx),
			models.SelectWhere.UserAccounts.Provider.In(filter.Providers...),
			// models.SelectWhere.UserAccounts.Provider.EQ(filter.Provider.MustGet()),

		)
	}
	if len(filter.Ids) > 0 {
		q.Apply(
			models.SelectWhere.Users.ID.In(filter.Ids...),
		)
	}
	if len(filter.Emails) > 0 {
		q.Apply(
			models.SelectWhere.Users.Email.In(filter.Emails...),
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
