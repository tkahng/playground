package core

import (
	"fmt"
	"log/slog"

	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/events"
	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"

	"github.com/tkahng/playground/internal/tools/filesystem"
	"github.com/tkahng/playground/internal/tools/logger"
	"github.com/tkahng/playground/internal/tools/sse"
)

var _ App = (*BaseApp)(nil)

type BaseApp struct {
	cfg *conf.EnvConfig

	lc Lifecycle

	db      database.Dbx
	adapter stores.StorageAdapterInterface

	logger      *slog.Logger
	mailService services.OtpMailService

	jobManager jobs.JobManager
	jobService services.JobService

	payment services.PaymentService

	auth    services.AuthService
	rbac    services.RBACService
	checker services.ConstraintChecker

	task services.TaskService

	team           services.TeamService
	teamInvitation services.TeamInvitationService

	notifierPublisher services.Notifier

	fs filesystem.FileSystem

	sseManager sse.Manager

	eventManager events.EventManager
}

// MailService implements App.
func (app *BaseApp) MailService() services.OtpMailService {
	if app.mailService == nil {
		panic("mail service not initialized")
	}
	return app.mailService
}

// EventManager implements App.
func (app *BaseApp) EventManager() events.EventManager {
	if app.eventManager == nil {
		panic("event manager not initialized")
	}
	return app.eventManager
}

// NotificationPublisher implements App.
func (app *BaseApp) NotificationPublisher() services.Notifier {
	if app.notifierPublisher == nil {
		panic("notifier not initialized")
	}
	return app.notifierPublisher
}

// SseManager implements App.
func (app *BaseApp) SseManager() sse.Manager {
	if app.sseManager == nil {
		panic("sse manager not initialized")
	}
	return app.sseManager
}

// check settings -------------------------------------------------------------------------------------
func (app *BaseApp) Config() *conf.EnvConfig {
	if app.cfg == nil {
		opts := conf.AppConfigGetter()
		app.cfg = &opts
	}
	return app.cfg
}

// check db -------------------------------------------------------------------------------------

func (app *BaseApp) Db() database.Dbx {
	return app.db
}

// Adapter implements App.
func (app *BaseApp) Adapter() stores.StorageAdapterInterface {
	if app.adapter == nil {
		panic("adapter not initialized")
	}
	return app.adapter
}

func (app *BaseApp) Lifecycle() Lifecycle {
	if app.lc == nil {
		app.lc = NewLifecycle(app.logger)
	}
	return app.lc
}

// check logging -------------------------------------------------------------------------------------
func (app *BaseApp) Logger() *slog.Logger {
	if app.logger == nil {
		app.logger = logger.GetDefaultLogger()
	}
	return app.logger
}

// BootStrap implements App.

// JobManager implements App.
func (app *BaseApp) JobManager() jobs.JobManager {
	if app.jobManager == nil {
		panic("job manager not initialized")
	}
	return app.jobManager
}

// JobService implements App.
func (app *BaseApp) JobService() services.JobService {
	if app.jobService == nil {
		panic("job service not initialized")
	}
	return app.jobService
}

// TeamInvitation implements App.
func (app *BaseApp) TeamInvitation() services.TeamInvitationService {
	if app.teamInvitation == nil {
		panic("team invitation not initialized")
	}
	return app.teamInvitation
}

func (app *BaseApp) Task() services.TaskService {
	return app.task
}

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

// Payment implements App.
func (a *BaseApp) Payment() services.PaymentService {
	return a.payment
}

// Settings implements App.

// Mailer implements App.
// RegisterBaseHooks implements App.

func BootstrappedApp(cfg conf.EnvConfig) *BaseApp {
	app := new(BaseApp)
	if err := app.Bootstrap(); err != nil {
		panic(fmt.Errorf("failed to bootstrap app: %w", err))
	}
	return app
}

func newApp(
	fs filesystem.FileSystem,
	pool database.Dbx,
	logger *slog.Logger,
	cfg conf.EnvConfig,
	authService services.AuthService,
	paymentService services.PaymentService,
	checker services.ConstraintChecker,
	rbacService services.RBACService,
	taskService services.TaskService,
	teamService services.TeamService,
	adapter stores.StorageAdapterInterface,
	invitation services.TeamInvitationService,
	jobManager jobs.JobManager,
	jobService services.JobService,
) *BaseApp {
	app := &BaseApp{
		fs:             fs,
		db:             pool,
		logger:         logger,
		cfg:            &cfg,
		auth:           authService,
		payment:        paymentService,
		checker:        checker,
		rbac:           rbacService,
		task:           taskService,
		team:           teamService,
		adapter:        adapter,
		teamInvitation: invitation,
		jobManager:     jobManager,
		jobService:     jobService,
	}
	return app
}
