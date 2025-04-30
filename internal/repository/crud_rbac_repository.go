package repository

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
)

func ListPermissionsOrderByFunc(ctx context.Context, q *psql.ViewQuery[*models.Permission, models.PermissionSlice], input *shared.PermissionsListParams) {
	if q == nil {
		return
	}
	if input == nil || input.SortBy == "" {
		q.Apply(
			sm.OrderBy(models.PermissionColumns.CreatedAt).Desc(),
			sm.OrderBy(models.PermissionColumns.ID).Desc(),
		)
		return
	}
	if slices.Contains(models.Permissions.Columns().Names(), input.SortBy) {
		if input.SortParams.SortOrder == "desc" {
			q.Apply(
				sm.OrderBy(input.SortBy).Desc(),
				sm.OrderBy(models.PermissionColumns.ID).Desc(),
			)
		} else if input.SortParams.SortOrder == "asc" || input.SortParams.SortOrder == "" {
			q.Apply(
				sm.OrderBy(input.SortBy).Asc(),
				sm.OrderBy(models.PermissionColumns.ID).Asc(),
			)
		}
		return
	}
}

func ListPermissionsFilterFunc(ctx context.Context, q *psql.ViewQuery[*models.Permission, models.PermissionSlice], filter *shared.PermissionsListFilter) {
	if filter == nil {
		return
	}
	if filter.Q != "" {
		q.Apply(
			psql.WhereOr(models.SelectWhere.Permissions.Name.ILike("%"+filter.Q+"%"),
				models.SelectWhere.Permissions.Description.ILike("%"+filter.Q+"%")),
		)
	}
	if len(filter.Names) > 0 {
		q.Apply(
			models.SelectWhere.Permissions.Name.In(filter.Names...),
		)
	}
	if len(filter.Ids) > 0 {
		var ids []uuid.UUID = ParseUUIDs(filter.Ids)
		q.Apply(
			models.SelectWhere.Permissions.ID.In(ids...),
		)
	}

	if filter.RoleId != "" {
		id, err := uuid.Parse(filter.RoleId)
		if err != nil {
			return
		}
		if filter.RoleReverse {
			q.Apply(
				sm.LeftJoin(models.RolePermissions.NameAs()).On(
					models.PermissionColumns.ID.EQ(models.RolePermissionColumns.PermissionID),
					models.RolePermissionColumns.RoleID.EQ(psql.Arg(id)),
				),
				sm.Where(models.RolePermissionColumns.PermissionID.IsNull()),
			)
		} else {
			q.Apply(
				models.SelectJoins.Permissions.InnerJoin.Roles(ctx),
				models.SelectWhere.Roles.ID.EQ(id),
			)
		}
	}
}

// ListPermissions implements AdminCrudActions.
func ListPermissions(ctx context.Context, db bob.Executor, input *shared.PermissionsListParams) ([]*models.Permission, error) {
	q := models.Permissions.Query()
	filter := input.PermissionsListFilter
	pageInput := &input.PaginatedInput

	ViewApplyPagination(q, pageInput)
	ListPermissionsOrderByFunc(ctx, q, input)
	ListPermissionsFilterFunc(ctx, q, &filter)
	query, args, err := q.Build(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Println(query, args)
	data, err := q.All(ctx, db)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CountPermissions implements AdminCrudActions.
func CountPermissions(ctx context.Context, db bob.Executor, filter *shared.PermissionsListFilter) (int64, error) {
	q := models.Permissions.Query()
	ListPermissionsFilterFunc(ctx, q, filter)
	return CountExec(ctx, db, q)
}
func ListRolesFilterFunc(ctx context.Context, q *psql.ViewQuery[*models.Role, models.RoleSlice], filter *shared.RoleListFilter) {
	if filter == nil {
		return
	}
	if filter.Q != "" {
		q.Apply(
			psql.WhereOr(models.SelectWhere.Roles.Name.ILike("%"+filter.Q+"%"),
				models.SelectWhere.Roles.Description.ILike("%"+filter.Q+"%")),
		)
	}
	if len(filter.Names) > 0 {
		q.Apply(
			models.SelectWhere.Roles.Name.In(filter.Names...),
		)
	}
	if len(filter.Ids) > 0 {
		var ids []uuid.UUID = ParseUUIDs(filter.Ids)
		q.Apply(
			models.SelectWhere.Roles.ID.In(ids...),
		)
	}

	if filter.UserId != "" {
		id, err := uuid.Parse(filter.UserId)
		if err != nil {
			return
		}
		if filter.Reverse == "user" {
			q.Apply(
				sm.LeftJoin(models.UserRoles.NameAs()).On(
					models.RoleColumns.ID.EQ(models.UserRoleColumns.RoleID),
					models.UserRoleColumns.UserID.EQ(psql.Arg(id)),
				),
				sm.Where(models.UserRoleColumns.RoleID.IsNull()),
			)
		} else {
			q.Apply(
				models.SelectJoins.Roles.InnerJoin.Users(ctx),
				models.SelectWhere.Users.ID.EQ(id),
			)
		}
	}
	if filter.ProductId != "" {
		if filter.Reverse == "product" {
			q.Apply(
				sm.LeftJoin(models.ProductRoles.NameAs()).On(
					models.RoleColumns.ID.EQ(models.ProductRoleColumns.RoleID),
					models.ProductRoleColumns.ProductID.EQ(psql.Arg(filter.ProductId)),
				),
				sm.Where(models.ProductRoleColumns.RoleID.IsNull()),
			)
		} else {
			q.Apply(
				models.SelectJoins.Roles.InnerJoin.StripeProducts(ctx),
				models.SelectWhere.StripeProducts.ID.EQ(filter.ProductId),
			)
		}
	}
}

func ListRolesOrderByFunc(ctx context.Context, q *psql.ViewQuery[*models.Role, models.RoleSlice], input *shared.RolesListParams) {
	if q == nil {
		return
	}
	if input == nil || input.SortBy == "" {
		q.Apply(
			sm.OrderBy(models.RoleColumns.CreatedAt).Desc(),
			sm.OrderBy(models.RoleColumns.ID).Desc(),
		)
		return
	}
	if slices.Contains(models.Roles.Columns().Names(), input.SortBy) {
		if input.SortParams.SortOrder == "desc" {
			q.Apply(
				sm.OrderBy(input.SortBy).Desc(),
				sm.OrderBy(models.RoleColumns.ID).Desc(),
			)
		} else if input.SortParams.SortOrder == "asc" || input.SortParams.SortOrder == "" {
			q.Apply(
				sm.OrderBy(input.SortBy).Asc(),
				sm.OrderBy(models.RoleColumns.ID).Asc(),
			)
		}
		return
	}
}

// ListRoles implements AdminCrudActions.
func ListRoles(ctx context.Context, db bob.Executor, input *shared.RolesListParams) (models.RoleSlice, error) {
	q := models.Roles.Query()
	filter := input.RoleListFilter
	pageInput := &input.PaginatedInput

	ViewApplyPagination(q, pageInput)
	ListRolesOrderByFunc(ctx, q, input)
	ListRolesFilterFunc(ctx, q, &filter)
	data, err := q.All(ctx, db)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CountRoles implements AdminCrudActions.
func CountRoles(ctx context.Context, db bob.Executor, filter *shared.RoleListFilter) (int64, error) {
	q := models.Roles.Query()
	ListRolesFilterFunc(ctx, q, filter)
	data, err := q.Count(ctx, db)
	if err != nil {
		return 0, err
	}
	return data, nil
}
