package stores_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/test"
)

func TestTokenStore_CRUD(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		userStore := stores.NewDbUserStore(dbxx)
		store := stores.NewPostgresTokenStore(dbxx)
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "user@example.com",
		})
		if err != nil {
			return err
		}
		tokenStr := "tok_test_123"
		tok := &shared.CreateTokenDTO{
			Type:       shared.TokenType(models.TokenTypesAccessToken),
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
			expiredTok := &shared.CreateTokenDTO{
				Type:       shared.TokenTypesAccessToken,
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
			tok2 := &shared.CreateTokenDTO{
				Type:       shared.TokenTypesVerificationToken,
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
