package repository_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/google/uuid"
	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/test"
)

func TestFindCustomerByUserId(t *testing.T) {
	ctx, dbx, pl := test.DbSetup()
	t.Cleanup(func() {
		repository.TruncateModels(ctx, dbx)
		pl.Close()
	})
	user, err := repository.CreateUser(ctx, dbx, &shared.AuthenticateUserParams{
		Email: "email",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = repository.UpsertCustomer(ctx, dbx, user.ID, "stripeid")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		repository.TruncateModels(ctx, dbx)
		pl.Close()
	})
	type args struct {
		ctx    context.Context
		dbx    bob.Executor
		userId uuid.UUID
	}
	tests := []struct {
		name    string
		args    args
		want    *models.StripeCustomer
		wantErr bool
	}{
		{
			name: "",
			args: args{
				ctx:    ctx,
				dbx:    dbx,
				userId: user.ID,
			},

			want: &models.StripeCustomer{
				ID:       user.ID,
				StripeID: "stripeid",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repository.FindCustomerByUserId(tt.args.ctx, tt.args.dbx, tt.args.userId)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindCustomerByUserId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindCustomerByUserId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpsertCustomer(t *testing.T) {
	ctx, dbx, pl := test.DbSetup()
	t.Cleanup(func() {
		repository.TruncateModels(ctx, dbx)
		pl.Close()
	})
	user, err := repository.CreateUser(ctx, dbx, &shared.AuthenticateUserParams{
		Email: "email",
	})
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		ctx              context.Context
		dbx              bob.Executor
		userId           uuid.UUID
		stripeCustomerId string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				ctx:              ctx,
				dbx:              dbx,
				userId:           user.ID,
				stripeCustomerId: "stripe",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := repository.UpsertCustomer(tt.args.ctx, tt.args.dbx, tt.args.userId, tt.args.stripeCustomerId); (err != nil) != tt.wantErr {
				t.Errorf("UpsertCustomer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpsertProduct(t *testing.T) {
	ctx, dbx, pl := test.DbSetup()
	t.Cleanup(func() {
		repository.TruncateModels(ctx, dbx)
		pl.Close()
	})
	type args struct {
		ctx     context.Context
		dbx     bob.Executor
		product *models.StripeProductSetter
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				ctx: ctx,
				dbx: dbx,
				product: &models.StripeProductSetter{
					ID:          omit.From("fsfkajfl;as"),
					Active:      omit.From(true),
					Name:        omitnull.From(" "),
					Description: omitnull.From(" "),
					Image:       omitnull.From(" "),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := repository.UpsertProduct(tt.args.ctx, tt.args.dbx, tt.args.product); (err != nil) != tt.wantErr {
				t.Errorf("UpsertProduct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpsertPrice(t *testing.T) {
	ctx, dbx, pl := test.DbSetup()
	t.Cleanup(func() {
		repository.TruncateModels(ctx, dbx)
		pl.Close()
	})
	type args struct {
		ctx   context.Context
		dbx   bob.Executor
		price *models.StripePriceSetter
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "",
			args: args{
				ctx: ctx,
				dbx: dbx,
				price: &models.StripePriceSetter{
					ID:        omit.From("fsfkajfl;as"),
					ProductID: omit.From("fsfkajfl;as"),
					Active:    omit.From(true),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := repository.UpsertPrice(tt.args.ctx, tt.args.dbx, tt.args.price); (err != nil) != tt.wantErr {
				t.Errorf("UpsertPrice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpsertSubscription(t *testing.T) {
	type args struct {
		ctx          context.Context
		dbx          bob.Executor
		subscription *models.StripeSubscriptionSetter
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := repository.UpsertSubscription(tt.args.ctx, tt.args.dbx, tt.args.subscription); (err != nil) != tt.wantErr {
				t.Errorf("UpsertSubscription() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
