package stores

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/repository"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/types"
)

func (a *DbRbacStore) AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error {
	if len(roleNames) > 0 {
		user, err := repository.User.GetOne(
			ctx,
			a.db,
			&map[string]any{
				models.UserTable.ID: map[string]any{
					"_eq": userId,
				},
			},
		)
		if err != nil {
			return fmt.Errorf("error finding user while assigning roles: %w", err)
		}
		if user == nil {
			return fmt.Errorf("user not found while assigning roles")
		}
		roles, err := repository.Role.Get(
			ctx,
			a.db,
			&map[string]any{
				models.RoleTable.Name: map[string]any{
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
			_, err = repository.UserRole.Post(ctx, a.db, userRoles)
			if err != nil {
				return fmt.Errorf("error assigning user role while assigning roles: %w", err)
			}
		}
	}
	return nil
}

func (p *DbRbacStore) CreateUserPermissions(ctx context.Context, userId uuid.UUID, permissionIds ...uuid.UUID) error {
	var dtos []models.UserPermission
	for _, id := range permissionIds {
		dtos = append(dtos, models.UserPermission{
			UserID:       userId,
			PermissionID: id,
		})
	}
	_, err := repository.UserPermission.Post(
		ctx,
		p.db,
		dtos,
	)
	if err != nil {
		return err
	}
	return nil
}

func (p *DbRbacStore) CreateUserRoles(ctx context.Context, userId uuid.UUID, roleIds ...uuid.UUID) error {
	var dtos []models.UserRole
	for _, id := range roleIds {
		dtos = append(dtos, models.UserRole{
			UserID: userId,
			RoleID: id,
		})
	}
	_, err := repository.UserRole.Post(
		ctx,
		p.db,
		dtos,
	)
	if err != nil {

		return err
	}
	return nil
}

func (p *DbRbacStore) ListUserNotPermissionsSource(ctx context.Context, userId uuid.UUID, limit int64, offset int64) ([]*models.PermissionSource, error) {
	const getuserNotPermissions = `
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
	res, err := database.QueryAll[*models.PermissionSource](ctx, p.db, getuserNotPermissions, userId, limit, offset)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *DbRbacStore) CountNotUserPermissionSource(ctx context.Context, userId uuid.UUID) (int64, error) {
	// q := psql.RawQuery(getuserNotPermissionCounts, userId, userId)
	const (
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

	data, err := database.Count(ctx, p.db, getuserNotPermissionCounts, userId)
	if err != nil {
		return 0, err
	}
	return data, nil
}

func (p *DbRbacStore) ListUserPermissionsSource(ctx context.Context, userId uuid.UUID, limit int64, offset int64) ([]*models.PermissionSource, error) {
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
	)
	data, err := database.QueryAll[*models.PermissionSource](ctx, p.db, QueryUserPermissionSource, userId, limit, offset)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (p *DbRbacStore) CountUserPermissionSource(ctx context.Context, userId uuid.UUID) (int64, error) {
	const QueryUserPermissionSourceCount string = `
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

	data, err := database.Count(ctx, p.db, QueryUserPermissionSourceCount, userId)
	if err != nil {
		return 0, err
	}
	return data, nil
}

func (p *DbRbacStore) DeleteUserRole(ctx context.Context, userId, roleId uuid.UUID) error {
	_, err := repository.RolePermission.Delete(
		ctx,
		p.db,
		&map[string]any{
			"role_id": map[string]any{
				"_eq": roleId,
			},
			"user_id": map[string]any{
				"_eq": userId,
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (p *DbRbacStore) GetUserRoles(ctx context.Context, userIds ...uuid.UUID) ([][]*models.Role, error) {
	const (
		GetUserRolesQuery = `
		SELECT rp.user_id as key,
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
	FROM public.user_roles rp
			LEFT JOIN public.roles p ON p.id = rp.role_id
			WHERE rp.user_id = ANY (
					$1::uuid []
			)
	GROUP BY rp.user_id;`
	)

	data, err := database.QueryAll[shared.JoinedResult[*models.Role, uuid.UUID]](
		ctx,
		p.db,
		GetUserRolesQuery,
		userIds,
	)
	if err != nil {
		return nil, err
	}
	return mapper.Map(mapper.MapTo(data, userIds, func(a shared.JoinedResult[*models.Role, uuid.UUID]) uuid.UUID {
		return a.Key
	}), func(a *shared.JoinedResult[*models.Role, uuid.UUID]) []*models.Role {
		if a == nil {
			return nil
		}
		return a.Data
	}), nil
}

func GetUserPermissions(ctx context.Context, db database.Dbx, userIds ...uuid.UUID) ([][]*models.Permission, error) {
	const (
		GetUserPermissionsQuery = `
	SELECT rp.user_id as key,
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
	FROM public.user_permissions rp
			LEFT JOIN public.permissions p ON p.id = rp.permission_id
	WHERE rp.user_id = ANY ($1::uuid [])
	GROUP BY rp.user_id;`
	)

	data, err := database.QueryAll[shared.JoinedResult[*models.Permission, uuid.UUID]](
		ctx,
		db,
		GetUserPermissionsQuery,
		userIds,
	)
	if err != nil {
		return nil, err
	}
	return mapper.Map(mapper.MapTo(data, userIds, func(a shared.JoinedResult[*models.Permission, uuid.UUID]) uuid.UUID {
		return a.Key
	}), func(a *shared.JoinedResult[*models.Permission, uuid.UUID]) []*models.Permission {
		if a == nil {
			return nil
		}
		return a.Data
	}), nil
}
