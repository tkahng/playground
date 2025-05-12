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
