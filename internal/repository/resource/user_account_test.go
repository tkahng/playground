package resource

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/test"
	"github.com/tkahng/authgo/internal/tools/logger"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

func TestNewUserAccountRepositoryResource_FilterFunc(t *testing.T) {
	db := &database.Queries{} // Mock or use a real database connection as needed
	repo := NewUserAccountRepositoryResource(db)

	filterFunc := repo.filter

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
		userResource := NewUserRepositoryResource(db)
		accountResource := NewUserAccountRepositoryResource(db)

		t.Run("Create with valid data", func(t *testing.T) {

			user, err := userResource.Create(ctx, &models.User{
				Email: "test@example.com",
			})
			userAccount := &models.UserAccount{
				UserID:            user.ID,
				Provider:          models.ProvidersCredentials,
				ProviderAccountID: user.ID.String(),
				Type:              models.ProviderTypeCredentials,
			}
			created, err := accountResource.Create(ctx, userAccount)
			assert.NoError(t, err)
			assert.NotNil(t, created)
			assert.Equal(t, userAccount.Provider, created.Provider)
			assert.Equal(t, userAccount.ProviderAccountID, created.ProviderAccountID)
			assert.Equal(t, userAccount.Type, created.Type)
			assert.Equal(t, user.ID, created.UserID)
		})

		t.Run("Create with duplicate provider", func(t *testing.T) {
			user, err := userResource.Create(ctx, &models.User{
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
			_, err = accountResource.Create(ctx, userAccount)
			assert.NoError(t, err)
			_, err = accountResource.Create(ctx, userAccount2)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "duplicate key value violates unique constraint")

		})
	})
}

func TestUserAccountRepsository_find(t *testing.T) {
	logger.SetDefaultLogger()
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		userResource := NewUserRepositoryResource(db)
		user, err := userResource.Create(ctx, &models.User{
			Email: "test@example.com",
		})
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		useraccountsInput := []*models.UserAccount{
			{
				UserID:            user.ID,
				Type:              models.ProviderTypeOAuth,
				Provider:          models.ProvidersApple,
				ProviderAccountID: user.ID.String(),
			},
			{
				UserID:            user.ID,
				Type:              models.ProviderTypeCredentials,
				Provider:          models.ProvidersCredentials,
				ProviderAccountID: user.ID.String(),
			},
			{
				UserID:            user.ID,
				Type:              models.ProviderTypeOAuth,
				Provider:          models.ProvidersFacebook,
				ProviderAccountID: user.ID.String(),
			},
			{
				UserID:            user.ID,
				Type:              models.ProviderTypeOAuth,
				Provider:          models.ProvidersGithub,
				ProviderAccountID: user.ID.String(),
			},
			{
				UserID:            user.ID,
				Type:              models.ProviderTypeOAuth,
				Provider:          models.ProvidersGoogle,
				ProviderAccountID: user.ID.String(),
			},
		}
		useraccountResource := NewUserAccountRepositoryResource(db)
		for _, useraccount := range useraccountsInput {
			_, err := useraccountResource.Create(ctx, useraccount)
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
						PaginatedInput: PaginatedInput{
							Page:    0,
							PerPage: 3,
						},
						SortParams: SortParams{
							SortBy:    "provider",
							SortOrder: "asc",
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
					CheckUserAccountOrderByName(t, got)
				},
			},

			{
				name: "find all useraccounts sorted by provider ascending, 3 per page, page 1",
				args: args{
					ctx: ctx,
					filter: &UserAccountFilter{
						PaginatedInput: PaginatedInput{
							Page:    1,
							PerPage: 3,
						},
						SortParams: SortParams{
							SortBy:    "provider",
							SortOrder: "asc",
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
					for i := 1; i < len(got)-1; i++ {
						firstName, secondName := got[i].Provider.String(), got[i+1].Provider.String()
						if firstName > secondName {
							t.Errorf("users are not in order. first name %s > second name %s", firstName, secondName)
						}
					}
				},
			},
			{
				name: "find all useraccounts sorted by name ascending, 3 per page, page 2",
				args: args{
					ctx: ctx,
					filter: &UserAccountFilter{
						PaginatedInput: PaginatedInput{
							Page:    2,
							PerPage: 3,
						},
						SortParams: SortParams{
							SortBy:    "provider",
							SortOrder: "asc",
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
				name: "find all useraccounts with 'ta' in name. sorted by name ascending, 10 per page, page 0",
				args: args{
					ctx: ctx,
					filter: &UserAccountFilter{
						PaginatedInput: PaginatedInput{
							Page:    0,
							PerPage: 10,
						},
						SortParams: SortParams{
							SortBy:    "provider",
							SortOrder: "asc",
						},
						// Q: "ta",
					},
				},
				predicate: func(t *testing.T, got []*models.UserAccount, err error) {
					if err != nil {
						t.Errorf("UserAccountRepository.find() error = %v", err)
					}
					if len(got) != 5 {
						t.Errorf("UserRepository.find() got = %d, want %d", len(got), 5)
					}
					for i := 1; i < len(got)-1; i++ {
						firstName, secondName := got[i].Provider.String(), got[i+1].Provider.String()
						if firstName > secondName {
							allproviders := mapper.Map(got, func(u *models.UserAccount) string {
								return u.Provider.String()
							})
							t.Errorf("users are not in order. first name %s > second name %s. all names: %v", firstName, secondName, allproviders)
						}
					}
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := useraccountResource.Find(tt.args.ctx, tt.args.filter)
				tt.predicate(t, got, err)
			})
		}
	})
}
