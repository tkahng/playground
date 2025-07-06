package stores

import (
	"context"
	"errors"
	"fmt"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/security"
	"github.com/tkahng/authgo/internal/tools/types"
)

type UserAccountFilter struct {
	repository.PaginatedInput
	repository.SortParams
	Providers     []models.Providers     `query:"providers,omitempty" required:"false" uniqueItems:"true" minimum:"1" maximum:"100" enum:"google,apple,facebook,github,credentials"`
	ProviderTypes []models.ProviderTypes `query:"provider_types,omitempty" required:"false" uniqueItems:"true" minimum:"1" maximum:"100" enum:"oauth,credentials"`
	Q             string                 `query:"q,omitempty" required:"false"`
	Ids           []uuid.UUID            `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	UserIds       []uuid.UUID            `query:"user_ids,omitempty" minimum:"1" maximum:"100" required:"false" format:"uuid"`
}
type DbAccountStoreInterface interface {
	FindUserAccount(ctx context.Context, filter *UserAccountFilter) (*models.UserAccount, error)
	CountUserAccounts(ctx context.Context, filter *UserAccountFilter) (int64, error)
	CreateUserAccount(ctx context.Context, account *models.UserAccount) (*models.UserAccount, error)
	GetUserAccounts(ctx context.Context, userIds ...uuid.UUID) ([][]*models.UserAccount, error)
	ListUserAccounts(ctx context.Context, input *UserAccountFilter) ([]*models.UserAccount, error)
	UnlinkAccount(ctx context.Context, userId uuid.UUID, provider models.Providers) error
	UpdateUserAccount(ctx context.Context, account *models.UserAccount) error
	UpdateUserPassword(ctx context.Context, userId uuid.UUID, password string) error
}

type DbAccountStore struct {
	db database.Dbx
}

func NewDbAccountStore(db database.Dbx) *DbAccountStore {
	return &DbAccountStore{
		db: db,
	}
}

func (s *DbAccountStore) WithTx(tx database.Dbx) *DbAccountStore {
	return &DbAccountStore{
		db: tx,
	}
}

var (
	// UserColumnNames = models.Users.Columns().Names()
	UserAccountColumnNames = []string{"id", "user_id", "type", "provider", "provider_account_id", "created_at", "updated_at"}
	// UserAccountColumnNames = models.UserAccounts.Columns().Names()
)

func (u *DbAccountStore) ListUserAccounts(ctx context.Context, input *UserAccountFilter) ([]*models.UserAccount, error) {
	where := u.filter(input)
	sort := repository.UserAccountBuilder.Sort(input)
	data, err := repository.UserAccount.Get(
		ctx,
		u.db,
		where,
		sort,
		types.Pointer(int(input.Page)),
		types.Pointer(int(input.PerPage)),
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CountUsers implements AdminCrudActions.
func (u *DbAccountStore) CountUserAccounts(ctx context.Context, filter *UserAccountFilter) (int64, error) {
	where := u.filter(filter)
	data, err := repository.UserAccount.Count(ctx, u.db, where)
	if err != nil {
		return 0, err
	}
	return data, nil
}

// FindUserAccountByUserIdAndProvider implements UserAccountStore.
func (u *DbAccountStore) FindUserAccountByUserIdAndProvider(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error) {
	return repository.UserAccount.GetOne(ctx, u.db, &map[string]any{
		models.UserAccountTable.UserID: map[string]any{
			"_eq": userId,
		},
		models.UserAccountTable.Provider: map[string]any{
			"_eq": provider,
		},
	})
}

func (u *DbAccountStore) filter(filter *UserAccountFilter) *map[string]any {
	where := make(map[string]any)
	if filter == nil {
		return &where
	}
	if len(filter.Providers) > 0 {
		where[models.UserAccountTable.Provider] = map[string]any{
			"_in": filter.Providers,
		}
	}
	if len(filter.ProviderTypes) > 0 {
		where[models.UserAccountTable.Type] = map[string]any{
			"_in": filter.ProviderTypes,
		}
	}
	if len(filter.Ids) > 0 {
		where[models.UserAccountTable.ID] = map[string]any{
			"_in": filter.Ids,
		}
	}
	if len(filter.UserIds) > 0 {
		where[models.UserAccountTable.UserID] = map[string]any{
			"_in": filter.UserIds,
		}
	}
	return &where
}

func (u *DbAccountStore) FindUserAccount(ctx context.Context, filter *UserAccountFilter) (*models.UserAccount, error) {
	where := u.filter(filter)
	return repository.UserAccount.GetOne(
		ctx,
		u.db,
		where,
	)
}

// CreateUserAccount implements UserAccountStore.
func (u *DbAccountStore) CreateUserAccount(ctx context.Context, account *models.UserAccount) (*models.UserAccount, error) {
	if account == nil {
		return nil, errors.New("account is nil")
	}
	createdAccount, err := repository.UserAccount.PostOne(
		ctx,
		u.db,
		account,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating user account: %w", err)
	}
	return createdAccount, nil
}

// UnlinkAccount implements UserAccountStore.
func (u *DbAccountStore) UnlinkAccount(ctx context.Context, userId uuid.UUID, provider models.Providers) error {
	_, err := repository.UserAccount.Delete(
		ctx,
		u.db,
		&map[string]any{
			models.UserAccountTable.UserID: map[string]any{
				"_eq": userId,
			},
			models.UserAccountTable.Provider: map[string]any{
				"_eq": provider,
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

// UpdateUserAccount implements UserAccountStore.
func (u *DbAccountStore) UpdateUserAccount(ctx context.Context, account *models.UserAccount) error {
	_, err := repository.UserAccount.PutOne(ctx, u.db, account)
	if err != nil {
		return fmt.Errorf("error updating user account: %w", err)
	}
	return nil
}

func (u *DbAccountStore) GetUserAccounts(ctx context.Context, userIds ...uuid.UUID) ([][]*models.UserAccount, error) {
	// var results []JoinedResult[*crudModels.Permission, uuid.UUID]
	ids := []string{}
	for _, id := range userIds {
		ids = append(ids, id.String())
	}
	data, err := repository.UserAccount.Get(
		ctx,
		u.db,
		&map[string]any{
			models.UserAccountTable.UserID: map[string]any{
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

func (u *DbAccountStore) UpdateUserPassword(ctx context.Context, userId uuid.UUID, password string) error {
	account, err := repository.UserAccount.GetOne(
		ctx,
		u.db,
		&map[string]any{
			models.UserAccountTable.UserID: map[string]any{
				"_eq": userId,
			},
			models.UserAccountTable.Provider: map[string]any{
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
	_, err = repository.UserAccount.PutOne(
		ctx,
		u.db,
		account,
	)
	if err != nil {
		return err
	}
	return nil
}
