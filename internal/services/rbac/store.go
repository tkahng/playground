package rbac

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
)

type RBACStore interface {
	FindPermissionByName(ctx context.Context, name string) (*models.Permission, error)
	CreateProductPermissions(ctx context.Context, productId string, permissionIds ...uuid.UUID) error
}

type PostgresRBACStore struct {
	db database.Dbx
}

func NewPostgresRBACStore(db database.Dbx) RBACStore {
	return &PostgresRBACStore{
		db: db,
	}
}

var _ RBACStore = &PostgresRBACStore{}

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
