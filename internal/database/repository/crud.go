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
	User               Repository[models.User]               = NewPostgresRepository(UserBuilder)
	Role               Repository[models.Role]               = NewPostgresRepository(RoleBuilder)
	Permission         Repository[models.Permission]         = NewPostgresRepository(PermissionBuilder)
	UserAccount        Repository[models.UserAccount]        = NewPostgresRepository(UserAccountBuilder)
	UserRole           Repository[models.UserRole]           = NewPostgresRepository(UserRoleBuilder)
	UserPermission     Repository[models.UserPermission]     = NewPostgresRepository(UserPermissionBuilder)
	RolePermission     Repository[models.RolePermission]     = NewPostgresRepository(RolePermissionBuilder)
	Token              Repository[models.Token]              = NewPostgresRepository(TokenBuilder)
	TaskProject        Repository[models.TaskProject]        = NewPostgresRepository(TaskProjectBuilder)
	Task               Repository[models.Task]               = NewPostgresRepository(TaskBuilder)
	ProductRole        Repository[models.ProductRole]        = NewPostgresRepository(ProductRoleBuilder)
	ProductPermission  Repository[models.ProductPermission]  = NewPostgresRepository(ProductPermissionBuilder)
	StripeProduct      Repository[models.StripeProduct]      = NewPostgresRepository(StripeProductBuilder)
	StripePrice        Repository[models.StripePrice]        = NewPostgresRepository(StripePriceBuilder)
	StripeCustomer     Repository[models.StripeCustomer]     = NewPostgresRepository(StripeCustomerBuilder)
	StripeSubscription Repository[models.StripeSubscription] = NewPostgresRepository(StripeSubscriptionBuilder)
	Media              Repository[models.Medium]             = NewPostgresRepository(MediaBuilder)
	AiUsage            Repository[models.AiUsage]            = NewPostgresRepository(AiUsageBuilder)
	Team               Repository[models.Team]               = NewPostgresRepository(TeamBuilder)
	TeamMember         Repository[models.TeamMember]         = NewPostgresRepository(TeamMemberBuilder)
	TeamInvitation     Repository[models.TeamInvitation]     = NewPostgresRepository(TeamInvitationBuilder)
	Notification       Repository[models.Notification]       = NewPostgresRepository(NotificationBuilder)
	Job                Repository[models.JobRow]             = NewPostgresRepository(JobBuilder)
	UserReaction       Repository[models.UserReaction]       = NewPostgresRepository(UserReactionBuilder)
)
