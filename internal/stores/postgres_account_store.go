package stores

import (
	"context"
	"errors"
	"fmt"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/security"
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

func (u *PostgresAccountStore) CreateUserAccount(ctx context.Context, account *models.UserAccount) (*models.UserAccount, error) {

	createdAccount, err := crudrepo.UserAccount.PostOne(ctx, u.db, account)
	if err != nil {
		return nil, fmt.Errorf("error creating user account: %w", err)
	}
	return createdAccount, nil
}

func (u *PostgresAccountStore) GetUserAccounts(ctx context.Context, userIds ...uuid.UUID) ([][]*models.UserAccount, error) {
	// var results []JoinedResult[*crudModels.Permission, uuid.UUID]
	ids := []string{}
	for _, id := range userIds {
		ids = append(ids, id.String())
	}
	data, err := crudrepo.UserAccount.Get(
		ctx,
		u.db,
		&map[string]any{
			"user_id": map[string]any{
				"_in": ids,
			},
		},
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return mapper.MapToManyPointer(data, userIds, func(a *models.UserAccount) uuid.UUID {
		return a.UserID
	}), nil
}

func (u *PostgresAccountStore) UpdateUserPassword(ctx context.Context, userId uuid.UUID, password string) error {
	account, err := crudrepo.UserAccount.GetOne(
		ctx,
		u.db,
		&map[string]any{
			"user_id": map[string]any{
				"_eq": userId.String(),
			},
			"provider": map[string]any{
				"_eq": string(models.ProvidersCredentials),
			},
		},
	)
	if err != nil {
		return err
	}
	if account == nil {
		return errors.New("user ProvidersCredentials account not found")
	}
	hash, err := security.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return err
	}
	account.Password = &hash
	_, err = crudrepo.UserAccount.PutOne(
		ctx,
		u.db,
		account,
	)
	if err != nil {
		return err
	}
	return nil
}
