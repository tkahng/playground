package stores_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

func TestTxManagerImpl_RunInTxCtx(t *testing.T) {
	t.Run("test transaction manager succeed in creating user and account", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			store := stores.NewTxManager(db)
			userStore := stores.NewDbUserStore(db)
			accountStore := stores.NewDbAccountStore(db)

			err := store.RunInTxCtx(ctx, func(newCtx context.Context) error {
				user := &models.User{
					Email: "test@example.com",
				}
				user, err := userStore.CreateUser(newCtx, user)
				assert.NoError(t, err)
				assert.NotNil(t, user)
				acc := &models.UserAccount{
					UserID:            user.ID,
					Provider:          models.ProvidersCredentials,
					Type:              models.ProviderTypeCredentials,
					ProviderAccountID: "test@example.com",
				}
				acc, err = accountStore.CreateUserAccount(newCtx, acc)
				assert.NoError(t, err)
				assert.NotNil(t, acc)
				count, err := userStore.CountUsers(newCtx, nil)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), count)
				accountCount, err := accountStore.CountUserAccounts(newCtx, nil)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), accountCount)
				return nil
			})
			assert.NoError(t, err)
			count, err := userStore.CountUsers(ctx, nil)
			assert.NoError(t, err)
			assert.Equal(t, int64(1), count)
			accountCount, err := accountStore.CountUserAccounts(ctx, nil)
			assert.NoError(t, err)
			assert.Equal(t, int64(1), accountCount)
			assert.NoError(t, err)
		})
	})
	t.Run("test transaction manager succeed in creating user and account but error at end", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			txManager := stores.NewTxManager(db)
			userStore := stores.NewDbUserStore(db)
			accountStore := stores.NewDbAccountStore(db)

			txErr := txManager.RunInTxCtx(ctx, func(newCtx context.Context) error {
				user := &models.User{
					Email: "test@example.com",
				}
				user, err := userStore.CreateUser(newCtx, user)
				assert.NoError(t, err)
				assert.NotNil(t, user)
				acc := &models.UserAccount{
					UserID:            user.ID,
					Provider:          models.ProvidersCredentials,
					Type:              models.ProviderTypeCredentials,
					ProviderAccountID: "test@example.com",
				}
				acc, err = accountStore.CreateUserAccount(newCtx, acc)
				assert.NoError(t, err)
				assert.NotNil(t, acc)
				count, err := userStore.CountUsers(newCtx, nil)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), count)
				accountCount, err := accountStore.CountUserAccounts(newCtx, nil)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), accountCount)
				return errors.New("error at end of transaction")
			})
			assert.Error(t, txErr)
			count, userCountErr := userStore.CountUsers(ctx, nil)
			assert.NoError(t, userCountErr)
			assert.Equal(t, int64(0), count)
			accountCount, userAccountCountErr := accountStore.CountUserAccounts(ctx, nil)
			assert.NoError(t, userAccountCountErr)
			assert.Equal(t, int64(0), accountCount)
		})
	})
	t.Run("test transaction manager succeed in creating user but error before account created", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			store := stores.NewTxManager(db)
			userStore := stores.NewDbUserStore(db)
			accountStore := stores.NewDbAccountStore(db)

			err := store.RunInTxCtx(ctx, func(newCtx context.Context) error {
				user := &models.User{
					Email: "test@example.com",
				}
				user, err := userStore.CreateUser(newCtx, user)
				assert.NoError(t, err)
				assert.NotNil(t, user)
				acc := &models.UserAccount{
					UserID:            user.ID,
					Provider:          models.ProvidersCredentials,
					Type:              models.ProviderTypeCredentials,
					ProviderAccountID: "test@example.com",
				}
				acc, err = accountStore.CreateUserAccount(newCtx, acc)
				assert.NoError(t, err)
				assert.NotNil(t, acc)
				count, err := userStore.CountUsers(newCtx, nil)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), count)
				accountCount, err := accountStore.CountUserAccounts(newCtx, nil)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), accountCount)
				return errors.New("error at end of transaction")
			})
			assert.Error(t, err)
			count, err := userStore.CountUsers(ctx, nil)
			assert.NoError(t, err)
			assert.Equal(t, int64(0), count)
			accountCount, err := accountStore.CountUserAccounts(ctx, nil)
			assert.NoError(t, err)
			assert.Equal(t, int64(0), accountCount)
			assert.NoError(t, err)
		})
	})
}
