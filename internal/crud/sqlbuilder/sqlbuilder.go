package sqlbuilder

import (
	"github.com/tkahng/authgo/internal/crud/models"
	"github.com/tkahng/authgo/internal/crud/repository"
)

var (
	User               = repository.NewSQLBuilder[models.User]()
	Role               = repository.NewSQLBuilder[models.Role]()
	Permission         = repository.NewSQLBuilder[models.Permission]()
	UserAccount        = repository.NewSQLBuilder[models.UserAccount]()
	UserRole           = repository.NewSQLBuilder[models.UserRole]()
	UserPermission     = repository.NewSQLBuilder[models.UserPermission]()
	RolePermission     = repository.NewSQLBuilder[models.RolePermission]()
	Token              = repository.NewSQLBuilder[models.Token]()
	TaskProject        = repository.NewSQLBuilder[models.TaskProject]()
	Task               = repository.NewSQLBuilder[models.Task]()
	ProductRole        = repository.NewSQLBuilder[models.ProductRole]()
	StripeProduct      = repository.NewSQLBuilder[models.StripeProduct]()
	StripePrice        = repository.NewSQLBuilder[models.StripePrice]()
	StripeCustomer     = repository.NewSQLBuilder[models.StripeCustomer]()
	StripeSubscription = repository.NewSQLBuilder[models.StripeSubscription]()
	Media              = repository.NewSQLBuilder[models.Medium]()
	AiUsage            = repository.NewSQLBuilder[models.AiUsage]()
)
