package authmodule

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

var _ AuthStore = (*PostgresAuthStore)(nil)

func NewAuthStore(
	tokenStore TokenStore,
	userStore UserStore,
	accountStore UserAccountStore,
) AuthStore {
	return &PostgresAuthStore{
		tokenStore:   tokenStore,
		userStore:    userStore,
		accountStore: accountStore,
	}
}

type PostgresAuthStore struct {
	tokenStore   TokenStore
	userStore    UserStore
	accountStore UserAccountStore
}

// AssignUserRoles implements AuthStore.
func (a *PostgresAuthStore) AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error {
	return a.userStore.AssignUserRoles(ctx, userId, roleNames...)
}

// CreateUser implements AuthStore.
func (a *PostgresAuthStore) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	return a.userStore.CreateUser(ctx, user)
}

// DeleteUser implements AuthStore.
func (a *PostgresAuthStore) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return a.userStore.DeleteUser(ctx, id)
}

// FindUserAccountByUserIdAndProvider implements AuthStore.
func (a *PostgresAuthStore) FindUserAccountByUserIdAndProvider(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error) {
	return a.accountStore.FindUserAccountByUserIdAndProvider(ctx, userId, provider)
}

// FindUserByEmail implements AuthStore.
func (a *PostgresAuthStore) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return a.userStore.FindUserByEmail(ctx, email)
}

// GetUserInfo implements AuthStore.
func (a *PostgresAuthStore) GetUserInfo(ctx context.Context, email string) (*shared.UserInfo, error) {
	return a.userStore.GetUserInfo(ctx, email)
}

// LinkAccount implements AuthStore.
func (a *PostgresAuthStore) LinkAccount(ctx context.Context, account *models.UserAccount) error {
	return a.accountStore.LinkAccount(ctx, account)
}

// UnlinkAccount implements AuthStore.
func (a *PostgresAuthStore) UnlinkAccount(ctx context.Context, userId uuid.UUID, provider models.Providers) error {
	return a.accountStore.UnlinkAccount(ctx, userId, provider)
}

// UpdateUser implements AuthStore.
func (a *PostgresAuthStore) UpdateUser(ctx context.Context, user *models.User) error {
	return a.userStore.UpdateUser(ctx, user)
}

// UpdateUserAccount implements AuthStore.
func (a *PostgresAuthStore) UpdateUserAccount(ctx context.Context, account *models.UserAccount) error {
	return a.accountStore.UpdateUserAccount(ctx, account)
}

// DeleteToken implements AuthStore.
func (a *PostgresAuthStore) DeleteToken(ctx context.Context, token string) error {
	return a.tokenStore.DeleteToken(ctx, token)
}

// GetToken implements AuthStore.
func (a *PostgresAuthStore) GetToken(ctx context.Context, token string) (*models.Token, error) {
	return a.tokenStore.GetToken(ctx, token)
}

// SaveToken implements AuthStore.
func (a *PostgresAuthStore) SaveToken(ctx context.Context, token *shared.CreateTokenDTO) error {
	return a.tokenStore.SaveToken(ctx, token)
}

// // VerifyTokenStorage implements AuthStore.
// func (a *PostgresAuthStore) VerifyTokenStorage(ctx context.Context, token string) error {
// 	return a.tokenStore.VerifyTokenStorage(ctx, token)
// }
