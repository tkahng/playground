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
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		// rbacStore := stores.NewPostgresRBACStore(dbxx)
		paymentStore := stores.NewPostgresPaymentStore(dbxx)
		permission, err := paymentStore.FindOrCreatePermission(ctx, "basic")
		if err != nil {
			t.Fatalf("failed to find or create permission: %v", err)
		}
		err = paymentStore.UpsertProduct(ctx, &models.StripeProduct{
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
				if err := paymentStore.CreateProductPermissions(tt.args.ctx, tt.args.productId, tt.args.permissionIds...); (err != nil) != tt.wantErr {
					t.Errorf("CreateProductPermissions() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
		return errors.New("rollback")
	})
}

func TestCreateUserRoles(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		userStore := stores.NewPostgresUserStore(dbxx)
		rbacStore := stores.NewPostgresRBACStore(dbxx)
		// Create a user
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Errorf("failed to create user: %v", err)
			return err
		}
		role, err := rbacStore.FindOrCreateRole(ctx, "basic")
		if err != nil {
			t.Errorf("failed to create role: %v", err)
			return err
		}
		type args struct {
			ctx     context.Context
			db      database.Dbx
			userId  uuid.UUID
			roleIds []uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "create user roles",
				args: args{
					ctx:     ctx,
					db:      dbxx,
					userId:  user.ID,
					roleIds: []uuid.UUID{role.ID},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := rbacStore.CreateUserRoles(tt.args.ctx, tt.args.userId, tt.args.roleIds...); (err != nil) != tt.wantErr {
					t.Errorf("createUserRoles() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
		return test.EndTestErr
	})
}
