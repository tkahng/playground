package stores_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
)

func TestTokenStore_CRUD(t *testing.T) {
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		userStore := stores.NewDbUserStore(dbxx)
		store := stores.NewPostgresTokenStore(dbxx)
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "user@example.com",
		})
		assert.NoError(t, err)
		tokenStr := "tok_test_123"
		tok := &stores.CreateTokenDTO{
			Type:       models.TokenTypes(models.TokenTypesAccessToken),
			Identifier: "user@example.com",
			Expires:    time.Now().Add(1 * time.Hour),
			Token:      tokenStr,
			UserID:     &user.ID,
			Otp:        nil,
		}

		t.Run("SaveToken", func(t *testing.T) {
			err := store.SaveToken(ctx, tok)
			assert.NoError(t, err)
		})

		t.Run("GetToken", func(t *testing.T) {
			got, err := store.GetToken(ctx, tokenStr)
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tokenStr, got.Token)
		})

		t.Run("DeleteToken", func(t *testing.T) {
			err := store.DeleteToken(ctx, tokenStr)
			assert.NoError(t, err)
			got, err := store.GetToken(ctx, tokenStr)
			assert.ErrorIs(t, err, shared.ErrTokenNotFound)
			assert.Nil(t, got)
		})

		t.Run("GetToken_expired", func(t *testing.T) {
			expiredTok := &stores.CreateTokenDTO{
				Type:       models.TokenTypesAccessToken,
				Identifier: "user2@example.com",
				Expires:    time.Now().Add(-1 * time.Hour),
				Token:      "tok_expired",
				UserID:     &user.ID,
				Otp:        nil,
			}
			err := store.SaveToken(ctx, expiredTok)
			assert.NoError(t, err)
			got, err := store.GetToken(ctx, "tok_expired")
			assert.ErrorIs(t, err, shared.ErrTokenExpired)
			assert.Nil(t, got)
		})

	})
}

func TestDbTokenStore_GetTokenByValueTypeExpires(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		store := stores.NewPostgresTokenStore(db)
		// opts := conf.ZeroEnvConfig()

		t.Run("GetTokenByValueTypeExpires-success", func(t *testing.T) {

			tok := &stores.CreateTokenDTO{
				Type:       models.TokenTypesAccessToken,
				Identifier: "user@example.com",
				Expires:    time.Now().Add(1 * time.Hour),
				Token:      "tok_test_123",
				UserID:     nil,
				Otp:        nil,
			}
			err := store.SaveToken(ctx, tok)
			assert.NoError(t, err)
			got, err := store.GetTokenByValueTypeExpires(ctx, tok.Token, tok.Type, time.Now())
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tok.Token, got.Token)
		})
		t.Run("GetTokenByValueTypeExpires-fail expired", func(t *testing.T) {

			tok := &stores.CreateTokenDTO{
				Type:       models.TokenTypesAccessToken,
				Identifier: "user@example.com",
				Expires:    time.Now().Add(-1 * time.Hour),
				Token:      "tok_test_1234",
				UserID:     nil,
				Otp:        nil,
			}
			err := store.SaveToken(ctx, tok)
			assert.NoError(t, err)
			got, err := store.GetTokenByValueTypeExpires(ctx, tok.Token, tok.Type, tok.Expires)
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tok.Token, got.Token)
		})
	})
}
