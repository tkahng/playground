package queries

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/aarondl/opt/null"
	"github.com/google/uuid"
	"github.com/stephenafamo/scan"
	"github.com/stephenafamo/scan/pgxscan"
	"github.com/tkahng/authgo/internal/crud/crudrepo"
	crudModels "github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

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

func LoadRolePermissions(ctx context.Context, db Queryer, roleIds ...uuid.UUID) ([][]*crudModels.Permission, error) {
	// var results []JoinedResult[*crudModels.Permission, uuid.UUID]
	ids := []string{}
	for _, id := range roleIds {
		ids = append(ids, id.String())
	}
	data, err := pgxscan.All(
		ctx,
		db,
		scan.StructMapper[shared.JoinedResult[*crudModels.Permission, uuid.UUID]](),
		GetRolePermissionsQuery,
		roleIds,
	)
	if err != nil {
		return nil, err
	}
	return mapper.Map(mapper.MapTo(data, roleIds, func(a shared.JoinedResult[*crudModels.Permission, uuid.UUID]) uuid.UUID {
		return a.Key
	}), func(a *shared.JoinedResult[*crudModels.Permission, uuid.UUID]) []*crudModels.Permission {
		if a == nil {
			return nil
		}
		return a.Data
	}), nil
}

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

func GetUserRoles(ctx context.Context, db Queryer, userIds ...uuid.UUID) ([][]*crudModels.Role, error) {
	// var results []JoinedResult[*crudModels.Permission, uuid.UUID]
	ids := []string{}
	for _, id := range userIds {
		ids = append(ids, id.String())
	}
	data, err := pgxscan.All(
		ctx,
		db,
		scan.StructMapper[shared.JoinedResult[*crudModels.Role, uuid.UUID]](),
		GetUserRolesQuery,
		userIds,
	)
	if err != nil {
		return nil, err
	}
	return mapper.Map(mapper.MapTo(data, userIds, func(a shared.JoinedResult[*crudModels.Role, uuid.UUID]) uuid.UUID {
		return a.Key
	}), func(a *shared.JoinedResult[*crudModels.Role, uuid.UUID]) []*crudModels.Role {
		if a == nil {
			return nil
		}
		return a.Data
	}), nil
}

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
        LEFT JOIN public.permission p ON p.id = rp.permission_id
        WHERE rp.user_id = ANY (
                $1::uuid []
        )
GROUP BY rp.user_id;`
)

func GetUserPermissions(ctx context.Context, db Queryer, userIds ...uuid.UUID) ([][]*crudModels.Permission, error) {
	// var results []JoinedResult[*crudModels.Permission, uuid.UUID]
	ids := []string{}
	for _, id := range userIds {
		ids = append(ids, id.String())
	}
	data, err := pgxscan.All(
		ctx,
		db,
		scan.StructMapper[shared.JoinedResult[*crudModels.Permission, uuid.UUID]](),
		GetUserRolesQuery,
		userIds,
	)
	if err != nil {
		return nil, err
	}
	return mapper.Map(mapper.MapTo(data, userIds, func(a shared.JoinedResult[*crudModels.Permission, uuid.UUID]) uuid.UUID {
		return a.Key
	}), func(a *shared.JoinedResult[*crudModels.Permission, uuid.UUID]) []*crudModels.Permission {
		if a == nil {
			return nil
		}
		return a.Data
	}), nil
}

func CreateRolePermissions(ctx context.Context, db Queryer, roleId uuid.UUID, permissionIds ...uuid.UUID) error {
	var permissions []crudModels.RolePermission
	for _, perm := range permissionIds {
		permissions = append(permissions, crudModels.RolePermission{
			RoleID:       roleId,
			PermissionID: perm,
		})
	}
	q := squirrel.Insert("role_permissions").Columns("role_id", "permission_id")
	for _, perm := range permissions {
		q = q.Values(perm.RoleID, perm.PermissionID)
	}
	q = q.Suffix("RETURNING *")
	// sql, args, err := q.PlaceholderFormat(squirrel.Dollar).ToSql()
	// if err != nil {
	// 	return err
	// }
	// fmt.Println(sql, args)
	// _, err = pgxscan.All(ctx, db, scan.StructMapper[crudModels.RolePermission](), sql, args...)
	_, err := QueryWithBuilder[crudModels.UserPermission](ctx, db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return err
	}
	return nil
}

func CreateProductRoles(ctx context.Context, db Queryer, productId string, roleIds ...uuid.UUID) error {
	var roles []crudModels.ProductRole
	for _, role := range roleIds {
		roles = append(roles, crudModels.ProductRole{
			ProductID: productId,
			RoleID:    role,
		})
	}
	q := squirrel.Insert("product_roles").Columns("product_id", "role_id")
	for _, perm := range roles {
		q = q.Values(perm.ProductID, perm.RoleID)
	}
	q = q.Suffix("RETURNING *")
	sql, args, err := q.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return err
	}
	fmt.Println(sql, args)
	_, err = pgxscan.All(ctx, db, scan.StructMapper[crudModels.ProductRole](), sql, args...)
	if err != nil {
		return err
	}
	return nil
}

func EnsureRoleAndPermissions(ctx context.Context, db Queryer, roleName string, permissionNames ...string) error {
	// find superuser role
	role, err := FindOrCreateRole(ctx, db, roleName)
	if err != nil {
		return err
	}
	for _, permissionName := range permissionNames {
		perm, err := FindOrCreatePermission(ctx, db, permissionName)
		if err != nil {
			continue
		}

		err = CreateRolePermissions(ctx, db, role.ID, perm.ID)
		if err != nil && !IsUniqConstraintErr(err) {
			log.Println(err)
		}
	}
	return nil
}

type CreateRoleDto struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func FindOrCreateRole(ctx context.Context, dbx Queryer, roleName string) (*crudModels.Role, error) {
	role, err := crudrepo.Role.GetOne(
		ctx,
		dbx,
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
		role, err = CreateRole(ctx, dbx, &CreateRoleDto{Name: roleName})
		if err != nil {
			return nil, err
		}
	}
	return role, nil
}

func CreateRole(ctx context.Context, dbx Queryer, role *CreateRoleDto) (*crudModels.Role, error) {
	data, err := crudrepo.Role.PostOne(ctx, dbx, &crudModels.Role{
		Name:        role.Name,
		Description: role.Description,
	})
	// data, err := models.Roles.Insert(
	// 	&models.RoleSetter{
	// 		Name:        omit.From(role.Name),
	// 		Description: omitnull.FromPtr(role.Description),
	// 	},
	// 	im.Returning("*"),
	// ).One(ctx, dbx)
	return data, err
}

type UpdateRoleDto struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func UpdateRole(ctx context.Context, dbx Queryer, id uuid.UUID, roledto *UpdateRoleDto) error {
	role, err := crudrepo.Role.GetOne(
		ctx,
		dbx,
		&map[string]any{
			"id": map[string]any{
				"_eq": id,
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
	_, err = crudrepo.Role.PutOne(ctx, dbx, role)

	// q := models.Roles.Update(
	// 	models.UpdateWhere.Roles.ID.EQ(id),
	// 	models.RoleSetter{
	// 		Name:        omit.From(roledto.Name),
	// 		Description: omitnull.FromPtr(roledto.Description),
	// 	}.UpdateMod(),
	// )
	// _, err := q.Exec(ctx, dbx)
	if err != nil {
		return err
	}
	return nil
}

type UpdatePermissionDto struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func UpdatePermission(ctx context.Context, dbx Queryer, id uuid.UUID, roledto *UpdatePermissionDto) error {
	permission, err := crudrepo.Permission.GetOne(
		ctx,
		dbx,
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
	_, err = crudrepo.Permission.PutOne(ctx, dbx, permission)

	// q := models.Roles.Update(
	// 	models.UpdateWhere.Roles.ID.EQ(id),
	// 	models.RoleSetter{
	// 		Name:        omit.From(roledto.Name),
	// 		Description: omitnull.FromPtr(roledto.Description),
	// 	}.UpdateMod(),
	// )
	// _, err := q.Exec(ctx, dbx)
	if err != nil {
		return err
	}
	return nil
}

func DeleteRole(ctx context.Context, dbx Queryer, id uuid.UUID) error {
	_, err := crudrepo.Role.DeleteReturn(
		ctx,
		dbx,
		&map[string]any{
			"id": map[string]any{
				"_eq": id,
			},
		},
	)
	// _, err := models.Roles.Delete(
	// 	models.DeleteWhere.Roles.ID.EQ(id),
	// ).Exec(ctx, dbx)
	return err
}

func DeleteRolePermissions(ctx context.Context, dbx Queryer, id uuid.UUID) error {
	_, err := crudrepo.RolePermission.DeleteReturn(
		ctx,
		dbx,
		&map[string]any{
			"role_id": map[string]any{
				"_eq": id.String(),
			},
		},
	)
	return err
}

type CreatePermissionDto struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func FindOrCreatePermission(ctx context.Context, dbx Queryer, permissionName string) (*crudModels.Permission, error) {
	permission, err := crudrepo.Permission.GetOne(
		ctx,
		dbx,
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
		permission, err = CreatePermission(ctx, dbx, &CreatePermissionDto{Name: permissionName})
		if err != nil {
			return nil, err
		}
	}
	return permission, nil
}

func CreatePermission(ctx context.Context, dbx Queryer, permission *CreatePermissionDto) (*crudModels.Permission, error) {
	data, err := crudrepo.Permission.PostOne(ctx, dbx, &crudModels.Permission{
		Name:        permission.Name,
		Description: permission.Description,
	})
	// data, err := models.Permissions.Insert(
	// 	&models.PermissionSetter{
	// 		Name:        omit.From(permission.Name),
	// 		Description: omitnull.FromPtr(permission.Description),
	// 	},
	// 	im.Returning("*"),
	// ).One(ctx, dbx)
	return data, err
}

func FindPermissionsByIds(ctx context.Context, dbx Queryer, params []uuid.UUID) ([]*crudModels.Permission, error) {
	newIds := make([]string, len(params))
	for i, id := range params {
		newIds[i] = id.String()
	}
	return crudrepo.Permission.Get(
		ctx,
		dbx,
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
	// return models.Permissions.Query(models.SelectWhere.Permissions.ID.In(params...)).All(ctx, dbx)
}

func DeletePermission(ctx context.Context, dbx Queryer, id uuid.UUID) error {
	_, err := crudrepo.Permission.DeleteReturn(
		ctx,
		dbx,
		&map[string]any{
			"id": map[string]any{
				"_eq": id.String(),
			},
		},
	)
	// _, err := models.Permissions.Delete(
	// 	models.DeleteWhere.Permissions.ID.EQ(id),
	// ).Exec(ctx, dbx)
	return err
}

func FindPermissionById(ctx context.Context, dbx Queryer, id uuid.UUID) (*crudModels.Permission, error) {
	data, err := crudrepo.Permission.GetOne(
		ctx,
		dbx,
		&map[string]any{
			"id": map[string]any{
				"_eq": id.String(),
			},
		},
	)
	// data, err := models.Permissions.Query(models.SelectWhere.Permissions.ID.EQ(id)).One(ctx, dbx)
	return OptionalRow(data, err)
}

const (
	QueryUserPermissionSource string = `
WITH -- Get permissions assigned through roles
role_based_permissions AS (
    SELECT p.*,
        rp.role_id,
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
        -- Null indicates not from a role
        up.user_id AS direct_assignment
    FROM public.user_permissions up
        JOIN public.permissions p ON up.permission_id = p.id
    WHERE up.user_id = $1
),
-- Combine both sources
combined_permissions AS (
    SELECT *
    FROM role_based_permissions
    UNION ALL
    SELECT *
    FROM direct_permissions
) -- Final result with aggregated role information
SELECT p.id,
    p.name,
    p.description,
    p.created_at,
    p.updated_at,
    -- Array of role IDs that grant this permission (empty if direct)
    array_remove(array_agg(DISTINCT rp.role_id), NULL) AS role_ids,
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
        -- Null indicates not from a role
        up.user_id AS direct_assignment
    FROM public.user_permissions up
        JOIN public.permissions p ON up.permission_id = p.id
    WHERE up.user_id = $1
),
-- Combine both sources
combined_permissions AS (
    SELECT *
    FROM role_based_permissions
    UNION ALL
    SELECT *
    FROM direct_permissions
) -- Final result with aggregated role information
SELECT COUNT(DISTINCT id)
FROM combined_permissions
	;`
)

type PermissionSource struct {
	ID          uuid.UUID        `db:"id,pk" json:"id"`
	Name        string           `db:"name" json:"name"`
	Description null.Val[string] `db:"description" json:"description"`
	CreatedAt   time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time        `db:"updated_at" json:"updated_at"`
	RoleIDs     []uuid.UUID      `db:"role_ids" json:"role_ids"`
	IsDirectly  bool             `db:"is_directly_assigned" json:"is_directly_assigned"`
}

func ListUserPermissionsSource(ctx context.Context, dbx Queryer, userId uuid.UUID, limit int64, offset int64) ([]PermissionSource, error) {
	// q := psql.RawQuery(QueryUserPermissionSource, userId, userId, limit, offset)

	// data, err := bob.All(ctx, dbx, q, scan.StructMapper[PermissionSource]())
	data, err := QueryAll[PermissionSource](ctx, dbx, QueryUserPermissionSource, userId, limit, offset)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func CountUserPermissionSource(ctx context.Context, dbx Queryer, userId uuid.UUID) (int64, error) {
	// q := psql.RawQuery(QueryUserPermissionSourceCount, userId, userId)

	// data, err := bob.One(ctx, dbx, q, scan.SingleColumnMapper[int64])
	data, err := Count(ctx, dbx, QueryUserPermissionSourceCount, userId)
	if err != nil {
		return 0, err
	}
	return data, nil
}

const (
	getuserNotPermissions = `WITH -- Get permissions assigned through roles
role_based_permissions AS (
    SELECT p.*,
        rp.role_id,
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
        -- Null indicates not from a role
        up.user_id AS direct_assignment
    FROM public.user_permissions up
        JOIN public.permissions p ON up.permission_id = p.id
    WHERE up.user_id = $1
),
-- Combine both sources
combined_permissions AS (
    SELECT *
    FROM role_based_permissions
    UNION ALL
    SELECT *
    FROM direct_permissions
) -- Final result with aggregated role information
SELECT p.id,
    p.name,
    p.description,
    p.created_at,
    p.updated_at,
    -- Array of role IDs that grant this permission (empty if direct)
    array []::uuid [] AS role_ids,
    -- Boolean indicating if permission is directly assigned
    false AS is_directly_assigned
FROM public.permissions p
    LEFT JOIN combined_permissions cp ON p.id = cp.id
WHERE cp.id IS NULL
GROUP BY p.id
ORDER BY p.name,
    p.id
LIMIT $2 OFFSET $3;`
	getuserNotPermissionCounts = `WITH -- Get permissions assigned through roles
role_based_permissions AS (
    SELECT p.*,
        rp.role_id,
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
        -- Null indicates not from a role
        up.user_id AS direct_assignment
    FROM public.user_permissions up
        JOIN public.permissions p ON up.permission_id = p.id
    WHERE up.user_id = $1
),
-- Combine both sources
combined_permissions AS (
    SELECT *
    FROM role_based_permissions
    UNION ALL
    SELECT *
    FROM direct_permissions
) -- Final result with aggregated role information
SELECT COUNT(DISTINCT p.id)
FROM public.permissions p
    LEFT JOIN combined_permissions cp ON p.id = cp.id
WHERE cp.id IS NULL;
;`
)

func ListUserNotPermissionsSource(ctx context.Context, dbx Queryer, userId uuid.UUID, limit int64, offset int64) ([]PermissionSource, error) {

	res, err := QueryAll[PermissionSource](ctx, dbx, getuserNotPermissions, userId, limit, offset)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func CountNotUserPermissionSource(ctx context.Context, dbx Queryer, userId uuid.UUID) (int64, error) {
	// q := psql.RawQuery(getuserNotPermissionCounts, userId, userId)

	data, err := Count(ctx, dbx, getuserNotPermissionCounts, userId)
	if err != nil {
		return 0, err
	}
	return data, nil
}
