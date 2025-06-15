package core

import (
	"context"
	"log/slog"

	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/jobs"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/stores"

	"github.com/tkahng/authgo/internal/tools/filesystem"
	"github.com/tkahng/authgo/internal/tools/logger"
	"github.com/tkahng/authgo/internal/tools/mailer"
	"github.com/tkahng/authgo/internal/tools/payment"
)

var _ App = (*BaseApp)(nil)

type BaseApp struct {
	cfg      *conf.EnvConfig
	db       database.Dbx
	settings *conf.AppOptions
	payment  services.PaymentService
	logger   *slog.Logger
	fs       filesystem.FileSystem
	mail     mailer.Mailer
	auth     services.AuthService
	team     services.TeamService
	checker  services.ConstraintChecker
	rbac     services.RBACService
	task     services.TaskService
	adapter  stores.StorageAdapterInterface
}

// Adapter implements App.
func (app *BaseApp) Adapter() stores.StorageAdapterInterface {
	return app.adapter
}

func (app *BaseApp) Task() services.TaskService {
	return app.task
}

// User implements App.

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
	l := logger.GetDefaultLogger()

	if err != nil {
		panic(err)
	}
	var mail mailer.Mailer
	if cfg.ResendConfig.ResendApiKey != "" {
		mail = mailer.NewResendMailer(cfg.ResendConfig)
	} else {
		mail = &mailer.LogMailer{}
	}

	adapter := stores.NewStorageAdapter(pool)

	enqueuer := jobs.NewDBEnqueuer(pool)

	rbacService := services.NewRBACService(adapter)

	taskService := services.NewTaskService(adapter)

	paymentClient := payment.NewPaymentClient(cfg.StripeConfig)
	paymentService := services.NewPaymentService(
		paymentClient,
		adapter,
	)

	tokenService := services.NewJwtService()
	passwordService := services.NewPasswordService()

	routineService := services.NewRoutineService()
	authMailService := services.NewMailService(mail)
	authService := services.NewAuthService(
		settings,
		authMailService,
		tokenService,
		passwordService,
		routineService,
		l,
		enqueuer,
		adapter,
	)
	checker := services.NewConstraintCheckerService(
		adapter,
	)

	teamService := services.NewTeamService(adapter)

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
		taskService,
		teamService,
		adapter,
	)
	return app
}

func NewApp(
	fs filesystem.FileSystem,
	pool database.Dbx,
	settings *conf.AppOptions,
	logger *slog.Logger,
	cfg conf.EnvConfig,
	mail mailer.Mailer,
	authService services.AuthService,
	paymentService services.PaymentService,
	checker services.ConstraintChecker,
	rbacService services.RBACService,
	taskService services.TaskService,
	teamService services.TeamService,
	adapter stores.StorageAdapterInterface,
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
		task:     taskService,
		team:     teamService,
		adapter:  adapter,
	}
	return app
}

func (app *BaseApp) Bootstrap() {
	ctx := context.Background()
	db := app.Db()
	queries.EnsureRoleAndPermissions(ctx, db, "superuser", "superuser")
	queries.EnsureRoleAndPermissions(ctx, db, "basic", "basic")
}
