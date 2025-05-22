package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type RBACStore interface {
	AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error
	CountNotUserPermissionSource(ctx context.Context, userId uuid.UUID) (int64, error)
	CountPermissions(ctx context.Context, filter *shared.PermissionsListFilter) (int64, error)
	CountRoles(ctx context.Context, filter *shared.RoleListFilter) (int64, error)
	CountUserPermissionSource(ctx context.Context, userId uuid.UUID) (int64, error)
	CreatePermission(ctx context.Context, name string, description *string) (*models.Permission, error)
	CreateProductPermissions(ctx context.Context, productId string, permissionIds ...uuid.UUID) error
	CreateRole(ctx context.Context, role *shared.CreateRoleDto) (*models.Role, error)
	CreateRolePermissions(ctx context.Context, roleId uuid.UUID, permissionIds ...uuid.UUID) error
	CreateUserPermissions(ctx context.Context, userId uuid.UUID, permissionIds ...uuid.UUID) error
	CreateUserRoles(ctx context.Context, userId uuid.UUID, roleIds ...uuid.UUID) error
	DeletePermission(ctx context.Context, id uuid.UUID) error
	DeleteRole(ctx context.Context, id uuid.UUID) error
	DeleteUserRole(ctx context.Context, userId, roleId uuid.UUID) error
	EnsureRoleAndPermissions(ctx context.Context, roleName string, permissionNames ...string) error
	FindOrCreatePermission(ctx context.Context, permissionName string) (*models.Permission, error)
	FindOrCreateRole(ctx context.Context, roleName string) (*models.Role, error)
	FindPermissionById(ctx context.Context, id uuid.UUID) (*models.Permission, error)
	FindPermissionByName(ctx context.Context, name string) (*models.Permission, error)
	FindPermissionsByIds(ctx context.Context, params []uuid.UUID) ([]*models.Permission, error)
	ListPermissions(ctx context.Context, input *shared.PermissionsListParams) ([]*models.Permission, error)
	ListRoles(ctx context.Context, input *shared.RolesListParams) ([]*models.Role, error)
	ListUserNotPermissionsSource(ctx context.Context, userId uuid.UUID, limit int64, offset int64) ([]shared.PermissionSource, error)
	ListUserPermissionsSource(ctx context.Context, userId uuid.UUID, limit int64, offset int64) ([]shared.PermissionSource, error)
	LoadRolePermissions(ctx context.Context, roleIds ...uuid.UUID) ([][]*models.Permission, error)
	UpdatePermission(ctx context.Context, id uuid.UUID, roledto *shared.UpdatePermissionDto) error
	UpdateRole(ctx context.Context, id uuid.UUID, roledto *shared.UpdateRoleDto) error
	FindRoleById(ctx context.Context, id uuid.UUID) (*models.Role, error)
	FindRoleByName(ctx context.Context, name string) (*models.Role, error)
	FindRolesByIds(ctx context.Context, params []uuid.UUID) ([]*models.Role, error)
	DeleteRolePermissions(ctx context.Context, roleId uuid.UUID, permissionIds ...uuid.UUID) error
	LoadProductPermissions(ctx context.Context, productIds ...string) ([][]*models.Permission, error)
	DeleteProductRoles(ctx context.Context, productId string, roleIds ...uuid.UUID) error
}

type RBACService interface {
	Store() RBACStore
}

type rbacService struct {
	store RBACStore
}

// Store implements RBACService.
func (r *rbacService) Store() RBACStore {
	return r.store
}

func NewRBACService(store RBACStore) RBACService {
	return &rbacService{
		store: store,
	}
}
