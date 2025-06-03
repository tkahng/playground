package stores

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/stephenafamo/scan"
	"github.com/stephenafamo/scan/pgxscan"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/types"
)

type PostgresRBACStore struct {
	db database.Dbx
}

// DeleteProductRoles implements services.RBACStore.
func (s *PostgresRBACStore) DeleteProductRoles(ctx context.Context, productId string, roleIds ...uuid.UUID) error {
	if len(roleIds) == 0 {
		return nil
	}
	newIds := make([]string, len(roleIds))
	for i, id := range roleIds {
		newIds[i] = id.String()
	}
	_, err := crudrepo.ProductRole.Delete(
		ctx,
		s.db,
		&map[string]any{
			"product_id": map[string]any{
				"_eq": productId,
			},
			"role_id": map[string]any{
				"_in": newIds,
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

type RBACStore struct {
	*PostgresRBACStore
}

var _ services.RBACStore = &PostgresRBACStore{}

func NewPostgresRBACStore(db database.Dbx) *PostgresRBACStore {
	return &PostgresRBACStore{
		db: db,
	}
}

func (s *PostgresRBACStore) FindRolesByIds(ctx context.Context, params []uuid.UUID) ([]*models.Role, error) {
	if len(params) == 0 {
		return nil, nil
	}
	newIds := make([]string, len(params))
	for i, id := range params {
		newIds[i] = id.String()
	}
	return crudrepo.Role.Get(
		ctx,
		s.db,
		&map[string]any{
			"id": map[string]any{
				"_in": newIds,
			},
		},
		&map[string]string{
			"name": "asc",
		},
		nil,
		nil,
	)
}

func (s *PostgresRBACStore) CreateProductRoles(ctx context.Context, productId string, roleIds ...uuid.UUID) error {
	var roles []models.ProductRole
	for _, role := range roleIds {
		roles = append(roles, models.ProductRole{
			ProductID: productId,
			RoleID:    role,
		})
	}
	_, err := crudrepo.ProductRole.Post(ctx, s.db, roles)
	if err != nil {
		return err
	}
	return nil

}

func (a *PostgresRBACStore) FindPermissionById(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	data, err := crudrepo.Permission.GetOne(
		ctx,
		a.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": id.String(),
			},
		},
	)
	return database.OptionalRow(data, err)
}

func (a *PostgresRBACStore) FindRoleById(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	return crudrepo.Role.GetOne(ctx, a.db, &map[string]any{
		"id": map[string]any{
			"_eq": id.String(),
		},
	})
}

func (a *PostgresRBACStore) FindRoleByName(ctx context.Context, name string) (*models.Role, error) {
	return crudrepo.Role.GetOne(
		ctx,
		a.db,
		&map[string]any{
			"name": map[string]any{
				"_eq": name,
			},
		},
	)
}

// var _ RBACStore = &PostgresRBACStore{}

func (a *PostgresRBACStore) AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error {
	if len(roleNames) > 0 {
		user, err := crudrepo.User.GetOne(
			ctx,
			a.db,
			&map[string]any{
				"id": map[string]any{
					"_eq": userId.String(),
				},
			},
		)
		if err != nil {
			return fmt.Errorf("error finding user while assigning roles: %w", err)
		}
		if user == nil {
			return fmt.Errorf("user not found while assigning roles")
		}
		roles, err := crudrepo.Role.Get(
			ctx,
			a.db,
			&map[string]any{
				"name": map[string]any{
					"_in": roleNames,
				},
			},
			nil,
			types.Pointer(10),
			nil,
		)
		if err != nil {
			return fmt.Errorf("error finding user role while assigning roles: %w", err)
		}
		// if len(roles) > 0 {
		// 	// var rolesIDs []uuid.UUID
		// 	// for _, role := range roles {
		// 	// 	rolesIDs = append(rolesIDs, role.ID)
		// 	// }
		// 	// err = queries.CreateUserRoles(ctx, a.db, user.ID, rolesIDs...)
		// 	// if err != nil {
		// 	// 	return fmt.Errorf("error assigning user role while assigning roles: %w", err)
		// 	// }
		// }
		if len(roles) > 0 {
			var userRoles []models.UserRole
			for _, role := range roles {
				userRoles = append(userRoles, models.UserRole{
					UserID: user.ID,
					RoleID: role.ID,
				})
			}
			_, err = crudrepo.UserRole.Post(ctx, a.db, userRoles)
			if err != nil {
				return fmt.Errorf("error assigning user role while assigning roles: %w", err)
			}
		}
	}
	return nil
}

// CreateProductPermissions implements RBACStore.
func (p *PostgresRBACStore) CreateProductPermissions(ctx context.Context, productId string, permissionIds ...uuid.UUID) error {
	var db database.Dbx = p.db
	var permissions []models.ProductPermission
	for _, permissionId := range permissionIds {
		permissions = append(permissions, models.ProductPermission{
			ProductID:    productId,
			PermissionID: permissionId,
		})
	}
	_, err := crudrepo.ProductPermission.Post(
		ctx,
		db,
		permissions,
	)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresRBACStore) CreateUserPermissions(ctx context.Context, userId uuid.UUID, permissionIds ...uuid.UUID) error {
	var dtos []models.UserPermission
	for _, id := range permissionIds {
		dtos = append(dtos, models.UserPermission{
			UserID:       userId,
			PermissionID: id,
		})
	}
	_, err := crudrepo.UserPermission.Post(
		ctx,
		p.db,
		dtos,
	)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresRBACStore) CreateUserRoles(ctx context.Context, userId uuid.UUID, roleIds ...uuid.UUID) error {
	var dtos []models.UserRole
	for _, id := range roleIds {
		dtos = append(dtos, models.UserRole{
			UserID: userId,
			RoleID: id,
		})
	}
	_, err := crudrepo.UserRole.Post(
		ctx,
		p.db,
		dtos,
	)
	if err != nil {

		return err
	}
	return nil
}

// FindPermissionByName implements RBACStore.
func (p *PostgresRBACStore) FindPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	data, err := crudrepo.Permission.GetOne(
		ctx,
		p.db,
		&map[string]any{
			"name": map[string]any{
				"_eq": name,
			},
		},
	)
	return database.OptionalRow(data, err)
}

func (p *PostgresRBACStore) FindPermissionsByIds(ctx context.Context, params []uuid.UUID) ([]*models.Permission, error) {
	if len(params) == 0 {
		return nil, nil
	}
	newIds := make([]string, len(params))
	for i, id := range params {
		newIds[i] = id.String()
	}
	return crudrepo.Permission.Get(
		ctx,
		p.db,
		&map[string]any{
			"id": map[string]any{
				"_in": newIds,
			},
		},
		&map[string]string{
			"name": "asc",
		},
		nil,
		nil,
	)
}

func (p *PostgresRBACStore) FindOrCreatePermission(ctx context.Context, permissionName string) (*models.Permission, error) {
	permission, err := crudrepo.Permission.GetOne(
		ctx,
		p.db,
		&map[string]any{
			"name": map[string]any{
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

func (p *PostgresRBACStore) CreatePermission(ctx context.Context, name string, description *string) (*models.Permission, error) {
	data, err := crudrepo.Permission.PostOne(ctx, p.db, &models.Permission{
		Name:        name,
		Description: description,
	})
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (p *PostgresRBACStore) UpdatePermission(ctx context.Context, id uuid.UUID, roledto *shared.UpdatePermissionDto) error {
	permission, err := crudrepo.Permission.GetOne(
		ctx,
		p.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": id.String(),
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
	_, err = crudrepo.Permission.PutOne(ctx, p.db, permission)

	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresRBACStore) CreateRole(ctx context.Context, role *shared.CreateRoleDto) (*models.Role, error) {
	if role == nil {
		return nil, fmt.Errorf("role is nil")
	}
	data, err := crudrepo.Role.PostOne(ctx, p.db, &models.Role{
		Name:        role.Name,
		Description: role.Description,
	})
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (p *PostgresRBACStore) UpdateRole(ctx context.Context, id uuid.UUID, roledto *shared.UpdateRoleDto) error {
	role, err := crudrepo.Role.GetOne(
		ctx,
		p.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": id.String(),
			},
		},
	)
	if err != nil {
		return err
	}
	if role == nil {
		return nil
	}
	role.Name = roledto.Name
	role.Description = roledto.Description
	_, err = crudrepo.Role.PutOne(ctx, p.db, role)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresRBACStore) DeleteRole(ctx context.Context, id uuid.UUID) error {
	_, err := crudrepo.Role.Delete(
		ctx,
		p.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": id.String(),
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresRBACStore) CreateRolePermissions(ctx context.Context, roleId uuid.UUID, permissionIds ...uuid.UUID) error {
	var permissions []models.RolePermission
	for _, perm := range permissionIds {
		permissions = append(permissions, models.RolePermission{
			RoleID:       roleId,
			PermissionID: perm,
		})
	}
	_, err := crudrepo.RolePermission.Post(ctx, p.db, permissions)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresRBACStore) DeleteUserRole(ctx context.Context, userId, roleId uuid.UUID) error {
	_, err := crudrepo.RolePermission.Delete(
		ctx,
		p.db,
		&map[string]any{
			"role_id": map[string]any{
				"_eq": roleId.String(),
			},
			"user_id": map[string]any{
				"_eq": userId.String(),
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresRBACStore) DeletePermission(ctx context.Context, id uuid.UUID) error {
	_, err := crudrepo.Permission.Delete(
		ctx,
		p.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": id.String(),
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresRBACStore) FindOrCreateRole(ctx context.Context, roleName string) (*models.Role, error) {
	role, err := crudrepo.Role.GetOne(
		ctx,
		p.db,
		&map[string]any{
			"name": map[string]any{
				"_eq": roleName,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if role == nil {
		role, err = p.CreateRole(ctx, &shared.CreateRoleDto{Name: roleName})
		if err != nil {
			return nil, err
		}
	}
	return role, nil
}

func (p *PostgresRBACStore) EnsureRoleAndPermissions(ctx context.Context, roleName string, permissionNames ...string) error {
	// find superuser role
	role, err := p.FindOrCreateRole(ctx, roleName)
	if err != nil {
		return err
	}
	for _, permissionName := range permissionNames {
		perm, err := p.FindOrCreatePermission(ctx, permissionName)
		if err != nil {
			slog.ErrorContext(ctx, "error finding or creating permission", "name", permissionName, "error", err)
			continue
		}
		if perm == nil {
			continue
		}
		err = p.CreateRolePermissions(ctx, role.ID, perm.ID)
		if err != nil && !database.IsUniqConstraintErr(err) {
			log.Println(err)
		}
	}
	return nil
}

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

func (p *PostgresRBACStore) DeleteRolePermissions(ctx context.Context, roleId uuid.UUID, permissionIds ...uuid.UUID) error {
	if len(permissionIds) == 0 {
		return nil
	}
	var ids []string
	for _, id := range permissionIds {
		ids = append(ids, id.String())
	}
	_, err := crudrepo.RolePermission.Delete(
		ctx,
		p.db,
		&map[string]any{
			"role_id": map[string]any{
				"_eq": roleId.String(),
			},
			"permission_id": map[string]any{
				"_in": ids,
			},
		},
	)
	return err
}

const (
	GetProductPermissionsQuery = `
	SELECT rp.product_id as key,
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
FROM public.product_permissions rp
        LEFT JOIN public.permissions p ON p.id = rp.permission_id
        WHERE rp.product_id = ANY (
                $1::text []
        )
GROUP BY rp.product_id;`
)

func (p *PostgresRBACStore) LoadProductPermissions(ctx context.Context, productIds ...string) ([][]*models.Permission, error) {

	data, err := pgxscan.All(
		ctx,
		p.db,
		scan.StructMapper[shared.JoinedResult[*models.Permission, string]](),
		GetProductPermissionsQuery,
		productIds,
	)
	if err != nil {
		return nil, err
	}
	return mapper.Map(mapper.MapTo(data, productIds, func(a shared.JoinedResult[*models.Permission, string]) string {
		return a.Key
	}), func(a *shared.JoinedResult[*models.Permission, string]) []*models.Permission {
		if a == nil {
			return nil
		}
		return a.Data
	}), nil
}

func (p *PostgresRBACStore) LoadRolePermissions(ctx context.Context, roleIds ...uuid.UUID) ([][]*models.Permission, error) {
	// var results []JoinedResult[*crudModels.Permission, uuid.UUID]
	ids := []string{}
	for _, id := range roleIds {
		ids = append(ids, id.String())
	}
	data, err := pgxscan.All(
		ctx,
		p.db,
		scan.StructMapper[shared.JoinedResult[*models.Permission, uuid.UUID]](),
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

func (p *PostgresRBACStore) CountRoles(ctx context.Context, filter *shared.RoleListFilter) (int64, error) {
	q := squirrel.Select("COUNT(roles.*)").From("roles")

	q = ListRolesFilterFuncQuery(q, filter)

	data, err := database.QueryWithBuilder[CountOutput](ctx, p.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil
	}

	return data[0].Count, nil
}
func (p *PostgresRBACStore) ListPermissions(ctx context.Context, input *shared.PermissionsListParams) ([]*models.Permission, error) {
	q := squirrel.Select("permissions.*").From("permissions")
	filter := input.PermissionsListFilter
	pageInput := &input.PaginatedInput

	// q = ViewApplyPagination(q, pageInput)
	q = ListPermissionsFilterFunc(q, &filter)
	q = database.Paginate(q, pageInput)
	if input.SortBy != "" && input.SortOrder != "" {
		q = q.OrderBy(input.SortBy + " " + strings.ToUpper(input.SortOrder))
	}
	data, err := database.QueryWithBuilder[*models.Permission](ctx, p.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CountPermissions implements AdminCrudActions.
func (p *PostgresRBACStore) CountPermissions(ctx context.Context, filter *shared.PermissionsListFilter) (int64, error) {
	q := squirrel.Select("COUNT(permissions.*)").From("permissions")

	// q = ViewApplyPagination(q, pageInput)
	q = ListPermissionsFilterFunc(q, filter)

	data, err := database.QueryWithBuilder[CountOutput](ctx, p.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil
	}

	return data[0].Count, nil
}
func ListPermissionsFilterFunc(sq squirrel.SelectBuilder, filter *shared.PermissionsListFilter) squirrel.SelectBuilder {
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

func (p *PostgresRBACStore) ListRoles(ctx context.Context, input *shared.RolesListParams) ([]*models.Role, error) {
	q := squirrel.Select("roles.*").From("roles")
	filter := input.RoleListFilter
	pageInput := &input.PaginatedInput

	// q = ViewApplyPagination(q, pageInput)
	q = ListRolesFilterFuncQuery(q, &filter)
	q = database.Paginate(q, pageInput)
	if input.SortBy != "" && input.SortOrder != "" {
		q = q.OrderBy(input.SortBy + " " + strings.ToUpper(input.SortOrder))
	}
	data, err := database.QueryWithBuilder[*models.Role](ctx, p.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return data, nil
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

const (
	getuserNotPermissions = `
	WITH -- Get permissions assigned through roles
role_based_permissions AS (
    SELECT p.*,
        rp.role_id,
		NULL::text as product_id,
        NULL::uuid AS direct_assignment -- Null indicates not directly assigned
    FROM public.user_roles ur
        JOIN public.role_permissions rp ON ur.role_id = rp.role_id
        JOIN public.permissions p ON rp.permission_id = p.id
    WHERE ur.user_id = $1
),
-- Get permissions assigned directly to user
direct_permissions AS (
    SELECT p.*,
        NULL::uuid AS role_id,
		NULL::text as product_id,
        -- Null indicates not from a role
        up.user_id AS direct_assignment
    FROM public.user_permissions up
        JOIN public.permissions p ON up.permission_id = p.id
    WHERE up.user_id = $1
),
-- Get permissions assigned through products
-- product_permissions AS (
-- 	SELECT p.*,
--         NULL::uuid AS role_id,
--         sprice.product_id AS product_id,
--         NULL::uuid AS direct_assignment -- Null indicates not directly assigned
-- FROM public.stripe_subscriptions ss
--         JOIN public.stripe_prices sprice ON ss.price_id = sprice.id
--         JOIN public.stripe_products sproduct ON sprice.product_id = sproduct.id
--         JOIN public.product_permissions pr ON sproduct.id = pr.product_id
--         JOIN public.permissions p ON pr.permission_id = p.id
-- WHERE ss.user_id = $1
--         AND ss.status IN ('active', 'trialing')
-- ),
-- Combine both sources
combined_permissions AS (
    SELECT *
    FROM role_based_permissions
    UNION ALL
    SELECT *
    FROM direct_permissions
	-- UNION ALL
	-- SELECT *
	-- FROM product_permissions
) -- Final result with aggregated role information
SELECT p.id,
    p.name,
    p.description,
    p.created_at,
    p.updated_at,
    -- Array of role IDs that grant this permission (empty if direct)
    array []::uuid [] AS role_ids,
	-- Array of product IDs that grant this permission (empty if direct)
	array []::text [] AS product_ids,
    -- Boolean indicating if permission is directly assigned
    false AS is_directly_assigned
FROM public.permissions p
    LEFT JOIN combined_permissions cp ON p.id = cp.id
WHERE cp.id IS NULL
GROUP BY p.id
ORDER BY p.name,
    p.id
LIMIT $2 OFFSET $3;`

	getuserNotPermissionCounts = `
	WITH -- Get permissions assigned through roles
role_based_permissions AS (
    SELECT p.*,
        rp.role_id,
		NULL::text as product_id,
        NULL::uuid AS direct_assignment -- Null indicates not directly assigned
    FROM public.user_roles ur
        JOIN public.role_permissions rp ON ur.role_id = rp.role_id
        JOIN public.permissions p ON rp.permission_id = p.id
    WHERE ur.user_id = $1
),
-- Get permissions assigned directly to user
direct_permissions AS (
    SELECT p.*,
        NULL::uuid AS role_id,
		NULL::text as product_id,
        -- Null indicates not from a role
        up.user_id AS direct_assignment
    FROM public.user_permissions up
        JOIN public.permissions p ON up.permission_id = p.id
    WHERE up.user_id = $1
),
-- Get permissions assigned through products
-- product_permissions AS (
-- 	SELECT p.*,
--         NULL::uuid AS role_id,
--         sprice.product_id AS product_id,
--         NULL::uuid AS direct_assignment -- Null indicates not directly assigned
-- FROM public.stripe_subscriptions ss
--         JOIN public.stripe_prices sprice ON ss.price_id = sprice.id
--         JOIN public.stripe_products sproduct ON sprice.product_id = sproduct.id
--         JOIN public.product_permissions pr ON sproduct.id = pr.product_id
--         JOIN public.permissions p ON pr.permission_id = p.id
-- WHERE ss.user_id = $1
--         AND ss.status IN ('active', 'trialing')
-- ),
-- Combine both sources
combined_permissions AS (
    SELECT *
    FROM role_based_permissions
    UNION ALL
    SELECT *
    FROM direct_permissions
    -- UNION ALL
    -- SELECT *
    -- FROM product_permissions
) -- Final result with aggregated role information
SELECT COUNT(DISTINCT p.id)
FROM public.permissions p
    LEFT JOIN combined_permissions cp ON p.id = cp.id
WHERE cp.id IS NULL;`
)

func (p *PostgresRBACStore) ListUserNotPermissionsSource(ctx context.Context, userId uuid.UUID, limit int64, offset int64) ([]shared.PermissionSource, error) {

	res, err := database.QueryAll[shared.PermissionSource](ctx, p.db, getuserNotPermissions, userId, limit, offset)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *PostgresRBACStore) CountNotUserPermissionSource(ctx context.Context, userId uuid.UUID) (int64, error) {
	// q := psql.RawQuery(getuserNotPermissionCounts, userId, userId)

	data, err := database.Count(ctx, p.db, getuserNotPermissionCounts, userId)
	if err != nil {
		return 0, err
	}
	return data, nil
}

const (
	QueryUserPermissionSource string = `
WITH -- Get permissions assigned through roles
role_based_permissions AS (
    SELECT p.*,
        rp.role_id,
		NULL::text as product_id,
        NULL::uuid AS direct_assignment -- Null indicates not directly assigned
    FROM public.user_roles ur
        JOIN public.role_permissions rp ON ur.role_id = rp.role_id
        JOIN public.permissions p ON rp.permission_id = p.id
    WHERE ur.user_id = $1
),
-- Get permissions assigned directly to user
direct_permissions AS (
    SELECT p.*,
        NULL::uuid AS role_id,
		NULL::text as product_id,
        -- Null indicates not from a role
        up.user_id AS direct_assignment
    FROM public.user_permissions up
        JOIN public.permissions p ON up.permission_id = p.id
    WHERE up.user_id = $1
),
-- Get permissions assigned through products
-- product_permissions AS (
-- 	SELECT p.*,
--         NULL::uuid AS role_id,
--         sprice.product_id AS product_id,
--         NULL::uuid AS direct_assignment -- Null indicates not directly assigned
-- FROM public.stripe_subscriptions ss
--         JOIN public.stripe_prices sprice ON ss.price_id = sprice.id
--         JOIN public.stripe_products sproduct ON sprice.product_id = sproduct.id
--         JOIN public.product_permissions pr ON sproduct.id = pr.product_id
--         JOIN public.permissions p ON pr.permission_id = p.id
-- WHERE ss.user_id = $1
--         AND ss.status IN ('active', 'trialing')
-- ),
-- Combine both sources
combined_permissions AS (
    SELECT *
    FROM role_based_permissions
    UNION ALL
    SELECT *
    FROM direct_permissions
	-- UNION ALL
    -- SELECT *
    -- FROM product_permissions
) -- Final result with aggregated role information
SELECT p.id,
    p.name,
    p.description,
    p.created_at,
    p.updated_at,
    -- Array of role IDs that grant this permission (empty if direct)
    array_remove(array_agg(DISTINCT rp.role_id), NULL) AS role_ids,
	-- Array of product IDs that grant this permission (empty if direct)
	array_remove(array_agg(DISTINCT rp.product_id), NULL) AS product_ids,
    -- Boolean indicating if permission is directly assigned
    bool_or(rp.direct_assignment IS NOT NULL) AS is_directly_assigned
FROM (
        SELECT DISTINCT id,
            name,
            description,
            created_at,
            updated_at
        FROM combined_permissions
    ) p
    LEFT JOIN combined_permissions rp ON p.id = rp.id
GROUP BY p.id,
    p.name,
    p.description,
    p.created_at,
    p.updated_at
ORDER BY p.name,
    p.id
LIMIT $2
OFFSET $3
	;`
	QueryUserPermissionSourceCount string = `
WITH -- Get permissions assigned through roles
role_based_permissions AS (
    SELECT p.*,
        rp.role_id,
		NULL::text as product_id,
        NULL::uuid AS direct_assignment -- Null indicates not directly assigned
    FROM public.user_roles ur
        JOIN public.role_permissions rp ON ur.role_id = rp.role_id
        JOIN public.permissions p ON rp.permission_id = p.id
    WHERE ur.user_id = $1
),
-- Get permissions assigned directly to user
direct_permissions AS (
    SELECT p.*,
        NULL::uuid AS role_id,
		NULL::text as product_id,
        -- Null indicates not from a role
        up.user_id AS direct_assignment
    FROM public.user_permissions up
        JOIN public.permissions p ON up.permission_id = p.id
    WHERE up.user_id = $1
),
-- Get permissions assigned through products
-- product_permissions AS (
-- 	SELECT p.*,
--         NULL::uuid AS role_id,
--         sprice.product_id AS product_id,
--         NULL::uuid AS direct_assignment -- Null indicates not directly assigned
-- FROM public.stripe_subscriptions ss
--         JOIN public.stripe_prices sprice ON ss.price_id = sprice.id
--         JOIN public.stripe_products sproduct ON sprice.product_id = sproduct.id
--         JOIN public.product_permissions pr ON sproduct.id = pr.product_id
--         JOIN public.permissions p ON pr.permission_id = p.id
-- WHERE ss.user_id = $1
--         AND ss.status IN ('active', 'trialing')
-- ),
-- Combine both sources
combined_permissions AS (
    SELECT *
    FROM role_based_permissions
    UNION ALL
    SELECT *
    FROM direct_permissions
    -- UNION ALL
    -- SELECT *
    -- FROM product_permissions
) -- Final result with aggregated role information
SELECT COUNT(DISTINCT id)
FROM combined_permissions
	;`
)

func (p *PostgresRBACStore) ListUserPermissionsSource(ctx context.Context, userId uuid.UUID, limit int64, offset int64) ([]shared.PermissionSource, error) {
	data, err := database.QueryAll[shared.PermissionSource](ctx, p.db, QueryUserPermissionSource, userId, limit, offset)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (p *PostgresRBACStore) CountUserPermissionSource(ctx context.Context, userId uuid.UUID) (int64, error) {
	data, err := database.Count(ctx, p.db, QueryUserPermissionSourceCount, userId)
	if err != nil {
		return 0, err
	}
	return data, nil
}
