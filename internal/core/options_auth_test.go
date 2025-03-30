package core_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/test"
	"github.com/tkahng/authgo/internal/tools/security"
)

func TestGetOrSetEncryptedAuthOptions(t *testing.T) {
	ctx, db, pl := test.DbSetup()
	encryptionKey := security.RandomString(32)
	t.Cleanup(func() {
		repository.TruncateModels(ctx, db)
		pl.Close()
	})
	type args struct {
		ctx           context.Context
		dbx           bob.DB
		encryptionKey string
	}
	tests := []struct {
		name    string
		args    args
		want    *core.AuthOptions
		wantErr bool
	}{
		{
			name: "",
			args: args{
				ctx:           ctx,
				dbx:           db,
				encryptionKey: encryptionKey,
			},
			want:    core.DefaultAuthSettings(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := core.GetOrSetEncryptedAuthOptions(tt.args.ctx, tt.args.dbx, tt.args.encryptionKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrSetEncryptedAuthOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOrSetEncryptedAuthOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}
