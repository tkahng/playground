package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type UserStore interface {
	// AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	DeleteUser(ctx context.Context, userId uuid.UUID) error
	FindUserByID(ctx context.Context, userId uuid.UUID) (*models.User, error)
	GetUserInfo(ctx context.Context, email string) (*shared.UserInfo, error)
	LoadUsersByUserIds(ctx context.Context, userIds ...uuid.UUID) ([]*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
}

// type UserService interface {
// }
// type userService struct {
// 	adapter stores.StorageAdapterInterface
// }

// func NewUserService(adapter stores.StorageAdapterInterface) UserService {
// 	return &userService{
// 		adapter: adapter,
// 	}
// }
