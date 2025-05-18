package useraccountmodule

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
)

type UserAccountStore interface {
	FindUserAccountByUserIdAndProvider(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error)
	UpdateUserAccount(ctx context.Context, account *models.UserAccount) error
	LinkAccount(ctx context.Context, account *models.UserAccount) error
	UnlinkAccount(ctx context.Context, userId uuid.UUID, provider models.Providers) error
}
