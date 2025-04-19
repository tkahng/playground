package repository

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/google/uuid"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
)

var (
	UserColumnNames = models.Users.Columns().Names()
)

// func CreateUser(ctx context.Context, db bob.Executor, params *shared.AuthenticateUserParams) (*models.User, error) {

func ListUserFilterFunc(ctx context.Context, q *psql.ViewQuery[*models.User, models.UserSlice], filter *shared.UserListFilter) {
	if filter == nil {
		return
	}
	if len(filter.Providers) > 0 {
		q.Apply(
			models.SelectJoins.Users.InnerJoin.UserAccounts(ctx),
			models.SelectWhere.UserAccounts.Provider.In(filter.Providers...),
			// models.SelectWhere.UserAccounts.Provider.EQ(filter.Provider.MustGet()),

		)
	}
	if len(filter.Ids) > 0 {
		var ids = ParseUUIDs(filter.Ids)
		q.Apply(
			models.SelectWhere.Users.ID.In(ids...),
		)
	}
	if len(filter.Emails) > 0 {
		q.Apply(
			models.SelectWhere.Users.Email.In(filter.Emails...),
		)
	}

	// if len(filter.PermissionIds) > 0 {
	// 	var ids = ParseUUIDs(filter.PermissionIds)
	// 	q.Apply(
	// 		models.SelectJoins.Users.InnerJoin.Permissions(ctx),
	// 		models.SelectWhere.Permissions.ID.In(ids...),
	// 	)
	// }
	if len(filter.RoleIds) > 0 {
		var ids = ParseUUIDs(filter.RoleIds)
		q.Apply(
			models.SelectJoins.Users.InnerJoin.Roles(ctx),
			models.SelectWhere.Roles.ID.In(ids...),
		)
	}
}

// ListUsers implements AdminCrudActions.
func ListUsers(ctx context.Context, db bob.Executor, input *shared.UserListParams) (models.UserSlice, error) {

	q := models.Users.Query()
	filter := input.UserListFilter
	pageInput := &input.PaginatedInput

	ViewApplyPagination(q, pageInput)
	ListUsersOrderByFunc(ctx, q, input)
	ListUserFilterFunc(ctx, q, &filter)
	data, err := q.All(ctx, db)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ListUsersOrderByFunc(ctx context.Context, q *psql.ViewQuery[*models.User, models.UserSlice], input *shared.UserListParams) {
	if q == nil {
		return
	}
	if input == nil || input.SortBy == "" {
		q.Apply(
			sm.OrderBy(models.UserColumns.CreatedAt).Desc(),
			sm.OrderBy(models.UserColumns.ID).Desc(),
		)
		return
	}
	if slices.Contains(UserColumnNames, input.SortBy) {
		if input.SortParams.SortOrder == "desc" {
			q.Apply(
				sm.OrderBy(input.SortBy).Desc(),
				sm.OrderBy(models.UserColumns.ID).Desc(),
			)
		} else if input.SortParams.SortOrder == "asc" {
			q.Apply(
				sm.OrderBy(input.SortBy).Asc(),
				sm.OrderBy(models.UserColumns.ID).Asc(),
			)
		}
	}
}

// CountUsers implements AdminCrudActions.
func CountUsers(ctx context.Context, db bob.Executor, filter *shared.UserListFilter) (int64, error) {
	q := models.Users.Query()
	ListUserFilterFunc(ctx, q, filter)
	data, err := q.Count(ctx, db)
	if err != nil {
		return 0, err
	}
	return data, nil
}

// delete users
func DeleteUsers(ctx context.Context, db bob.Executor, userId uuid.UUID) error {
	user, err := models.FindUser(ctx, db, userId)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	return user.Delete(ctx, db)
}

type UpdateUserInput struct {
	Email           string
	Name            *string
	AvatarUrl       *string
	EmailVerifiedAt *time.Time
}

// update users by id
func UpdateUser(ctx context.Context, db bob.Executor, userId uuid.UUID, input *UpdateUserInput) error {
	user, err := models.FindUser(ctx, db, userId)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	return user.Update(ctx, db, &models.UserSetter{
		Email:           omit.From(input.Email),
		Name:            omitnull.FromPtr(input.Name),
		Image:           omitnull.FromPtr(input.AvatarUrl),
		EmailVerifiedAt: omitnull.FromPtr(input.EmailVerifiedAt),
	})
}
