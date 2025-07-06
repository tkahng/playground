package resource

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/test"
	"github.com/tkahng/authgo/internal/tools/types"
)

// mockDbx is a minimal mock for database.Dbx

func TestNewTokenRepositoryResource_FilterFunc(t *testing.T) {
	db := &database.Queries{} // Mock or use a real database connection as needed
	repo := NewTokenRepositoryResource(db)

	// reflect to get the filterFunc

	filterFunc := repo.filter

	t.Run("nil filter returns nil", func(t *testing.T) {
		result := filterFunc(nil)
		assert.NotNil(t, result)
		assert.Empty(t, *result)
	})

	t.Run("empty filter returns empty map", func(t *testing.T) {
		filter := &TokenFilter{}
		result := filterFunc(filter)
		assert.NotNil(t, result)
		assert.Empty(t, *result)
	})

	t.Run("filter with UserIds", func(t *testing.T) {
		uid := uuid.New()
		filter := &TokenFilter{UserIds: []uuid.UUID{uid}}
		result := filterFunc(filter)
		assert.Contains(t, *result, "user_id")
		assert.Equal(t, map[string]any{"_in": []uuid.UUID{uid}}, (*result)["user_id"])
	})

	t.Run("filter with Ids", func(t *testing.T) {
		id := uuid.New()
		filter := &TokenFilter{Ids: []uuid.UUID{id}}
		result := filterFunc(filter)
		assert.Contains(t, *result, "id")
		assert.Equal(t, map[string]any{"_in": []uuid.UUID{id}}, (*result)["id"])
	})

	t.Run("filter with Types", func(t *testing.T) {
		filter := &TokenFilter{Types: []models.TokenTypes{models.TokenTypesAccessToken}}
		result := filterFunc(filter)
		assert.Contains(t, *result, "type")
		assert.Equal(t, map[string]any{"_in": []models.TokenTypes{models.TokenTypesAccessToken}}, (*result)["type"])
	})

	t.Run("filter with Identifiers", func(t *testing.T) {
		filter := &TokenFilter{Identifiers: []string{"foo", "bar"}}
		result := filterFunc(filter)
		assert.Contains(t, *result, "identifier")
		assert.Equal(t, map[string]any{"_in": []string{"foo", "bar"}}, (*result)["identifier"])
	})

	t.Run("filter with Tokens", func(t *testing.T) {
		filter := &TokenFilter{Tokens: []string{"tok1", "tok2"}}
		result := filterFunc(filter)
		assert.Contains(t, *result, "token")
		assert.Equal(t, map[string]any{"_in": []string{"tok1", "tok2"}}, (*result)["token"])
	})

	t.Run("filter with all fields", func(t *testing.T) {
		uid := uuid.New()
		id := uuid.New()
		filter := &TokenFilter{
			UserIds:     []uuid.UUID{uid},
			Ids:         []uuid.UUID{id},
			Types:       []models.TokenTypes{models.TokenTypesRefreshToken},
			Identifiers: []string{"id1"},
			Tokens:      []string{"tok"},
		}
		result := filterFunc(filter)
		assert.Len(t, *result, 5)
		assert.Equal(t, map[string]any{"_in": []uuid.UUID{uid}}, (*result)["user_id"])
		assert.Equal(t, map[string]any{"_in": []uuid.UUID{id}}, (*result)["id"])
		assert.Equal(t, map[string]any{"_in": []models.TokenTypes{models.TokenTypesRefreshToken}}, (*result)["type"])
		assert.Equal(t, map[string]any{"_in": []string{"id1"}}, (*result)["identifier"])
		assert.Equal(t, map[string]any{"_in": []string{"tok"}}, (*result)["token"])
	})
}

func TestTokenRepositoryResource_Create(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		resource := NewTokenRepositoryResource(db)

		t.Run("Create with valid data", func(t *testing.T) {
			token := &models.Token{
				Type:       models.TokenTypesAccessToken,
				Identifier: "test-identifier",
				Token:      "test-token",
			}
			created, err := resource.Create(ctx, token)
			assert.NoError(t, err)
			assert.NotNil(t, created)
			assert.Equal(t, token.UserID, created.UserID)
			assert.Equal(t, token.Type, created.Type)
			assert.Equal(t, token.Identifier, created.Identifier)
			assert.Equal(t, token.Token, created.Token)
		})

		t.Run("Create with duplicate identifier", func(t *testing.T) {
			token := &models.Token{
				Type:       models.TokenTypesRefreshToken,
				Identifier: "duplicate-identifier",
				Token:      "duplicate-token",
			}
			_, err := resource.Create(ctx, token)
			assert.NoError(t, err)

			// Attempt to create with the same identifier
			_, err = resource.Create(ctx, token)
			assert.Error(t, err)
		})
	})
}

func TestTokenRepositoryResource_Filter(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		resource := NewTokenRepositoryResource(db)
		userResource := NewUserRepositoryResource(db)
		user, _ := userResource.Create(ctx, &models.User{
			Email: "test@example.com",
		})
		inputTokens := []models.Token{

			{
				UserID:     &user.ID,
				Type:       models.TokenTypesAccessToken,
				Identifier: user.Email,
				Token:      "user-access-token-expired",
				Expires:    time.Now().Add(-1 * time.Minute),
			},
			{
				UserID:     &user.ID,
				Type:       models.TokenTypesRefreshToken,
				Identifier: user.Email,
				Token:      "user-refresh-token-expired",
				Expires:    time.Now().Add(-1 * time.Minute),
			},
			{
				UserID:     &user.ID,
				Type:       models.TokenTypesPasswordResetToken,
				Identifier: user.Email,
				Token:      "user-password-reset-token",
				Expires:    time.Now().Add(10 * time.Minute),
			},
			{
				Type:       models.TokenTypesStateToken,
				Identifier: user.Email,
				Token:      "state-token",
				Expires:    time.Now().Add(10 * time.Minute),
			},
			{
				Type:       models.TokenTypesInviteToken,
				Identifier: user.Email,
				Token:      "user-invite-token",
				Expires:    time.Now().Add(10 * time.Minute),
			},
		}
		// 3 user id tokens, 2 expired, 5 email
		for _, token := range inputTokens {
			_, err := resource.Create(ctx, &token)
			assert.NoError(t, err)
		}
		t.Run("Filter with valid UserIds", func(t *testing.T) {
			filter := &TokenFilter{UserIds: []uuid.UUID{user.ID}}
			tokens, err := resource.Find(ctx, filter)
			assert.NoError(t, err, "Failed to filter tokens by UserIds")
			assert.NotEmpty(t, tokens)
			assert.Len(t, tokens, 3, "Expected 3 tokens to be returned")
			for _, token := range tokens {
				assert.Equal(t, user.ID, *token.UserID)
			}
		})
		t.Run("Filter with valid Types", func(t *testing.T) {
			filter := &TokenFilter{Types: []models.TokenTypes{models.TokenTypesAccessToken, models.TokenTypesRefreshToken}}
			tokens, err := resource.Find(ctx, filter)
			assert.NoError(t, err, "Failed to filter tokens by Types")
			assert.NotEmpty(t, tokens)
			assert.Len(t, tokens, 2, "Expected 2 tokens to be returned")
			for _, token := range tokens {
				assert.Contains(t, []models.TokenTypes{models.TokenTypesAccessToken, models.TokenTypesRefreshToken}, token.Type)
			}
		})
		t.Run("Filter with valid Identifiers", func(t *testing.T) {
			filter := &TokenFilter{Identifiers: []string{user.Email}}
			tokens, err := resource.Find(ctx, filter)
			assert.NoError(t, err, "Failed to filter tokens by Identifiers")
			assert.NotEmpty(t, tokens)
			assert.Len(t, tokens, 5, "Expected 5 tokens to be returned")
			for _, token := range tokens {
				assert.Equal(t, user.Email, token.Identifier)
			}
		})
		t.Run("Filter with valid Tokens", func(t *testing.T) {
			filter := &TokenFilter{Tokens: []string{"user-access-token-expired", "user-refresh-token-expired"}}
			tokens, err := resource.Find(ctx, filter)
			assert.NoError(t, err, "Failed to filter tokens by Tokens")
			assert.NotEmpty(t, tokens)
			assert.Len(t, tokens, 2, "Expected 2 tokens to be returned")
			for _, token := range tokens {
				assert.Contains(t, []string{"user-access-token-expired", "user-refresh-token-expired"}, token.Token)
			}
		})
		t.Run("Filter expires after now", func(t *testing.T) {
			filter := &TokenFilter{
				ExpiresAfter: types.OptionalParam[time.Time]{Value: time.Now(), IsSet: true},
			}
			tokens, err := resource.Find(ctx, filter)
			assert.NoError(t, err, "Failed to filter tokens by ExpiresAfter")
			assert.NotEmpty(t, tokens)
			assert.Len(t, tokens, 3, "Expected 3 tokens to be returned")
			for _, token := range tokens {
				assert.GreaterOrEqual(t, token.Expires, time.Now(), "Token should expire after now")
			}
		})
		t.Run("Filter expires before now", func(t *testing.T) {
			filter := &TokenFilter{
				ExpiresBefore: types.OptionalParam[time.Time]{Value: time.Now(), IsSet: true},
			}
			tokens, err := resource.Find(ctx, filter)
			assert.NoError(t, err, "Failed to filter tokens by ExpiresBefore")
			assert.NotEmpty(t, tokens)
			assert.Len(t, tokens, 2, "Expected 2 tokens to be returned")
			for _, token := range tokens {
				assert.Less(t, token.Expires, time.Now(), "Token should expire before now")
			}
		})
	})
}
