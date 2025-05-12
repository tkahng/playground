package repository

import "github.com/tkahng/authgo/internal/models"

var (
	UserBuilder        = NewSQLBuilder[models.User]()
	RoleBuilder        = NewSQLBuilder[models.Role]()
	PermissionBuilder  = NewSQLBuilder[models.Permission]()
	UserAccountBuilder = NewSQLBuilder[models.UserAccount]()
	UserRoleBuilder    = NewSQLBuilder(
		func(s *SQLBuilder[models.UserRole]) error {
			s.skipIdInsert = false
			return nil
		},
	)
	UserPermissionBuilder = NewSQLBuilder[models.UserPermission]()
	RolePermissionBuilder = NewSQLBuilder[models.RolePermission]()
	TokenBuilder          = NewSQLBuilder[models.Token]()
	TaskProjectBuilder    = NewSQLBuilder[models.TaskProject]()
	TaskBuilder           = NewSQLBuilder[models.Task]()
	ProductRoleBuilder    = NewSQLBuilder[models.ProductRole]()
	StripeProductBuilder  = NewSQLBuilder(
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
	MediaBuilder   = NewSQLBuilder[models.Medium]()
	AiUsageBuilder = NewSQLBuilder[models.AiUsage]()
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
	StripeProduct      = NewPostgresRepository(StripeProductBuilder)
	StripePrice        = NewPostgresRepository(StripePriceBuilder)
	StripeCustomer     = NewPostgresRepository(StripeCustomerBuilder)
	StripeSubscription = NewPostgresRepository(StripeSubscriptionBuilder)
	Media              = NewPostgresRepository(MediaBuilder)
	AiUsage            = NewPostgresRepository(AiUsageBuilder)
)
