package stores

import (
	"context"
	"slices"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/utils"
)

type PermissionFilter struct {
	PaginatedInput
	SortParams
	Q           string      `query:"q,omitempty" required:"false"`
	Ids         []uuid.UUID `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Names       []string    `query:"names,omitempty" required:"false" minimum:"1" maximum:"100"`
	RoleId      uuid.UUID   `query:"role_id,omitempty" required:"false" format:"uuid"`
	RoleReverse bool        `query:"role_reverse,omitempty" required:"false" doc:"When role_id is provided, if this is true, it will return the permissions that the role does not have"`
}

func (p *DbRbacStore) ListPermissions(ctx context.Context, input *PermissionFilter) ([]*models.Permission, error) {
	q := squirrel.Select("permissions.*").From("permissions")

	// q = ViewApplyPagination(q, pageInput)
	q = ListPermissionsFilterFunc(q, input)
	q = queryPagination(q, input)
	if input.SortBy != "" && slices.Contains(repository.PermissionBuilder.ColumnNames(), utils.Quote(input.SortBy)) {
		q = q.OrderBy(input.SortBy + " " + strings.ToUpper(input.SortOrder))
	}
	data, err := database.QueryWithBuilder[*models.Permission](ctx, p.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CountPermissions implements AdminCrudActions.
func (p *DbRbacStore) CountPermissions(ctx context.Context, filter *PermissionFilter) (int64, error) {
	q := squirrel.Select("COUNT(permissions.*)").From("permissions")

	// q = ViewApplyPagination(q, pageInput)
	q = ListPermissionsFilterFunc(q, filter)

	data, err := database.QueryWithBuilder[database.CountOutput](ctx, p.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil
	}

	return data[0].Count, nil
}
func ListPermissionsFilterFunc(sq squirrel.SelectBuilder, filter *PermissionFilter) squirrel.SelectBuilder {

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

	if filter.RoleId != uuid.Nil {
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

func (p *DbRbacStore) FindPermission(ctx context.Context, filter *PermissionFilter) (*models.Permission, error) {
	q := squirrel.Select("permissions.*").From("permissions")
	q = ListPermissionsFilterFunc(q, filter)
	q = q.Limit(1)
	data, err := database.QueryWithBuilder[*models.Permission](ctx, p.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return data[0], nil
}

// FindPermissionByName implements RBACStore.
func (p *DbRbacStore) FindPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	data, err := repository.Permission.GetOne(
		ctx,
		p.db,
		&map[string]any{
			models.PermissionTable.Name: map[string]any{
				"_eq": name,
			},
		},
	)
	return database.OptionalRow(data, err)
}

func (a *DbRbacStore) FindPermissionById(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	data, err := repository.Permission.GetOne(
		ctx,
		a.db,
		&map[string]any{
			models.PermissionTable.ID: map[string]any{
				"_eq": id,
			},
		},
	)
	return database.OptionalRow(data, err)
}

func (p *DbRbacStore) FindPermissionsByIds(ctx context.Context, params []uuid.UUID) ([]*models.Permission, error) {
	if len(params) == 0 {
		return nil, nil
	}
	newIds := make([]string, len(params))
	for i, id := range params {
		newIds[i] = id.String()
	}
	return repository.Permission.Get(
		ctx,
		p.db,
		&map[string]any{
			models.PermissionTable.ID: map[string]any{
				"_in": newIds,
			},
		},
		&map[string]string{
			models.PermissionTable.Name: "asc",
		},
		nil,
		nil,
	)
}

func (p *DbRbacStore) FindOrCreatePermission(ctx context.Context, permissionName string) (*models.Permission, error) {
	permission, err := repository.Permission.GetOne(
		ctx,
		p.db,
		&map[string]any{
			models.PermissionTable.Name: map[string]any{
				"_eq": permissionName,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		permission, err = p.CreatePermission(ctx, permissionName, nil)
		if err != nil {
			return nil, err
		}
	}
	return permission, nil
}

func (p *DbRbacStore) CreatePermission(ctx context.Context, name string, description *string) (*models.Permission, error) {
	data, err := repository.Permission.PostOne(ctx, p.db, &models.Permission{
		Name:        name,
		Description: description,
	})
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (p *DbRbacStore) UpdatePermission(ctx context.Context, id uuid.UUID, roledto *shared.UpdatePermissionDto) error {
	permission, err := repository.Permission.GetOne(
		ctx,
		p.db,
		&map[string]any{
			models.PermissionTable.ID: map[string]any{
				"_eq": id,
			},
		},
	)
	if err != nil {
		return err
	}
	if permission == nil {
		return nil
	}
	permission.Name = roledto.Name
	permission.Description = roledto.Description
	_, err = repository.Permission.PutOne(ctx, p.db, permission)

	if err != nil {
		return err
	}
	return nil
}

func (p *DbRbacStore) CreateRolePermissions(ctx context.Context, roleId uuid.UUID, permissionIds ...uuid.UUID) error {
	var permissions []models.RolePermission
	for _, perm := range permissionIds {
		permissions = append(permissions, models.RolePermission{
			RoleID:       roleId,
			PermissionID: perm,
		})
	}
	_, err := repository.RolePermission.Post(ctx, p.db, permissions)
	if err != nil {
		return err
	}
	return nil
}

func (p *DbRbacStore) LoadRolePermissions(ctx context.Context, roleIds ...uuid.UUID) ([][]*models.Permission, error) {
	const (
		GetRolePermissionsQuery = `
		SELECT rp.role_id as key,
			COALESCE(
					json_agg(
							jsonb_build_object(
									'id',
									p.id,
									'name',
									p.name,
									'description',
									p.description,
									'created_at',
									p.created_at,
									'updated_at',
									p.updated_at
							)
					) FILTER (
							WHERE p.id IS NOT NULL
					),
					'[]'
			) AS data
	FROM public.role_permissions rp
			LEFT JOIN public.permissions p ON p.id = rp.permission_id
			WHERE rp.role_id = ANY (
					$1::uuid []
			)
	GROUP BY rp.role_id;`
	)

	data, err := database.QueryAll[shared.JoinedResult[*models.Permission, uuid.UUID]](
		ctx,
		p.db,
		GetRolePermissionsQuery,
		roleIds,
	)
	if err != nil {
		return nil, err
	}
	return mapper.Map(mapper.MapTo(data, roleIds, func(a shared.JoinedResult[*models.Permission, uuid.UUID]) uuid.UUID {
		return a.Key
	}), func(a *shared.JoinedResult[*models.Permission, uuid.UUID]) []*models.Permission {
		if a == nil {
			return nil
		}
		return a.Data
	}), nil
}

func (p *DbRbacStore) DeletePermission(ctx context.Context, id uuid.UUID) error {
	_, err := repository.Permission.Delete(
		ctx,
		p.db,
		&map[string]any{
			models.PermissionTable.ID: map[string]any{
				"_eq": id,
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (p *DbRbacStore) DeleteRolePermissions(ctx context.Context, roleId uuid.UUID, permissionIds ...uuid.UUID) error {
	if len(permissionIds) == 0 {
		return nil
	}
	var ids []string
	for _, id := range permissionIds {
		ids = append(ids, id.String())
	}
	_, err := repository.RolePermission.Delete(
		ctx,
		p.db,
		&map[string]any{
			"role_id": map[string]any{
				"_eq": roleId,
			},
			"permission_id": map[string]any{
				"_in": ids,
			},
		},
	)
	return err
}
