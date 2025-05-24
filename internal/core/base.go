package core

import (
	"context"
	"log/slog"

	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/stores"

	"github.com/tkahng/authgo/internal/tools/filesystem"
	"github.com/tkahng/authgo/internal/tools/logger"
	"github.com/tkahng/authgo/internal/tools/mailer"
	"github.com/tkahng/authgo/internal/tools/payment"
)

type AppFactory interface {
	New(
		fs filesystem.FileSystem,
		pool *database.Queries,
		settings *conf.AppOptions,
		cfg conf.EnvConfig,
		mail mailer.Mailer,
		authService services.AuthService,
		paymentService services.PaymentService,
		checker *services.ConstraintCheckerService,
		rbacService services.RBACService,
		userService services.UserService,
		userAccountService services.UserAccountService,
		taskService services.TaskService,
	) *BaseApp
}
type NewAppFunc func(
	fs filesystem.FileSystem,
	pool *database.Queries,
	settings *conf.AppOptions,
	cfg conf.EnvConfig,
	mail mailer.Mailer,
	authService services.AuthService,
	paymentService services.PaymentService,
	checker *services.ConstraintCheckerService,
	rbacService services.RBACService,
	userService services.UserService,
	userAccountService services.UserAccountService,
	taskService services.TaskService,
) *BaseApp

var _ App = (*BaseApp)(nil)

type BaseApp struct {
	cfg      *conf.EnvConfig
	db       *database.Queries
	settings *conf.AppOptions
	payment  services.PaymentService
	logger   *slog.Logger
	fs       filesystem.FileSystem
	mail     mailer.Mailer
	auth     services.AuthService
	team     services.TeamService
	checker  services.ConstraintChecker
	rbac     services.RBACService
	user     services.UserService
	userAcc  services.UserAccountService
	task     services.TaskService
}

// UserAccount implements App.
func (app *BaseApp) UserAccount() services.UserAccountService {
	return app.userAcc
}

func (app *BaseApp) Task() services.TaskService {
	return app.task
}

// User implements App.
func (app *BaseApp) User() services.UserService {
	return app.user
}

// Rbac implements App.
func (app *BaseApp) Rbac() services.RBACService {
	return app.rbac
}

func (app *BaseApp) Team() services.TeamService {
	return app.team
}

// Checker implements App.
func (a *BaseApp) Checker() services.ConstraintChecker {
	return a.checker
}

// Auth implements App.
func (a *BaseApp) Auth() services.AuthService {
	return a.auth
}

func (app *BaseApp) Fs() filesystem.FileSystem {
	return app.fs
}

func (app *BaseApp) Logger() *slog.Logger {
	return app.logger
}
func (app *BaseApp) Db() database.Dbx {
	return app.db
}

// Payment implements App.
func (a *BaseApp) Payment() services.PaymentService {
	return a.payment
}

// Settings implements App.
func (a *BaseApp) Settings() *conf.AppOptions {
	return a.settings
}

func (app *BaseApp) Cfg() *conf.EnvConfig {
	return app.cfg
}

// Mailer implements App.
func (app *BaseApp) Mailer() mailer.Mailer {
	return app.mail
}

func NewBaseApp(ctx context.Context, cfg conf.EnvConfig) *BaseApp {
	settings := cfg.ToSettings()
	pool := database.CreateQueries(ctx, cfg.Db.DatabaseUrl)
	fs, err := filesystem.NewFileSystem(cfg.StorageConfig)
	l := logger.GetDefaultLogger(slog.LevelInfo)

	if err != nil {
		panic(err)
	}
	var mail mailer.Mailer
	if cfg.ResendConfig.ResendApiKey != "" {
		mail = mailer.NewResendMailer(cfg.ResendConfig)
	} else {
		mail = &mailer.LogMailer{}
	}
	userStore := stores.NewPostgresUserStore(pool)
	userService := services.NewUserService(userStore)
	userAccountStore := stores.NewPostgresUserAccountStore(pool)
	userAccountService := services.NewUserAccountService(userAccountStore)
	rbac := stores.NewPostgresRBACStore(pool)
	rbacService := services.NewRBACService(rbac)
	taskStore := stores.NewTaskStore(pool)
	taskService := services.NewTaskService(taskStore)
	paymentStore := stores.NewPostgresPaymentStore(pool)
	paymentClient := payment.NewPaymentClient(cfg.StripeConfig)
	paymentService := services.NewPaymentService(
		paymentClient,
		paymentStore,
	)

	tokenService := services.NewJwtService()
	passwordService := services.NewPasswordService()
	authStore := stores.NewPostgresAuthStore(pool)
	workerService := services.NewRoutineService()
	authMailService := services.NewMailService(mail)
	authService := services.NewAuthService(
		settings,
		authStore,
		authMailService,
		tokenService,
		passwordService,
		workerService,
		l,
	)
	checkerStore := stores.NewPostgresConstraintStore(pool)
	checker := services.NewConstraintCheckerService(
		checkerStore,
	)

	teamStore := stores.NewPostgresTeamServiceStore(pool)
	teamService := services.NewTeamService(teamStore)

	app := NewApp(
		fs,
		pool,
		settings,
		l,
		cfg,
		mail,
		authService,
		paymentService,
		checker, // pass as ConstraintChecker
		rbacService,
		userService,
		userAccountService,
		taskService,
		teamService,
	)
	return app
}

func NewApp(
	fs filesystem.FileSystem,
	pool *database.Queries,
	settings *conf.AppOptions,
	logger *slog.Logger,
	cfg conf.EnvConfig,
	mail mailer.Mailer,
	authService services.AuthService,
	paymentService services.PaymentService,
	checker services.ConstraintChecker,
	rbacService services.RBACService,
	userService services.UserService,
	userAccountService services.UserAccountService,
	taskService services.TaskService,
	teamService services.TeamService,
) *BaseApp {
	app := &BaseApp{
		fs:       fs,
		db:       pool,
		settings: settings,
		logger:   logger,
		cfg:      &cfg,
		mail:     mail,
		auth:     authService,
		payment:  paymentService,
		checker:  checker,
		rbac:     rbacService,
		user:     userService,
		userAcc:  userAccountService,
		task:     taskService,
		team:     teamService,
	}
	return app
}

func (app *BaseApp) Bootstrap() {
	ctx := context.Background()
	db := app.Db()
	queries.EnsureRoleAndPermissions(ctx, db, "superuser", "superuser")
	queries.EnsureRoleAndPermissions(ctx, db, "basic", "basic")
}
