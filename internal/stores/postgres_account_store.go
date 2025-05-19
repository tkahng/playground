package stores

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
)

type PostgresAccountStore struct {
	db database.Dbx
}

func NewPostgresUserAccountStore(db database.Dbx) *PostgresAccountStore {
	return &PostgresAccountStore{
		db: db,
	}
}

// FindUserAccountByUserIdAndProvider implements UserAccountStore.
func (u *PostgresAccountStore) FindUserAccountByUserIdAndProvider(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error) {
	return crudrepo.UserAccount.GetOne(ctx, u.db, &map[string]any{
		"user_id": map[string]any{
			"_eq": userId.String(),
		},
		"provider": map[string]any{
			"_eq": provider.String(),
		},
	})
}

// LinkAccount implements UserAccountStore.
func (u *PostgresAccountStore) LinkAccount(ctx context.Context, account *models.UserAccount) error {
	if account == nil {
		return errors.New("account is nil")
	}
	_, err := crudrepo.UserAccount.PostOne(ctx,
		u.db,
		account,
	)
	if err != nil {
		return err
	}
	return nil
}

// UnlinkAccount implements UserAccountStore.
func (u *PostgresAccountStore) UnlinkAccount(ctx context.Context, userId uuid.UUID, provider models.Providers) error {
	_, err := crudrepo.UserAccount.Delete(
		ctx,
		u.db,
		&map[string]any{
			"user_id": map[string]any{
				"_eq": userId.String(),
			},
			"provider": map[string]any{
				"_eq": provider.String(),
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

// UpdateUserAccount implements UserAccountStore.
func (u *PostgresAccountStore) UpdateUserAccount(ctx context.Context, account *models.UserAccount) error {
	_, err := crudrepo.UserAccount.PutOne(ctx, u.db, account)
	if err != nil {
		return fmt.Errorf("error updating user account: %w", err)
	}
	return nil
}
