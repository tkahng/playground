package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
)

func ListPermissionsFilterFunc(ctx context.Context, q *psql.ViewQuery[*models.Permission, models.PermissionSlice], filter *shared.PermissionsListFilter) {
	if filter == nil {
		return
	}
	if len(filter.Names) > 0 {
		q.Apply(
			models.SelectWhere.Permissions.Name.In(filter.Names...),
		)
	}
	if len(filter.Ids) > 0 {
		var ids []uuid.UUID
		for _, id := range filter.Ids {
			parsed, err := uuid.Parse(id)
			if err != nil {
				continue
			}
			ids = append(ids, parsed)
		}
		q.Apply(
			models.SelectWhere.Permissions.ID.In(ids...),
		)
	}

	if len(filter.RoleIds) > 0 {
		var ids []uuid.UUID
		for _, id := range filter.RoleIds {
			parsed, err := uuid.Parse(id)
			if err != nil {
				continue
			}
			ids = append(ids, parsed)
		}
		q.Apply(
			models.SelectJoins.Permissions.InnerJoin.Roles(ctx),
			models.SelectWhere.Roles.ID.In(ids...),
		)
	}
}

// ListPermissions implements AdminCrudActions.
func ListPermissions(ctx context.Context, db bob.DB, input *shared.PermissionsListParams) ([]*models.Permission, error) {
	q := models.Permissions.Query()
	filter := input.PermissionsListFilter
	pageInput := &input.PaginatedInput

	ViewApplyPagination(q, pageInput)
	ListPermissionsFilterFunc(ctx, q, &filter)
	data, err := q.All(ctx, db)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CountPermissions implements AdminCrudActions.
func CountPermissions(ctx context.Context, db bob.DB, filter *shared.PermissionsListFilter) (int64, error) {
	q := models.Permissions.Query()
	return CountExec(ctx, db, q)
}
func ListRolesFilterFunc(ctx context.Context, q *psql.ViewQuery[*models.Role, models.RoleSlice], filter *shared.RoleListFilter) {
	if filter == nil {
		return
	}
	if len(filter.Names) > 0 {
		q.Apply(
			models.SelectWhere.Roles.Name.In(filter.Names...),
		)
	}
	if len(filter.Ids) > 0 {
		var ids []uuid.UUID
		for _, id := range filter.Ids {
			parsed, err := uuid.Parse(id)
			if err != nil {
				continue
			}
			ids = append(ids, parsed)
		}
		q.Apply(
			models.SelectWhere.Roles.ID.In(ids...),
		)
	}

	if len(filter.PermissionIds) > 0 {
		var ids []uuid.UUID
		for _, id := range filter.PermissionIds {
			parsed, err := uuid.Parse(id)
			if err != nil {
				continue
			}
			ids = append(ids, parsed)
		}
		q.Apply(
			models.SelectJoins.Roles.InnerJoin.Permissions(ctx),
			models.SelectWhere.Permissions.ID.In(ids...),
		)
	}
}

// ListRoles implements AdminCrudActions.
func ListRoles(ctx context.Context, db bob.DB, input *shared.RolesListParams) (models.RoleSlice, error) {
	q := models.Roles.Query()
	filter := input.RoleListFilter
	pageInput := &input.PaginatedInput

	ViewApplyPagination(q, pageInput)
	ListRolesFilterFunc(ctx, q, &filter)
	data, err := q.All(ctx, db)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CountRoles implements AdminCrudActions.
func CountRoles(ctx context.Context, db bob.DB, filter *shared.RoleListFilter) (int64, error) {
	q := models.Roles.Query()
	ListRolesFilterFunc(ctx, q, filter)
	data, err := q.Count(ctx, db)
	if err != nil {
		return 0, err
	}
	return data, nil
}
