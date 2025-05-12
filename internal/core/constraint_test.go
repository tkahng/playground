package core_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/seeders"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/test"
)

func TestConstraintCheckerService_CannotHaveValidSubscription(t *testing.T) {
	ctx, dbx := test.DbSetup()

	dbx.RunInTransaction(ctx, func(tx db.Dbx) error {
		user, err := queries.CreateUser(ctx, tx, &shared.AuthenticationInput{
			Email: "test@example.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		prods, err := seeders.CreateStripeProductPrices(ctx, tx, 1)
		if err != nil {
			t.Fatalf("failed to create product prices: %v", err)
		}
		err = queries.UpsertSubscription(
			ctx,
			tx,
			&models.StripeSubscription{
				ID:      "sub_123",
				UserID:  user.ID,
				PriceID: prods[0].Prices[0].ID,
				Status:  models.StripeSubscriptionStatusActive,
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
			db  db.Dbx
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
				c := core.NewConstraintCheckerService(tt.fields.ctx, tt.fields.db)
				if err := c.CannotHaveValidSubscription(tt.args.userId); (err != nil) != tt.wantErr {
					t.Errorf("ConstraintCheckerService.CannotHaveValidSubscription() error = %v, wantErr %v", err, tt.wantErr)
					if err.Error() != "Cannot perform this action on a user with a valid subscription" {
						t.Errorf("unexpected error message: %v", err.Error())
					}
				}
			})
		}
		return test.EndTestErr
	})
}
