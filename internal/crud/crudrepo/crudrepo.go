package crudrepo

import (
	"github.com/tkahng/authgo/internal/crud/models"
	"github.com/tkahng/authgo/internal/crud/repository"
)

var (
	User           = repository.NewPostgresRepository[models.User]()
	Role           = repository.NewPostgresRepository[models.Role]()
	Permission     = repository.NewPostgresRepository[models.Permission]()
	UserAccount    = repository.NewPostgresRepository[models.UserAccount]()
	UserRole       = repository.NewPostgresRepository[models.UserRole]()
	RolePermission = repository.NewPostgresRepository[models.RolePermission]()
	Token          = repository.NewPostgresRepository[models.Token]()
	TaskProject    = repository.NewPostgresRepository[models.TaskProject]()
	Task           = repository.NewPostgresRepository[models.Task]()
)
