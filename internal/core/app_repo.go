package core

import (
	"github.com/tkahng/authgo/internal/crud/models"
	"github.com/tkahng/authgo/internal/crud/repository"
)

type AppRepo struct {
	dbx            repository.DBTX
	user           *repository.PostgresRepository[models.User]
	role           *repository.PostgresRepository[models.Role]
	permission     *repository.PostgresRepository[models.Permission]
	userAccount    *repository.PostgresRepository[models.UserAccount]
	userRole       *repository.PostgresRepository[models.UserRole]
	rolePermission *repository.PostgresRepository[models.RolePermission]
	token          *repository.PostgresRepository[models.Token]
}

func NewAppRepo(dbx repository.DBTX) *AppRepo {
	return &AppRepo{
		dbx:            dbx,
		user:           repository.NewPostgresRepository[models.User](),
		role:           repository.NewPostgresRepository[models.Role](),
		permission:     repository.NewPostgresRepository[models.Permission](),
		userAccount:    repository.NewPostgresRepository[models.UserAccount](),
		userRole:       repository.NewPostgresRepository[models.UserRole](),
		rolePermission: repository.NewPostgresRepository[models.RolePermission](),
		token:          repository.NewPostgresRepository[models.Token](),
	}
}
