package tokenmodule

import (
	"context"

	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type TokenStore interface {
	VerifyTokenStorage(ctx context.Context, token string) error
	GetToken(ctx context.Context, token string) (*models.Token, error)
	SaveToken(ctx context.Context, token *shared.CreateTokenDTO) error
	DeleteToken(ctx context.Context, token string) error
}
