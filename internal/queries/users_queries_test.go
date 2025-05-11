package queries_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db"
	crudModels "github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/test"
	"github.com/tkahng/authgo/internal/tools/types"
)

func TestCreateUser(t *testing.T) {
	ctx, dbx := test.DbSetup()

	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		type args struct {
			ctx    context.Context
			db     db.Dbx
			params *shared.AuthenticationInput
		}
		tests := []struct {
			name    string
			args    args
			want    *crudModels.User
			wantErr bool
		}{
			{
				name: "create user",
				args: args{
					ctx: ctx,
					db:  dbxx,
					params: &shared.AuthenticationInput{
						Email: "tkahng@gmail.com",
						Name:  types.Pointer("tchunoo"),
					},
				},
				want: &crudModels.User{
					Email: "tkahng@gmail.com",
					Name:  types.Pointer("tchunoo"),
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.CreateUser(tt.args.ctx, tt.args.db, tt.args.params)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got.Email, tt.want.Email) {
					t.Errorf("CreateUser() = %v, want %v", got.Email, tt.want.Email)
				}
			})
		}
		return errors.New("test error")
	})

}

func TestCreateUserRoles(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create a user
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Errorf("failed to create user: %v", err)
			return err
		}
		role, err := queries.FindOrCreateRole(ctx, dbxx, "basic")
		if err != nil {
			t.Errorf("failed to create role: %v", err)
			return err
		}
		type args struct {
			ctx     context.Context
			db      db.Dbx
			userId  uuid.UUID
			roleIds []uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "create user roles",
				args: args{
					ctx:     ctx,
					db:      dbxx,
					userId:  user.ID,
					roleIds: []uuid.UUID{role.ID},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := queries.CreateUserRoles(tt.args.ctx, tt.args.db, tt.args.userId, tt.args.roleIds...); (err != nil) != tt.wantErr {
					t.Errorf("createUserRoles() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
		return errors.New("test error")
	})
}

func TestCreateUserPermissions(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create a user
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Errorf("failed to create user: %v", err)
			return err
		}
		permission, err := queries.FindOrCreatePermission(ctx, dbxx, "basic")
		if err != nil {
			t.Errorf("failed to create role: %v", err)
			return err
		}
		type args struct {
			ctx           context.Context
			db            db.Dbx
			userId        uuid.UUID
			permissionIds []uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "create user permissions",
				args: args{
					ctx:           ctx,
					db:            dbxx,
					userId:        user.ID,
					permissionIds: []uuid.UUID{permission.ID},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := queries.CreateUserPermissions(tt.args.ctx, tt.args.db, tt.args.userId, tt.args.permissionIds...); (err != nil) != tt.wantErr {
					t.Errorf("CreateUserPermissions() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
		return errors.New("test error")
	})
}

func TestCreateAccount(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create a user
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "testuser@example.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
			return err
		}

		type args struct {
			ctx    context.Context
			db     db.Dbx
			userId uuid.UUID
			params *shared.AuthenticationInput
		}
		tests := []struct {
			name    string
			args    args
			want    *crudModels.UserAccount
			wantErr bool
		}{
			{
				name: "create credentials account",
				args: args{
					ctx:    ctx,
					db:     dbxx,
					userId: user.ID,
					params: &shared.AuthenticationInput{
						Type:              shared.ProviderTypeCredentials,
						Provider:          shared.ProvidersCredentials,
						ProviderAccountID: "credentials-account-id",
						Email:             "tkahng@gmail.com",
						Password:          types.Pointer("password"),
						HashPassword:      types.Pointer("password"),
					},
				},
				want: &crudModels.UserAccount{
					UserID:            user.ID,
					Type:              crudModels.ProviderTypeCredentials,
					Provider:          crudModels.ProvidersCredentials,
					ProviderAccountID: "credentials-account-id",
					Password:          types.Pointer("password"),
				},
				wantErr: false,
			},
			{
				name: "create oauth account",
				args: args{
					ctx:    ctx,
					db:     dbxx,
					userId: user.ID,
					params: &shared.AuthenticationInput{
						Type:              shared.ProviderTypeOAuth,
						Provider:          shared.ProvidersGoogle,
						ProviderAccountID: "google-account-id",
						AccessToken:       types.Pointer("google-access-token"),
						RefreshToken:      types.Pointer("google-refresh-token"),
					},
				},
				want: &crudModels.UserAccount{
					UserID:            user.ID,
					Type:              crudModels.ProviderTypeOAuth,
					Provider:          crudModels.ProvidersGoogle,
					ProviderAccountID: "google-account-id",
					AccessToken:       types.Pointer("google-access-token"),
					RefreshToken:      types.Pointer("google-refresh-token"),
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.CreateAccount(tt.args.ctx, tt.args.db, tt.args.userId, tt.args.params)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateAccount() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr {
					if got == nil {
						t.Errorf("CreateAccount() got = nil, want non-nil")
						return
					}
					if !reflect.DeepEqual(got.UserID, tt.want.UserID) {
						t.Errorf("CreateAccount() UserID = %v, want %v", got.UserID, tt.want.UserID)
					}
					if !reflect.DeepEqual(got.Type, tt.want.Type) {
						t.Errorf("CreateAccount() Type = %v, want %v", got.Type, tt.want.Type)
					}
					if !reflect.DeepEqual(got.Provider, tt.want.Provider) {
						t.Errorf("CreateAccount() Provider = %v, want %v", got.Provider, tt.want.Provider)
					}
					if !reflect.DeepEqual(got.ProviderAccountID, tt.want.ProviderAccountID) {
						t.Errorf("CreateAccount() ProviderAccountID = %v, want %v", got.ProviderAccountID, tt.want.ProviderAccountID)
					}
					if !reflect.DeepEqual(got.AccessToken, tt.want.AccessToken) {
						t.Errorf("CreateAccount() AccessToken = %v, want %v", got.AccessToken, tt.want.AccessToken)
					}
					if !reflect.DeepEqual(got.RefreshToken, tt.want.RefreshToken) {
						t.Errorf("CreateAccount() RefreshToken = %v, want %v", got.RefreshToken, tt.want.RefreshToken)
					}
					if !reflect.DeepEqual(got.Password, tt.want.Password) {
						t.Errorf("CreateAccount() Password = %v, want %v", got.Password, tt.want.Password)
					}
				}
			})
		}
		return errors.New("rollback transaction")
	})
}
func TestFindUserByEmail(t *testing.T) {
	ctx, dbx := test.DbSetup()

	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create a test user first
		testUser, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "test@example.com",
			Name:  types.Pointer("Test User"),
		})
		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}

		type args struct {
			ctx   context.Context
			db    db.Dbx
			email string
		}
		tests := []struct {
			name    string
			args    args
			want    *crudModels.User
			wantErr bool
		}{
			{
				name: "find existing user",
				args: args{
					ctx:   ctx,
					db:    dbxx,
					email: "test@example.com",
				},
				want:    testUser,
				wantErr: false,
			},
			{
				name: "user not found",
				args: args{
					ctx:   ctx,
					db:    dbxx,
					email: "nonexistent@example.com",
				},
				want:    nil,
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.FindUserByEmail(tt.args.ctx, tt.args.db, tt.args.email)
				if (err != nil) != tt.wantErr {
					t.Errorf("FindUserByEmail() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.want == nil && got != nil {
					t.Errorf("FindUserByEmail() = %v, want nil", got)
				}
				if tt.want != nil && got != nil {
					if !reflect.DeepEqual(got.Email, tt.want.Email) {
						t.Errorf("FindUserByEmail() Email = %v, want %v", got.Email, tt.want.Email)
					}
					if !reflect.DeepEqual(got.Name, tt.want.Name) {
						t.Errorf("FindUserByEmail() Name = %v, want %v", got.Name, tt.want.Name)
					}
				}
			})
		}
		return errors.New("rollback transaction")
	})
}
func TestFindUserById(t *testing.T) {
	ctx, dbx := test.DbSetup()

	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create a test user first
		testUser, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "test@example.com",
			Name:  types.Pointer("Test User"),
		})
		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}

		type args struct {
			ctx    context.Context
			db     db.Dbx
			userId uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			want    *crudModels.User
			wantErr bool
		}{
			{
				name: "find existing user",
				args: args{
					ctx:    ctx,
					db:     dbxx,
					userId: testUser.ID,
				},
				want:    testUser,
				wantErr: false,
			},
			{
				name: "user not found",
				args: args{
					ctx:    ctx,
					db:     dbxx,
					userId: uuid.New(),
				},
				want:    nil,
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.FindUserById(tt.args.ctx, tt.args.db, tt.args.userId)
				if (err != nil) != tt.wantErr {
					t.Errorf("FindUserById() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.want == nil && got != nil {
					t.Errorf("FindUserById() = %v, want nil", got)
				}
				if tt.want != nil && got != nil {
					if !reflect.DeepEqual(got.ID, tt.want.ID) {
						t.Errorf("FindUserById() ID = %v, want %v", got.ID, tt.want.ID)
					}
					if !reflect.DeepEqual(got.Email, tt.want.Email) {
						t.Errorf("FindUserById() Email = %v, want %v", got.Email, tt.want.Email)
					}
					if !reflect.DeepEqual(got.Name, tt.want.Name) {
						t.Errorf("FindUserById() Name = %v, want %v", got.Name, tt.want.Name)
					}
				}
			})
		}
		return errors.New("rollback transaction")
	})
}
func TestUpdateUserPassword(t *testing.T) {
	ctx, dbx := test.DbSetup()

	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// Create test user
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "test@example.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		// Create credentials account
		_, err = queries.CreateAccount(ctx, dbxx, user.ID, &shared.AuthenticationInput{
			Type:              shared.ProviderTypeCredentials,
			Provider:          shared.ProvidersCredentials,
			ProviderAccountID: "test-account",
			HashPassword:      types.Pointer("initial-password"),
		})
		if err != nil {
			t.Fatalf("failed to create account: %v", err)
		}

		type args struct {
			ctx      context.Context
			db       db.Dbx
			userId   uuid.UUID
			password string
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "update existing user password",
				args: args{
					ctx:      ctx,
					db:       dbxx,
					userId:   user.ID,
					password: "new-password",
				},
				wantErr: false,
			},
			{
				name: "update password for non-existent user",
				args: args{
					ctx:      ctx,
					db:       dbxx,
					userId:   uuid.New(),
					password: "new-password",
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := queries.UpdateUserPassword(tt.args.ctx, tt.args.db, tt.args.userId, tt.args.password)
				if (err != nil) != tt.wantErr {
					t.Errorf("UpdateUserPassword() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
		return errors.New("rollback transaction")
	})
}
