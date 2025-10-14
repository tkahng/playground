package repository

import (
	"github.com/tkahng/playground/internal/models"
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
		InsertID,
	)
	UserPermissionBuilder = NewSQLBuilder[models.UserPermission](
		InsertID,
	)
	RolePermissionBuilder = NewSQLBuilder[models.RolePermission](
		InsertID,
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
		InsertID,
	)
	ProductRoleBuilder = NewSQLBuilder[models.ProductRole](
		InsertID,
	)
	StripeProductBuilder = NewSQLBuilder[models.StripeProduct](
		InsertID,
	)
	StripePriceBuilder = NewSQLBuilder[models.StripePrice](
		InsertID,
	)
	StripeCustomerBuilder = NewSQLBuilder[models.StripeCustomer](
		InsertID,
	)
	StripeSubscriptionBuilder = NewSQLBuilder[models.StripeSubscription](
		InsertID,
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
	UserReactionBuilder = NewSQLBuilder[models.UserReaction](
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
	UserReaction       = NewPostgresRepository(UserReactionBuilder)
)
