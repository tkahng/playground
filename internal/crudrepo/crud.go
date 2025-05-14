package crudrepo

import (
	"github.com/tkahng/authgo/internal/models"
)

var (
	UserBuilder = NewSQLBuilder[models.User](
		UuidV7Generator,
	)
	RoleBuilder = NewSQLBuilder[models.Role](
		UuidV7Generator,
	)
	PermissionBuilder = NewSQLBuilder[models.Permission](
		UuidV7Generator,
	)
	UserAccountBuilder = NewSQLBuilder[models.UserAccount](
		UuidV7Generator,
	)
	UserRoleBuilder = NewSQLBuilder(
		func(s *SQLBuilder[models.UserRole]) error {
			s.skipIdInsert = false
			return nil
		},
	)
	UserPermissionBuilder = NewSQLBuilder(
		func(s *SQLBuilder[models.UserPermission]) error {
			s.skipIdInsert = false
			return nil
		},
	)
	RolePermissionBuilder = NewSQLBuilder(
		func(s *SQLBuilder[models.RolePermission]) error {
			s.skipIdInsert = false
			return nil
		},
	)
	TokenBuilder = NewSQLBuilder[models.Token](
		UuidV7Generator,
	)
	TaskProjectBuilder = NewSQLBuilder[models.TaskProject](
		UuidV7Generator,
	)
	TaskBuilder = NewSQLBuilder[models.Task](
		UuidV7Generator,
	)
	ProductPermissionBuilder = NewSQLBuilder(
		func(s *SQLBuilder[models.ProductPermission]) error {
			s.skipIdInsert = false
			return nil
		},
	)
	ProductRoleBuilder = NewSQLBuilder(
		func(s *SQLBuilder[models.ProductRole]) error {
			s.skipIdInsert = false
			return nil
		},
	)
	StripeProductBuilder = NewSQLBuilder(
		func(s *SQLBuilder[models.StripeProduct]) error {
			s.skipIdInsert = false
			return nil
		},
	)
	StripePriceBuilder = NewSQLBuilder(
		func(s *SQLBuilder[models.StripePrice]) error {
			s.skipIdInsert = false
			return nil
		},
	)
	StripeCustomerBuilder = NewSQLBuilder(
		func(s *SQLBuilder[models.StripeCustomer]) error {
			s.skipIdInsert = false
			return nil
		},
	)
	StripeSubscriptionBuilder = NewSQLBuilder(
		func(s *SQLBuilder[models.StripeSubscription]) error {
			s.skipIdInsert = false
			return nil
		},
	)
	MediaBuilder = NewSQLBuilder[models.Medium](
		UuidV7Generator,
	)
	AiUsageBuilder = NewSQLBuilder[models.AiUsage](
		UuidV7Generator,
	)
)

var (
	User               = NewPostgresRepository(UserBuilder)
	Role               = NewPostgresRepository(RoleBuilder)
	Permission         = NewPostgresRepository(PermissionBuilder)
	UserAccount        = NewPostgresRepository(UserAccountBuilder)
	UserRole           = NewPostgresRepository(UserRoleBuilder)
	UserPermission     = NewPostgresRepository(UserPermissionBuilder)
	RolePermission     = NewPostgresRepository(RolePermissionBuilder)
	Token              = NewPostgresRepository(TokenBuilder)
	TaskProject        = NewPostgresRepository(TaskProjectBuilder)
	Task               = NewPostgresRepository(TaskBuilder)
	ProductRole        = NewPostgresRepository(ProductRoleBuilder)
	ProductPermission  = NewPostgresRepository(ProductPermissionBuilder)
	StripeProduct      = NewPostgresRepository(StripeProductBuilder)
	StripePrice        = NewPostgresRepository(StripePriceBuilder)
	StripeCustomer     = NewPostgresRepository(StripeCustomerBuilder)
	StripeSubscription = NewPostgresRepository(StripeSubscriptionBuilder)
	Media              = NewPostgresRepository(MediaBuilder)
	AiUsage            = NewPostgresRepository(AiUsageBuilder)
)
