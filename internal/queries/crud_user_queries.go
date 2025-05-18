package queries

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

var (
	UserColumnNames = []string{"id", "email", "email_verified_at", "created_at", "updated_at"}
)

func ListUserFilterFunc(filter *shared.UserListFilter) *map[string]any {
	if filter == nil {
		return nil
	}
	where := map[string]any{}
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
		where["accounts"] = map[string]any{
			"provider": map[string]any{
				"_in": filter.Providers,
			},
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
	if len(where) == 0 {
		return nil
	}
	return &where
}

func ListUsers(ctx context.Context, dbx database.Dbx, input *shared.UserListParams) ([]*models.User, error) {
	if input == nil {
		input = &shared.UserListParams{}
	}
	filter := input.UserListFilter
	pageInput := &input.PaginatedInput

	where := ListUserFilterFunc(&filter)
	orderBy := ListUsersOrderByFunc(input)

	limit, offset := database.PaginateRepo(pageInput)
	data, err := crudrepo.User.Get(
		ctx,
		dbx,
		where,
		orderBy,
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

func ListUsersOrderByFunc(input *shared.UserListParams) *map[string]string {
	if input == nil || input.SortBy == "" || input.SortOrder == "" {
		return nil
	}
	order := make(map[string]string)

	if slices.Contains(UserColumnNames, input.SortBy) {
		order[input.SortBy] = strings.ToUpper(input.SortOrder)
	}

	return &order
}

// CountUsers returns the number of users in the database that match the given filter.
//
// If the filter is nil, it returns the total number of users in the database.
//
// The method returns an error if the count operation fails.
func CountUsers(ctx context.Context, db database.Dbx, filter *shared.UserListFilter) (int64, error) {
	where := ListUserFilterFunc(filter)
	data, err := crudrepo.User.Count(ctx, db, where)
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

func DeleteUsers(ctx context.Context, db database.Dbx, userId uuid.UUID) error {
	_, err := crudrepo.User.Delete(
		ctx,
		db,
		&map[string]any{
			"id": map[string]any{
				"_eq": userId.String(),
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

// UpdateUser updates an existing user by id.
//
// It only updates the email, name, image, and email_verified_at fields.
//
// If the user is not found, it returns an error.
//
// It returns an error if the update fails.
func UpdateUser(ctx context.Context, db database.Dbx, userId uuid.UUID, input *shared.UserMutationInput) error {
	user, err := crudrepo.User.GetOne(
		ctx,
		db,
		&map[string]any{
			"id": map[string]any{
				"_eq": userId.String(),
			},
		},
	)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	user.Email = input.Email
	user.Name = input.Name
	user.Image = input.Image
	user.EmailVerifiedAt = input.EmailVerifiedAt
	_, err = crudrepo.User.PutOne(
		ctx,
		db,
		user,
	)
	if err != nil {
		return err
	}

	return nil
}
