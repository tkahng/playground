package rbac

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
)

type RBACStore interface {
	CreatePermission(ctx context.Context, name string, description *string) (*models.Permission, error)
	CreateProductPermissions(ctx context.Context, productId string, permissionIds ...uuid.UUID) error
	FindPermissionByName(ctx context.Context, name string) (*models.Permission, error)
	FindOrCreatePermission(ctx context.Context, permissionName string) (*models.Permission, error)
}
