package core

import (
	"context"
	"log/slog"

	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/jobs"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/stores"

	"github.com/tkahng/authgo/internal/tools/filesystem"
	"github.com/tkahng/authgo/internal/tools/hook"
	"github.com/tkahng/authgo/internal/tools/logger"
	"github.com/tkahng/authgo/internal/tools/mailer"
)

var _ App = (*BaseApp)(nil)

type BaseApp struct {
	cfg            *conf.EnvConfig
	db             database.Dbx
	settings       *conf.AppOptions
	payment        services.PaymentService
	logger         *slog.Logger
	fs             filesystem.FileSystem
	mail           mailer.Mailer
	auth           services.AuthService
	team           services.TeamService
	checker        services.ConstraintChecker
	rbac           services.RBACService
	task           services.TaskService
	adapter        stores.StorageAdapterInterface
	teamInvitation services.TeamInvitationService
	jobManager     jobs.JobManager
	jobService     services.JobService
	notifier       services.NotifierService

	lc Lifecycle
}

// check settings -------------------------------------------------------------------------------------
func (app *BaseApp) Config() *conf.EnvConfig {
	if app.cfg == nil {
		opts := conf.AppConfigGetter()
		app.cfg = &opts
		app.settings = opts.ToSettings()
	}
	return app.cfg
}
func (a *BaseApp) Settings() *conf.AppOptions {
	return a.settings
}

// check db -------------------------------------------------------------------------------------

func (app *BaseApp) Db() database.Dbx {
	return app.db
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

// Notifier implements App.
func (app *BaseApp) Notifier() services.NotifierService {
	if app.notifier == nil {
		panic("notifier not initialized")
	}
	return app.notifier
}
func (app *BaseApp) IsBootstrapped() (isBootStrapped bool) {
	if app.cfg == nil {
		return
	}
	if app.db == nil {
		return
	}
	if app.settings == nil {
		return
	}
	if app.mail == nil {
		return
	}
	if app.auth == nil {
		return
	}
	if app.team == nil {
		return
	}
	if app.checker == nil {
		return
	}
	if app.rbac == nil {
		return
	}
	if app.task == nil {
		return
	}
	if app.adapter == nil {
		return
	}
	if app.teamInvitation == nil {
		return
	}
	if app.jobManager == nil {
		return
	}
	if app.jobService == nil {
		return
	}
	if app.notifier == nil {
		return
	}
	return true
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

// Adapter implements App.
func (app *BaseApp) Adapter() stores.StorageAdapterInterface {
	if app.adapter == nil {
		panic("adapter not initialized")
	}
	return app.adapter
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
func (app *BaseApp) RegisterBaseHooks() {
	app.Lifecycle().OnStart().Bind(&hook.Handler[*StartEvent]{
		Func: func(se *StartEvent) error {
			return nil
		},
		Priority: -99,
	})

}

func PrepApp(preApp *BaseApp) {
	if err := preApp.Config(); err != nil {
		panic(err)
	}
	if err := preApp.initDb(); err != nil {
		panic(err)
	}
	if err := preApp.initAdapter(); err != nil {
		panic(err)
	}
	if err := preApp.initPayment(); err != nil {
		panic(err)
	}

	// fs, err := filesystem.NewFileSystem(cfg.StorageConfig)
	if err := preApp.initJobs(); err != nil {
		panic(err)
	}
	if err := preApp.initMail(); err != nil {
		panic(err)
	}

	if err := preApp.initNotifier(); err != nil {
		panic(err)
	}
	if err := preApp.initAuth(); err != nil {
		panic(err)
	}
	if err := preApp.initTeams(); err != nil {
		panic(err)
	}
	if err := preApp.initTasks(); err != nil {
		panic(err)
	}
	if err := preApp.initWorkers(); err != nil {
		panic(err)
	}
}
func NewBaseApp(ctx context.Context, cfg conf.EnvConfig) *BaseApp {
	preApp := &BaseApp{}
	PrepApp(preApp)
	return preApp
}

func newApp(
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
	invitation services.TeamInvitationService,
	jobManager jobs.JobManager,
	jobService services.JobService,
) *BaseApp {
	app := &BaseApp{
		fs:             fs,
		db:             pool,
		settings:       settings,
		logger:         logger,
		cfg:            &cfg,
		mail:           mail,
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
