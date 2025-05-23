package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/shared"
)

type PostgresAuthStore struct {
	db database.Dbx
	*PostgresAccountStore
	*PostgresUserStore
	*PostgresTokenStore
}

func (s *PostgresAuthStore) WithTx(dbx database.Dbx) services.AuthStore {
	return &PostgresAuthStore{
		db:                   dbx,
		PostgresAccountStore: s.PostgresAccountStore,
		PostgresUserStore:    s.PostgresUserStore,
		PostgresTokenStore:   s.PostgresTokenStore,
	}
}

func NewPostgresAuthStore(db database.Dbx) *PostgresAuthStore {
	return &PostgresAuthStore{
		db:                   db,
		PostgresAccountStore: NewPostgresUserAccountStore(db),
		PostgresUserStore:    NewPostgresUserStore(db),
		PostgresTokenStore:   NewPostgresTokenStore(db),
	}
}

var _ services.AuthStore = (*PostgresAuthStore)(nil)

func (s *PostgresAuthStore) RunInTransaction(
	ctx context.Context,
	fn func(store services.AuthStore) error,
) error {
	return s.db.RunInTransaction(ctx, func(tx database.Dbx) error {
		store := s.WithTx(tx)
		return fn(store)
	})
}

type AuthStoreDecorator struct {
	Delegate                               services.AuthStore
	AssignUserRolesFunc                    func(ctx context.Context, userId uuid.UUID, roleNames ...string) error
	CreateUserFunc                         func(ctx context.Context, user *models.User) (*models.User, error)
	DeleteTokenFunc                        func(ctx context.Context, token string) error
	DeleteUserFunc                         func(ctx context.Context, id uuid.UUID) error
	FindUserAccountByUserIdAndProviderFunc func(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error)
	FindUserByEmailFunc                    func(ctx context.Context, email string) (*models.User, error)
	GetTokenFunc                           func(ctx context.Context, token string) (*models.Token, error)
	GetUserInfoFunc                        func(ctx context.Context, email string) (*shared.UserInfo, error)
	LinkAccountFunc                        func(ctx context.Context, account *models.UserAccount) (*models.UserAccount, error)
	RunInTransactionFunc                   func(ctx context.Context, fn func(store services.AuthStore) error) error
	SaveTokenFunc                          func(ctx context.Context, token *shared.CreateTokenDTO) error
	UnlinkAccountFunc                      func(ctx context.Context, userId uuid.UUID, provider models.Providers) error
	UpdateUserFunc                         func(ctx context.Context, user *models.User) error
	UpdateUserAccountFunc                  func(ctx context.Context, account *models.UserAccount) error
}

var _ services.AuthStore = (*AuthStoreDecorator)(nil)

func (a *AuthStoreDecorator) WithTx(dbx database.Dbx) services.AuthStore {
	return &AuthStoreDecorator{
		Delegate:                               a.Delegate.WithTx(dbx),
		AssignUserRolesFunc:                    a.AssignUserRolesFunc,
		CreateUserFunc:                         a.CreateUserFunc,
		DeleteTokenFunc:                        a.DeleteTokenFunc,
		DeleteUserFunc:                         a.DeleteUserFunc,
		FindUserAccountByUserIdAndProviderFunc: a.FindUserAccountByUserIdAndProviderFunc,
		FindUserByEmailFunc:                    a.FindUserByEmailFunc,
		GetTokenFunc:                           a.GetTokenFunc,
		GetUserInfoFunc:                        a.GetUserInfoFunc,
		LinkAccountFunc:                        a.LinkAccountFunc,
		RunInTransactionFunc:                   a.RunInTransactionFunc,
		SaveTokenFunc:                          a.SaveTokenFunc,
		UnlinkAccountFunc:                      a.UnlinkAccountFunc,
		UpdateUserFunc:                         a.UpdateUserFunc,
		UpdateUserAccountFunc:                  a.UpdateUserAccountFunc,
	}
}

// AssignUserRoles implements services.AuthStore.
func (a *AuthStoreDecorator) AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error {
	if a.AssignUserRolesFunc != nil {
		return a.AssignUserRolesFunc(ctx, userId, roleNames...)
	}
	return a.Delegate.AssignUserRoles(ctx, userId, roleNames...)
}

// CreateUser implements services.AuthStore.
func (a *AuthStoreDecorator) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	if a.CreateUserFunc != nil {
		return a.CreateUserFunc(ctx, user)
	}
	return a.Delegate.CreateUser(ctx, user)
}

// DeleteToken implements services.AuthStore.
func (a *AuthStoreDecorator) DeleteToken(ctx context.Context, token string) error {
	if a.DeleteTokenFunc != nil {
		return a.DeleteTokenFunc(ctx, token)
	}
	return a.Delegate.DeleteToken(ctx, token)
}

// DeleteUser implements services.AuthStore.
func (a *AuthStoreDecorator) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if a.DeleteUserFunc != nil {
		return a.DeleteUserFunc(ctx, id)
	}
	return a.Delegate.DeleteUser(ctx, id)
}

// FindUserAccountByUserIdAndProvider implements services.AuthStore.
func (a *AuthStoreDecorator) FindUserAccountByUserIdAndProvider(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error) {
	if a.FindUserAccountByUserIdAndProviderFunc != nil {
		return a.FindUserAccountByUserIdAndProviderFunc(ctx, userId, provider)
	}
	return a.Delegate.FindUserAccountByUserIdAndProvider(ctx, userId, provider)
}

// FindUserByEmail implements services.AuthStore.
func (a *AuthStoreDecorator) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if a.FindUserByEmailFunc != nil {
		return a.FindUserByEmailFunc(ctx, email)
	}
	return a.Delegate.FindUserByEmail(ctx, email)
}

// GetToken implements services.AuthStore.
func (a *AuthStoreDecorator) GetToken(ctx context.Context, token string) (*models.Token, error) {
	if a.GetTokenFunc != nil {
		return a.GetTokenFunc(ctx, token)
	}
	return a.Delegate.GetToken(ctx, token)
}

// GetUserInfo implements services.AuthStore.
func (a *AuthStoreDecorator) GetUserInfo(ctx context.Context, email string) (*shared.UserInfo, error) {
	if a.GetUserInfoFunc != nil {
		return a.GetUserInfoFunc(ctx, email)
	}
	return a.Delegate.GetUserInfo(ctx, email)
}

// LinkAccount implements services.AuthStore.
func (a *AuthStoreDecorator) LinkAccount(ctx context.Context, account *models.UserAccount) (*models.UserAccount, error) {
	if a.LinkAccountFunc != nil {
		return a.LinkAccountFunc(ctx, account)
	}
	return a.Delegate.LinkAccount(ctx, account)
}

// RunInTransaction implements services.AuthStore.
func (a *AuthStoreDecorator) RunInTransaction(ctx context.Context, fn func(store services.AuthStore) error) error {
	if a.RunInTransactionFunc != nil {
		return a.RunInTransactionFunc(ctx, fn)
	}
	return a.Delegate.RunInTransaction(ctx, fn)
}

// SaveToken implements services.AuthStore.
func (a *AuthStoreDecorator) SaveToken(ctx context.Context, token *shared.CreateTokenDTO) error {
	if a.SaveTokenFunc != nil {
		return a.SaveTokenFunc(ctx, token)
	}
	return a.Delegate.SaveToken(ctx, token)
}

// UnlinkAccount implements services.AuthStore.
func (a *AuthStoreDecorator) UnlinkAccount(ctx context.Context, userId uuid.UUID, provider models.Providers) error {
	if a.UnlinkAccountFunc != nil {
		return a.UnlinkAccountFunc(ctx, userId, provider)
	}
	return a.Delegate.UnlinkAccount(ctx, userId, provider)
}

// UpdateUser implements services.AuthStore.
func (a *AuthStoreDecorator) UpdateUser(ctx context.Context, user *models.User) error {
	if a.UpdateUserFunc != nil {
		return a.UpdateUserFunc(ctx, user)
	}
	return a.Delegate.UpdateUser(ctx, user)
}

// UpdateUserAccount implements services.AuthStore.
func (a *AuthStoreDecorator) UpdateUserAccount(ctx context.Context, account *models.UserAccount) error {
	if a.UpdateUserAccountFunc != nil {
		return a.UpdateUserAccountFunc(ctx, account)
	}
	return a.Delegate.UpdateUserAccount(ctx, account)
}
