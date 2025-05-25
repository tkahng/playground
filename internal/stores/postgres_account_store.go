package stores

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/security"
	"github.com/tkahng/authgo/internal/tools/types"
)

type PostgresAccountStore struct {
	db database.Dbx
}

// CreateUserAccount implements services.UserAccountStore.

func NewPostgresUserAccountStore(db database.Dbx) *PostgresAccountStore {
	return &PostgresAccountStore{
		db: db,
	}
}

var (
	// UserColumnNames = models.Users.Columns().Names()
	UserAccountColumnNames = []string{"id", "user_id", "type", "provider", "provider_account_id", "created_at", "updated_at"}
	// UserAccountColumnNames = models.UserAccounts.Columns().Names()
)

// ListUserAccounts implements AdminCrudActions.
// ListUsers implements AdminCrudActions.

func (u *PostgresAccountStore) ListUserAccounts(ctx context.Context, input *shared.UserAccountListParams) ([]*models.UserAccount, error) {
	where := UserAccountWhere(&input.UserAccountListFilter)
	sort := UserAccountOrderBy(&input.SortParams)
	data, err := crudrepo.UserAccount.Get(
		ctx,
		u.db,
		where,
		sort,
		types.Pointer(int(input.PaginatedInput.Page)),
		types.Pointer(int(input.PaginatedInput.PerPage)),
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func UserAccountOrderBy(params *shared.SortParams) *map[string]string {
	if params == nil {
		return nil
	}
	if slices.Contains(UserAccountColumnNames, params.SortBy) {
		return &map[string]string{
			params.SortBy: params.SortOrder,
		}
	}
	return nil
}

// func CreateUser(ctx context.Context, db db.Dbx, params *shared.AuthenticateUserParams) (*models.User, error) {
func UserAccountWhere(filter *shared.UserAccountListFilter) *map[string]any {
	where := make(map[string]any)
	if filter == nil {
		return &where
	}
	if len(filter.Providers) > 0 {
		var providers []string
		for _, p := range filter.Providers {
			providers = append(providers, p.String())
		}
		where["provider"] = map[string]any{
			"_in": providers,
		}
	}
	if len(filter.ProviderTypes) > 0 {
		var providerTypes []string
		for _, pt := range filter.ProviderTypes {
			providerTypes = append(providerTypes, pt.String())
		}
		where["type"] = map[string]any{
			"_in": providerTypes,
		}
	}
	if len(filter.Ids) > 0 {
		where["id"] = map[string]any{
			"_in": filter.Ids,
		}
	}
	if len(filter.UserIds) > 0 {
		where["user_id"] = map[string]any{
			"_in": filter.UserIds,
		}
	}
	return &where
}

// CountUsers implements AdminCrudActions.
func (u *PostgresAccountStore) CountUserAccounts(ctx context.Context, filter *shared.UserAccountListFilter) (int64, error) {
	where := UserAccountWhere(filter)
	data, err := crudrepo.UserAccount.Count(ctx, u.db, where)
	if err != nil {
		return 0, err
	}
	return data, nil
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

// CreateUserAccount implements UserAccountStore.
func (u *PostgresAccountStore) CreateUserAccount(ctx context.Context, account *models.UserAccount) (*models.UserAccount, error) {
	if account == nil {
		return nil, errors.New("account is nil")
	}
	createdAccount, err := crudrepo.UserAccount.PostOne(ctx, u.db, account)
	if err != nil {
		return nil, fmt.Errorf("error creating user account: %w", err)
	}
	return createdAccount, nil
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
