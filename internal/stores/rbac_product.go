package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/repository"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/tools/mapper"
)

func (s *DbRbacStore) CreateProductRoles(ctx context.Context, productId string, roleIds ...uuid.UUID) error {
	var roles []models.ProductRole
	for _, role := range roleIds {
		roles = append(roles, models.ProductRole{
			ProductID: productId,
			RoleID:    role,
		})
	}
	_, err := repository.ProductRole.Post(ctx, s.db, roles)
	if err != nil {
		return err
	}
	return nil

}

// CreateProductPermissions implements RBACStore.
func (p *DbRbacStore) CreateProductPermissions(ctx context.Context, productId string, permissionIds ...uuid.UUID) error {
	db := p.db
	var permissions []models.ProductPermission
	for _, permissionId := range permissionIds {
		permissions = append(permissions, models.ProductPermission{
			ProductID:    productId,
			PermissionID: permissionId,
		})
	}
	_, err := repository.ProductPermission.Post(
		ctx,
		db,
		permissions,
	)
	if err != nil {
		return err
	}
	return nil
}

func (p *DbRbacStore) LoadProductPermissions(ctx context.Context, productIds ...string) ([][]*models.Permission, error) {
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
	data, err := database.QueryAll[shared.JoinedResult[*models.Permission, string]](
		ctx,
		p.db,
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

func (s *DbRbacStore) DeleteProductPermissions(ctx context.Context, productId string, permissionIds ...uuid.UUID) error {
	if len(permissionIds) == 0 {
		return nil
	}
	_, err := repository.ProductPermission.Delete(
		ctx,
		s.db,
		&map[string]any{
			"product_id": map[string]any{
				"_eq": productId,
			},
			"permission_id": map[string]any{
				"_in": permissionIds,
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *DbRbacStore) DeleteProductRoles(ctx context.Context, productId string, roleIds ...uuid.UUID) error {
	if len(roleIds) == 0 {
		return nil
	}
	newIds := make([]string, len(roleIds))
	for i, id := range roleIds {
		newIds[i] = id.String()
	}
	_, err := repository.ProductRole.Delete(
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
