package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type UserAccountStore interface {
	CountUserAccounts(ctx context.Context, filter *shared.UserAccountListFilter) (int64, error)
	CreateUserAccount(ctx context.Context, account *models.UserAccount) (*models.UserAccount, error)
	FindUserAccountByUserIdAndProvider(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error)
	GetUserAccounts(ctx context.Context, userIds ...uuid.UUID) ([][]*models.UserAccount, error)
	LinkAccount(ctx context.Context, account *models.UserAccount) error
	ListUserAccounts(ctx context.Context, input *shared.UserAccountListParams) ([]*models.UserAccount, error)
	UnlinkAccount(ctx context.Context, userId uuid.UUID, provider models.Providers) error
	UpdateUserAccount(ctx context.Context, account *models.UserAccount) error
	UpdateUserPassword(ctx context.Context, userId uuid.UUID, password string) error
}

type UserAccountService interface {
	Store() UserAccountStore
}

type userAccountService struct {
	store UserAccountStore
}

// Store implements UserAccountService.
func (u *userAccountService) Store() UserAccountStore {
	return u.store
}
func NewUserAccountService(store UserAccountStore) UserAccountService {
	return &userAccountService{
		store: store,
	}
}
