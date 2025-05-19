package stores

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
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
