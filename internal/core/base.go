package core

import (
	"context"

	"log/slog"

	"github.com/tkahng/playground/internal/auth"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/events"
	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/token"

	"github.com/tkahng/playground/internal/tools/filesystem"
	"github.com/tkahng/playground/internal/tools/logger"
	"github.com/tkahng/playground/internal/tools/mailer"
	"github.com/tkahng/playground/internal/tools/sse"
)

var _ App = (*BaseApp)(nil)

type BaseApp struct {
	cfg *conf.EnvConfig

	db      database.Dbx
	adapter stores.StorageAdapterInterface

	logger *slog.Logger

	mailer      mailer.Mailer
	mailService services.OtpMailService

	jwt services.JwtService

	jobManager jobs.JobManager
	jobService services.JobService

	paymentClient services.PaymentClient
	payment       services.PaymentService
	password      services.PasswordService

	auth  services.AuthService
	auth2 auth.AuthService

	rbac    services.RBACService
	checker services.ConstraintChecker

	task services.TaskService

	token token.TokenService

	team           services.TeamService
	teamInvitation services.TeamInvitationService

	notifierPublisher services.Notifier

	fs filesystem.FileSystem

	sseManager sse.Manager

	eventManager events.EventManager
}

// PaymentClient implements App.
func (b *BaseApp) PaymentClient() services.PaymentClient {
	if b.paymentClient == nil {
		panic("payment client not initialized")
	}
	return b.paymentClient
}

// Start implements App.

func (b *BaseApp) Close() {
	slog.Info("closing app.")
	b.db.Close()
}

func (app *BaseApp) Jwt() services.JwtService {
	if app.jwt == nil {
		panic("jwt not initialized")
	}
	return app.jwt
}

func (app *BaseApp) Password() services.PasswordService {
	if app.password == nil {
		panic("password not initialized")
	}
	return app.password
}

// Mailer implements App.
func (app *BaseApp) Mailer() mailer.Mailer {
	if app.mailer == nil {
		panic("mailer not initialized")
	}
	return app.mailer
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
		app.cfg = opts
	}
	return app.cfg
}

// check db -------------------------------------------------------------------------------------

func (app *BaseApp) Db() database.Dbx {
	if app.db == nil {
		panic("db not initialized")

	}
	return app.db
}

// Adapter implements App.
func (app *BaseApp) Adapter() stores.StorageAdapterInterface {
	if app.adapter == nil {
		if app.db != nil {
			app.adapter = stores.NewStorageAdapter(app.db)
		} else {
			panic("db not initialized")
		}
	}
	return app.adapter
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

func (app *BaseApp) Token() token.TokenService {
	if app.token == nil {
		panic("token not initialized")
	}
	return app.token
}

// TeamInvitation implements App.
func (app *BaseApp) TeamInvitation() services.TeamInvitationService {
	if app.teamInvitation == nil {
		panic("team invitation not initialized")
	}
	return app.teamInvitation
}

func (app *BaseApp) Task() services.TaskService {
	if app.task == nil {
		panic("task not initialized")
	}
	return app.task
}

func (app *BaseApp) Rbac() services.RBACService {
	if app.rbac == nil {
		panic("rbac not initialized")
	}
	return app.rbac
}

func (app *BaseApp) Team() services.TeamService {
	if app.team == nil {
		panic("team not initialized")
	}
	return app.team
}

// Checker implements App.
func (a *BaseApp) Checker() services.ConstraintChecker {
	if a.checker == nil {
		panic("checker not initialized")
	}
	return a.checker
}

// Auth implements App.
func (a *BaseApp) Auth() services.AuthService {
	if a.auth == nil {
		panic("auth not initialized")
	}
	return a.auth
}

// Auth2 implements App.
func (a *BaseApp) Auth2() auth.AuthService {
	if a.auth2 == nil {
		panic("auth2 not initialized")
	}
	return a.auth2
}

func (app *BaseApp) Fs() filesystem.FileSystem {
	if app.fs == nil {
		panic("fs not initialized")
	}
	return app.fs
}

// Payment implements App.
func (a *BaseApp) Payment() services.PaymentService {
	if a.payment == nil {
		panic("payment not initialized")
	}
	return a.payment
}
func (app *BaseApp) RunBackgroundProcesses(firstCtx context.Context) {
	go func() {
		app.Logger().Info("Starting poller")
		if err := app.JobManager().Run(firstCtx); err != nil {
			app.Logger().ErrorContext(
				firstCtx,
				"error starting poller",
				slog.Any("error", err),
			)
			return
		}
	}()

	go func() {
		app.Logger().Info("Starting sse manager")
		app.SseManager().Run(firstCtx)
	}()
	go func() {
		app.Logger().Info("Starting event manager")
		if err := app.EventManager().Run(firstCtx); err != nil {
			app.Logger().ErrorContext(
				firstCtx,
				"error starting event manager",
				slog.Any("error", err),
			)
			return
		}
	}()
}

func NewApp(config *conf.EnvConfig) *BaseApp {
	app := new(BaseApp)
	db := database.CreateNewQueriesContext(context.Background(), config.Db.GetDatabaseUrl())
	payment := services.NewPaymentClient(config.StripeConfig)
	mailer := mailer.NewResendMailer(config.ResendConfig)
	logger := logger.GetDefaultLogger()
	app.db = db
	app.logger = logger
	app.cfg = config
	app.paymentClient = payment
	app.mailer = mailer
	assembler := NewAssembler()
	assembler.AssembleApp(app)
	return app
}

func NewTestBaseApp(config *conf.EnvConfig, db database.Dbx) *BaseApp {
	app := new(BaseApp)
	payment := services.NewTestPaymentClient()
	mailer := mailer.NewTestMailer()
	logger := logger.GetDefaultLogger()
	app.logger = logger
	app.db = db
	app.cfg = config
	app.paymentClient = payment
	app.mailer = mailer
	assembler := NewAssembler()
	assembler.AssembleApp(app)
	return app
}
