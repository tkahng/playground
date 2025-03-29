package repository_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/test"
	"github.com/tkahng/authgo/internal/tools/security"
)

func TestCreateToken(t *testing.T) {
	ctx, dbx, pl := test.DbSetup()
	t.Cleanup(func() {
		repository.TruncateModels(ctx, dbx)
		pl.Close()
	})
	type args struct {
		params *repository.TokenDTO
	}
	tests := []struct {
		name    string
		args    args
		want    *models.Token
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				params: &repository.TokenDTO{
					Type:       models.TokenTypesReauthenticationToken,
					Identifier: "1",
					Expires:    time.Now().Add(time.Hour * 24),
					Token:      "1",
				},
			},
			want: &models.Token{
				Type:       models.TokenTypesReauthenticationToken,
				Identifier: "1",
				Expires:    time.Now().Add(time.Hour * 24),
				Token:      "1",
			},
			wantErr: false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.want.Expires = tt.args.params.Expires
			got, err := repository.CreateToken(ctx, dbx, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Identifier != tt.want.Identifier {
				t.Errorf("CreateToken() = %v, want %v", got.Identifier, tt.want.Identifier)
			}
			if got.Type != tt.want.Type {
				t.Errorf("CreateToken() = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Token != tt.want.Token {
				t.Errorf("CreateToken() = %v, want %v", got.Token, tt.want.Token)
			}
		})
	}
}

func TestUseToken(t *testing.T) {
	ctx, dbx, pl := test.DbSetup()
	t.Cleanup(func() {
		repository.TruncateModels(ctx, dbx)
		pl.Close()
	})
	type args struct {
		params string
	}
	tests := []struct {
		name    string
		args    args
		want    *models.Token
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repository.UseToken(ctx, dbx, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("UseToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UseToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func randomToken(ctx context.Context, db bob.DB, email string, tokenType models.TokenTypes, otp string) (*models.Token, error) {
	token := &models.TokenSetter{
		Identifier: omit.From(email),
		Type:       omit.From(tokenType),
		Token:      omit.From(security.GenerateTokenKey()),
		Expires:    omit.From(time.Now().Add(time.Duration(259200) * time.Second)),
		Otp:        omitnull.From(otp),
	}
	return models.Tokens.Insert(token).One(ctx, db)
}

func TestDeleteTokensByUser(t *testing.T) {
	ctx, dbx, pl := test.DbSetup()
	t.Cleanup(func() {
		repository.TruncateModels(ctx, dbx)
		pl.Close()
	})
	identifier := "email@example.com"
	t1, _ := randomToken(ctx, dbx, identifier, models.TokenTypesReauthenticationToken, security.GenerateOtp(6))
	t2, _ := randomToken(ctx, dbx, identifier, models.TokenTypesPasswordResetToken, security.GenerateOtp(6))
	t3, _ := randomToken(ctx, dbx, identifier, models.TokenTypesRefreshToken, security.GenerateOtp(6))
	t4, _ := randomToken(ctx, dbx, identifier, models.TokenTypesStateToken, security.GenerateOtp(6))
	t5, _ := randomToken(ctx, dbx, identifier, models.TokenTypesVerificationToken, security.GenerateOtp(6))
	type args struct {
		ctx    context.Context
		db     bob.Executor
		params *repository.OtpDto
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "",
			args: args{
				ctx: ctx,
				db:  dbx,
				params: &repository.OtpDto{
					Type:       t5.Type,
					Identifier: t5.Identifier,
					Otp:        t5.Otp.Ptr(),
				},
			},
			want:    4,
			wantErr: false,
		},
		{
			name: "",
			args: args{
				ctx: ctx,
				db:  dbx,
				params: &repository.OtpDto{
					Type:       t4.Type,
					Identifier: t4.Identifier,
					Otp:        t4.Otp.Ptr(),
				},
			},
			want:    3,
			wantErr: false,
		},
		{
			name: "",
			args: args{
				ctx: ctx,
				db:  dbx,
				params: &repository.OtpDto{
					Type:       t3.Type,
					Identifier: t3.Identifier,
					Otp:        t3.Otp.Ptr(),
				},
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "",
			args: args{
				ctx: ctx,
				db:  dbx,
				params: &repository.OtpDto{
					Type:       t2.Type,
					Identifier: t2.Identifier,
					Otp:        t2.Otp.Ptr(),
				},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "",
			args: args{
				ctx: ctx,
				db:  dbx,
				params: &repository.OtpDto{
					Type:       t1.Type,
					Identifier: t1.Identifier,
					Otp:        t1.Otp.Ptr(),
				},
			},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := repository.DeleteTokensByUser(tt.args.ctx, tt.args.db, tt.args.params); (err != nil) != tt.wantErr {
				t.Errorf("DeleteTokensByUser() error = %v, wantErr %v", err, tt.wantErr)
			}
			count, _ := models.Tokens.Query().Count(tt.args.ctx, tt.args.db)
			if count != tt.want {
				t.Errorf("DeleteTokensByUser() = %v, want %v", count, tt.want)
			}
		})
	}
}
