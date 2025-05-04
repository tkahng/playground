package core

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/shared"
)

func TestNewAuthStorage(t *testing.T) {
	type args struct {
		dbtx db.Dbx
	}
	tests := []struct {
		name string
		args args
		want *AuthAdapterBase
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAuthStorage(tt.args.dbtx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAuthStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthAdapterBase_GetToken(t *testing.T) {
	type fields struct {
		db db.Dbx
	}
	type args struct {
		ctx   context.Context
		token string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *shared.Token
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthAdapterBase{
				db: tt.fields.db,
			}
			got, err := a.GetToken(tt.args.ctx, tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthAdapterBase.GetToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AuthAdapterBase.GetToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthAdapterBase_SaveToken(t *testing.T) {
	type fields struct {
		db db.Dbx
	}
	type args struct {
		ctx   context.Context
		token *shared.CreateTokenDTO
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthAdapterBase{
				db: tt.fields.db,
			}
			if err := a.SaveToken(tt.args.ctx, tt.args.token); (err != nil) != tt.wantErr {
				t.Errorf("AuthAdapterBase.SaveToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthAdapterBase_DeleteToken(t *testing.T) {
	type fields struct {
		db db.Dbx
	}
	type args struct {
		ctx   context.Context
		token string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthAdapterBase{
				db: tt.fields.db,
			}
			if err := a.DeleteToken(tt.args.ctx, tt.args.token); (err != nil) != tt.wantErr {
				t.Errorf("AuthAdapterBase.DeleteToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthAdapterBase_VerifyTokenStorage(t *testing.T) {
	type fields struct {
		db db.Dbx
	}
	type args struct {
		ctx   context.Context
		token string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthAdapterBase{
				db: tt.fields.db,
			}
			if err := a.VerifyTokenStorage(tt.args.ctx, tt.args.token); (err != nil) != tt.wantErr {
				t.Errorf("AuthAdapterBase.VerifyTokenStorage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthAdapterBase_FindUserByEmail(t *testing.T) {
	type fields struct {
		db db.Dbx
	}
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *shared.User
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthAdapterBase{
				db: tt.fields.db,
			}
			got, err := a.FindUserByEmail(tt.args.ctx, tt.args.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthAdapterBase.FindUserByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AuthAdapterBase.FindUserByEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthAdapterBase_FindUserAccountByUserIdAndProvider(t *testing.T) {
	type fields struct {
		db db.Dbx
	}
	type args struct {
		ctx      context.Context
		userId   uuid.UUID
		provider shared.Providers
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *shared.UserAccount
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthAdapterBase{
				db: tt.fields.db,
			}
			got, err := a.FindUserAccountByUserIdAndProvider(tt.args.ctx, tt.args.userId, tt.args.provider)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthAdapterBase.FindUserAccountByUserIdAndProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AuthAdapterBase.FindUserAccountByUserIdAndProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthAdapterBase_UpdateUserAccount(t *testing.T) {
	type fields struct {
		db db.Dbx
	}
	type args struct {
		ctx     context.Context
		account *shared.UserAccount
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthAdapterBase{
				db: tt.fields.db,
			}
			if err := a.UpdateUserAccount(tt.args.ctx, tt.args.account); (err != nil) != tt.wantErr {
				t.Errorf("AuthAdapterBase.UpdateUserAccount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthAdapterBase_GetUserInfo(t *testing.T) {
	type fields struct {
		db db.Dbx
	}
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *shared.UserInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthAdapterBase{
				db: tt.fields.db,
			}
			got, err := a.GetUserInfo(tt.args.ctx, tt.args.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthAdapterBase.GetUserInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AuthAdapterBase.GetUserInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthAdapterBase_CreateUser(t *testing.T) {
	type fields struct {
		db db.Dbx
	}
	type args struct {
		ctx  context.Context
		user *shared.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *shared.User
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthAdapterBase{
				db: tt.fields.db,
			}
			got, err := a.CreateUser(tt.args.ctx, tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthAdapterBase.CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AuthAdapterBase.CreateUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthAdapterBase_DeleteUser(t *testing.T) {
	type fields struct {
		db db.Dbx
	}
	type args struct {
		ctx context.Context
		id  uuid.UUID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthAdapterBase{
				db: tt.fields.db,
			}
			if err := a.DeleteUser(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("AuthAdapterBase.DeleteUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthAdapterBase_LinkAccount(t *testing.T) {
	type fields struct {
		db db.Dbx
	}
	type args struct {
		ctx     context.Context
		account *shared.UserAccount
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthAdapterBase{
				db: tt.fields.db,
			}
			if err := a.LinkAccount(tt.args.ctx, tt.args.account); (err != nil) != tt.wantErr {
				t.Errorf("AuthAdapterBase.LinkAccount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthAdapterBase_UnlinkAccount(t *testing.T) {
	type fields struct {
		db db.Dbx
	}
	type args struct {
		ctx      context.Context
		userId   uuid.UUID
		provider shared.Providers
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthAdapterBase{
				db: tt.fields.db,
			}
			if err := a.UnlinkAccount(tt.args.ctx, tt.args.userId, tt.args.provider); (err != nil) != tt.wantErr {
				t.Errorf("AuthAdapterBase.UnlinkAccount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthAdapterBase_UpdateUser(t *testing.T) {
	type fields struct {
		db db.Dbx
	}
	type args struct {
		ctx  context.Context
		user *shared.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthAdapterBase{
				db: tt.fields.db,
			}
			if err := a.UpdateUser(tt.args.ctx, tt.args.user); (err != nil) != tt.wantErr {
				t.Errorf("AuthAdapterBase.UpdateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthAdapterBase_AssignUserRoles(t *testing.T) {
	type fields struct {
		db db.Dbx
	}
	type args struct {
		ctx       context.Context
		userId    uuid.UUID
		roleNames []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthAdapterBase{
				db: tt.fields.db,
			}
			if err := a.AssignUserRoles(tt.args.ctx, tt.args.userId, tt.args.roleNames...); (err != nil) != tt.wantErr {
				t.Errorf("AuthAdapterBase.AssignUserRoles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthAdapterBase_RunInTransaction(t *testing.T) {
	type fields struct {
		db db.Dbx
	}
	type args struct {
		ctx context.Context
		fn  func(AuthStorage) error
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthAdapterBase{
				db: tt.fields.db,
			}
			if err := a.RunInTransaction(tt.args.ctx, tt.args.fn); (err != nil) != tt.wantErr {
				t.Errorf("AuthAdapterBase.RunInTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
