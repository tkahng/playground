package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
)

type AccountStoreDecorator struct {
	Delegate                               *DbAccountStore
	CountUserAccountsFunc                  func(ctx context.Context, filter *UserAccountFilter) (int64, error)
	CreateUserAccountFunc                  func(ctx context.Context, account *models.UserAccount) (*models.UserAccount, error)
	FindUserAccountByUserIdAndProviderFunc func(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error)
	GetUserAccountsFunc                    func(ctx context.Context, userIds ...uuid.UUID) ([][]*models.UserAccount, error)
	ListUserAccountsFunc                   func(ctx context.Context, input *UserAccountFilter) ([]*models.UserAccount, error)
	UnlinkAccountFunc                      func(ctx context.Context, userId uuid.UUID, provider models.Providers) error
	UpdateUserAccountFunc                  func(ctx context.Context, account *models.UserAccount) error
	UpdateUserPasswordFunc                 func(ctx context.Context, userId uuid.UUID, password string) error
	WithTxFunc                             func(dbx database.Dbx) *AccountStoreDecorator
	FindUserAccountFunc                    func(ctx context.Context, filter *UserAccountFilter) (*models.UserAccount, error)
}

func NewAccountStoreDecorator(db database.Dbx) *AccountStoreDecorator {
	delegate := NewDbAccountStore(db)
	return &AccountStoreDecorator{
		Delegate: delegate,
	}
}

// FindUserAccount implements DbAccountStoreInterface.
func (a *AccountStoreDecorator) FindUserAccount(ctx context.Context, filter *UserAccountFilter) (*models.UserAccount, error) {
	if a.FindUserAccountFunc != nil {
		return a.FindUserAccountFunc(ctx, filter)
	}
	if a.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return a.Delegate.FindUserAccount(ctx, filter)
}

func (a *AccountStoreDecorator) Cleanup() {
	a.WithTxFunc = nil
	a.CountUserAccountsFunc = nil
	a.CreateUserAccountFunc = nil
	a.FindUserAccountByUserIdAndProviderFunc = nil
	a.GetUserAccountsFunc = nil
	a.ListUserAccountsFunc = nil
	a.UnlinkAccountFunc = nil
	a.UpdateUserAccountFunc = nil
	a.UpdateUserPasswordFunc = nil
}

func (a *AccountStoreDecorator) WithTx(dbx database.Dbx) *AccountStoreDecorator {
	if a.WithTxFunc != nil {
		return a.WithTxFunc(dbx)
	}
	return &AccountStoreDecorator{
		Delegate: a.Delegate.WithTx(dbx),
	}
}

// CountUserAccounts implements DbAccountStoreInterface.
func (a *AccountStoreDecorator) CountUserAccounts(ctx context.Context, filter *UserAccountFilter) (int64, error) {
	if a.CountUserAccountsFunc != nil {
		return a.CountUserAccountsFunc(ctx, filter)
	}
	if a.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return a.Delegate.CountUserAccounts(ctx, filter)
}

// CreateUserAccount implements DbAccountStoreInterface.
func (a *AccountStoreDecorator) CreateUserAccount(ctx context.Context, account *models.UserAccount) (*models.UserAccount, error) {
	if a.CreateUserAccountFunc != nil {
		return a.CreateUserAccountFunc(ctx, account)
	}
	if a.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return a.Delegate.CreateUserAccount(ctx, account)
}

// FindUserAccountByUserIdAndProvider implements DbAccountStoreInterface.

// GetUserAccounts implements DbAccountStoreInterface.
func (a *AccountStoreDecorator) GetUserAccounts(ctx context.Context, userIds ...uuid.UUID) ([][]*models.UserAccount, error) {
	if a.GetUserAccountsFunc != nil {
		return a.GetUserAccountsFunc(ctx, userIds...)
	}
	if a.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return a.Delegate.GetUserAccounts(ctx, userIds...)
}

// ListUserAccounts implements DbAccountStoreInterface.
func (a *AccountStoreDecorator) ListUserAccounts(ctx context.Context, input *UserAccountFilter) ([]*models.UserAccount, error) {
	if a.ListUserAccountsFunc != nil {
		return a.ListUserAccountsFunc(ctx, input)
	}
	if a.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return a.Delegate.ListUserAccounts(ctx, input)
}

// UnlinkAccount implements DbAccountStoreInterface.
func (a *AccountStoreDecorator) UnlinkAccount(ctx context.Context, userId uuid.UUID, provider models.Providers) error {
	if a.UnlinkAccountFunc != nil {
		return a.UnlinkAccountFunc(ctx, userId, provider)
	}
	if a.Delegate == nil {
		return ErrDelegateNil
	}
	return a.Delegate.UnlinkAccount(ctx, userId, provider)
}

// UpdateUserAccount implements DbAccountStoreInterface.
func (a *AccountStoreDecorator) UpdateUserAccount(ctx context.Context, account *models.UserAccount) error {
	if a.UpdateUserAccountFunc != nil {
		return a.UpdateUserAccountFunc(ctx, account)
	}
	if a.Delegate == nil {
		return ErrDelegateNil
	}
	return a.Delegate.UpdateUserAccount(ctx, account)
}

// UpdateUserPassword implements DbAccountStoreInterface.
func (a *AccountStoreDecorator) UpdateUserPassword(ctx context.Context, userId uuid.UUID, password string) error {
	if a.UpdateUserPasswordFunc != nil {
		return a.UpdateUserPasswordFunc(ctx, userId, password)
	}
	if a.Delegate == nil {
		return ErrDelegateNil
	}
	return a.Delegate.UpdateUserPassword(ctx, userId, password)
}

var _ DbAccountStoreInterface = (*AccountStoreDecorator)(nil)
