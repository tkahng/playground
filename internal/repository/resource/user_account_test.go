package resource

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/test"
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

	// t.Run("Q filter", func(t *testing.T) {
	// 	filter := &UserAccountFilter{
	// 		Q: "test",
	// 	}
	// 	where := filterFunc(filter)
	// 	assert.NotNil(t, where)
	// 	expected := map[string]any{
	// 		"_or": []map[string]any{
	// 			{"email": map[string]any{"_ilike": "%test%"}},
	// 			{"name": map[string]any{"_ilike": "%test%"}},
	// 		},
	// 	}
	// 	assert.Equal(t, expected, *where)
	// })
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
			// userAccount2 := &models.UserAccount{
			// 	UserID:            user.ID,
			// 	Provider:          models.ProvidersGoogle,
			// 	ProviderAccountID: user.ID.String(),
			// 	Type:              models.ProviderTypeOAuth,
			// }
			_, err = accountResource.Create(ctx, userAccount)
		})
	})
}

// func TestNewUserAccountRepositoryResource_SortFunc(t *testing.T) {
// 	db := &database.Queries{}
// 	repo := NewUserAccountRepositoryResource(db)
// 	sortFunc := repo.sort

// 	t.Run("nil filter returns nil", func(t *testing.T) {
// 		assert.Nil(t, sortFunc(nil))
// 	})

// 	t.Run("empty sort fields returns default", func(t *testing.T) {
// 		filter := &UserAccountFilter{}
// 		s, b := filter.Sort()
// 		fmt.Println("haa", s, b)
// 		order := sortFunc(filter)
// 		assert.NotNil(t, order)
// 		assert.Equal(t, map[string]string{"created_at": "desc"}, *order)
// 	})

// 	t.Run("invalid sort by returns nil map", func(t *testing.T) {
// 		filter := &UserAccountFilter{SortParams: SortParams{
// 			SortBy:    "notacol",
// 			SortOrder: "asc",
// 		}}
// 		order := sortFunc(filter)
// 		assert.Nil(t, order)
// 	})

// 	t.Run("valid sort by returns map", func(t *testing.T) {
// 		filter := &UserAccountFilter{SortParams: SortParams{
// 			SortBy:    "email",
// 			SortOrder: "desc",
// 		}}
// 		order := sortFunc(filter)
// 		assert.NotNil(t, order)
// 		assert.Equal(t, map[string]string{"email": "desc"}, *order)
// 	})
// }

// func TestNewUserAccountRepositoryResource_PaginationFunc(t *testing.T) {
// 	db := &database.Queries{}
// 	repo := NewUserAccountRepositoryResource(db)
// 	paginationFunc := repo.pagination

// 	t.Run("nil input returns default", func(t *testing.T) {
// 		limit, offset := paginationFunc(nil)
// 		assert.Equal(t, 10, limit)
// 		assert.Equal(t, 0, offset)
// 	})

// 	t.Run("negative page returns page 0", func(t *testing.T) {
// 		input := &UserAccountFilter{PaginatedInput: PaginatedInput{Page: -2, PerPage: 5}}
// 		limit, offset := paginationFunc(input)
// 		assert.Equal(t, 5, limit)
// 		assert.Equal(t, 0, offset)
// 	})

// 	t.Run("perPage < 1 returns default", func(t *testing.T) {
// 		input := &UserAccountFilter{PaginatedInput: PaginatedInput{Page: 2, PerPage: 0}}
// 		limit, offset := paginationFunc(input)
// 		assert.Equal(t, 10, limit)
// 		assert.Equal(t, 20, offset)
// 	})

// 	t.Run("normal values", func(t *testing.T) {
// 		input := &UserAccountFilter{PaginatedInput: PaginatedInput{Page: 3, PerPage: 15}}
// 		limit, offset := paginationFunc(input)
// 		assert.Equal(t, 15, limit)
// 		assert.Equal(t, 45, offset)
// 	})
// }

// func TestUserAccountRepository_create(t *testing.T) {
// 	test.DbSetup()
// 	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
// 		useraccountResource := NewUserAccountRepositoryResource(db)
// 		useraccount, err := useraccountResource.Create(ctx, &models.UserAccount{
// 			Name:  types.Pointer("Test UserAccount"),
// 			Email: "duplicate@email.com",
// 		})
// 		if err != nil || useraccount == nil {
// 			t.Fatalf("Failed to create useraccount: %v", err)
// 		}
// 		type args struct {
// 			ctx   context.Context
// 			model *models.UserAccount
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			want    *models.UserAccount
// 			wantErr bool
// 			err     error
// 		}{
// 			{
// 				name: "successfully create useraccount email ",
// 				args: args{
// 					ctx: ctx,
// 					model: &models.UserAccount{
// 						Name:  types.Pointer("Test UserAccount"),
// 						Email: "test@example.com",
// 					},
// 				},
// 				want: &models.UserAccount{
// 					Name:  types.Pointer("Test UserAccount"),
// 					Email: "test@example.com",
// 				},
// 			},
// 			{
// 				name: "successfully create useraccount with email and image",
// 				args: args{
// 					ctx: ctx,
// 					model: &models.UserAccount{
// 						Name:  types.Pointer("Test UserAccount With Image"),
// 						Email: "test-with-image@example.com",
// 					},
// 				},
// 				want: &models.UserAccount{
// 					Name:  types.Pointer("Test UserAccount With Image"),
// 					Email: "test-with-image@example.com",
// 				},
// 			},
// 			{
// 				name: "error creating useraccount with same mail",
// 				args: args{
// 					ctx: ctx,
// 					model: &models.UserAccount{
// 						Email: "duplicate@email.com",
// 					},
// 				},
// 				wantErr: true,
// 				err:     errors.New("duplicate key value violates unique constraint \"useraccounts_email_key\" (SQLSTATE 23505)"),
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				got, err := useraccountResource.Create(tt.args.ctx, tt.args.model)
// 				if err != nil {
// 					if !tt.wantErr {
// 						t.Errorf("UserAccountRepository.create() error = %v, wantErr %v", err, tt.wantErr)
// 					} else if !strings.Contains(err.Error(), tt.err.Error()) {
// 						t.Errorf("UserAccountRepository.create() error = %v, want %v", err, tt.err)
// 					}
// 				}
// 				if got != nil && tt.want != nil {
// 					if got.Name == nil && tt.want.Name != nil {
// 						t.Errorf("UserAccountRepository.create() got = %v, want %v", got.Name, tt.want.Name)
// 					}
// 					if got.Name != nil && tt.want.Name != nil && *got.Name != *tt.want.Name {
// 						t.Errorf("UserAccountRepository.create() got = %s, want %s", *got.Name, *tt.want.Name)
// 					}
// 					if got.Email != tt.want.Email {
// 						t.Errorf("UserAccountRepository.create() got = %s, want %s", got.Email, tt.want.Email)
// 					}
// 					if got.EmailVerifiedAt != tt.want.EmailVerifiedAt {
// 						t.Errorf("UserAccountRepository.create() got = %v, want %v", got.EmailVerifiedAt, tt.want.EmailVerifiedAt)
// 					}
// 					if got.Image != tt.want.Image {
// 						t.Errorf("UserAccountRepository.create() got = %v, want %v", got.Image, tt.want.Image)
// 					}
// 				}
// 			})
// 		}
// 	})
// }

// func TestUserAccountRepsository_find(t *testing.T) {
// 	test.DbSetup()
// 	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
// 		useraccountsInput := []*models.UserAccount{
// 			{
// 				Name:            types.Pointer("Alpha UserAccount"),
// 				Email:           "alpha@example.com",
// 				EmailVerifiedAt: types.Pointer(time.Now()),
// 			},
// 			{
// 				Name:  types.Pointer("Beta UserAccount"),
// 				Email: "beta@example.com",
// 			},
// 			{
// 				Name:  types.Pointer("Charlie UserAccount"),
// 				Email: "charlie@example.com",
// 			},
// 			{
// 				Name:            types.Pointer("Delta UserAccount"),
// 				Email:           "delta@example.com",
// 				EmailVerifiedAt: types.Pointer(time.Now()),
// 			},
// 			{
// 				Name:  types.Pointer("Echo UserAccount"),
// 				Email: "echo@example.com",
// 			},
// 			{
// 				Name:  types.Pointer("Foxtrot UserAccount"),
// 				Email: "foxtrot@example.com",
// 			},
// 			{
// 				Name:            types.Pointer("Gamma UserAccount"),
// 				Email:           "gamma@example.com",
// 				EmailVerifiedAt: types.Pointer(time.Now()),
// 			},
// 			{
// 				Name:  types.Pointer("Hotel UserAccount"),
// 				Email: "hotel@example.com",
// 			},
// 			{
// 				Name:  types.Pointer("Yankee UserAccount"),
// 				Email: "yankee@example.com",
// 			},
// 			{
// 				Name:  types.Pointer("Zeta UserAccount"),
// 				Email: "zeta@example.com",
// 			},
// 		}
// 		useraccountResource := NewUserAccountRepositoryResource(db)
// 		for _, useraccount := range useraccountsInput {
// 			_, err := useraccountResource.Create(ctx, useraccount)
// 			if err != nil {
// 				t.Fatalf("Failed to create useraccount: %v", err)
// 			}
// 		}
// 		type args struct {
// 			ctx    context.Context
// 			filter *UserAccountFilter
// 		}
// 		tests := []struct {
// 			name      string
// 			args      args
// 			predicate func(t *testing.T, got []*models.UserAccount, err error)
// 		}{
// 			{
// 				name: "find all useraccounts sorted by name ascending",
// 				args: args{
// 					ctx: ctx,
// 					filter: &UserAccountFilter{
// 						PaginatedInput: PaginatedInput{
// 							Page:    0,
// 							PerPage: 10,
// 						},
// 						SortParams: SortParams{
// 							SortBy:    "name",
// 							SortOrder: "asc",
// 						},
// 					},
// 				},
// 				predicate: func(t *testing.T, got []*models.UserAccount, err error) {
// 					if err != nil {
// 						t.Errorf("UserAccountRepository.find() error = %v", err)
// 					}
// 					CheckSliceLength(t, got, 10)
// 					for i := 1; i < len(got)-1; i++ {
// 						firstName, secondName := *got[i].Name, *got[i+1].Name
// 						if firstName > secondName {
// 							t.Errorf("useraccounts are not in order. first name %s > second name %s", firstName, secondName)
// 						}
// 					}
// 				},
// 			},
// 			{
// 				name: "find all useraccounts sorted by name ascending, 3 per page, page 0",
// 				args: args{
// 					ctx: ctx,
// 					filter: &UserAccountFilter{
// 						PaginatedInput: PaginatedInput{
// 							Page:    0,
// 							PerPage: 3,
// 						},
// 						SortParams: SortParams{
// 							SortBy:    "name",
// 							SortOrder: "asc",
// 						},
// 					},
// 				},
// 				predicate: func(t *testing.T, got []*models.UserAccount, err error) {
// 					if err != nil {
// 						t.Errorf("UserAccountRepository.find() error = %v", err)
// 					}
// 					CheckSliceLength(t, got, 3)
// 					CheckUserAccountOrderByName(t, got)
// 				},
// 			},

// 			{
// 				name: "find all useraccounts sorted by name ascending, 3 per page, page 1",
// 				args: args{
// 					ctx: ctx,
// 					filter: &UserAccountFilter{
// 						PaginatedInput: PaginatedInput{
// 							Page:    1,
// 							PerPage: 3,
// 						},
// 						SortParams: SortParams{
// 							SortBy:    "name",
// 							SortOrder: "asc",
// 						},
// 					},
// 				},
// 				predicate: func(t *testing.T, got []*models.UserAccount, err error) {
// 					if err != nil {
// 						t.Errorf("UserAccountRepository.find() error = %v", err)
// 					}
// 					CheckSliceLength(t, got, 3)
// 					CheckUserAccountOrderByName(t, got)
// 				},
// 			},
// 			{
// 				name: "find all useraccounts sorted by name ascending, 3 per page, page 2",
// 				args: args{
// 					ctx: ctx,
// 					filter: &UserAccountFilter{
// 						PaginatedInput: PaginatedInput{
// 							Page:    2,
// 							PerPage: 3,
// 						},
// 						SortParams: SortParams{
// 							SortBy:    "name",
// 							SortOrder: "asc",
// 						},
// 					},
// 				},
// 				predicate: func(t *testing.T, got []*models.UserAccount, err error) {
// 					if err != nil {
// 						t.Errorf("UserAccountRepository.find() error = %v", err)
// 					}
// 					CheckSliceLength(t, got, 3)
// 					CheckUserAccountOrderByName(t, got)
// 				},
// 			},
// 			{
// 				name: "find all useraccounts sorted by name ascending, 3 per page, page 3",
// 				args: args{
// 					ctx: ctx,
// 					filter: &UserAccountFilter{
// 						PaginatedInput: PaginatedInput{
// 							Page:    3,
// 							PerPage: 3,
// 						},
// 						SortParams: SortParams{
// 							SortBy:    "name",
// 							SortOrder: "asc",
// 						},
// 					},
// 				},
// 				predicate: func(t *testing.T, got []*models.UserAccount, err error) {
// 					if err != nil {
// 						t.Errorf("UserAccountRepository.find() error = %v", err)
// 					}
// 					CheckSliceLength(t, got, 1)
// 					if got[0].Name == nil || *got[0].Name != "Zeta UserAccount" {
// 						t.Errorf("UserAccountRepository.find() got = %s, want %s", *got[0].Name, "Zeta UserAccount")
// 					}
// 				},
// 			},
// 			{
// 				name: "find all useraccounts with 'ta' in name. sorted by name ascending, 10 per page, page 0",
// 				args: args{
// 					ctx: ctx,
// 					filter: &UserAccountFilter{
// 						PaginatedInput: PaginatedInput{
// 							Page:    0,
// 							PerPage: 10,
// 						},
// 						SortParams: SortParams{
// 							SortBy:    "name",
// 							SortOrder: "asc",
// 						},
// 						Q: "ta",
// 					},
// 				},
// 				predicate: func(t *testing.T, got []*models.UserAccount, err error) {
// 					if err != nil {
// 						t.Errorf("UserAccountRepository.find() error = %v", err)
// 					}
// 					CheckSliceLength(t, got, 3)
// 					CheckUserAccountOrderByName(t, got)
// 				},
// 			},
// 			{
// 				name: "find all useraccounts that are verified. sorted by name ascending, 10 per page, page 0",
// 				args: args{
// 					ctx: ctx,
// 					filter: &UserAccountFilter{
// 						PaginatedInput: PaginatedInput{
// 							Page:    0,
// 							PerPage: 10,
// 						},
// 						SortParams: SortParams{
// 							SortBy:    "name",
// 							SortOrder: "asc",
// 						},
// 						EmailVerified: types.OptionalParam[bool]{IsSet: true, Value: true},
// 					},
// 				},
// 				predicate: func(t *testing.T, got []*models.UserAccount, err error) {
// 					if err != nil {
// 						t.Errorf("UserAccountRepository.find() error = %v", err)
// 					}
// 					CheckSliceLength(t, got, 3)
// 					CheckUserAccountOrderByName(t, got)
// 				},
// 			},
// 			{
// 				name: "find all useraccounts that are verified. sorted by name ascending, 10 per page, page 0",
// 				args: args{
// 					ctx: ctx,
// 					filter: &UserAccountFilter{
// 						PaginatedInput: PaginatedInput{
// 							Page:    0,
// 							PerPage: 10,
// 						},
// 						SortParams: SortParams{
// 							SortBy:    "name",
// 							SortOrder: "asc",
// 						},
// 						EmailVerified: types.OptionalParam[bool]{IsSet: true, Value: true},
// 					},
// 				},
// 				predicate: func(t *testing.T, got []*models.UserAccount, err error) {
// 					if err != nil {
// 						t.Errorf("UserAccountRepository.find() error = %v", err)
// 					}
// 					CheckSliceLength(t, got, 3)
// 					CheckUserAccountOrderByName(t, got)
// 				},
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				got, err := useraccountResource.Find(tt.args.ctx, tt.args.filter)
// 				tt.predicate(t, got, err)
// 			})
// 		}
// 	})
// }

// func CheckSliceLength(t *testing.T, got []*models.UserAccount, expected int) {
// 	if len(got) != expected {
// 		t.Errorf("UserAccountRepository.find() got = %d, want %d", len(got), expected)
// 	}
// }

// func CheckUserAccountOrderByName(t *testing.T, got []*models.UserAccount) {
// 	for i := 1; i < len(got)-1; i++ {
// 		firstName, secondName := *got[i].Name, *got[i+1].Name
// 		if firstName > secondName {
// 			t.Errorf("useraccounts are not in order. first name %s > second name %s", firstName, secondName)
// 		}
// 	}
// }
