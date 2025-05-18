package authmodule

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type AuthStore interface {
	UserStore
	TokenStore
	UserAccountStore
}

type UserStore interface {
	GetUserInfo(ctx context.Context, email string) (*shared.UserInfo, error)
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type TokenStore interface {
	VerifyTokenStorage(ctx context.Context, token string) error
	GetToken(ctx context.Context, token string) (*models.Token, error)
	SaveToken(ctx context.Context, token *shared.CreateTokenDTO) error
	DeleteToken(ctx context.Context, token string) error
}

type UserAccountStore interface {
	FindUserAccountByUserIdAndProvider(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error)
	UpdateUserAccount(ctx context.Context, account *models.UserAccount) error
	LinkAccount(ctx context.Context, account *models.UserAccount) error
	UnlinkAccount(ctx context.Context, userId uuid.UUID, provider models.Providers) error
}
