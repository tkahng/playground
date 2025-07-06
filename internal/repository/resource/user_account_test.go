package resource

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/test"
	"github.com/tkahng/authgo/internal/tools/logger"
)

func TestNewUserAccountRepositoryResource_FilterFunc(t *testing.T) {
	filterFunc := UserAccount.filter

	t.Run("nil filter returns empty map", func(t *testing.T) {
		where := filterFunc(nil)
		assert.NotNil(t, where)
		assert.Equal(t, 0, len(*where))
	})

	t.Run("Providers filter", func(t *testing.T) {
		filter := &UserAccountFilter{
			Providers: []models.Providers{"google", "github"},
		}
		where := filterFunc(filter)
		assert.NotNil(t, where)
		assert.Equal(t, map[string]any{
			"provider": map[string]any{"_in": []models.Providers{"google", "github"}},
		}, *where)
	})

	t.Run("Ids filter", func(t *testing.T) {
		id1 := uuid.New()
		id2 := uuid.New()
		filter := &UserAccountFilter{
			Ids: []uuid.UUID{id1, id2},
		}
		where := filterFunc(filter)
		assert.NotNil(t, where)
		assert.Equal(t, map[string]any{
			"id": map[string]any{"_in": []uuid.UUID{id1, id2}},
		}, *where)
	})

}

func TestUserAccountRepositoryResource_Create(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		userResource := User
		accountResource := UserAccount

		t.Run("Create with valid data", func(t *testing.T) {

			user, err := userResource.Create(ctx, db, &models.User{
				Email: "test@example.com",
			})
			assert.NoError(t, err)
			userAccount := &models.UserAccount{
				UserID:            user.ID,
				Provider:          models.ProvidersCredentials,
				ProviderAccountID: user.ID.String(),
				Type:              models.ProviderTypeCredentials,
			}
			created, err := accountResource.Create(ctx, db, userAccount)
			assert.NoError(t, err)
			assert.NotNil(t, created)
			assert.Equal(t, userAccount.Provider, created.Provider)
			assert.Equal(t, userAccount.ProviderAccountID, created.ProviderAccountID)
			assert.Equal(t, userAccount.Type, created.Type)
			assert.Equal(t, user.ID, created.UserID)
		})

		t.Run("Create with duplicate provider", func(t *testing.T) {
			user, err := userResource.Create(ctx, db, &models.User{
				Email: "test-duplicate-google@example.com",
			})
			if err != nil {
				t.Fatalf("Failed to create user: %v", err)
			}
			userAccount := &models.UserAccount{
				UserID:            user.ID,
				Provider:          models.ProvidersGoogle,
				ProviderAccountID: user.ID.String(),
				Type:              models.ProviderTypeOAuth,
			}
			userAccount2 := &models.UserAccount{
				UserID:            user.ID,
				Provider:          models.ProvidersGoogle,
				ProviderAccountID: user.ID.String(),
				Type:              models.ProviderTypeOAuth,
			}
			_, err = accountResource.Create(ctx, db, userAccount)
			assert.NoError(t, err)
			_, err = accountResource.Create(ctx, db, userAccount2)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "duplicate key value violates unique constraint")

		})
	})
}

func TestUserAccountRepsository_find(t *testing.T) {
	logger.SetDefaultLogger()
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		userResource := User
		user1, err := userResource.Create(ctx, db, &models.User{
			Email: "test@example.com",
		})
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		if user1 == nil {
			t.Fatal("User should not be nil")
		}
		user2, err := userResource.Create(ctx, db, &models.User{
			Email: "test2@example.com",
		})
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		if user2 == nil {
			t.Fatal("User should not be nil")
		}
		useraccountsInput := []*models.UserAccount{
			{
				UserID:            user1.ID,
				Type:              models.ProviderTypeOAuth,
				Provider:          models.ProvidersApple,
				ProviderAccountID: user1.ID.String(),
			},
			{
				UserID:            user2.ID,
				Type:              models.ProviderTypeCredentials,
				Provider:          models.ProvidersCredentials,
				ProviderAccountID: user2.ID.String(),
			},
			{
				UserID:            user1.ID,
				Type:              models.ProviderTypeOAuth,
				Provider:          models.ProvidersFacebook,
				ProviderAccountID: user1.ID.String(),
			},
			{
				UserID:            user1.ID,
				Type:              models.ProviderTypeOAuth,
				Provider:          models.ProvidersGithub,
				ProviderAccountID: user1.ID.String(),
			},
			{
				UserID:            user1.ID,
				Type:              models.ProviderTypeOAuth,
				Provider:          models.ProvidersGoogle,
				ProviderAccountID: user1.ID.String(),
			},
		}
		useraccountResource := UserAccount
		for _, useraccount := range useraccountsInput {
			_, err := useraccountResource.Create(ctx, db, useraccount)
			if err != nil {
				t.Fatalf("Failed to create useraccount: %v", err)
			}
		}
		type args struct {
			ctx    context.Context
			filter *UserAccountFilter
		}
		tests := []struct {
			name      string
			args      args
			predicate func(t *testing.T, got []*models.UserAccount, err error)
		}{
			{
				name: "find all useraccounts sorted by provider ascending, 3 per page, page 0",
				args: args{
					ctx: ctx,
					filter: &UserAccountFilter{
						PaginatedInput: repository.PaginatedInput{
							Page:    0,
							PerPage: 3,
						},
					},
				},
				predicate: func(t *testing.T, got []*models.UserAccount, err error) {
					if err != nil {
						t.Errorf("UserAccountRepository.find() error = %v", err)
					}

					if len(got) != 3 {
						t.Errorf("UserRepository.find() got = %d, want %d", len(got), 3)
					}
				},
			},

			{
				name: "find all useraccounts sorted by provider ascending, 3 per page, page 1",
				args: args{
					ctx: ctx,
					filter: &UserAccountFilter{
						PaginatedInput: repository.PaginatedInput{
							Page:    1,
							PerPage: 3,
						},
					},
				},
				predicate: func(t *testing.T, got []*models.UserAccount, err error) {
					if err != nil {
						t.Errorf("UserAccountRepository.find() error = %v", err)
					}
					if len(got) != 2 {
						t.Errorf("UserRepository.find() got = %d, want %d", len(got), 2)
					}

				},
			},
			{
				name: "find all useraccounts sorted by name ascending, 3 per page, page 2",
				args: args{
					ctx: ctx,
					filter: &UserAccountFilter{
						PaginatedInput: repository.PaginatedInput{
							Page:    2,
							PerPage: 3,
						},
					},
				},
				predicate: func(t *testing.T, got []*models.UserAccount, err error) {
					if err != nil {
						t.Errorf("UserAccountRepository.find() error = %v", err)
					}
					if len(got) != 0 {
						t.Errorf("UserRepository.find() got = %d, want %d", len(got), 0)
					}
				},
			},

			{
				name: "find accounts of user 1",
				args: args{
					ctx: ctx,
					filter: &UserAccountFilter{
						UserIds: []uuid.UUID{user1.ID},
					},
				},
				predicate: func(t *testing.T, got []*models.UserAccount, err error) {
					if err != nil {
						t.Errorf("UserAccountRepository.find() error = %v", err)
					}
					if len(got) != 4 {
						t.Errorf("UserRepository.find() got = %d, want %d", len(got), 4)
					}
					for _, account := range got {
						if account.UserID != user1.ID {
							t.Errorf("UserAccountRepository.find() got user ID = %s, want %s", account.UserID, user1.ID)
						}
					}
				},
			},
			{
				name: "find accounts of user 2",
				args: args{
					ctx: ctx,
					filter: &UserAccountFilter{
						UserIds: []uuid.UUID{user2.ID},
					},
				},
				predicate: func(t *testing.T, got []*models.UserAccount, err error) {
					if err != nil {
						t.Errorf("UserAccountRepository.find() error = %v", err)
					}
					if len(got) != 1 {
						t.Errorf("UserRepository.find() got = %d, want %d", len(got), 1)
					}
					for _, account := range got {
						if account.UserID != user2.ID {
							t.Errorf("UserAccountRepository.find() got user ID = %s, want %s", account.UserID, user2.ID)
						}
					}
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := useraccountResource.Find(tt.args.ctx, db, tt.args.filter)
				tt.predicate(t, got, err)
			})
		}
	})
}
