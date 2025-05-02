package core

import (
	"github.com/tkahng/authgo/internal/crud/crudModels"
	"github.com/tkahng/authgo/internal/crud/repository"
)

type AppRepo struct {
	dbx            repository.DBTX
	user           *repository.PostgresRepository[crudModels.User]
	role           *repository.PostgresRepository[crudModels.Role]
	permission     *repository.PostgresRepository[crudModels.Permission]
	userAccount    *repository.PostgresRepository[crudModels.UserAccount]
	userRole       *repository.PostgresRepository[crudModels.UserRole]
	rolePermission *repository.PostgresRepository[crudModels.RolePermission]
	token          *repository.PostgresRepository[crudModels.Token]
}

func NewAppRepo(dbx repository.DBTX) *AppRepo {
	return &AppRepo{
		dbx:            dbx,
		user:           repository.NewPostgresRepository[crudModels.User](),
		role:           repository.NewPostgresRepository[crudModels.Role](),
		permission:     repository.NewPostgresRepository[crudModels.Permission](),
		userAccount:    repository.NewPostgresRepository[crudModels.UserAccount](),
		userRole:       repository.NewPostgresRepository[crudModels.UserRole](),
		rolePermission: repository.NewPostgresRepository[crudModels.RolePermission](),
		token:          repository.NewPostgresRepository[crudModels.Token](),
	}
}
