package resource

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
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
		filter := &UserListFilter{
			EmailVerified: types.OptionalParam[bool]{IsSet: true, Value: true},
		}
		where := filterFunc(filter)
		assert.NotNil(t, where)
		assert.Equal(t, map[string]any{
			"email_verified_at": map[string]any{"_neq": nil},
		}, *where)
	})

	t.Run("EmailVerified false", func(t *testing.T) {
		filter := &UserListFilter{
			EmailVerified: types.OptionalParam[bool]{IsSet: true, Value: false},
		}
		where := filterFunc(filter)
		assert.NotNil(t, where)
		assert.Equal(t, map[string]any{
			"email_verified_at": map[string]any{"_eq": nil},
		}, *where)
	})

	t.Run("EmailVerified false not set", func(t *testing.T) {
		id := uuid.New()
		filter := &UserListFilter{
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
		filter := &UserListFilter{
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
		filter := &UserListFilter{
			Ids: []uuid.UUID{id1, id2},
		}
		where := filterFunc(filter)
		assert.NotNil(t, where)
		assert.Equal(t, map[string]any{
			"id": map[string]any{"_in": []uuid.UUID{id1, id2}},
		}, *where)
	})

	t.Run("Providers filter", func(t *testing.T) {
		filter := &UserListFilter{
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
		filter := &UserListFilter{
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
		filter := &UserListFilter{
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
		filter := &UserListFilter{
			Emails:        []string{"a@example.com"},
			RoleIds:       []uuid.UUID{role},
			EmailVerified: types.OptionalParam[bool]{IsSet: true, Value: true},
		}
		where := filterFunc(filter)
		assert.NotNil(t, where)
		assert.Equal(t, map[string]any{
			"email":             map[string]any{"_in": []string{"a@example.com"}},
			"roles":             map[string]any{"id": map[string]any{"_in": []uuid.UUID{role}}},
			"email_verified_at": map[string]any{"_neq": nil},
		}, *where)
	})

	t.Run("Empty filter returns nil", func(t *testing.T) {
		filter := &UserListFilter{}
		where := filterFunc(filter)
		assert.Nil(t, where)
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
		filter := &UserListFilter{}
		s, b := filter.Sort()
		fmt.Println("haa", s, b)
		order := sortFunc(filter)
		assert.NotNil(t, order)
		assert.Equal(t, map[string]string{"created_at": "desc"}, *order)
	})

	t.Run("invalid sort by returns nil map", func(t *testing.T) {
		filter := &UserListFilter{SortParams: SortParams{
			SortBy:    "notacol",
			SortOrder: "asc",
		}}
		order := sortFunc(filter)
		assert.Nil(t, order)
	})

	t.Run("valid sort by returns map", func(t *testing.T) {
		filter := &UserListFilter{SortParams: SortParams{
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
		input := &UserListFilter{PaginatedInput: PaginatedInput{Page: -2, PerPage: 5}}
		limit, offset := paginationFunc(input)
		assert.Equal(t, 5, limit)
		assert.Equal(t, 0, offset)
	})

	t.Run("perPage < 1 returns default", func(t *testing.T) {
		input := &UserListFilter{PaginatedInput: PaginatedInput{Page: 2, PerPage: 0}}
		limit, offset := paginationFunc(input)
		assert.Equal(t, 10, limit)
		assert.Equal(t, 20, offset)
	})

	t.Run("normal values", func(t *testing.T) {
		input := &UserListFilter{PaginatedInput: PaginatedInput{Page: 3, PerPage: 15}}
		limit, offset := paginationFunc(input)
		assert.Equal(t, 15, limit)
		assert.Equal(t, 45, offset)
	})
}
