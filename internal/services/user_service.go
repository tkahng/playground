package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type UserStore interface {
	AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	DeleteUser(ctx context.Context, userId uuid.UUID) error
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)
	FindUserById(ctx context.Context, userId uuid.UUID) (*models.User, error)
	GetUserInfo(ctx context.Context, email string) (*shared.UserInfo, error)
	LoadUsersByUserIds(ctx context.Context, userIds ...uuid.UUID) ([]*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
}

type UserService interface {
	Store() UserStore
}
type userService struct {
	store UserStore
}

// Store implements UserService.
func (u *userService) Store() UserStore {
	return u.store
}

func NewUserService(store UserStore) UserService {
	return &userService{
		store: store,
	}
}
