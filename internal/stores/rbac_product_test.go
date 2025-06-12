package stores_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/test"
)

func TestCreateProductPermissions(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		adapter := stores.NewStorageAdapter(dbxx)
		permission, err := adapter.Rbac().FindOrCreatePermission(ctx, "basic")
		if err != nil {
			t.Fatalf("failed to find or create permission: %v", err)
		}
		err = adapter.Product().UpsertProduct(ctx, &models.StripeProduct{
			ID:          "stripe-product-id",
			Active:      true,
			Name:        "Test Product",
			Description: new(string),
			Image:       new(string),
			Metadata:    map[string]string{},
		})
		if err != nil {
			t.Fatalf("failed to upsert product: %v", err)
		}
		type args struct {
			ctx           context.Context
			db            database.Dbx
			productId     string
			permissionIds []uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "create product permission",
				args: args{
					ctx:           ctx,
					db:            dbxx,
					productId:     "stripe-product-id",
					permissionIds: []uuid.UUID{permission.ID},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := adapter.Rbac().CreateProductPermissions(tt.args.ctx, tt.args.productId, tt.args.permissionIds...); (err != nil) != tt.wantErr {
					t.Errorf("CreateProductPermissions() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
		return errors.New("rollback")
	})
}

func TestCreateProductRoles(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		adapter := stores.NewStorageAdapter(dbxx)
		role, err := adapter.Rbac().FindOrCreateRole(ctx, "basic")
		if err != nil {
			t.Fatalf("failed to find or create role: %v", err)
		}
		err = adapter.Product().UpsertProduct(ctx, &models.StripeProduct{
			ID:          "stripe-product-id",
			Active:      true,
			Name:        "Test Product",
			Description: new(string),
			Image:       new(string),
			Metadata:    map[string]string{},
		})
		if err != nil {
			t.Fatalf("failed to upsert product: %v", err)
		}
		type args struct {
			ctx       context.Context
			productId string
			roleIds   []uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "create product role",
				args: args{
					ctx: ctx,

					productId: "stripe-product-id",
					roleIds:   []uuid.UUID{role.ID},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := adapter.Rbac().CreateProductRoles(tt.args.ctx, tt.args.productId, tt.args.roleIds...); (err != nil) != tt.wantErr {
					t.Errorf("CreateProductRoles() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
		return errors.New("rollback")
	})
}
