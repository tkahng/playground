package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
)

type RbacStoreDecorator struct {
	Delegate                         *DbRbacStore
	AssignRoleFunc                   func(ctx context.Context, userId uuid.UUID, roleNames ...string) error
	CountNotUserPermissionSourceFunc func(ctx context.Context, userId uuid.UUID) (int64, error)
	CountPermissionsFunc             func(ctx context.Context, filter *PermissionFilter) (int64, error)
	CountRolesFunc                   func(ctx context.Context, filter *RoleListFilter) (int64, error)
	CountUserPermissionSourceFunc    func(ctx context.Context, userId uuid.UUID) (int64, error)
	CreatePermissionFunc             func(ctx context.Context, name string, description *string) (*models.Permission, error)
	CreateProductPermissionsFunc     func(ctx context.Context, productId string, permissionIds ...uuid.UUID) error
	CreateProductRolesFunc           func(ctx context.Context, productId string, roleIds ...uuid.UUID) error
	CreateRoleFunc                   func(ctx context.Context, role *CreateRoleDto) (*models.Role, error)
	CreateRolePermissionsFunc        func(ctx context.Context, roleId uuid.UUID, permissionIds ...uuid.UUID) error
	CreateUserPermissionsFunc        func(ctx context.Context, userId uuid.UUID, permissionIds ...uuid.UUID) error

	CreateUserRolesFunc          func(ctx context.Context, userId uuid.UUID, roleIds ...uuid.UUID) error
	DeletePermissionFunc         func(ctx context.Context, id uuid.UUID) error
	DeleteProductRolesFunc       func(ctx context.Context, productId string, roleIds ...uuid.UUID) error
	DeleteProductPermissionsFunc func(ctx context.Context, productId string, permissionIds ...uuid.UUID) error
	DeleteRoleFunc               func(ctx context.Context, id uuid.UUID) error
	DeleteRolePermissionsFunc    func(ctx context.Context, roleId uuid.UUID, permissionIds ...uuid.UUID) error
	DeleteUserRoleFunc           func(ctx context.Context, userId uuid.UUID, roleId uuid.UUID) error
	EnsureRoleAndPermissionsFunc func(ctx context.Context, roleName string, permissionNames ...string) error
	FindOrCreatePermissionFunc   func(ctx context.Context, permissionName string) (*models.Permission, error)
	FindOrCreateRoleFunc         func(ctx context.Context, roleName string) (*models.Role, error)
	FindPermissionFunc           func(ctx context.Context, filter *PermissionFilter) (*models.Permission, error)
	FindPermissionByIdFunc       func(ctx context.Context, id uuid.UUID) (*models.Permission, error)
	FindPermissionByNameFunc     func(ctx context.Context, name string) (*models.Permission, error)
	FindPermissionsByIdsFunc     func(ctx context.Context, params []uuid.UUID) ([]*models.Permission, error)
	FindRoleByIdFunc             func(ctx context.Context, id uuid.UUID) (*models.Role, error)
	FindRoleByNameFunc           func(ctx context.Context, name string) (*models.Role, error)
	FindRolesByIdsFunc           func(ctx context.Context, params []uuid.UUID) ([]*models.Role, error)
	GetUserRolesFunc             func(ctx context.Context, userIds ...uuid.UUID) ([][]*models.Role, error)

	ListPermissionsFunc              func(ctx context.Context, input *PermissionFilter) ([]*models.Permission, error)
	ListRolesFunc                    func(ctx context.Context, input *RoleListFilter) ([]*models.Role, error)
	ListUserNotPermissionsSourceFunc func(ctx context.Context, userId uuid.UUID, limit int64, offset int64) ([]*models.PermissionSource, error)
	ListUserPermissionsSourceFunc    func(ctx context.Context, userId uuid.UUID, limit int64, offset int64) ([]*models.PermissionSource, error)
	LoadProductPermissionsFunc       func(ctx context.Context, productIds ...string) ([][]*models.Permission, error)
	LoadRolePermissionsFunc          func(ctx context.Context, roleIds ...uuid.UUID) ([][]*models.Permission, error)
	UpdatePermissionFunc             func(ctx context.Context, id uuid.UUID, roledto *UpdatePermissionDto) error
	UpdateRoleFunc                   func(ctx context.Context, id uuid.UUID, roledto *UpdateRoleDto) error
}

func NewRbacStoreDecorator(db database.Dbx) *RbacStoreDecorator {
	delegate := NewDbRBACStore(db)
	return &RbacStoreDecorator{
		Delegate: delegate,
	}
}

func (r *RbacStoreDecorator) Cleanup() {
	r.AssignRoleFunc = nil
	r.CountNotUserPermissionSourceFunc = nil
	r.CountPermissionsFunc = nil
	r.CountRolesFunc = nil
	r.CountUserPermissionSourceFunc = nil
	r.CreatePermissionFunc = nil
	r.CreateProductPermissionsFunc = nil
	r.CreateProductRolesFunc = nil
	r.CreateRoleFunc = nil
	r.CreateRolePermissionsFunc = nil
	r.CreateUserPermissionsFunc = nil
	r.CreateUserRolesFunc = nil
	r.DeletePermissionFunc = nil
	r.DeleteProductRolesFunc = nil
	r.DeleteRoleFunc = nil
	r.DeleteRolePermissionsFunc = nil
	r.DeleteUserRoleFunc = nil
	r.EnsureRoleAndPermissionsFunc = nil
	r.FindOrCreatePermissionFunc = nil
	r.FindOrCreateRoleFunc = nil
	r.FindPermissionFunc = nil
	r.FindPermissionByIdFunc = nil
	r.FindPermissionByNameFunc = nil
	r.FindPermissionsByIdsFunc = nil
	r.FindRoleByIdFunc = nil
	r.FindRoleByNameFunc = nil
	r.FindRolesByIdsFunc = nil
	r.GetUserRolesFunc = nil
	r.ListPermissionsFunc = nil
	r.ListRolesFunc = nil
	r.ListUserNotPermissionsSourceFunc = nil
	r.ListUserPermissionsSourceFunc = nil
	r.LoadProductPermissionsFunc = nil
	r.LoadRolePermissionsFunc = nil
	r.UpdatePermissionFunc = nil
	r.UpdateRoleFunc = nil

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
func (r *RbacStoreDecorator) CountPermissions(ctx context.Context, filter *PermissionFilter) (int64, error) {
	if r.CountPermissionsFunc != nil {
		return r.CountPermissionsFunc(ctx, filter)
	}
	if r.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return r.Delegate.CountPermissions(ctx, filter)
}

// CountRoles implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) CountRoles(ctx context.Context, filter *RoleListFilter) (int64, error) {
	if r.CountRolesFunc != nil {
		return r.CountRolesFunc(ctx, filter)
	}
	if r.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return r.Delegate.CountRoles(ctx, filter)
}

// CountUserPermissionSource implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) CountUserPermissionSource(ctx context.Context, userId uuid.UUID) (int64, error) {
	if r.CountUserPermissionSourceFunc != nil {
		return r.CountUserPermissionSourceFunc(ctx, userId)
	}
	if r.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return r.Delegate.CountUserPermissionSource(ctx, userId)
}

// CreatePermission implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) CreatePermission(ctx context.Context, name string, description *string) (*models.Permission, error) {
	if r.CreatePermissionFunc != nil {
		return r.CreatePermissionFunc(ctx, name, description)
	}
	if r.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return r.Delegate.CreatePermission(ctx, name, description)
}

// CreateProductPermissions implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) CreateProductPermissions(ctx context.Context, productId string, permissionIds ...uuid.UUID) error {
	if r.CreateProductPermissionsFunc != nil {
		return r.CreateProductPermissionsFunc(ctx, productId, permissionIds...)
	}
	if r.Delegate == nil {
		return ErrDelegateNil
	}
	return r.Delegate.CreateProductPermissions(ctx, productId, permissionIds...)
}

// CreateProductRoles implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) CreateProductRoles(ctx context.Context, productId string, roleIds ...uuid.UUID) error {
	if r.CreateProductRolesFunc != nil {
		return r.CreateProductRolesFunc(ctx, productId, roleIds...)
	}
	if r.Delegate == nil {
		return ErrDelegateNil
	}
	return r.Delegate.CreateProductRoles(ctx, productId, roleIds...)
}

// CreateRole implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) CreateRole(ctx context.Context, role *CreateRoleDto) (*models.Role, error) {
	if r.CreateRoleFunc != nil {
		return r.CreateRoleFunc(ctx, role)
	}
	if r.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return r.Delegate.CreateRole(ctx, role)
}

// CreateRolePermissions implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) CreateRolePermissions(ctx context.Context, roleId uuid.UUID, permissionIds ...uuid.UUID) error {
	if r.CreateRolePermissionsFunc != nil {
		return r.CreateRolePermissionsFunc(ctx, roleId, permissionIds...)
	}
	if r.Delegate == nil {
		return ErrDelegateNil
	}
	return r.Delegate.CreateRolePermissions(ctx, roleId, permissionIds...)
}

// CreateUserPermissions implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) CreateUserPermissions(ctx context.Context, userId uuid.UUID, permissionIds ...uuid.UUID) error {
	if r.CreateUserPermissionsFunc != nil {
		return r.CreateUserPermissionsFunc(ctx, userId, permissionIds...)
	}
	if r.Delegate == nil {
		return ErrDelegateNil
	}
	return r.Delegate.CreateUserPermissions(ctx, userId, permissionIds...)
}

// CreateUserRoles implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) CreateUserRoles(ctx context.Context, userId uuid.UUID, roleIds ...uuid.UUID) error {
	if r.CreateUserRolesFunc != nil {
		return r.CreateUserRolesFunc(ctx, userId, roleIds...)
	}
	if r.Delegate == nil {
		return ErrDelegateNil
	}
	return r.Delegate.CreateUserRoles(ctx, userId, roleIds...)
}

// DeletePermission implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) DeletePermission(ctx context.Context, id uuid.UUID) error {
	if r.DeletePermissionFunc != nil {
		return r.DeletePermissionFunc(ctx, id)
	}
	if r.Delegate == nil {
		return ErrDelegateNil
	}
	return r.Delegate.DeletePermission(ctx, id)
}

// DeleteProductRoles implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) DeleteProductRoles(ctx context.Context, productId string, roleIds ...uuid.UUID) error {
	if r.DeleteProductRolesFunc != nil {
		return r.DeleteProductRolesFunc(ctx, productId, roleIds...)
	}
	if r.Delegate == nil {
		return ErrDelegateNil
	}
	return r.Delegate.DeleteProductRoles(ctx, productId, roleIds...)
}

// DeleteProductPermissions implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) DeleteProductPermissions(ctx context.Context, productId string, permissionIds ...uuid.UUID) error {
	if r.DeleteProductPermissionsFunc != nil {
		return r.DeleteProductPermissionsFunc(ctx, productId, permissionIds...)
	}
	if r.Delegate == nil {
		return ErrDelegateNil
	}
	return r.Delegate.DeleteProductPermissions(ctx, productId, permissionIds...)
}

// DeleteRole implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) DeleteRole(ctx context.Context, id uuid.UUID) error {
	if r.DeleteRoleFunc != nil {
		return r.DeleteRoleFunc(ctx, id)
	}
	if r.Delegate == nil {
		return ErrDelegateNil
	}
	return r.Delegate.DeleteRole(ctx, id)
}

// DeleteRolePermissions implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) DeleteRolePermissions(ctx context.Context, roleId uuid.UUID, permissionIds ...uuid.UUID) error {
	if r.DeleteRolePermissionsFunc != nil {
		return r.DeleteRolePermissionsFunc(ctx, roleId, permissionIds...)
	}
	if r.Delegate == nil {
		return ErrDelegateNil
	}
	return r.Delegate.DeleteRolePermissions(ctx, roleId, permissionIds...)
}

// DeleteUserRole implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) DeleteUserRole(ctx context.Context, userId uuid.UUID, roleId uuid.UUID) error {
	if r.DeleteUserRoleFunc != nil {
		return r.DeleteUserRoleFunc(ctx, userId, roleId)
	}
	if r.Delegate == nil {
		return ErrDelegateNil
	}
	return r.Delegate.DeleteUserRole(ctx, userId, roleId)
}

// EnsureRoleAndPermissions implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) EnsureRoleAndPermissions(ctx context.Context, roleName string, permissionNames ...string) error {
	if r.EnsureRoleAndPermissionsFunc != nil {
		return r.EnsureRoleAndPermissionsFunc(ctx, roleName, permissionNames...)
	}
	if r.Delegate == nil {
		return ErrDelegateNil
	}
	return r.Delegate.EnsureRoleAndPermissions(ctx, roleName, permissionNames...)
}

// FindOrCreatePermission implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) FindOrCreatePermission(ctx context.Context, permissionName string) (*models.Permission, error) {
	if r.FindOrCreatePermissionFunc != nil {
		return r.FindOrCreatePermissionFunc(ctx, permissionName)
	}
	if r.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return r.Delegate.FindOrCreatePermission(ctx, permissionName)
}

// FindOrCreateRole implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) FindOrCreateRole(ctx context.Context, roleName string) (*models.Role, error) {
	if r.FindOrCreateRoleFunc != nil {
		return r.FindOrCreateRoleFunc(ctx, roleName)
	}
	if r.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return r.Delegate.FindOrCreateRole(ctx, roleName)
}

func (r *RbacStoreDecorator) FindPermission(ctx context.Context, filter *PermissionFilter) (*models.Permission, error) {
	if r.FindPermissionFunc != nil {
		return r.FindPermissionFunc(ctx, filter)
	}
	if r.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return r.Delegate.FindPermission(ctx, filter)
}

// FindPermissionById implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) FindPermissionById(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	if r.FindPermissionByIdFunc != nil {
		return r.FindPermissionByIdFunc(ctx, id)
	}
	if r.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return r.Delegate.FindPermissionById(ctx, id)
}

// FindPermissionByName implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) FindPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	if r.FindPermissionByNameFunc != nil {
		return r.FindPermissionByNameFunc(ctx, name)
	}
	if r.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return r.Delegate.FindPermissionByName(ctx, name)
}

// FindPermissionsByIds implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) FindPermissionsByIds(ctx context.Context, params []uuid.UUID) ([]*models.Permission, error) {
	if r.FindPermissionsByIdsFunc != nil {
		return r.FindPermissionsByIdsFunc(ctx, params)
	}
	if r.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return r.Delegate.FindPermissionsByIds(ctx, params)
}

// FindRoleById implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) FindRoleById(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	if r.FindRoleByIdFunc != nil {
		return r.FindRoleByIdFunc(ctx, id)
	}
	if r.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return r.Delegate.FindRoleById(ctx, id)
}

// FindRoleByName implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) FindRoleByName(ctx context.Context, name string) (*models.Role, error) {
	if r.FindRoleByNameFunc != nil {
		return r.FindRoleByNameFunc(ctx, name)
	}
	if r.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return r.Delegate.FindRoleByName(ctx, name)
}

// FindRolesByIds implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) FindRolesByIds(ctx context.Context, params []uuid.UUID) ([]*models.Role, error) {
	if r.FindRolesByIdsFunc != nil {
		return r.FindRolesByIdsFunc(ctx, params)
	}
	if r.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return r.Delegate.FindRolesByIds(ctx, params)
}

// GetUserRoles implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) GetUserRoles(ctx context.Context, userIds ...uuid.UUID) ([][]*models.Role, error) {
	if r.GetUserRolesFunc != nil {
		return r.GetUserRolesFunc(ctx, userIds...)
	}
	if r.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return r.Delegate.GetUserRoles(ctx, userIds...)
}

// ListPermissions implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) ListPermissions(ctx context.Context, input *PermissionFilter) ([]*models.Permission, error) {
	if r.ListPermissionsFunc != nil {
		return r.ListPermissionsFunc(ctx, input)
	}
	if r.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return r.Delegate.ListPermissions(ctx, input)
}

// ListRoles implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) ListRoles(ctx context.Context, input *RoleListFilter) ([]*models.Role, error) {
	if r.ListRolesFunc != nil {
		return r.ListRolesFunc(ctx, input)
	}
	if r.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return r.Delegate.ListRoles(ctx, input)
}

// ListUserNotPermissionsSource implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) ListUserNotPermissionsSource(ctx context.Context, userId uuid.UUID, limit int64, offset int64) ([]*models.PermissionSource, error) {
	if r.ListUserNotPermissionsSourceFunc != nil {
		return r.ListUserNotPermissionsSourceFunc(ctx, userId, limit, offset)
	}
	if r.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return r.Delegate.ListUserNotPermissionsSource(ctx, userId, limit, offset)
}

// ListUserPermissionsSource implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) ListUserPermissionsSource(ctx context.Context, userId uuid.UUID, limit int64, offset int64) ([]*models.PermissionSource, error) {
	if r.ListUserPermissionsSourceFunc != nil {
		return r.ListUserPermissionsSourceFunc(ctx, userId, limit, offset)
	}
	if r.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return r.Delegate.ListUserPermissionsSource(ctx, userId, limit, offset)
}

// LoadProductPermissions implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) LoadProductPermissions(ctx context.Context, productIds ...string) ([][]*models.Permission, error) {
	if r.LoadProductPermissionsFunc != nil {
		return r.LoadProductPermissionsFunc(ctx, productIds...)
	}
	if r.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return r.Delegate.LoadProductPermissions(ctx, productIds...)
}

// LoadRolePermissions implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) LoadRolePermissions(ctx context.Context, roleIds ...uuid.UUID) ([][]*models.Permission, error) {
	if r.LoadRolePermissionsFunc != nil {
		return r.LoadRolePermissionsFunc(ctx, roleIds...)
	}
	if r.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return r.Delegate.LoadRolePermissions(ctx, roleIds...)
}

// UpdatePermission implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) UpdatePermission(ctx context.Context, id uuid.UUID, roledto *UpdatePermissionDto) error {
	if r.UpdatePermissionFunc != nil {
		return r.UpdatePermissionFunc(ctx, id, roledto)
	}
	if r.Delegate == nil {
		return ErrDelegateNil
	}
	return r.Delegate.UpdatePermission(ctx, id, roledto)
}

// UpdateRole implements DbRbacStoreInterface.
func (r *RbacStoreDecorator) UpdateRole(ctx context.Context, id uuid.UUID, roledto *UpdateRoleDto) error {
	if r.UpdateRoleFunc != nil {
		return r.UpdateRoleFunc(ctx, id, roledto)
	}
	if r.Delegate == nil {
		return ErrDelegateNil
	}
	return r.Delegate.UpdateRole(ctx, id, roledto)
}

var _ DbRbacStoreInterface = (*RbacStoreDecorator)(nil)
