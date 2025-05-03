package crudrepo

import (
	"github.com/tkahng/authgo/internal/crud/repository"
	"github.com/tkahng/authgo/internal/crud/sqlbuilder"
)

var (
	User               = repository.NewPostgresRepository2(sqlbuilder.User)
	Role               = repository.NewPostgresRepository2(sqlbuilder.Role)
	Permission         = repository.NewPostgresRepository2(sqlbuilder.Permission)
	UserAccount        = repository.NewPostgresRepository2(sqlbuilder.UserAccount)
	UserRole           = repository.NewPostgresRepository2(sqlbuilder.UserRole)
	UserPermission     = repository.NewPostgresRepository2(sqlbuilder.UserPermission)
	RolePermission     = repository.NewPostgresRepository2(sqlbuilder.RolePermission)
	Token              = repository.NewPostgresRepository2(sqlbuilder.Token)
	TaskProject        = repository.NewPostgresRepository2(sqlbuilder.TaskProject)
	Task               = repository.NewPostgresRepository2(sqlbuilder.Task)
	ProductRole        = repository.NewPostgresRepository2(sqlbuilder.ProductRole)
	StripeProduct      = repository.NewPostgresRepository2(sqlbuilder.StripeProduct)
	StripePrice        = repository.NewPostgresRepository2(sqlbuilder.StripePrice)
	StripeCustomer     = repository.NewPostgresRepository2(sqlbuilder.StripeCustomer)
	StripeSubscription = repository.NewPostgresRepository2(sqlbuilder.StripeSubscription)
	Media              = repository.NewPostgresRepository2(sqlbuilder.Media)
	AiUsage            = repository.NewPostgresRepository2(sqlbuilder.AiUsage)
)
