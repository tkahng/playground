package crudrepo

import (
	"github.com/tkahng/authgo/internal/crud/crudModels"
	"github.com/tkahng/authgo/internal/crud/repository"
)

var (
	User               = repository.NewPostgresRepository[crudModels.User]()
	Role               = repository.NewPostgresRepository[crudModels.Role]()
	Permission         = repository.NewPostgresRepository[crudModels.Permission]()
	UserAccount        = repository.NewPostgresRepository[crudModels.UserAccount]()
	UserRole           = repository.NewPostgresRepository[crudModels.UserRole]()
	UserPermission     = repository.NewPostgresRepository[crudModels.UserPermission]()
	RolePermission     = repository.NewPostgresRepository[crudModels.RolePermission]()
	Token              = repository.NewPostgresRepository[crudModels.Token]()
	TaskProject        = repository.NewPostgresRepository[crudModels.TaskProject]()
	Task               = repository.NewPostgresRepository[crudModels.Task]()
	ProductRole        = repository.NewPostgresRepository[crudModels.ProductRole]()
	StripeProduct      = repository.NewPostgresRepository[crudModels.StripeProduct]()
	StripePrice        = repository.NewPostgresRepository[crudModels.StripePrice]()
	StripeCustomer     = repository.NewPostgresRepository[crudModels.StripeCustomer]()
	StripeSubscription = repository.NewPostgresRepository[crudModels.StripeSubscription]()
)
