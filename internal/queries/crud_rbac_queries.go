package queries

import (
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/tkahng/authgo/internal/crud/crudModels"
	"github.com/tkahng/authgo/internal/shared"
)

var (
	PermissionColumnNames = []string{"id", "name", "description", "created_at", "updated_at"}
)

func ListPermissionsFilterFunc(sq squirrel.SelectBuilder, filter *shared.PermissionsListFilter) squirrel.SelectBuilder {
	// where := make(map[string]any)
	if filter == nil {
		return sq
	}
	if filter.Q != "" {
		sq = sq.Where(
			squirrel.Or{
				squirrel.ILike{"name": "%" + filter.Q + "%"},
				squirrel.ILike{"description": "%" + filter.Q + "%"},
			},
		)

	}
	if len(filter.Names) > 0 {
		sq = sq.Where(squirrel.Eq{"name": filter.Names})
	}
	if len(filter.Ids) > 0 {
		sq = sq.Where(squirrel.Eq{"id": filter.Ids})
	}

	if filter.RoleId != "" {

		if filter.RoleReverse {
			sq = sq.LeftJoin(
				"role_permissions"+" on "+"permissions.id"+" = "+"role_permissions"+"."+"permission_id"+" and "+"role_permissions"+"."+"role_id"+" = ?",
				filter.RoleId,
			)
			sq = sq.Where("role_permissions.permission_id is null")

		} else {
			sq = sq.Join("role_permissions on permissions.id = role_permissions.permission_id and role_permissions.role_id = ?", filter.RoleId).
				Where(squirrel.Eq{"role_permissions.role_id": filter.RoleId})

		}
	}
	return sq
}
func ListPermissionsFilterFunc2(filter *shared.PermissionsListFilter) *map[string]any {
	where := make(map[string]any)
	if filter == nil {
		return nil
	}
	if filter.Q != "" {
		where["_or"] = []map[string]any{
			{
				"name": map[string]any{
					"_ilike": fmt.Sprintf("%%%s%%", filter.Q),
				},
			},
			{
				"description": map[string]any{
					"_ilike": fmt.Sprintf("%%%s%%", filter.Q),
				},
			},
		}
	}
	if len(filter.Names) > 0 {
		where["name"] = map[string]any{
			"_in": filter.Names,
		}
	}
	if len(filter.Ids) > 0 {
		where["id"] = map[string]any{
			"_in": filter.Ids,
		}
	}

	if filter.RoleId != "" {
		// id, err := uuid.Parse(filter.RoleId)
		// if err != nil {
		// 	return &where
		// }
		if filter.RoleReverse {
			// q.Apply(
			// 	sm.LeftJoin(models.RolePermissions.NameAs()).On(
			// 		models.PermissionColumns.ID.EQ(models.RolePermissionColumns.PermissionID),
			// 		models.RolePermissionColumns.RoleID.EQ(psql.Arg(id)),
			// 	),
			// 	sm.Where(models.RolePermissionColumns.PermissionID.IsNull()),
			// )
		} else {
			// q.Apply(
			// 	models.SelectJoins.Permissions.InnerJoin.Roles(ctx),
			// 	models.SelectWhere.Roles.ID.EQ(id),
			// )
			where["role_id"] = map[string]any{
				"_eq": filter.RoleId,
			}
		}
	}
	return &where
}

// ListPermissions implements AdminCrudActions.
func ListPermissions(ctx context.Context, db Queryer, input *shared.PermissionsListParams) ([]*crudModels.Permission, error) {
	q := squirrel.Select("permissions.*").From("permissions")
	filter := input.PermissionsListFilter
	pageInput := &input.PaginatedInput

	// q = ViewApplyPagination(q, pageInput)
	q = ListPermissionsFilterFunc(q, &filter)
	q = Paginate(q, pageInput)
	if input.SortBy != "" && input.SortOrder != "" {
		q = q.OrderBy(input.SortBy + " " + strings.ToUpper(input.SortOrder))
	}
	data, err := QueryWithBuilder[*crudModels.Permission](ctx, db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return data, nil
}

type CountOutput struct {
	Count int64
}

// CountPermissions implements AdminCrudActions.
func CountPermissions(ctx context.Context, db Queryer, filter *shared.PermissionsListFilter) (int64, error) {
	q := squirrel.Select("COUNT(permissions.*)").From("permissions")

	// q = ViewApplyPagination(q, pageInput)
	q = ListPermissionsFilterFunc(q, filter)

	data, err := QueryWithBuilder[CountOutput](ctx, db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil
	}

	return data[0].Count, nil
}
func ListRolesFilterFuncQuery(sq squirrel.SelectBuilder, filter *shared.RoleListFilter) squirrel.SelectBuilder {
	// where := make(map[string]any)
	if filter == nil {
		return sq
	}
	if filter.Q != "" {
		sq = sq.Where(
			squirrel.Or{
				squirrel.ILike{"name": "%" + filter.Q + "%"},
				squirrel.ILike{"description": "%" + filter.Q + "%"},
			},
		)

	}
	if len(filter.Names) > 0 {
		sq = sq.Where(squirrel.Eq{"name": filter.Names})
	}
	if len(filter.Ids) > 0 {
		sq = sq.Where(squirrel.Eq{"id": filter.Ids})
	}

	if filter.UserId != "" {

		if filter.Reverse == "user" {
			sq = sq.LeftJoin(
				"user_roles"+" on "+"roles.id"+" = "+"user_roles"+"."+"role_id"+" and "+"user_roles"+"."+"user_id"+" = ?",
				filter.UserId,
			)
			sq = sq.Where("user_roles.role_id is null")

		} else {
			sq = sq.Join("user_roles on roles.id = user_roles.role_id").
				Where(squirrel.Eq{"user_roles.user_id": filter.UserId})

		}
	}
	return sq
}

// ListRoles implements AdminCrudActions.
func ListRoles(ctx context.Context, db Queryer, input *shared.RolesListParams) ([]*crudModels.Role, error) {
	q := squirrel.Select("roles.*").From("roles")
	filter := input.RoleListFilter
	pageInput := &input.PaginatedInput

	// q = ViewApplyPagination(q, pageInput)
	q = ListRolesFilterFuncQuery(q, &filter)
	q = Paginate(q, pageInput)
	if input.SortBy != "" && input.SortOrder != "" {
		q = q.OrderBy(input.SortBy + " " + strings.ToUpper(input.SortOrder))
	}
	data, err := QueryWithBuilder[*crudModels.Role](ctx, db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CountRoles implements AdminCrudActions.
func CountRoles(ctx context.Context, db Queryer, filter *shared.RoleListFilter) (int64, error) {
	q := squirrel.Select("COUNT(roles.*)").From("roles")

	// q = ViewApplyPagination(q, pageInput)
	q = ListRolesFilterFuncQuery(q, filter)

	data, err := QueryWithBuilder[CountOutput](ctx, db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil
	}

	return data[0].Count, nil
}
