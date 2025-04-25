package core_test

import (
	"github.com/tkahng/authgo/internal/core"
)

var (
	authOptions = core.DefaultAuthSettings()
)

// func TestCreateAuthenticationToken(t *testing.T) {
// 	config := authOptions.AccessToken

// 	type args struct {
// 		payload *core.AuthenticationPayload
// 		config  core.TokenOption
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		// want    string
// 		wantErr bool
// 	}{
// 		struct {
// 			name    string
// 			args    args
// 			wantErr bool
// 		}{
// 			name: "no error",
// 			args: args{
// 				payload: &core.AuthenticationPayload{
// 					UserId: uuid.New(),
// 					Email:  "tkahng@gmail.com",
// 				},
// 				config: config,
// 			},
// 			wantErr: false,
// 		},

// 		{
// 			name: "error payload is nil",
// 			args: args{
// 				payload: nil,
// 				config:  config,
// 			},
// 			wantErr: true,
// 		},
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			_, err := core.CreateAuthenticationToken(tt.args.payload, tt.args.config)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("CreateAuthenticationToken() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			// if got != tt.want {
// 			// 	t.Errorf("CreateAuthenticationToken() = %v, want %v", got, tt.want)
// 			// }
// 		})
// 	}
// }

// func TestVerifyAuthenticationToken(t *testing.T) {
// 	config := authOptions.AccessToken
// 	fake := faker.New()
// 	type args struct {
// 		token  string
// 		config core.TokenOption
// 	}
// 	type test struct {
// 		name    string
// 		args    args
// 		want    *core.AuthenticationPayload
// 		wantErr bool
// 	}
// 	var tests []test
// 	for range 10 {
// 		payload := &core.AuthenticationPayload{
// 			UserId: uuid.New(),
// 			Email:  fake.Internet().Email(),
// 		}
// 		token, err := core.CreateAuthenticationToken(payload, config)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		tests = append(tests, test{
// 			name: "no error",
// 			args: args{
// 				token:  token,
// 				config: config,
// 			},
// 			want:    payload,
// 			wantErr: false,
// 		})
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := core.VerifyAuthenticationToken(tt.args.token, tt.args.config)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("VerifyAuthenticationToken() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if got.UserId != tt.want.UserId {
// 				t.Errorf("VerifyAuthenticationToken() = %v, want %v", got.UserId, tt.want.UserId)
// 			}
// 			if got.Email != tt.want.Email {
// 				t.Errorf("VerifyAuthenticationToken() = %v, want %v", got.Email, tt.want.Email)
// 			}
// 		})
// 	}
// }

// func TestCreateRefreshToken(t *testing.T) {
// 	ctx, db, pl := test.DbSetup()
// 	t.Cleanup(func() {
// 		repository.TruncateModels(ctx, db)
// 		pl.Close()
// 	})
// 	config := authOptions.RefreshToken
// 	fake := faker.New()
// 	t.Cleanup(func() {
// 		models.Tokens.Delete().Exec(ctx, db)
// 		models.UserAccounts.Delete().Exec(ctx, db)
// 		models.Users.Delete().Exec(ctx, db)
// 	})
// 	user, err := repository.CreateUser(ctx, db, &shared.AuthenticateUserParams{
// 		Email: fake.Internet().Email(),
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	type args struct {
// 		ctx     context.Context
// 		db      bob.Executor
// 		payload *core.RefreshTokenPayload
// 		config  core.TokenOption
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name: "",
// 			args: args{
// 				ctx: ctx,
// 				db:  db,
// 				payload: &core.RefreshTokenPayload{
// 					UserId: user.ID,
// 					Email:  user.Email,
// 					Token:  "a",
// 				},
// 				config: config,
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			_, err := core.CreateAndPersistRefreshToken(tt.args.ctx, tt.args.db, tt.args.payload, tt.args.config)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("CreateRefreshToken() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			count, err := models.Tokens.Query(models.SelectWhere.Tokens.Token.EQ(tt.args.payload.Token)).Count(ctx, db)
// 			if err != nil {
// 				t.Errorf("CreateRefreshToken() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			if count != 1 {
// 				t.Errorf("CreateRefreshToken() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestVerifyRefreshToken(t *testing.T) {
// 	config := authOptions.RefreshToken
// 	fake := faker.New()
// 	ctx, db, pl := test.DbSetup()
// 	t.Cleanup(func() {
// 		repository.TruncateModels(ctx, db)
// 		pl.Close()
// 	})
// 	user, err := repository.CreateUser(ctx, db, &shared.AuthenticateUserParams{
// 		Email: fake.Internet().Email(),
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	type args struct {
// 		ctx    context.Context
// 		db     bob.Executor
// 		token  string
// 		config core.TokenOption
// 	}
// 	type test struct {
// 		name    string
// 		args    args
// 		want    *core.RefreshTokenClaims
// 		wantErr bool
// 	}
// 	var tests []test
// 	for range 5 {
// 		key := uuid.NewString()
// 		token, err := core.CreateAndPersistRefreshToken(ctx, db, &core.RefreshTokenPayload{
// 			UserId: user.ID,
// 			Email:  user.Email,
// 			Token:  key,
// 		}, config)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		tests = append(tests, test{
// 			name: "no error",
// 			args: args{
// 				ctx:    ctx,
// 				db:     db,
// 				token:  token,
// 				config: config,
// 			},
// 			want: &core.RefreshTokenClaims{
// 				RefreshTokenPayload: core.RefreshTokenPayload{
// 					UserId: user.ID,
// 					Email:  user.Email,
// 					Token:  key,
// 				},
// 			},
// 			wantErr: false,
// 		})
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := core.VerifyRefreshToken(tt.args.ctx, tt.args.db, tt.args.token, tt.args.config)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("VerifyRefreshToken() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			count, err := models.Tokens.Query(models.SelectWhere.Tokens.Token.EQ(tt.want.Token)).Count(ctx, db)
// 			if err != nil {
// 				t.Errorf("CreateRefreshToken() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			if got.Token != tt.want.Token {
// 				t.Errorf("CreateRefreshToken() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			if got.Email != tt.want.Email {
// 				t.Errorf("CreateRefreshToken() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			if got.UserId != tt.want.UserId {
// 				t.Errorf("CreateRefreshToken() error = %v, wantErr %v", err, tt.wantErr)
// 				if count != 0 {
// 					t.Errorf("CreateRefreshToken() error = %v, wantErr %v", err, tt.wantErr)
// 				}
// 			}
// 		})
// 	}
// }

// func TestNewProviderStateClaims(t *testing.T) {
// 	type args struct {
// 		payload *ProviderStatePayload
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want ProviderStateClaims
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := NewProviderStateClaims(tt.args.payload); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("NewProviderStateClaims() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestNewPasswordResetClaims(t *testing.T) {
// 	type args struct {
// 		payload PasswordResetPayload
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want PasswordResetClaims
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := NewPasswordResetClaims(tt.args.payload); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("NewPasswordResetClaims() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestNewEmailVerificationClaims(t *testing.T) {
// 	type args struct {
// 		payload EmailVerificationPayload
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want EmailVerificationClaims
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := NewEmailVerificationClaims(tt.args.payload); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("NewEmailVerificationClaims() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
