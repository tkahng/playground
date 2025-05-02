package queries

import (
	"context"
	"errors"
	"slices"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/google/uuid"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/tkahng/authgo/internal/crud/crudModels"
	"github.com/tkahng/authgo/internal/crud/crudrepo"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
)

var (
	UserColumnNames = models.Users.Columns().Names()
)

func ListUserFilterFunc(ctx context.Context, q *psql.ViewQuery[*models.User, models.UserSlice], filter *shared.UserListFilter) {
	if filter == nil {
		return
	}
	if filter.EmailVerified != "" {
		if filter.EmailVerified == shared.Verified {
			q.Apply(
				models.SelectWhere.Users.EmailVerifiedAt.IsNotNull(),
			)
		} else if filter.EmailVerified == shared.UnVerified {
			q.Apply(
				models.SelectWhere.Users.EmailVerifiedAt.IsNull(),
			)
		}
	}
	if len(filter.Providers) > 0 {
		var providers []models.Providers
		for _, p := range filter.Providers {
			providers = append(providers, shared.ToModelProvider(p))
		}
		q.Apply(
			models.SelectJoins.Users.InnerJoin.UserAccounts(ctx),
			models.SelectWhere.UserAccounts.Provider.In(providers...),
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

	if len(filter.RoleIds) > 0 {
		var ids = ParseUUIDs(filter.RoleIds)
		q.Apply(
			models.SelectJoins.Users.InnerJoin.Roles(ctx),
			models.SelectWhere.Roles.ID.In(ids...),
		)
	}
}

func ListUsers(ctx context.Context, db Queryer, input *shared.UserListParams) ([]*crudModels.User, error) {

	// q := models.Users.Query()
	filter := input.UserListFilter
	pageInput := &input.PaginatedInput

	where := map[string]any{}
	orderBy := map[string]string{}
	if filter.EmailVerified != "" {
		if filter.EmailVerified == shared.Verified {
			where["email_verified_at"] = map[string]any{
				"_neq": nil,
			}
		} else if filter.EmailVerified == shared.UnVerified {
			where["email_verified_at"] = map[string]any{
				"_eq": nil,
			}
		}
	}
	if len(filter.Emails) > 0 {
		where["email"] = map[string]any{
			"_in": filter.Emails,
		}
	}
	if len(filter.Ids) > 0 {
		where["id"] = map[string]any{
			"_in": filter.Ids,
		}
	}
	if len(filter.Providers) > 0 {
		var providers []models.Providers
		for _, p := range filter.Providers {
			providers = append(providers, shared.ToModelProvider(p))
		}
		where["provider"] = map[string]any{
			"_in": providers,
		}
	}
	if len(filter.RoleIds) > 0 {
		where["roles"] = map[string]any{
			"id": map[string]any{
				"_in": filter.RoleIds,
			},
		}
	}
	if filter.Q != "" {
		where["_or"] = []map[string]any{
			{
				"email": map[string]any{
					"_ilike": "%" + filter.Q + "%",
				},
			},
			{
				"name": map[string]any{
					"_ilike": "%" + filter.Q + "%",
				},
			},
		}
	}
	if input.SortBy != "" {
		orderBy[input.SortBy] = input.SortOrder
	}
	limit, offset := PaginateRepo(pageInput)
	data, err := crudrepo.User.Get(
		ctx,
		db,
		&where,
		&orderBy,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ListUsersOrderByFunc applies sorting to the user query based on the input parameters.
// It first checks if the query or input is nil, returning early if either is true.
// If no specific sorting column is provided, it defaults to sorting by the CreatedAt
// and ID columns in descending order. If a valid SortBy column is specified in the input
// and exists in UserColumnNames, it applies the specified sorting order (either ascending
// or descending) to that column, followed by sorting by the ID column.

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

// CountUsers returns the number of users in the database that match the given filter.
//
// If the filter is nil, it returns the total number of users in the database.
//
// The method returns an error if the count operation fails.
func CountUsers(ctx context.Context, db Queryer, filter *shared.UserListFilter) (int64, error) {
	q := models.Users.Query()
	ListUserFilterFunc(ctx, q, filter)
	data, err := q.Count(ctx, db)
	if err != nil {
		return 0, err
	}
	return data, nil
}

// DeleteUsers deletes the user with the given ID.
//
// It first finds the user and checks if it exists. If the user does not exist,
// it returns an error. If the user exists, it calls the user's Delete method
// to delete the user.
//
// The method returns an error if the user could not be deleted.

func DeleteUsers(ctx context.Context, db Queryer, userId uuid.UUID) error {
	user, err := models.FindUser(ctx, db, userId)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	return user.Delete(ctx, db)
}

// UpdateUser updates an existing user by id.
//
// It only updates the email, name, image, and email_verified_at fields.
//
// If the user is not found, it returns an error.
//
// It returns an error if the update fails.
func UpdateUser(ctx context.Context, db Queryer, userId uuid.UUID, input *shared.UserMutationInput) error {
	q := models.Users.Update(
		models.UpdateWhere.Users.ID.EQ(userId),
		models.UserSetter{
			Email:           omit.From(input.Email),
			Name:            omitnull.FromPtr(input.Name),
			Image:           omitnull.FromPtr(input.Image),
			EmailVerifiedAt: omitnull.FromPtr(input.EmailVerifiedAt),
		}.UpdateMod(),
	)
	_, err := q.Exec(ctx, db)
	if err != nil {
		return err
	}
	return nil
}
