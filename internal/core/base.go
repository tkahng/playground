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
	"github.com/tkahng/authgo/internal/tools/payment"
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

// RegisterHooks implements App.
func (app *BaseApp) RegisterHooks() {
	app.Lifecycle().OnStart().Bind(&hook.Handler[*StartEvent]{
		Func: func(se *StartEvent) error {
			return nil
		},
		Priority: -99,
	})

}

func (app *BaseApp) Lifecycle() Lifecycle {
	if app.lc == nil {
		app.lc = NewLifecycle(app.logger)
	}
	return app.lc
}

// Notifier implements App.
func (app *BaseApp) Notifier() services.NotifierService {
	if app.notifier == nil {
		panic("notifier not initialized")
	}
	return app.notifier
}

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

func (app *BaseApp) Db() database.Dbx {
	return app.db
}
func (app *BaseApp) SetDb(db database.Dbx) {
	app.db = db
}

// Payment implements App.
func (a *BaseApp) Payment() services.PaymentService {
	return a.payment
}

// Settings implements App.
func (a *BaseApp) Settings() *conf.AppOptions {
	return a.settings
}

func (a *BaseApp) SetSettings(settings *conf.AppOptions) {
	a.settings = settings
}

func (app *BaseApp) Cfg() *conf.EnvConfig {
	return app.cfg
}

func (a *BaseApp) SetCfg(cfg *conf.EnvConfig) {
	a.cfg = cfg
	a.SetSettings(cfg.ToSettings())
}

// Mailer implements App.

func NewBaseApp(ctx context.Context, cfg conf.EnvConfig) *BaseApp {
	settings := cfg.ToSettings()
	pool := database.CreateQueries(ctx, cfg.Db.DatabaseUrl)
	fs, err := filesystem.NewFileSystem(cfg.StorageConfig)
	l := logger.GetDefaultLogger()

	if err != nil {
		panic(err)
	}
	var mail mailer.Mailer
	if cfg.ResendApiKey != "" {
		mail = mailer.NewResendMailer(cfg.ResendConfig)
	} else {
		mail = &mailer.LogMailer{}
	}
	adapter := stores.NewStorageAdapter(pool)
	jobManager := jobs.NewDbJobManager(pool)
	jobService := services.NewJobService(jobManager)
	mailServiece := services.NewOtpMailService(
		settings,
		mail,
		adapter,
	)

	rbacService := services.NewRBACService(adapter)
	taskService := services.NewTaskService(adapter)

	paymentClient := payment.NewPaymentClient(cfg.StripeConfig)
	paymentService := services.NewPaymentService(
		paymentClient,
		adapter,
	)
	jobService.RegisterWorkers(mailServiece, paymentService)
	authService := services.NewAuthService(
		settings,
		jobService,
		adapter,
	)
	checker := services.NewConstraintCheckerService(
		adapter,
	)

	teamService := services.NewTeamService(adapter)
	teamInvitationService := services.NewInvitationService(
		adapter,
		*settings,
		jobService,
	)
	app := newApp(
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
		teamInvitationService,
		jobManager,
		jobService,
	)
	return app
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
