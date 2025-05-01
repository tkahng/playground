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
		user:           repository.NewPostgresRepository[models.User](dbx),
		role:           repository.NewPostgresRepository[models.Role](dbx),
		permission:     repository.NewPostgresRepository[models.Permission](dbx),
		userAccount:    repository.NewPostgresRepository[models.UserAccount](dbx),
		userRole:       repository.NewPostgresRepository[models.UserRole](dbx),
		rolePermission: repository.NewPostgresRepository[models.RolePermission](dbx),
		token:          repository.NewPostgresRepository[models.Token](dbx),
	}
}
