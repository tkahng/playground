package core

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/shared"
)

type AuthAdapter interface {
	GetUserInfo(ctx context.Context, email string) (*shared.UserInfo, error)
	CreateUser(ctx context.Context, user *shared.User) (*shared.User, error)
	AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error
	FindUserByEmail(ctx context.Context, email string) (*shared.User, error)
	FindUserAccountByUserIdAndProvider(ctx context.Context, userId uuid.UUID, provider shared.Providers) (*shared.UserAccount, error)
	UpdateUser(ctx context.Context, user *shared.User) error
	UpdateUserAccount(ctx context.Context, account *shared.UserAccount) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	LinkAccount(ctx context.Context, account *shared.UserAccount) error
	UnlinkAccount(ctx context.Context, userId uuid.UUID, provider shared.Providers) error
	VerifyTokenStorage(ctx context.Context, token string) error
	GetToken(ctx context.Context, token string) (*shared.Token, error)
	SaveToken(ctx context.Context, token *shared.CreateTokenDTO) error
	DeleteToken(ctx context.Context, token string) error
}
