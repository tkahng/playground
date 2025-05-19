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

func TestPostgresAccountStore_GetUserAccounts(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		userStore := NewPostgresUserStore(dbxx)
		store := NewPostgresUserAccountStore(dbxx)
		user1, err := userStore.CreateUser(ctx, &models.User{Email: "user1@example.com"})
		assert.NoError(t, err)
		user2, err := userStore.CreateUser(ctx, &models.User{Email: "user2@example.com"})
		assert.NoError(t, err)
		acc1 := &models.UserAccount{UserID: user1.ID, Provider: models.ProvidersGoogle, Type: "oauth", ProviderAccountID: "g1"}
		acc2 := &models.UserAccount{UserID: user2.ID, Provider: models.ProvidersGoogle, Type: "oauth", ProviderAccountID: "g2"}
		err = store.LinkAccount(ctx, acc1)
		assert.NoError(t, err)
		err = store.LinkAccount(ctx, acc2)
		assert.NoError(t, err)
		results, err := store.GetUserAccounts(ctx, user1.ID, user2.ID)
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, user1.ID, results[0][0].UserID)
		assert.Equal(t, user2.ID, results[1][0].UserID)
		return errors.New("rollback")
	})
}

func TestPostgresAccountStore_UpdateUserPassword(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		userStore := NewPostgresUserStore(dbxx)
		store := NewPostgresUserAccountStore(dbxx)
		user, err := userStore.CreateUser(ctx, &models.User{Email: "pwuser@example.com"})
		assert.NoError(t, err)
		acc := &models.UserAccount{UserID: user.ID, Provider: models.ProvidersCredentials, Type: "credentials", ProviderAccountID: "pwuser"}
		err = store.LinkAccount(ctx, acc)
		assert.NoError(t, err)
		newPassword := "newpassword123"
		err = store.UpdateUserPassword(ctx, user.ID, newPassword)
		assert.NoError(t, err)
		updated, err := store.FindUserAccountByUserIdAndProvider(ctx, user.ID, models.ProvidersCredentials)
		assert.NoError(t, err)
		assert.NotNil(t, updated.Password)
		return errors.New("rollback")
	})
}
