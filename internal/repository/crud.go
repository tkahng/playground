package repository

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
	UserRoleBuilder = NewSQLBuilder[models.UserRole](
		SkipIdInsert,
	)
	UserPermissionBuilder = NewSQLBuilder[models.UserPermission](
		SkipIdInsert,
	)
	RolePermissionBuilder = NewSQLBuilder[models.RolePermission](
		SkipIdInsert,
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
	ProductPermissionBuilder = NewSQLBuilder[models.ProductPermission](
		SkipIdInsert,
	)
	ProductRoleBuilder = NewSQLBuilder[models.ProductRole](
		SkipIdInsert,
	)
	StripeProductBuilder = NewSQLBuilder[models.StripeProduct](
		SkipIdInsert,
	)
	StripePriceBuilder = NewSQLBuilder[models.StripePrice](
		SkipIdInsert,
	)
	StripeCustomerBuilder = NewSQLBuilder[models.StripeCustomer](
		SkipIdInsert,
	)
	StripeSubscriptionBuilder = NewSQLBuilder[models.StripeSubscription](
		SkipIdInsert,
	)
	MediaBuilder = NewSQLBuilder[models.Medium](
		UuidV7Generator,
	)
	AiUsageBuilder = NewSQLBuilder[models.AiUsage](
		UuidV7Generator,
	)
	TeamBuilder = NewSQLBuilder[models.Team](
		UuidV7Generator,
	)
	TeamMemberBuilder = NewSQLBuilder[models.TeamMember](
		UuidV7Generator,
	)
	TeamInvitationBuilder = NewSQLBuilder[models.TeamInvitation](
		UuidV7Generator,
	)
	NotificationBuilder = NewSQLBuilder[models.Notification](
		UuidV7Generator,
	)
	JobBuilder = NewSQLBuilder[models.JobRow](
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
	Team               = NewPostgresRepository(TeamBuilder)
	TeamMember         = NewPostgresRepository(TeamMemberBuilder)
	TeamInvitation     = NewPostgresRepository(TeamInvitationBuilder)
	Notification       = NewPostgresRepository(NotificationBuilder)
	Job                = NewPostgresRepository(JobBuilder)
)
