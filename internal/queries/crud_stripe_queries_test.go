package queries_test

import (
	"context"
	"testing"

	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/seeders"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/test"
)

func TestListProducts(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		_, err := seeders.CreateStripeProductPrices(ctx, dbxx, 5)
		if err != nil {
			t.Fatalf("failed to create stripe products: %v", err)
		}
		type args struct {
			ctx   context.Context
			db    db.Dbx
			input *shared.StripeProductListParams
		}
		tests := []struct {
			name      string
			args      args
			wantCount int
			wantErr   bool
		}{
			{
				name: "List all products",
				args: args{
					ctx:   ctx,
					db:    dbxx,
					input: &shared.StripeProductListParams{},
				},
				wantCount: 5,
				wantErr:   false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.ListProducts(tt.args.ctx, tt.args.db, tt.args.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("ListProducts() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if len(got) != tt.wantCount {
					t.Errorf("ListProducts() got = %v, want %v", len(got), tt.wantCount)
				}
				if len(got) > 0 {
					if got[0].ID == "" {
						t.Errorf("ListProducts() got = %v, want %v", got[0].ID, "not empty")
					}
					if got[0].Name == "" {
						t.Errorf("ListProducts() got = %v, want %v", got[0].Name, "not empty")
					}
				}
			})
		}
		return test.EndTestErr
	})
}
func TestLoadProductRoles(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		products, err := seeders.CreateStripeProductPrices(ctx, dbxx, 2)
		if err != nil {
			t.Fatalf("failed to create stripe products: %v", err)
		}

		// Create roles and product_roles relationships
		role, err := queries.FindOrCreateRole(ctx, dbxx, "basic")
		if err != nil {
			t.Fatalf("failed to create roles: %v", err)
		}
		role2, err := queries.FindOrCreateRole(ctx, dbxx, "premium")
		if err != nil {
			t.Fatalf("failed to create roles: %v", err)
		}
		// Assign roles to first product
		for _, product := range products {
			err = queries.CreateProductRoles(ctx, dbxx, product.ID, role.ID, role2.ID)
			if err != nil {
				t.Fatalf("failed to assign role to product: %v", err)
			}
		}

		productIds := []string{products[0].ID, products[1].ID}

		type args struct {
			ctx        context.Context
			db         db.Dbx
			productIds []string
		}
		tests := []struct {
			name      string
			args      args
			wantCount []int
			wantErr   bool
		}{
			{
				name: "Load product roles",
				args: args{
					ctx:        ctx,
					db:         dbxx,
					productIds: productIds,
				},
				wantCount: []int{2, 2}, // First product has 2 roles, second has 2 roles
				wantErr:   false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.LoadProductRoles(tt.args.ctx, tt.args.db, tt.args.productIds...)
				if (err != nil) != tt.wantErr {
					t.Errorf("LoadProductRoles() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				for i, roles := range got {
					count := tt.wantCount[i]
					if len(roles) != count {
						t.Errorf("LoadProductRoles() got = %v, want %v", len(roles), count)
					}
				}
			})
		}
		return test.EndTestErr
	})
}
