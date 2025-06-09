package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type RbacStoreDecorator struct {
	Delegate                         *DbRbacStore
	AssignRoleFunc                   func(ctx context.Context, userId uuid.UUID, roleNames ...string) error
	CountNotUserPermissionSourceFunc func(ctx context.Context, userId uuid.UUID) (int64, error)
	CountPermissionsFunc             func(ctx context.Context, filter *shared.PermissionsListFilter) (int64, error)
	CountRolesFunc                   func(ctx context.Context, filter *shared.RoleListFilter) (int64, error)
	CountUserPermissionSourceFunc    func(ctx context.Context, userId uuid.UUID) (int64, error)
	CreatePermissionFunc             func(ctx context.Context, name string, description *string) (*models.Permission, error)
	CreateProductPermissionsFunc     func(ctx context.Context, productId string, permissionIds ...uuid.UUID) error
	CreateProductRolesFunc           func(ctx context.Context, productId string, roleIds ...uuid.UUID) error
	CreateRoleFunc                   func(ctx context.Context, role *shared.CreateRoleDto) (*models.Role, error)
	CreateRolePermissionsFunc        func(ctx context.Context, roleId uuid.UUID, permissionIds ...uuid.UUID) error
	CreateUserPermissionsFunc        func(ctx context.Context, userId uuid.UUID, permissionIds ...uuid.UUID) error

	CreateUserRolesFunc          func(ctx context.Context, userId uuid.UUID, roleIds ...uuid.UUID) error
	DeletePermissionFunc         func(ctx context.Context, id uuid.UUID) error
	DeleteProductRolesFunc       func(ctx context.Context, productId string, roleIds ...uuid.UUID) error
	DeleteRoleFunc               func(ctx context.Context, id uuid.UUID) error
	DeleteRolePermissionsFunc    func(ctx context.Context, roleId uuid.UUID, permissionIds ...uuid.UUID) error
	DeleteUserRoleFunc           func(ctx context.Context, userId uuid.UUID, roleId uuid.UUID) error
	EnsureRoleAndPermissionsFunc func(ctx context.Context, roleName string, permissionNames ...string) error
	FindOrCreatePermissionFunc   func(ctx context.Context, permissionName string) (*models.Permission, error)
	FindOrCreateRoleFunc         func(ctx context.Context, roleName string) (*models.Role, error)
	FindPermissionByIdFunc       func(ctx context.Context, id uuid.UUID) (*models.Permission, error)
	FindPermissionByNameFunc     func(ctx context.Context, name string) (*models.Permission, error)
	FindPermissionsByIdsFunc     func(ctx context.Context, params []uuid.UUID) ([]*models.Permission, error)
	FindRoleByIdFunc             func(ctx context.Context, id uuid.UUID) (*models.Role, error)
	FindRoleByNameFunc           func(ctx context.Context, name string) (*models.Role, error)
	FindRolesByIdsFunc           func(ctx context.Context, params []uuid.UUID) ([]*models.Role, error)
	GetUserRolesFunc             func(ctx context.Context, userIds ...uuid.UUID) ([][]*models.Role, error)
	ListPermissionsFunc          func(ctx context.Context, input *shared.PermissionsListParams) ([]*models.Permission, error)
}

// AssignUserRoles implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error {
	if r.AssignRoleFunc != nil {
		return r.AssignRoleFunc(ctx, userId, roleNames...)
	}
	if r.Delegate == nil {
		return ErrDelegateNil
	}
	return r.Delegate.AssignUserRoles(ctx, userId, roleNames...)
}

// CountNotUserPermissionSource implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) CountNotUserPermissionSource(ctx context.Context, userId uuid.UUID) (int64, error) {
	if r.CountNotUserPermissionSourceFunc != nil {
		return r.CountNotUserPermissionSourceFunc(ctx, userId)
	}
	if r.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return r.Delegate.CountNotUserPermissionSource(ctx, userId)
}

// CountPermissions implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) CountPermissions(ctx context.Context, filter *shared.PermissionsListFilter) (int64, error) {
	panic("unimplemented")
}

// CountRoles implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) CountRoles(ctx context.Context, filter *shared.RoleListFilter) (int64, error) {
	panic("unimplemented")
}

// CountUserPermissionSource implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) CountUserPermissionSource(ctx context.Context, userId uuid.UUID) (int64, error) {
	panic("unimplemented")
}

// CreatePermission implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) CreatePermission(ctx context.Context, name string, description *string) (*models.Permission, error) {
	panic("unimplemented")
}

// CreateProductPermissions implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) CreateProductPermissions(ctx context.Context, productId string, permissionIds ...uuid.UUID) error {
	panic("unimplemented")
}

// CreateProductRoles implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) CreateProductRoles(ctx context.Context, productId string, roleIds ...uuid.UUID) error {
	panic("unimplemented")
}

// CreateRole implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) CreateRole(ctx context.Context, role *shared.CreateRoleDto) (*models.Role, error) {
	panic("unimplemented")
}

// CreateRolePermissions implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) CreateRolePermissions(ctx context.Context, roleId uuid.UUID, permissionIds ...uuid.UUID) error {
	panic("unimplemented")
}

// CreateUserPermissions implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) CreateUserPermissions(ctx context.Context, userId uuid.UUID, permissionIds ...uuid.UUID) error {
	panic("unimplemented")
}

// CreateUserRoles implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) CreateUserRoles(ctx context.Context, userId uuid.UUID, roleIds ...uuid.UUID) error {
	panic("unimplemented")
}

// DeletePermission implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) DeletePermission(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

// DeleteProductRoles implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) DeleteProductRoles(ctx context.Context, productId string, roleIds ...uuid.UUID) error {
	panic("unimplemented")
}

// DeleteRole implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) DeleteRole(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

// DeleteRolePermissions implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) DeleteRolePermissions(ctx context.Context, roleId uuid.UUID, permissionIds ...uuid.UUID) error {
	panic("unimplemented")
}

// DeleteUserRole implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) DeleteUserRole(ctx context.Context, userId uuid.UUID, roleId uuid.UUID) error {
	panic("unimplemented")
}

// EnsureRoleAndPermissions implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) EnsureRoleAndPermissions(ctx context.Context, roleName string, permissionNames ...string) error {
	panic("unimplemented")
}

// FindOrCreatePermission implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) FindOrCreatePermission(ctx context.Context, permissionName string) (*models.Permission, error) {
	panic("unimplemented")
}

// FindOrCreateRole implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) FindOrCreateRole(ctx context.Context, roleName string) (*models.Role, error) {
	panic("unimplemented")
}

// FindPermissionById implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) FindPermissionById(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	panic("unimplemented")
}

// FindPermissionByName implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) FindPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	panic("unimplemented")
}

// FindPermissionsByIds implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) FindPermissionsByIds(ctx context.Context, params []uuid.UUID) ([]*models.Permission, error) {
	panic("unimplemented")
}

// FindRoleById implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) FindRoleById(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	panic("unimplemented")
}

// FindRoleByName implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) FindRoleByName(ctx context.Context, name string) (*models.Role, error) {
	panic("unimplemented")
}

// FindRolesByIds implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) FindRolesByIds(ctx context.Context, params []uuid.UUID) ([]*models.Role, error) {
	panic("unimplemented")
}

// GetUserRoles implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) GetUserRoles(ctx context.Context, userIds ...uuid.UUID) ([][]*models.Role, error) {
	panic("unimplemented")
}

// ListPermissions implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) ListPermissions(ctx context.Context, input *shared.PermissionsListParams) ([]*models.Permission, error) {
	panic("unimplemented")
}

// ListRoles implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) ListRoles(ctx context.Context, input *shared.RolesListParams) ([]*models.Role, error) {
	panic("unimplemented")
}

// ListUserNotPermissionsSource implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) ListUserNotPermissionsSource(ctx context.Context, userId uuid.UUID, limit int64, offset int64) ([]shared.PermissionSource, error) {
	panic("unimplemented")
}

// ListUserPermissionsSource implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) ListUserPermissionsSource(ctx context.Context, userId uuid.UUID, limit int64, offset int64) ([]shared.PermissionSource, error) {
	panic("unimplemented")
}

// LoadProductPermissions implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) LoadProductPermissions(ctx context.Context, productIds ...string) ([][]*models.Permission, error) {
	panic("unimplemented")
}

// LoadRolePermissions implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) LoadRolePermissions(ctx context.Context, roleIds ...uuid.UUID) ([][]*models.Permission, error) {
	panic("unimplemented")
}

// UpdatePermission implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) UpdatePermission(ctx context.Context, id uuid.UUID, roledto *shared.UpdatePermissionDto) error {
	panic("unimplemented")
}

// UpdateRole implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) UpdateRole(ctx context.Context, id uuid.UUID, roledto *shared.UpdateRoleDto) error {
	panic("unimplemented")
}

var _ DbRbacStoreInterface = (*RbacStoreDecorator)(nil)
