package resource

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/test"
	"github.com/tkahng/authgo/internal/tools/types"
)

func TestNewUserRepositoryResource_FilterFunc(t *testing.T) {
	db := &database.Queries{} // Mock or use a real database connection as needed
	repo := NewUserRepositoryResource(db)

	// get the filter function
	filterFunc := repo.filterFn

	t.Run("nil filter returns empty map", func(t *testing.T) {
		where := filterFunc(nil)
		assert.NotNil(t, where)
		assert.Equal(t, 0, len(*where))
	})

	t.Run("EmailVerified true", func(t *testing.T) {
		filter := &UserFilter{
			EmailVerified: types.OptionalParam[bool]{IsSet: true, Value: true},
		}
		where := filterFunc(filter)
		assert.NotNil(t, where)
		assert.Equal(t, map[string]any{
			"email_verified_at": map[string]any{"_isnotnull": nil},
		}, *where)
	})

	t.Run("EmailVerified false", func(t *testing.T) {
		filter := &UserFilter{
			EmailVerified: types.OptionalParam[bool]{IsSet: true, Value: false},
		}
		where := filterFunc(filter)
		assert.NotNil(t, where)
		assert.Equal(t, map[string]any{
			"email_verified_at": map[string]any{"_isnull": nil},
		}, *where)
	})

	t.Run("EmailVerified false not set", func(t *testing.T) {
		id := uuid.New()
		filter := &UserFilter{
			EmailVerified: types.OptionalParam[bool]{IsSet: false, Value: false},
			Ids: uuid.UUIDs{
				id,
			},
		}
		where := filterFunc(filter)
		assert.NotNil(t, where)
		assert.Equal(t, map[string]any{
			"id": map[string]any{"_in": []uuid.UUID{id}},
		}, *where)
	})

	t.Run("Emails filter", func(t *testing.T) {
		filter := &UserFilter{
			Emails: []string{"a@example.com", "b@example.com"},
		}
		where := filterFunc(filter)
		assert.NotNil(t, where)
		assert.Equal(t, map[string]any{
			"email": map[string]any{"_in": []string{"a@example.com", "b@example.com"}},
		}, *where)
	})

	t.Run("Ids filter", func(t *testing.T) {
		id1 := uuid.New()
		id2 := uuid.New()
		filter := &UserFilter{
			Ids: []uuid.UUID{id1, id2},
		}
		where := filterFunc(filter)
		assert.NotNil(t, where)
		assert.Equal(t, map[string]any{
			"id": map[string]any{"_in": []uuid.UUID{id1, id2}},
		}, *where)
	})

	t.Run("Providers filter", func(t *testing.T) {
		filter := &UserFilter{
			Providers: []models.Providers{"google", "github"},
		}
		where := filterFunc(filter)
		assert.NotNil(t, where)
		assert.Equal(t, map[string]any{
			"accounts": map[string]any{
				"provider": map[string]any{
					"_in": []models.Providers{"google", "github"},
				},
			},
		}, *where)
	})

	t.Run("RoleIds filter", func(t *testing.T) {
		role1 := uuid.New()
		role2 := uuid.New()
		filter := &UserFilter{
			RoleIds: []uuid.UUID{role1, role2},
		}
		where := filterFunc(filter)
		assert.NotNil(t, where)
		assert.Equal(t, map[string]any{
			"roles": map[string]any{
				"id": map[string]any{
					"_in": []uuid.UUID{role1, role2},
				},
			},
		}, *where)
	})

	t.Run("Q filter", func(t *testing.T) {
		filter := &UserFilter{
			Q: "foo",
		}
		where := filterFunc(filter)
		assert.NotNil(t, where)
		expected := map[string]any{
			"_or": []map[string]any{
				{"email": map[string]any{"_ilike": "%foo%"}},
				{"name": map[string]any{"_ilike": "%foo%"}},
			},
		}
		assert.Equal(t, expected, *where)
	})

	t.Run("Multiple filters combined", func(t *testing.T) {
		role := uuid.New()
		filter := &UserFilter{
			Emails:        []string{"a@example.com"},
			RoleIds:       []uuid.UUID{role},
			EmailVerified: types.OptionalParam[bool]{IsSet: true, Value: true},
		}
		where := filterFunc(filter)
		assert.NotNil(t, where)
		assert.Equal(t, map[string]any{
			"email":             map[string]any{"_in": []string{"a@example.com"}},
			"roles":             map[string]any{"id": map[string]any{"_in": []uuid.UUID{role}}},
			"email_verified_at": map[string]any{"_isnotnull": nil},
		}, *where)
	})

	t.Run("Empty filter returns nil", func(t *testing.T) {
		filter := &UserFilter{}
		where := filterFunc(filter)
		assert.Nil(t, where)
	})
	t.Run("Email verified at nil", func(t *testing.T) {
		filter := &UserFilter{
			EmailVerified: types.OptionalParam[bool]{IsSet: true, Value: false},
		}
		where := filterFunc(filter)
		assert.NotNil(t, where)
		assert.Equal(t, map[string]any{
			"email_verified_at": map[string]any{"_isnull": nil},
		}, *where)
	})
}

func TestNewUserRepositoryResource_SortFunc(t *testing.T) {
	db := &database.Queries{}
	repo := NewUserRepositoryResource(db)
	sortFunc := repo.sort

	t.Run("nil filter returns nil", func(t *testing.T) {
		assert.Nil(t, sortFunc(nil))
	})

	t.Run("empty sort fields returns default", func(t *testing.T) {
		filter := &UserFilter{}
		s, b := filter.Sort()
		fmt.Println("haa", s, b)
		order := sortFunc(filter)
		assert.Nil(t, order)
		// assert.Equal(t, map[string]string{"created_at": "desc"}, *order)
	})

	t.Run("invalid sort by returns nil map", func(t *testing.T) {
		filter := &UserFilter{SortParams: repository.SortParams{
			SortBy:    "notacol",
			SortOrder: "asc",
		}}
		order := sortFunc(filter)
		assert.Nil(t, order)
	})

	t.Run("valid sort by returns map", func(t *testing.T) {
		filter := &UserFilter{SortParams: repository.SortParams{
			SortBy:    "email",
			SortOrder: "desc",
		}}
		order := sortFunc(filter)
		assert.NotNil(t, order)
		assert.Equal(t, map[string]string{"email": "desc"}, *order)
	})
}

func TestNewUserRepositoryResource_PaginationFunc(t *testing.T) {
	db := &database.Queries{}
	repo := NewUserRepositoryResource(db)
	paginationFunc := repo.pagination

	t.Run("nil input returns default", func(t *testing.T) {
		limit, offset := paginationFunc(nil)
		assert.Equal(t, 10, limit)
		assert.Equal(t, 0, offset)
	})

	t.Run("negative page returns page 0", func(t *testing.T) {
		input := &UserFilter{PaginatedInput: repository.PaginatedInput{Page: -2, PerPage: 5}}
		limit, offset := paginationFunc(input)
		assert.Equal(t, 5, limit)
		assert.Equal(t, 0, offset)
	})

	t.Run("perPage < 1 returns default", func(t *testing.T) {
		input := &UserFilter{PaginatedInput: repository.PaginatedInput{Page: 2, PerPage: 0}}
		limit, offset := paginationFunc(input)
		assert.Equal(t, 10, limit)
		assert.Equal(t, 20, offset)
	})

	t.Run("normal values", func(t *testing.T) {
		input := &UserFilter{PaginatedInput: repository.PaginatedInput{Page: 3, PerPage: 15}}
		limit, offset := paginationFunc(input)
		assert.Equal(t, 15, limit)
		assert.Equal(t, 45, offset)
	})
}

func TestUserRepository_create(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		userResource := NewUserRepositoryResource(db)
		user, err := userResource.Create(ctx, &models.User{
			Name:  types.Pointer("Test User"),
			Email: "duplicate@email.com",
		})
		if err != nil || user == nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		type args struct {
			ctx   context.Context
			model *models.User
		}
		tests := []struct {
			name    string
			args    args
			want    *models.User
			wantErr bool
			err     error
		}{
			{
				name: "successfully create user email ",
				args: args{
					ctx: ctx,
					model: &models.User{
						Name:  types.Pointer("Test User"),
						Email: "test@example.com",
					},
				},
				want: &models.User{
					Name:  types.Pointer("Test User"),
					Email: "test@example.com",
				},
			},
			{
				name: "successfully create user with email and image",
				args: args{
					ctx: ctx,
					model: &models.User{
						Name:  types.Pointer("Test User With Image"),
						Email: "test-with-image@example.com",
					},
				},
				want: &models.User{
					Name:  types.Pointer("Test User With Image"),
					Email: "test-with-image@example.com",
				},
			},
			{
				name: "error creating user with same mail",
				args: args{
					ctx: ctx,
					model: &models.User{
						Email: "duplicate@email.com",
					},
				},
				wantErr: true,
				err:     errors.New("duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)"),
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := userResource.Create(tt.args.ctx, tt.args.model)
				if err != nil {
					if !tt.wantErr {
						t.Errorf("UserRepository.create() error = %v, wantErr %v", err, tt.wantErr)
					} else if !strings.Contains(err.Error(), tt.err.Error()) {
						t.Errorf("UserRepository.create() error = %v, want %v", err, tt.err)
					}
				}
				if got != nil && tt.want != nil {
					if got.Name == nil && tt.want.Name != nil {
						t.Errorf("UserRepository.create() got = %v, want %v", got.Name, tt.want.Name)
					}
					if got.Name != nil && tt.want.Name != nil && *got.Name != *tt.want.Name {
						t.Errorf("UserRepository.create() got = %s, want %s", *got.Name, *tt.want.Name)
					}
					if got.Email != tt.want.Email {
						t.Errorf("UserRepository.create() got = %s, want %s", got.Email, tt.want.Email)
					}
					if got.EmailVerifiedAt != tt.want.EmailVerifiedAt {
						t.Errorf("UserRepository.create() got = %v, want %v", got.EmailVerifiedAt, tt.want.EmailVerifiedAt)
					}
					if got.Image != tt.want.Image {
						t.Errorf("UserRepository.create() got = %v, want %v", got.Image, tt.want.Image)
					}
				}
			})
		}
	})
}

func TestUserRepsository_find(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		usersInput := []*models.User{
			{
				Name:            types.Pointer("Alpha User"),
				Email:           "alpha@example.com",
				EmailVerifiedAt: types.Pointer(time.Now()),
			},
			{
				Name:  types.Pointer("Beta User"),
				Email: "beta@example.com",
			},
			{
				Name:  types.Pointer("Charlie User"),
				Email: "charlie@example.com",
			},
			{
				Name:            types.Pointer("Delta User"),
				Email:           "delta@example.com",
				EmailVerifiedAt: types.Pointer(time.Now()),
			},
			{
				Name:  types.Pointer("Echo User"),
				Email: "echo@example.com",
			},
			{
				Name:  types.Pointer("Foxtrot User"),
				Email: "foxtrot@example.com",
			},
			{
				Name:            types.Pointer("Gamma User"),
				Email:           "gamma@example.com",
				EmailVerifiedAt: types.Pointer(time.Now()),
			},
			{
				Name:  types.Pointer("Hotel User"),
				Email: "hotel@example.com",
			},
			{
				Name:  types.Pointer("Yankee User"),
				Email: "yankee@example.com",
			},
			{
				Name:  types.Pointer("Zeta User"),
				Email: "zeta@example.com",
			},
		}
		userResource := NewUserRepositoryResource(db)
		for _, user := range usersInput {
			_, err := userResource.Create(ctx, user)
			if err != nil {
				t.Fatalf("Failed to create user: %v", err)
			}
		}
		type args struct {
			ctx    context.Context
			filter *UserFilter
		}
		tests := []struct {
			name      string
			args      args
			predicate func(t *testing.T, got []*models.User, err error)
		}{
			{
				name: "find all users sorted by name ascending",
				args: args{
					ctx: ctx,
					filter: &UserFilter{
						PaginatedInput: repository.PaginatedInput{
							Page:    0,
							PerPage: 10,
						},
						SortParams: repository.SortParams{
							SortBy:    "name",
							SortOrder: "asc",
						},
					},
				},
				predicate: func(t *testing.T, got []*models.User, err error) {
					if err != nil {
						t.Errorf("UserRepository.find() error = %v", err)
					}
					if len(got) != 10 {
						t.Errorf("UserRepository.find() got = %d, want %d", len(got), 10)
					}
					for i := 1; i < len(got)-1; i++ {
						firstName, secondName := *got[i].Name, *got[i+1].Name
						if firstName > secondName {
							t.Errorf("users are not in order. first name %s > second name %s", firstName, secondName)
						}
					}
				},
			},
			{
				name: "find all users sorted by name ascending, 3 per page, page 0",
				args: args{
					ctx: ctx,
					filter: &UserFilter{
						PaginatedInput: repository.PaginatedInput{
							Page:    0,
							PerPage: 3,
						},
						SortParams: repository.SortParams{
							SortBy:    "name",
							SortOrder: "asc",
						},
					},
				},
				predicate: func(t *testing.T, got []*models.User, err error) {
					if err != nil {
						t.Errorf("UserRepository.find() error = %v", err)
					}
					if len(got) != 3 {
						t.Errorf("UserRepository.find() got = %d, want %d", len(got), 3)
					}
					CheckUserOrderByName(t, got)
				},
			},

			{
				name: "find all users sorted by name ascending, 3 per page, page 1",
				args: args{
					ctx: ctx,
					filter: &UserFilter{
						PaginatedInput: repository.PaginatedInput{
							Page:    1,
							PerPage: 3,
						},
						SortParams: repository.SortParams{
							SortBy:    "name",
							SortOrder: "asc",
						},
					},
				},
				predicate: func(t *testing.T, got []*models.User, err error) {
					if err != nil {
						t.Errorf("UserRepository.find() error = %v", err)
					}
					if len(got) != 3 {
						t.Errorf("UserRepository.find() got = %d, want %d", len(got), 3)
					}
					CheckUserOrderByName(t, got)
				},
			},
			{
				name: "find all users sorted by name ascending, 3 per page, page 2",
				args: args{
					ctx: ctx,
					filter: &UserFilter{
						PaginatedInput: repository.PaginatedInput{
							Page:    2,
							PerPage: 3,
						},
						SortParams: repository.SortParams{
							SortBy:    "name",
							SortOrder: "asc",
						},
					},
				},
				predicate: func(t *testing.T, got []*models.User, err error) {
					if err != nil {
						t.Errorf("UserRepository.find() error = %v", err)
					}
					if len(got) != 3 {
						t.Errorf("UserRepository.find() got = %d, want %d", len(got), 3)
					}
					CheckUserOrderByName(t, got)
				},
			},
			{
				name: "find all users sorted by name ascending, 3 per page, page 3",
				args: args{
					ctx: ctx,
					filter: &UserFilter{
						PaginatedInput: repository.PaginatedInput{
							Page:    3,
							PerPage: 3,
						},
						SortParams: repository.SortParams{
							SortBy:    "name",
							SortOrder: "asc",
						},
					},
				},
				predicate: func(t *testing.T, got []*models.User, err error) {
					if err != nil {
						t.Errorf("UserRepository.find() error = %v", err)
					}
					if len(got) != 1 {
						t.Errorf("UserRepository.find() got = %d, want %d", len(got), 1)
					}
					if got[0].Name == nil || *got[0].Name != "Zeta User" {
						t.Errorf("UserRepository.find() got = %s, want %s", *got[0].Name, "Zeta User")
					}
				},
			},
			{
				name: "find all users with 'ta' in name. sorted by name ascending, 10 per page, page 0",
				args: args{
					ctx: ctx,
					filter: &UserFilter{
						PaginatedInput: repository.PaginatedInput{
							Page:    0,
							PerPage: 10,
						},
						SortParams: repository.SortParams{
							SortBy:    "name",
							SortOrder: "asc",
						},
						Q: "ta",
					},
				},
				predicate: func(t *testing.T, got []*models.User, err error) {
					if err != nil {
						t.Errorf("UserRepository.find() error = %v", err)
					}
					if len(got) != 3 {
						t.Errorf("UserRepository.find() got = %d, want %d", len(got), 3)
					}
					CheckUserOrderByName(t, got)
				},
			},
			{
				name: "find all users that are verified. sorted by name ascending, 10 per page, page 0",
				args: args{
					ctx: ctx,
					filter: &UserFilter{
						PaginatedInput: repository.PaginatedInput{
							Page:    0,
							PerPage: 10,
						},
						SortParams: repository.SortParams{
							SortBy:    "name",
							SortOrder: "asc",
						},
						EmailVerified: types.OptionalParam[bool]{IsSet: true, Value: true},
					},
				},
				predicate: func(t *testing.T, got []*models.User, err error) {
					if err != nil {
						t.Errorf("UserRepository.find() error = %v", err)
					}
					if len(got) != 3 {
						t.Errorf("UserRepository.find() got = %d, want %d", len(got), 3)
					}
					CheckUserOrderByName(t, got)
				},
			},
			{
				name: "find all users that are verified. sorted by name ascending, 10 per page, page 0",
				args: args{
					ctx: ctx,
					filter: &UserFilter{
						PaginatedInput: repository.PaginatedInput{
							Page:    0,
							PerPage: 10,
						},
						SortParams: repository.SortParams{
							SortBy:    "name",
							SortOrder: "asc",
						},
						EmailVerified: types.OptionalParam[bool]{IsSet: true, Value: true},
					},
				},
				predicate: func(t *testing.T, got []*models.User, err error) {
					if err != nil {
						t.Errorf("UserRepository.find() error = %v", err)
					}
					if len(got) != 3 {
						t.Errorf("UserRepository.find() got = %d, want %d", len(got), 3)
					}
					CheckUserOrderByName(t, got)
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := userResource.Find(tt.args.ctx, tt.args.filter)
				tt.predicate(t, got, err)
			})
		}
	})
}

func CheckSliceLength[M any](t *testing.T, got []*M, expected int) {
	if len(got) != expected {
		t.Errorf("UserRepository.find() got = %d, want %d", len(got), expected)
	}
}

func CheckUserOrderByName(t *testing.T, got []*models.User) {
	for i := 1; i < len(got)-1; i++ {
		firstName, secondName := *got[i].Name, *got[i+1].Name
		if firstName > secondName {
			t.Errorf("users are not in order. first name %s > second name %s", firstName, secondName)
		}
	}
}
func CheckUserAccountOrderByName(t *testing.T, got []*models.UserAccount) {
	for i := 1; i < len(got)-1; i++ {
		firstName, secondName := got[i].Provider.String(), got[i+1].Provider.String()
		if firstName > secondName {
			t.Errorf("users are not in order. first name %s > second name %s", firstName, secondName)
		}
	}
}
