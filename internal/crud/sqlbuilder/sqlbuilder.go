package sqlbuilder

import (
	"github.com/tkahng/authgo/internal/crud/crudModels"
	"github.com/tkahng/authgo/internal/crud/repository"
)

var (
	User               = repository.NewSQLBuilder[crudModels.User]()
	Role               = repository.NewSQLBuilder[crudModels.Role]()
	Permission         = repository.NewSQLBuilder[crudModels.Permission]()
	UserAccount        = repository.NewSQLBuilder[crudModels.UserAccount]()
	UserRole           = repository.NewSQLBuilder[crudModels.UserRole]()
	UserPermission     = repository.NewSQLBuilder[crudModels.UserPermission]()
	RolePermission     = repository.NewSQLBuilder[crudModels.RolePermission]()
	Token              = repository.NewSQLBuilder[crudModels.Token]()
	TaskProject        = repository.NewSQLBuilder[crudModels.TaskProject]()
	Task               = repository.NewSQLBuilder[crudModels.Task]()
	ProductRole        = repository.NewSQLBuilder[crudModels.ProductRole]()
	StripeProduct      = repository.NewSQLBuilder[crudModels.StripeProduct]()
	StripePrice        = repository.NewSQLBuilder[crudModels.StripePrice]()
	StripeCustomer     = repository.NewSQLBuilder[crudModels.StripeCustomer]()
	StripeSubscription = repository.NewSQLBuilder[crudModels.StripeSubscription]()
	Media              = repository.NewSQLBuilder[crudModels.Medium]()
	AiUsage            = repository.NewSQLBuilder[crudModels.AiUsage]()
)
