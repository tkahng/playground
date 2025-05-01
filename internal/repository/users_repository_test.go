package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/db/models/factory"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/test"
)

func TestUpdateUserEmailConfirm(t *testing.T) {
	f := factory.New()
	// f.AddBaseUserMod(
	// 	factory.UserMods.RandomEmail(nil),
	// )

	ctx, dbx, pl := test.DbSetup()
	t.Cleanup(func() {
		repository.TruncateModels(ctx, dbx)
		pl.Close()
	})
	type args struct {
		ctx    context.Context
		db     repository.Queryer
		userId uuid.UUID
	}
	tests := []struct {
		name    string
		args    args
		want    *models.User
		wantErr bool
	}{
		{
			name: "",
			args: args{
				ctx:    ctx,
				db:     dbx,
				userId: uuid.New(),
			},
			want:    &models.User{},
			wantErr: false,
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			f.NewUser(
				factory.UserMods.ID(tt.args.userId),
				factory.UserMods.RandomEmail(nil),
			).Create(ctx, dbx)
			got, err := repository.UpdateUserEmailConfirm(tt.args.ctx, tt.args.db, tt.args.userId, time.Now())
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateUserEmailConfirm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.EmailVerifiedAt.IsNull() {
				t.Errorf("UpdateUserEmailConfirm() = %v, want %v", got.EmailVerifiedAt, tt.want.EmailVerifiedAt)
			}
		})
	}
}
