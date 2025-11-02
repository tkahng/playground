package core

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mailer"
	"github.com/tkahng/playground/internal/tools/types"
)

func ExtractTestMailer(t testing.TB, testApi App) *mailer.TestMailer {
	var testMailer *mailer.TestMailer
	if m, ok := testApi.Mailer().(*mailer.TestMailer); ok {
		testMailer = m
	} else {
		t.Fatal("mailer is not a TestMailer")
	}
	return testMailer
}
func ExtractTestPaymentClient(t testing.TB, app App) *services.MockPaymentClient {
	var paymenClient *services.MockPaymentClient
	if m, ok := app.PaymentClient().(*services.MockPaymentClient); ok {
		paymenClient = m
	} else {
		t.Fatal("mailer is not a TestMailer")
	}
	return paymenClient
}

func ExtractAdapterDecorator(t testing.TB, app App) *stores.StorageAdapterDecorator {
	var adapter *stores.StorageAdapterDecorator
	if m, ok := app.Adapter().(*stores.StorageAdapterDecorator); ok {
		adapter = m
	} else {
		t.Fatal("mailer is not a TestMailer")
	}
	return adapter
}

func CreateTokenHeader(t testing.TB, app App, email string) string {
	t.Helper()
	ctx := context.Background()
	tokensVerifiedTokens, err := app.Auth().GenerateAuthTokens(ctx, email)
	if err != nil {
		t.Errorf("Error creating auth tokens: %v", err)
	}
	VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)
	return VerifiedHeader
}
func CreateAccessHeaderAndRefreshToken(t testing.TB, app App, email string) (header string, refreshToken string) {
	t.Helper()
	ctx := context.Background()
	tokensVerifiedTokens, err := app.Auth().GenerateAuthTokens(ctx, email)
	if err != nil {
		t.Errorf("Error creating auth tokens: %v", err)
	}
	header = fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)
	refreshToken = tokensVerifiedTokens.Tokens.RefreshToken
	return
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

func CreateTeamAndMemberWithOptions(t testing.TB, app App, user *models.User, optFunc ...TeamOptionFunc) *models.TeamInfoModel {
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
	team, err := app.Adapter().TeamGroup().CreateTeam(ctx, teamName, strings.TrimSpace(teamName))
	if err != nil {
		t.Fatalf("Error creating team: %v", err)
	}
	_, err = app.Payment().CreateTeamCustomer(ctx, team, user)
	if err != nil {
		t.Fatalf("Error creating team: %v", err)
	}
	member, err := app.Adapter().TeamMember().CreateTeamMember(ctx, team.ID, user.ID, option.role, option.billing)
	if err != nil {
		t.Fatalf("Error creating team member: %v", err)
	}
	return &models.TeamInfoModel{
		Team:   *team,
		User:   *user,
		Member: *member,
	}
}

func CreateTeamMemberWithOptions(t testing.TB, app App, teamID uuid.UUID, userId uuid.UUID, optFunc ...TeamOptionFunc) *models.TeamMember {
	ctx := context.Background()
	option := &CreateTeamOptions{
		role:    models.TeamMemberRoleOwner,
		billing: true,
	}
	for _, optFunc := range optFunc {
		optFunc(option)
	}
	member, err := app.Adapter().TeamMember().CreateTeamMember(ctx, teamID, userId, option.role, option.billing)
	if err != nil {
		t.Fatalf("Error creating team member: %v", err)
	}
	return member
}

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
		if provider == models.ProvidersCredentials {
			opt.account.Type = models.ProviderTypeCredentials
		} else {
			opt.account.Type = models.ProviderTypeOAuth
			opt.user.EmailVerifiedAt = types.Pointer(time.Now())
		}
	}
}
func UserWithName(name string) UserOptionFunc {
	return func(opt *CreateUserOption) {
		opt.user.Name = &name
	}
}

func UserWithPassword(password string) UserOptionFunc {
	return func(opt *CreateUserOption) {
		hashService := services.NewHashService()
		hp, err := hashService.Hash(password)
		if err != nil {
			panic(err)
		}
		opt.account.Password = &hp
	}
}
func UserWithVerifiedNow() UserOptionFunc {
	return func(opt *CreateUserOption) {
		opt.user.EmailVerifiedAt = types.Pointer(time.Now())
	}
}
func UserWithVerified(emailVerifiedAt time.Time) UserOptionFunc {
	return func(opt *CreateUserOption) {
		opt.user.EmailVerifiedAt = &emailVerifiedAt
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

func CreateUserWithOptions(t testing.TB, app App, options ...UserOptionFunc) *models.UserInfo {
	t.Helper()
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

	user, err := app.Adapter().User().CreateUser(ctx, &models.User{
		Email:           opts.user.Email,
		Name:            opts.user.Name,
		EmailVerifiedAt: opts.user.EmailVerifiedAt,
		Image:           opts.user.Image,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	account, err := app.Adapter().UserAccount().CreateUserAccount(ctx, &models.UserAccount{
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
			perm, err := app.Adapter().Rbac().FindOrCreatePermission(ctx, perm)
			if err != nil {
				t.Fatalf("FindOrCreatePermission() error = %v", err)
			}
			err = app.Adapter().Rbac().CreateUserPermissions(ctx, user.ID, perm.ID)
			if err != nil {
				t.Fatalf("CreateUserAccount() error = %v", err)
			}
		}
	}
	if len(opts.roleNames) > 0 {
		for _, roleName := range opts.roleNames {
			role, err := app.Adapter().Rbac().FindOrCreateRole(ctx, roleName)
			if err != nil {
				t.Fatalf("FindOrCreatePermission() error = %v", err)
			}
			err = app.Adapter().Rbac().CreateUserRoles(ctx, user.ID, role.ID)
			if err != nil {
				t.Fatalf("CreateUserAccount() error = %v", err)
			}
		}
	}
	if opts.user.EmailVerifiedAt != nil {
		_, err := app.Payment().CreateUserCustomer(ctx, user)
		if err != nil {
			t.Fatalf("CreateUserCustomer() error = %v", err)
		}
	}
	return &models.UserInfo{
		User: *user,
	}

}
