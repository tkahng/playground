package stores_test

import (
	"errors"
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
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		userStore := stores.NewDbUserStore(dbxx)
		store := stores.NewPostgresTokenStore(dbxx)
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "user@example.com",
		})
		if err != nil {
			return err
		}
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

		t.Run("VerifyTokenStorage", func(t *testing.T) {
			tok2 := &stores.CreateTokenDTO{
				Type:       models.TokenTypesVerificationToken,
				Identifier: "user3@example.com",
				Expires:    time.Now().Add(1 * time.Hour),
				Token:      "tok_verify",
				UserID:     &user.ID,
				Otp:        nil,
			}
			err := store.SaveToken(ctx, tok2)
			assert.NoError(t, err)
			err = store.VerifyTokenStorage(ctx, "tok_verify")
			assert.NoError(t, err)
			got, err := store.GetToken(ctx, "tok_verify")
			assert.ErrorIs(t, err, shared.ErrTokenNotFound)
			assert.Nil(t, got)
		})

		return errors.New("rollback")
	})
}
