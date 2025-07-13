package stores

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
)

func TestAccountStore_CRUD(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		userStore := NewDbUserStore(dbxx)
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "test@example.com",
		})
		assert.NoError(t, err)
		userID := user.ID
		store := NewDbAccountStore(dbxx)
		account := &models.UserAccount{
			UserID:            userID,
			Provider:          models.ProvidersGoogle,
			Type:              "oauth",
			ProviderAccountID: "google-123",
		}

		assert.NoError(t, err)

		t.Run("LinkAccount", func(t *testing.T) {
			linkedAccount, err := store.CreateUserAccount(ctx, account)
			assert.NoError(t, err)
			assert.NotNil(t, linkedAccount)
		})

		t.Run("FindUserAccount", func(t *testing.T) {
			got, err := store.FindUserAccount(ctx, &UserAccountFilter{
				UserIds:   []uuid.UUID{userID},
				Providers: []models.Providers{models.ProvidersGoogle},
			})
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
			linkedAccount, err := store.CreateUserAccount(ctx, nil)
			assert.Error(t, err)
			assert.Nil(t, linkedAccount)
		})

		return errors.New("rollback")
	})
}

func TestAccountStore_GetUserAccounts(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		userStore := NewDbUserStore(dbxx)
		store := NewDbAccountStore(dbxx)
		user1, err := userStore.CreateUser(ctx, &models.User{Email: "user1@example.com"})
		assert.NoError(t, err)
		user2, err := userStore.CreateUser(ctx, &models.User{Email: "user2@example.com"})
		assert.NoError(t, err)
		acc1 := &models.UserAccount{UserID: user1.ID, Provider: models.ProvidersGoogle, Type: "oauth", ProviderAccountID: "g1"}
		acc2 := &models.UserAccount{UserID: user2.ID, Provider: models.ProvidersGoogle, Type: "oauth", ProviderAccountID: "g2"}
		linkedAccount, err := store.CreateUserAccount(ctx, acc1)
		assert.NoError(t, err)
		assert.NotNil(t, linkedAccount)
		_, err = store.CreateUserAccount(ctx, acc2)
		assert.NoError(t, err)
		results, err := store.GetUserAccounts(ctx, user1.ID, user2.ID)
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, user1.ID, results[0][0].UserID)
		assert.Equal(t, user2.ID, results[1][0].UserID)
		return errors.New("rollback")
	})
}

func TestAccountStore_UpdateUserPassword(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		userStore := NewDbUserStore(dbxx)
		store := NewDbAccountStore(dbxx)
		user, err := userStore.CreateUser(ctx, &models.User{Email: "pwuser@example.com"})
		assert.NoError(t, err)
		acc := &models.UserAccount{UserID: user.ID, Provider: models.ProvidersCredentials, Type: "credentials", ProviderAccountID: "pwuser"}
		linkedAccount, err := store.CreateUserAccount(ctx, acc)
		assert.NoError(t, err)
		assert.NotNil(t, linkedAccount)
		newPassword := "newpassword123"
		err = store.UpdateUserPassword(ctx, user.ID, newPassword)
		assert.NoError(t, err)
		updated, err := store.FindUserAccountByUserIdAndProvider(ctx, user.ID, models.ProvidersCredentials)
		assert.NoError(t, err)
		assert.NotNil(t, updated.Password)
		return errors.New("rollback")
	})
}
