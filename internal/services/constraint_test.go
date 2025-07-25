package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/seeders"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/types"
)

func TestConstraintCheckerService_CannotHaveValidSubscription(t *testing.T) {
	ctx, dbx := test.DbSetup()

	_ = dbx.RunInTx(func(tx database.Dbx) error {
		adapter := stores.NewStorageAdapter(tx)
		userStore := adapter.User()

		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "test@example.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		customer, err := adapter.Customer().CreateCustomer(ctx, &models.StripeCustomer{
			UserID:       types.Pointer(user.ID),
			Email:        user.Email,
			CustomerType: models.StripeCustomerTypeUser,
		})
		if err != nil {
			t.Fatalf("failed to create customer: %v", err)
		}
		if customer == nil {
			t.Fatalf("expected customer to be created, got nil")
		}

		prods, err := seeders.CreateStripeProductPrices(ctx, tx, 1)
		if err != nil {
			t.Fatalf("failed to create product prices: %v", err)
		}
		err = adapter.Subscription().UpsertSubscription(
			ctx,
			&models.StripeSubscription{
				ID:               "sub_123",
				StripeCustomerID: customer.ID,
				PriceID:          prods[0].Prices[0].ID,
				Status:           models.StripeSubscriptionStatusActive,
				Metadata: map[string]string{
					"key": "value",
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		)
		if err != nil {
			t.Fatalf("failed to upsert subscription: %v", err)
		}
		type fields struct {
			db  database.Dbx
			ctx context.Context
		}
		type args struct {
			userId uuid.UUID
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			wantErr bool
		}{
			{
				name:    "valid user",
				fields:  fields{db: tx, ctx: ctx},
				args:    args{userId: user.ID},
				wantErr: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c := services.NewConstraintCheckerService(adapter)
				ok, err := c.CannotHaveValidUserSubscription(tt.fields.ctx, tt.args.userId)
				if (err != nil) != tt.wantErr {
					t.Errorf("ConstraintCheckerService.CannotHaveValidSubscription() error = %v, wantErr %v", err, tt.wantErr)
					if err != nil && err.Error() != "Cannot perform this action on a user with a valid subscription" {
						t.Errorf("unexpected error message: %v", err.Error())
					}
				}
				if tt.wantErr && ok {
					t.Errorf("expected ok to be false when wantErr is true")
				}
				if !tt.wantErr && !ok {
					t.Errorf("expected ok to be true when wantErr is false")
				}
			})
		}
		return test.ErrEndTest
	})
}
func TestConstraintCheckerService_CannotBeAdminOrBasicName(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, dbx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbx)
		type fields struct {
			db  database.Dbx
			ctx context.Context
		}
		type args struct {
			permissionName string
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			wantErr bool
		}{
			{
				name:    "admin permission name",
				fields:  fields{db: dbx, ctx: ctx},
				args:    args{permissionName: shared.PermissionNameAdmin},
				wantErr: true,
			},
			{
				name:    "basic permission name",
				fields:  fields{db: dbx, ctx: ctx},
				args:    args{permissionName: shared.PermissionNameBasic},
				wantErr: true,
			},
			{
				name:    "other permission name",
				fields:  fields{db: dbx, ctx: ctx},
				args:    args{permissionName: "other"},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c := services.NewConstraintCheckerService(adapter)
				ok, err := c.CannotBeAdminOrBasicName(tt.fields.ctx, tt.args.permissionName)
				if (err != nil) != tt.wantErr {
					t.Errorf("ConstraintCheckerService.CannotBeAdminOrBasicName() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.wantErr && err != nil && err.Error() != "Cannot perform this action on the admin or basic permission" {
					t.Errorf("unexpected error message: %v", err.Error())
				}
				if tt.wantErr && ok {
					t.Errorf("expected ok to be false when wantErr is true")
				}
				if !tt.wantErr && !ok {
					t.Errorf("expected ok to be true when wantErr is false")
				}
			})
		}
	})
}
func TestConstraintCheckerService_CannotBeAdminOrBasicRoleAndPermissionName(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, dbx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbx)
		type fields struct {
			db  database.Dbx
			ctx context.Context
		}
		type args struct {
			roleName       string
			permissionName string
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			wantErr bool
		}{
			{
				name:   "admin role and permission",
				fields: fields{db: dbx, ctx: ctx},
				args: args{
					roleName:       shared.PermissionNameAdmin,
					permissionName: shared.PermissionNameAdmin,
				},
				wantErr: true,
			},
			{
				name:   "basic role and permission",
				fields: fields{db: dbx, ctx: ctx},
				args: args{
					roleName:       shared.PermissionNameBasic,
					permissionName: shared.PermissionNameBasic,
				},
				wantErr: true,
			},
			{
				name:   "admin role with different permission",
				fields: fields{db: dbx, ctx: ctx},
				args: args{
					roleName:       shared.PermissionNameAdmin,
					permissionName: "other",
				},
				wantErr: false,
			},
			{
				name:   "different role and permission",
				fields: fields{db: dbx, ctx: ctx},
				args: args{
					roleName:       "other",
					permissionName: "other",
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c := services.NewConstraintCheckerService(adapter)
				ok, err := c.CannotBeAdminOrBasicRoleAndPermissionName(tt.fields.ctx, tt.args.roleName, tt.args.permissionName)
				if (err != nil) != tt.wantErr {
					t.Errorf("ConstraintCheckerService.CannotBeAdminOrBasicRoleAndPermissionName() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.wantErr && err != nil && err.Error() != "Cannot perform this action on the admin role and permission" {
					t.Errorf("unexpected error message: %v", err.Error())
				}
				if tt.wantErr && ok {
					t.Errorf("expected ok to be false when wantErr is true")
				}
				if !tt.wantErr && !ok {
					t.Errorf("expected ok to be true when wantErr is false")
				}
			})
		}
	})

}
func TestConstraintCheckerService_CannotBeSuperUserEmailAndRoleName(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, dbx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbx)
		type fields struct {
			db  database.Dbx
			ctx context.Context
		}
		type args struct {
			email    string
			roleName string
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			wantErr bool
		}{
			{
				name:   "super user email and admin role",
				fields: fields{db: dbx, ctx: ctx},
				args: args{
					email:    shared.SuperUserEmail,
					roleName: shared.PermissionNameAdmin,
				},
				wantErr: true,
			},
			{
				name:   "super user email with different role",
				fields: fields{db: dbx, ctx: ctx},
				args: args{
					email:    shared.SuperUserEmail,
					roleName: "other",
				},
				wantErr: false,
			},
			{
				name:   "different email with admin role",
				fields: fields{db: dbx, ctx: ctx},
				args: args{
					email:    "other@example.com",
					roleName: shared.PermissionNameAdmin,
				},
				wantErr: false,
			},
			{
				name:   "different email and role",
				fields: fields{db: dbx, ctx: ctx},
				args: args{
					email:    "other@example.com",
					roleName: "other",
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c := services.NewConstraintCheckerService(adapter)
				ok, err := c.CannotBeSuperUserEmailAndRoleName(tt.fields.ctx, tt.args.email, tt.args.roleName)
				if (err != nil) != tt.wantErr {
					t.Errorf("ConstraintCheckerService.CannotBeSuperUserEmailAndRoleName() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.wantErr && err != nil && err.Error() != "Cannot perform this action on the super user email and admin role" {
					t.Errorf("unexpected error message: %v", err.Error())
				}
				if tt.wantErr && ok {
					t.Errorf("expected ok to be false when wantErr is true")
				}
				if !tt.wantErr && !ok {
					t.Errorf("expected ok to be true when wantErr is false")
				}
			})
		}
	})
}
func TestConstraintCheckerService_CannotBeSuperUserID(t *testing.T) {
	ctx, dbx := test.DbSetup()

	_ = dbx.RunInTx(func(tx database.Dbx) error {
		adapter := stores.NewStorageAdapter(tx)
		rbacStore := adapter.Rbac()
		userStore := adapter.User()

		err := rbacStore.EnsureRoleAndPermissions(
			ctx,
			shared.PermissionNameAdmin,
			shared.PermissionNameAdmin,
			shared.PermissionNameBasic,
		)
		if err != nil {
			t.Fatalf("failed to ensure roles and permissions: %v", err)
		}
		err = rbacStore.EnsureRoleAndPermissions(
			ctx,
			shared.PermissionNameBasic,
			shared.PermissionNameBasic,
		)
		if err != nil {
			t.Fatalf("failed to ensure roles and permissions: %v", err)
		}
		superUserRole, err := rbacStore.FindOrCreateRole(ctx, shared.PermissionNameAdmin)
		if err != nil {
			t.Fatalf("failed to find or create role: %v", err)
		}
		basicRole, err := rbacStore.FindOrCreateRole(ctx, shared.PermissionNameBasic)
		if err != nil {
			t.Fatalf("failed to find or create role: %v", err)
		}
		superUser, err := userStore.CreateUser(ctx, &models.User{
			Email: shared.SuperUserEmail,
		})
		if err != nil {
			t.Fatalf("failed to create super user: %v", err)
		}
		err = rbacStore.CreateUserRoles(ctx, superUser.ID, superUserRole.ID)
		if err != nil {
			t.Fatalf("failed to create user roles: %v", err)
		}
		regularUser, err := userStore.CreateUser(ctx, &models.User{
			Email: "regular@example.com",
		})
		if err != nil {
			t.Fatalf("failed to create regular user: %v", err)
		}
		err = rbacStore.CreateUserRoles(ctx, regularUser.ID, basicRole.ID)
		if err != nil {
			t.Fatalf("failed to create user roles: %v", err)
		}
		type fields struct {
			db  database.Dbx
			ctx context.Context
		}
		type args struct {
			userId uuid.UUID
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			wantErr bool
		}{
			{
				name:    "super user",
				fields:  fields{db: tx, ctx: ctx},
				args:    args{userId: superUser.ID},
				wantErr: true,
			},
			{
				name:    "regular user",
				fields:  fields{db: tx, ctx: ctx},
				args:    args{userId: regularUser.ID},
				wantErr: false,
			},
			{
				name:    "non-existent user",
				fields:  fields{db: tx, ctx: ctx},
				args:    args{userId: uuid.New()},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c := services.NewConstraintCheckerService(adapter)
				ok, err := c.CannotBeSuperUserID(tt.fields.ctx, tt.args.userId)
				if (err != nil) != tt.wantErr {
					t.Errorf("ConstraintCheckerService.CannotBeSuperUserID() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.wantErr && err != nil && err.Error() != "Cannot perform this action on the super user" {
					t.Errorf("unexpected error message: %v", err.Error())
				}
				if tt.wantErr && ok {
					t.Errorf("expected ok to be false when wantErr is true")
				}
				if !tt.wantErr && !ok {
					t.Errorf("expected ok to be true when wantErr is false")
				}
			})
		}
		return test.ErrEndTest
	})
}
func TestConstraintCheckerService_CannotBeSuperUserEmail(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, dbx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbx)

		type fields struct {
			db  database.Dbx
			ctx context.Context
		}
		type args struct {
			email string
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			wantErr bool
		}{
			{
				name:   "super user email",
				fields: fields{db: dbx, ctx: ctx},
				args: args{
					email: shared.SuperUserEmail,
				},
				wantErr: true,
			},
			{
				name:   "regular email",
				fields: fields{db: dbx, ctx: ctx},
				args: args{
					email: "regular@example.com",
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c := services.NewConstraintCheckerService(adapter)
				ok, err := c.CannotBeSuperUserEmail(tt.fields.ctx, tt.args.email)
				if (err != nil) != tt.wantErr {
					t.Errorf("ConstraintCheckerService.CannotBeSuperUserEmail() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.wantErr && err != nil && err.Error() != "Cannot perform this action on the super user" {
					t.Errorf("unexpected error message: %v", err.Error())
				}
				if tt.wantErr && ok {
					t.Errorf("expected ok to be false when wantErr is true")
				}
				if !tt.wantErr && !ok {
					t.Errorf("expected ok to be true when wantErr is false")
				}
			})
		}
	})
}
