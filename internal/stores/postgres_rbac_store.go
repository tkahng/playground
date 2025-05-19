package stores

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/stephenafamo/scan"
	"github.com/stephenafamo/scan/pgxscan"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/types"
)

type PostgresRBACStore struct {
	db database.Dbx
}

type RBACStore struct {
	*PostgresRBACStore
}

func NewPostgresRBACStore(db database.Dbx) *PostgresRBACStore {
	return &PostgresRBACStore{
		db: db,
	}
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

func (p *PostgresRBACStore) CreateUserPermissions(ctx context.Context, db database.Dbx, userId uuid.UUID, permissionIds ...uuid.UUID) error {
	var dtos []models.UserPermission
	for _, id := range permissionIds {
		dtos = append(dtos, models.UserPermission{
			UserID:       userId,
			PermissionID: id,
		})
	}
	_, err := crudrepo.UserPermission.Post(
		ctx,
		db,
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

type UpdatePermissionDto struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func (p *PostgresRBACStore) UpdatePermission(ctx context.Context, dbx database.Dbx, id uuid.UUID, roledto *UpdatePermissionDto) error {
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

type CreateRoleDto struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func (p *PostgresRBACStore) CreateRole(ctx context.Context, role *CreateRoleDto) (*models.Role, error) {
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

type UpdateRoleDto struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func (p *PostgresRBACStore) UpdateRole(ctx context.Context, id uuid.UUID, roledto *UpdateRoleDto) error {
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

func (p *PostgresRBACStore) DeleteRolePermissions(ctx context.Context, id uuid.UUID) error {
	_, err := crudrepo.RolePermission.Delete(
		ctx,
		p.db,
		&map[string]any{
			"role_id": map[string]any{
				"_eq": id.String(),
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
		role, err = p.CreateRole(ctx, &CreateRoleDto{Name: roleName})
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
