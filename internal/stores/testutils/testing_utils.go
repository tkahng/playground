package testutils

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/types"
)

type UserOptionFunc func(opt *CreateUserOption)
type CreateUserOption struct {
	user      *models.User
	account   *models.UserAccount
	perms     []string
	roleNames []string
}

func UserWithEmail(email string) UserOptionFunc {
	return func(opt *CreateUserOption) {
		opt.user.Email = email
		opt.account.ProviderAccountID = email
	}
}
func UserWithProviderType(providerType models.ProviderTypes) UserOptionFunc {
	return func(opt *CreateUserOption) {
		opt.account.Type = providerType
	}
}

func UserWithProvider(provider models.Providers) UserOptionFunc {
	return func(opt *CreateUserOption) {
		opt.account.Provider = provider
	}
}
func UserWithName(name string) UserOptionFunc {
	return func(opt *CreateUserOption) {
		opt.user.Name = &name
	}
}

func UserWithVerified(emailVerifiedAt *time.Time) UserOptionFunc {
	return func(opt *CreateUserOption) {
		opt.user.EmailVerifiedAt = emailVerifiedAt
	}
}

func UserWithPermission(perms ...string) UserOptionFunc {
	return func(opt *CreateUserOption) {
		opt.perms = perms
	}
}
func UserWithRoles(roleNames ...string) UserOptionFunc {
	return func(opt *CreateUserOption) {
		opt.roleNames = roleNames
	}
}

func CreateUserWithOptions(t testing.TB, adapter stores.StorageAdapterInterface, options ...UserOptionFunc) *models.UserInfo {
	ctx := context.Background()
	opts := &CreateUserOption{
		user: &models.User{
			Email: "tkahng+01@gmail.com",
			Name:  types.Pointer("Test User"),
		},
		account: &models.UserAccount{
			Provider:          models.ProvidersCredentials,
			Type:              models.ProviderTypeCredentials,
			ProviderAccountID: "tkahng+01@gmail.com",
		},
	}
	for _, option := range options {
		option(opts)
	}

	user, err := adapter.User().CreateUser(ctx, &models.User{
		Email:           opts.user.Email,
		Name:            opts.user.Name,
		EmailVerifiedAt: opts.user.EmailVerifiedAt,
		Image:           opts.user.Image,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	account, err := adapter.UserAccount().CreateUserAccount(ctx, &models.UserAccount{
		UserID:            user.ID,
		Provider:          opts.account.Provider,
		Type:              opts.account.Type,
		ProviderAccountID: opts.account.ProviderAccountID,
		Password:          opts.account.Password,
		AccessToken:       opts.account.AccessToken,
		RefreshToken:      opts.account.RefreshToken,
		IDToken:           opts.account.IDToken,
		ExpiresAt:         opts.account.ExpiresAt,
		Scope:             opts.account.Scope,
		SessionState:      opts.account.SessionState,
		TokenType:         opts.account.TokenType,
	})
	if err != nil {
		t.Fatalf("CreateUserAccount() error = %v", err)
	}
	user.Accounts = append(user.Accounts, account)
	if len(opts.perms) > 0 {
		for _, perm := range opts.perms {
			perm, err := adapter.Rbac().FindOrCreatePermission(ctx, perm)
			if err != nil {
				t.Fatalf("FindOrCreatePermission() error = %v", err)
			}
			err = adapter.Rbac().CreateUserPermissions(ctx, user.ID, perm.ID)
			if err != nil {
				t.Fatalf("CreateUserAccount() error = %v", err)
			}
		}
	}
	if len(opts.roleNames) > 0 {
		for _, roleName := range opts.roleNames {
			role, err := adapter.Rbac().FindOrCreateRole(ctx, roleName)
			if err != nil {
				t.Fatalf("FindOrCreatePermission() error = %v", err)
			}
			err = adapter.Rbac().CreateUserRoles(ctx, user.ID, role.ID)
			if err != nil {
				t.Fatalf("CreateUserAccount() error = %v", err)
			}
		}
	}

	return &models.UserInfo{
		User: *user,
	}
}

type TeamOptionFunc func(opt *CreateTeamOptions)

type CreateTeamOptions struct {
	teamName string
	role     models.TeamMemberRole
	billing  bool
}

func TeamWithName(name string) TeamOptionFunc {
	return func(opt *CreateTeamOptions) {
		opt.teamName = name
	}
}

func TeamWithRole(role models.TeamMemberRole) TeamOptionFunc {
	return func(opt *CreateTeamOptions) {
		opt.role = role
	}
}

func TeamWithBilling(billing bool) TeamOptionFunc {
	return func(opt *CreateTeamOptions) {
		opt.billing = billing
	}
}

func CreateTeamAndMemberWithOptions(t testing.TB, adapter stores.StorageAdapterInterface, user *models.User, optFunc ...TeamOptionFunc) *models.TeamInfoModel {
	ctx := context.Background()
	option := &CreateTeamOptions{
		teamName: user.Email,
		role:     models.TeamMemberRoleOwner,
		billing:  true,
	}
	for _, optFunc := range optFunc {
		optFunc(option)
	}
	teamName := option.teamName
	team, err := adapter.TeamGroup().CreateTeam(ctx, teamName, strings.TrimSpace(teamName))
	if err != nil {
		t.Fatalf("Error creating team: %v", err)
	}
	member, err := adapter.TeamMember().CreateTeamMember(ctx, &models.TeamMember{
		TeamID:           team.ID,
		UserID:           &user.ID,
		Role:             option.role,
		Active:           true,
		HasBillingAccess: option.billing,
	})
	if err != nil {
		t.Fatalf("Error creating team member: %v", err)
	}
	return &models.TeamInfoModel{
		Team: *team,
		User: models.User{
			ID:              user.ID,
			Name:            user.Name,
			EmailVerifiedAt: user.EmailVerifiedAt,
		},
		Member: *member,
	}
}

func CreateUser(adapter stores.StorageAdapterInterface, ctx context.Context, email string) *models.User {
	user, err := adapter.User().CreateUser(ctx, &models.User{
		Email: email,
	})
	if err != nil {
		panic(err)
	}
	return user
}

func CreateTeam(adapter stores.StorageAdapterInterface, ctx context.Context, slug string) *models.Team {
	team, err := adapter.TeamGroup().CreateTeam(ctx, slug, slug)
	if err != nil {
		panic(err)
	}
	return team
}

func CreateTeamMember(adapter stores.StorageAdapterInterface, ctx context.Context, team *models.Team, user *models.User, role models.TeamMemberRole, billingAccess bool) *models.TeamMember {
	member, err := adapter.TeamMember().CreateTeamMember(ctx, &models.TeamMember{
		TeamID:           team.ID,
		UserID:           &user.ID,
		Role:             role,
		Active:           true,
		HasBillingAccess: billingAccess,
	})
	if err != nil {
		panic(err)
	}
	return member
}

func CreateTeamProject(adapter stores.StorageAdapterInterface, ctx context.Context, member *models.TeamMember, name string, description string) *models.TaskProject {
	taskProject, err := adapter.Task().CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
		Name:        name,
		Status:      models.TaskProjectStatusDone,
		TeamID:      member.TeamID,
		MemberID:    member.ID,
		Description: &description,
	})
	if err != nil {
		panic(err)
	}
	return taskProject
}

func CreateTask(adapter stores.StorageAdapterInterface, ctx context.Context, task *models.Task) *models.Task {
	task, err := adapter.Task().CreateTask(ctx, task)
	if err != nil {
		panic(err)
	}
	return task
}
