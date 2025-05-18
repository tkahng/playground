package usermodule

import (
	"context"

	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type UserStore interface {
	CreateUser(ctx context.Context, params *shared.AuthenticationInput) (*models.User, error)
}
