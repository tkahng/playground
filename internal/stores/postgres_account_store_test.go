package stores

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/test"
)

func TestPostgresAccountStore_CRUD(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		userStore := NewPostgresUserStore(dbxx)
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "test@example.com",
		})
		assert.NoError(t, err)
		userID := user.ID
		store := NewPostgresUserAccountStore(dbxx)
		account := &models.UserAccount{
			UserID:            userID,
			Provider:          models.ProvidersGoogle,
			Type:              "oauth",
			ProviderAccountID: "google-123",
		}
		// account, err := store.CreateUserAccount(ctx, &models.UserAccount{
		// 	UserID:            userID,
		// 	Provider:          models.ProvidersGoogle,
		// 	Type:              "oauth",
		// 	ProviderAccountID: "google-123",
		// })
		assert.NoError(t, err)

		t.Run("LinkAccount", func(t *testing.T) {
			err := store.LinkAccount(ctx, account)
			assert.NoError(t, err)
		})

		t.Run("FindUserAccountByUserIdAndProvider", func(t *testing.T) {
			got, err := store.FindUserAccountByUserIdAndProvider(ctx, userID, models.ProvidersGoogle)
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, account.ProviderAccountID, got.ProviderAccountID)
		})

		t.Run("UpdateUserAccount", func(t *testing.T) {
			account, err = store.FindUserAccountByUserIdAndProvider(ctx, userID, models.ProvidersGoogle)
			assert.NoError(t, err)
			account.ProviderAccountID = "google-456"
			err := store.UpdateUserAccount(ctx, account)
			assert.NoError(t, err)
			got, err := store.FindUserAccountByUserIdAndProvider(ctx, userID, models.ProvidersGoogle)
			assert.NoError(t, err)
			assert.Equal(t, "google-456", got.ProviderAccountID)
		})

		t.Run("UnlinkAccount", func(t *testing.T) {
			err := store.UnlinkAccount(ctx, userID, models.ProvidersGoogle)
			assert.NoError(t, err)
			got, err := store.FindUserAccountByUserIdAndProvider(ctx, userID, models.ProvidersGoogle)
			assert.NoError(t, err)
			assert.Nil(t, got)
		})

		t.Run("LinkAccount_nil", func(t *testing.T) {
			err := store.LinkAccount(ctx, nil)
			assert.Error(t, err)
		})

		return errors.New("rollback")
	})
}
