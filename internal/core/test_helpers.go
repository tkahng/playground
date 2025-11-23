package core

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mailer"
	"github.com/tkahng/playground/internal/tools/mapper"
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
	slug     string
	teamName string
	active   bool
	role     models.TeamMemberRole
	billing  bool
}

func TeamWithActive(active bool) TeamOptionFunc {
	return func(opt *CreateTeamOptions) {
		opt.active = active
	}
}
func TeamWithName(name string) TeamOptionFunc {
	return func(opt *CreateTeamOptions) {
		opt.teamName = name
	}
}
func TeamWithSlug(slug string) TeamOptionFunc {
	return func(opt *CreateTeamOptions) {
		opt.slug = slug
	}
}

func TeamWithRole(role models.TeamMemberRole) TeamOptionFunc {
	return func(opt *CreateTeamOptions) {
		opt.role = role
		if role == models.TeamMemberRoleOwner {
			opt.billing = true
		} else {
			opt.billing = false
		}
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
		active:   true,
	}
	for _, optFunc := range optFunc {
		optFunc(option)
	}
	option.slug = strings.TrimSpace(option.teamName)
	team, err := app.Adapter().TeamGroup().CreateTeam(ctx, option.teamName, option.slug)
	if err != nil {
		t.Fatalf("Error creating team: %v", err)
	}
	_, err = app.Payment().CreateTeamCustomer(ctx, team, user)
	if err != nil {
		t.Fatalf("Error creating team: %v", err)
	}
	member, err := app.Adapter().TeamMember().CreateTeamMember(ctx, &models.TeamMember{
		TeamID:           team.ID,
		UserID:           types.Pointer(user.ID),
		Role:             option.role,
		HasBillingAccess: option.billing,
		Active:           option.active,
	})
	if err != nil {
		t.Fatalf("Error creating team member: %v", err)
	}
	return &models.TeamInfoModel{
		Team:   *team,
		User:   *user,
		Member: *member,
	}
}

type SubscriptionOptionFunc func(opt *models.StripeSubscription)

func SubscriptionWithItemID(itemID string) SubscriptionOptionFunc {
	return func(opt *models.StripeSubscription) {
		opt.ItemID = itemID
	}
}
func SubscriptionWithID(ID string) SubscriptionOptionFunc {
	return func(opt *models.StripeSubscription) {
		opt.ID = ID
	}
}
func SubscriptionWithPriceID(priceID string) SubscriptionOptionFunc {
	return func(opt *models.StripeSubscription) {
		opt.PriceID = priceID
	}
}
func SubscriptionWithQuantity(quantity int) SubscriptionOptionFunc {
	return func(opt *models.StripeSubscription) {
		opt.Quantity = int64(quantity)
	}
}
func SubscriptionWithCancelAtPeriodEnd(cancelAtPeriodEnd bool) SubscriptionOptionFunc {
	return func(opt *models.StripeSubscription) {
		opt.CancelAtPeriodEnd = cancelAtPeriodEnd
	}
}

func SubscriptionWithStatus(status models.StripeSubscriptionStatus) SubscriptionOptionFunc {
	return func(opt *models.StripeSubscription) {
		opt.Status = status
	}
}

func CreateStripeSubscriptionWithOptions(t testing.TB, app App, customerId string, optFunc ...SubscriptionOptionFunc) *models.StripeSubscription {
	t.Helper()
	sub := &models.StripeSubscription{
		ID:                 "sub_1",
		StripeCustomerID:   customerId,
		Status:             models.StripeSubscriptionStatusActive,
		Metadata:           map[string]string{},
		ItemID:             "item_1",
		PriceID:            "price_pro_month_usd_5000",
		Quantity:           1,
		CancelAtPeriodEnd:  false,
		Created:            time.Now(),
		CurrentPeriodStart: time.Now(),
		CurrentPeriodEnd:   time.Now().UTC().Add(30 * 24 * time.Hour),
	}
	for _, optFunc := range optFunc {
		optFunc(sub)
	}
	return repository.MustCreateOneCtx(t, t.Context(), repository.StripeSubscription, app.Db(), sub)
}

func CreateTeamMemberWithOptions(t testing.TB, app App, teamID uuid.UUID, userId uuid.UUID, optFunc ...TeamOptionFunc) *models.TeamMember {
	ctx := context.Background()
	option := &CreateTeamOptions{
		role:    models.TeamMemberRoleOwner,
		billing: true,
		active:  true,
	}
	for _, optFunc := range optFunc {
		optFunc(option)
	}
	member, err := app.Adapter().TeamMember().CreateTeamMember(ctx, &models.TeamMember{
		TeamID:           teamID,
		UserID:           types.Pointer(userId),
		Role:             option.role,
		HasBillingAccess: option.billing,
		Active:           option.active,
	})
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
func UserWithProviderAccountId(providerId string) UserOptionFunc {
	return func(opt *CreateUserOption) {
		opt.account.ProviderAccountID = providerId
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

func CreateProductsAndPrices(t testing.TB, app *BaseApp) {
	if err := app.Payment().FindAndUpsertAllProducts(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := app.Payment().FindAndUpsertAllPrices(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func WithProjectStatus(status models.TaskProjectStatus) func(*projectTaskOption) {
	return func(r *projectTaskOption) {
		r.project.Status = status
	}
}

func WithProjectDescription(description string) func(*projectTaskOption) {
	return func(r *projectTaskOption) {
		r.project.Description = types.Pointer(description)
	}
}

func WithProjectName(name string) func(*projectTaskOption) {
	return func(r *projectTaskOption) {
		r.project.Name = name
	}
}

func WithTaskByCountAndStatus(count int, status models.TaskStatus) func(*projectTaskOption) {
	return func(r *projectTaskOption) {
		for range count {
			task := &models.Task{
				Name:        fmt.Sprintf("Task %d", r.taskNum+1),
				Description: types.Pointer(fmt.Sprintf("Description for task %d", r.taskNum+1)),
				Status:      status,
				Rank:        float64(r.taskNum + 1),
			}
			r.tasks = append(r.tasks, task)
			r.taskNum++
		}
	}
}

type projectTaskOption struct {
	taskNum int
	project *models.TaskProject
	tasks   []*models.Task
}

func CreateProjectAndTasks(t testing.TB, app App, owner *models.TeamMember, fns ...func(*projectTaskOption)) *models.TaskProject {
	ctx := context.Background()
	projectTaskOptions := &projectTaskOption{
		project: &models.TaskProject{
			TeamID:            owner.TeamID,
			Name:              fmt.Sprintf("Project %s", uuid.NewString()),
			Description:       types.Pointer("description"),
			Status:            models.TaskProjectStatusTodo,
			CreatedByMemberID: &owner.ID,
		},
		tasks: []*models.Task{},
	}
	for _, fn := range fns {
		fn(projectTaskOptions)
	}

	taskProject, err := app.Adapter().Task().CreateTaskProjectWithTasks(ctx, &stores.CreateTaskProjectWithTasksDTO{
		CreateTaskProjectDTO: stores.CreateTaskProjectDTO{
			TeamID:      projectTaskOptions.project.TeamID,
			MemberID:    *projectTaskOptions.project.CreatedByMemberID,
			Name:        projectTaskOptions.project.Name,
			Status:      projectTaskOptions.project.Status,
			Description: projectTaskOptions.project.Description,
		},
		Tasks: mapper.Map(projectTaskOptions.tasks, func(task *models.Task) stores.CreateTaskProjectTaskDTO {
			return stores.CreateTaskProjectTaskDTO{
				Name:        task.Name,
				Description: task.Description,
				Status:      models.TaskStatus(task.Status),
				Rank:        task.Rank,
			}
		}),
	})
	if err != nil {
		t.Fatalf("failed to create project with tasks: %v", err)
	}
	return taskProject
}
