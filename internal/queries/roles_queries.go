package queries

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/stephenafamo/scan"
	"github.com/stephenafamo/scan/pgxscan"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	crudModels "github.com/tkahng/authgo/internal/models"
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

func LoadRolePermissions(ctx context.Context, db database.Dbx, roleIds ...uuid.UUID) ([][]*crudModels.Permission, error) {
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

func GetUserRoles(ctx context.Context, db database.Dbx, userIds ...uuid.UUID) ([][]*crudModels.Role, error) {
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
        LEFT JOIN public.permissions p ON p.id = rp.permission_id
WHERE rp.user_id = ANY ($1::uuid [])
GROUP BY rp.user_id;`
)

func GetUserPermissions(ctx context.Context, db database.Dbx, userIds ...uuid.UUID) ([][]*crudModels.Permission, error) {
	ids := []string{}
	for _, id := range userIds {
		ids = append(ids, id.String())
	}
	data, err := pgxscan.All(
		ctx,
		db,
		scan.StructMapper[shared.JoinedResult[*crudModels.Permission, uuid.UUID]](),
		GetUserPermissionsQuery,
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

func CreateRolePermissions(ctx context.Context, db database.Dbx, roleId uuid.UUID, permissionIds ...uuid.UUID) error {
	var permissions []crudModels.RolePermission
	for _, perm := range permissionIds {
		permissions = append(permissions, crudModels.RolePermission{
			RoleID:       roleId,
			PermissionID: perm,
		})
	}
	// q := squirrel.Insert("role_permissions").Columns("role_id", "permission_id")
	// for _, perm := range permissions {
	// 	q = q.Values(perm.RoleID, perm.PermissionID)
	// }
	// q = q.Suffix("RETURNING *")
	// _, err := QueryWithBuilder[crudModels.RolePermission](ctx, db, q.PlaceholderFormat(squirrel.Dollar))
	_, err := crudrepo.RolePermission.Post(ctx, db, permissions)
	if err != nil {
		return err
	}
	return nil
}

func CreateProductPermissions(ctx context.Context, db database.Dbx, productId string, permissionIds ...uuid.UUID) error {
	var permissions []crudModels.ProductPermission
	for _, permissionId := range permissionIds {
		permissions = append(permissions, crudModels.ProductPermission{
			ProductID:    productId,
			PermissionID: permissionId,
		})
	}
	_, err := crudrepo.ProductPermission.Post(
		ctx,
		db,
		permissions,
	)
	// q := squirrel.Insert("product_permissions").Columns("product_id", "permission_id")
	// for _, perm := range permissions {
	// 	q = q.Values(perm.ProductID, perm.PermissionID)
	// }
	// q = q.Suffix("RETURNING *")
	// sql, args, err := q.PlaceholderFormat(squirrel.Dollar).ToSql()
	// if err != nil {
	// 	return err
	// }
	// fmt.Println(sql, args)
	// _, err = pgxscan.All(ctx, db, scan.StructMapper[crudModels.ProductPermission](), sql, args...)
	if err != nil {
		return err
	}
	return nil
}

func EnsureRoleAndPermissions(ctx context.Context, db database.Dbx, roleName string, permissionNames ...string) error {
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

func FindOrCreateRole(ctx context.Context, dbx database.Dbx, roleName string) (*crudModels.Role, error) {
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

func CreateRole(ctx context.Context, dbx database.Dbx, role *CreateRoleDto) (*crudModels.Role, error) {
	data, err := crudrepo.Role.PostOne(ctx, dbx, &crudModels.Role{
		Name:        role.Name,
		Description: role.Description,
	})
	return data, err
}

type UpdateRoleDto struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func UpdateRole(ctx context.Context, dbx database.Dbx, id uuid.UUID, roledto *UpdateRoleDto) error {
	role, err := crudrepo.Role.GetOne(
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
	if role == nil {
		return nil
	}
	role.Name = roledto.Name
	role.Description = roledto.Description
	_, err = crudrepo.Role.PutOne(ctx, dbx, role)
	if err != nil {
		return err
	}
	return nil
}

type UpdatePermissionDto struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func UpdatePermission(ctx context.Context, dbx database.Dbx, id uuid.UUID, roledto *UpdatePermissionDto) error {
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

	if err != nil {
		return err
	}
	return nil
}

func DeleteRole(ctx context.Context, dbx database.Dbx, id uuid.UUID) error {
	_, err := crudrepo.Role.Delete(
		ctx,
		dbx,
		&map[string]any{
			"id": map[string]any{
				"_eq": id.String(),
			},
		},
	)
	return err
}

func DeleteRolePermissions(ctx context.Context, dbx database.Dbx, id uuid.UUID) error {
	_, err := crudrepo.RolePermission.Delete(
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

func CreatePermission(ctx context.Context, dbx database.Dbx, permission *CreatePermissionDto) (*crudModels.Permission, error) {
	data, err := crudrepo.Permission.PostOne(ctx, dbx, &crudModels.Permission{
		Name:        permission.Name,
		Description: permission.Description,
	})
	return data, err
}
func FindOrCreatePermission(ctx context.Context, dbx database.Dbx, permissionName string) (*crudModels.Permission, error) {
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

func DeletePermission(ctx context.Context, dbx database.Dbx, id uuid.UUID) error {
	_, err := crudrepo.Permission.Delete(
		ctx,
		dbx,
		&map[string]any{
			"id": map[string]any{
				"_eq": id.String(),
			},
		},
	)
	return err
}

func FindPermissionById(ctx context.Context, dbx database.Dbx, id uuid.UUID) (*crudModels.Permission, error) {
	data, err := crudrepo.Permission.GetOne(
		ctx,
		dbx,
		&map[string]any{
			"id": map[string]any{
				"_eq": id.String(),
			},
		},
	)
	return database.OptionalRow(data, err)
}
func FindPermissionByName(ctx context.Context, dbx database.Dbx, name string) (*crudModels.Permission, error) {
	data, err := crudrepo.Permission.GetOne(
		ctx,
		dbx,
		&map[string]any{
			"name": map[string]any{
				"_eq": name,
			},
		},
	)
	return database.OptionalRow(data, err)
}

func FindRoleByName(ctx context.Context, dbx database.Dbx, name string) (*crudModels.Role, error) {
	data, err := crudrepo.Role.GetOne(
		ctx,
		dbx,
		&map[string]any{
			"name": map[string]any{
				"_eq": name,
			},
		},
	)
	return database.OptionalRow(data, err)
}
