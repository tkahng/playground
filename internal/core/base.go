package core

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/tkahng/playground/internal/auth"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/events"
	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/token"
	"github.com/tkahng/playground/internal/workers"

	"github.com/tkahng/playground/internal/tools/filesystem"
	"github.com/tkahng/playground/internal/tools/logger"
	"github.com/tkahng/playground/internal/tools/mailer"
	"github.com/tkahng/playground/internal/tools/sse"
	"github.com/tkahng/playground/internal/tools/ticket"
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
	hash          services.HashService
	encrypt       services.Encryptor
	auth          auth.AuthService

	rbac    services.RBACService
	checker services.ConstraintChecker

	task    services.TaskService
	aiUsage services.AiUsageService

	token token.TokenService

	team           services.TeamService
	teamInvitation services.TeamInvitationService

	notifierPublisher       services.Notifier
	playerNotifierPublisher services.PlayerNotifier

	fs filesystem.FileSystem

	sseManager sse.Manager
	sseTickets ticket.Storer

	eventManager events.EventManager

	rpsGame services.RpsGameService
	ledger  services.LedgerService
	betting services.BettingService

	bgWg sync.WaitGroup
}

// RpsGame implements [App].
func (b *BaseApp) RpsGame() services.RpsGameService {
	if b.rpsGame == nil {
		panic("rps game not initialized")
	}
	return b.rpsGame
}

// Ledger implements [App].
func (b *BaseApp) Ledger() services.LedgerService {
	if b.ledger == nil {
		panic("ledger service not initialized")
	}
	return b.ledger
}

// Betting implements [App].
func (b *BaseApp) Betting() services.BettingService {
	if b.betting == nil {
		panic("betting service not initialized")
	}
	return b.betting
}

func (b *BaseApp) Encrypt() services.Encryptor {
	if b.encrypt == nil {
		panic("encrypt not initialized")
	}
	return b.encrypt
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
	b.bgWg.Wait()
	b.db.Close()
}

func (app *BaseApp) Jwt() services.JwtService {
	if app.jwt == nil {
		panic("jwt not initialized")
	}
	return app.jwt
}

func (app *BaseApp) Hash() services.HashService {
	if app.hash == nil {
		panic("hash not initialized")
	}
	return app.hash
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

// PlayerNotificationPublisher implements App.
func (app *BaseApp) PlayerNotificationPublisher() services.PlayerNotifier {
	if app.playerNotifierPublisher == nil {
		panic("player notifier not initialized")
	}
	return app.playerNotifierPublisher
}

// SseManager implements App.
func (app *BaseApp) SseManager() sse.Manager {
	if app.sseManager == nil {
		panic("sse manager not initialized")
	}
	return app.sseManager
}

// SseTickets implements App.
func (app *BaseApp) SseTickets() ticket.Storer {
	if app.sseTickets == nil {
		panic("sse tickets not initialized")
	}
	return app.sseTickets
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

func (app *BaseApp) AiUsage() services.AiUsageService {
	if app.aiUsage == nil {
		panic("ai usage service not initialized")
	}
	return app.aiUsage
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
func (a *BaseApp) Auth() auth.AuthService {
	if a.auth == nil {
		panic("auth2 not initialized")
	}
	return a.auth
}

func (app *BaseApp) Fs() filesystem.FileSystem {
	if app.fs == nil {
		panic("fs not initialized")
	}
	return app.fs
}

// SetFs replaces the filesystem used by the app. Intended for tests only.
func (app *BaseApp) SetFs(fs filesystem.FileSystem) {
	app.fs = fs
}

// Payment implements App.
func (a *BaseApp) Payment() services.PaymentService {
	if a.payment == nil {
		panic("payment not initialized")
	}
	return a.payment
}
func (app *BaseApp) syncPlanFeatures(ctx context.Context) {
	if err := services.SyncPlanFeatures(ctx, app.Adapter()); err != nil {
		app.Logger().ErrorContext(ctx, "plan features sync failed", slog.Any("error", err))
		return
	}
	app.Logger().InfoContext(ctx, "plan features sync complete")
}

func (app *BaseApp) seedHousePlayer(ctx context.Context) {
	if err := services.SeedHousePlayer(ctx, app.Adapter()); err != nil {
		app.Logger().ErrorContext(ctx, "house player seed failed", slog.Any("error", err))
		return
	}
	app.Logger().InfoContext(ctx, "house player ready")
}

func (app *BaseApp) RunBackgroundProcesses(ctx context.Context) {
	app.syncPlanFeatures(ctx)
	app.seedHousePlayer(ctx)

	run := func(name string, fn func()) {
		app.bgWg.Add(1)
		go func() {
			defer app.bgWg.Done()
			defer func() {
				if r := recover(); r != nil {
					app.Logger().ErrorContext(ctx, "panic in background process",
						slog.String("process", name),
						slog.Any("error", r),
					)
				}
			}()
			fn()
		}()
	}

	run("poller", func() {
		app.Logger().Info("Starting poller")
		if err := app.JobManager().Run(ctx); err != nil {
			app.Logger().ErrorContext(ctx, "error starting poller", slog.Any("error", err))
		}
	})

	run("sse manager", func() {
		app.Logger().Info("Starting sse manager")
		app.SseManager().Run(ctx)
	})

	run("event manager", func() {
		app.Logger().Info("Starting event manager")
		if err := app.EventManager().Run(ctx); err != nil {
			app.Logger().ErrorContext(ctx, "error starting event manager", slog.Any("error", err))
		}
	})

	if err := workers.SeedRpsGameExpiryJob(ctx, app.JobManager()); err != nil {
		app.Logger().ErrorContext(ctx, "failed to seed rps expiry job", slog.Any("error", err))
	}
	if err := workers.SeedRpsRematchExpiryJob(ctx, app.JobManager()); err != nil {
		app.Logger().ErrorContext(ctx, "failed to seed rps rematch expiry job", slog.Any("error", err))
	}
	if err := workers.SeedRpsExpiryWarningJob(ctx, app.JobManager()); err != nil {
		app.Logger().ErrorContext(ctx, "failed to seed rps expiry warning job", slog.Any("error", err))
	}

	run("task notification scheduler", func() {
		app.Logger().Info("Starting task notification scheduler")
		scheduler := services.NewTaskNotificationScheduler(app.Adapter().Task(), app.JobService())
		scheduler.Run(ctx)
	})
}

func NewApp(config *conf.EnvConfig) *BaseApp {
	app := new(BaseApp)

	db, err := database.CreateNewQueriesContext(context.Background(), config.Db.GetDatabaseURL())
	if err != nil {
		slog.Error("failed to connect to database", slog.Any("error", err))
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}
	adapter := stores.NewStorageAdapter(db)

	payment := services.NewPaymentClient(config.StripeConfig)

	mailer := mailer.NewSmtpMailer(config.SmtpConfig)

	logger := logger.GetDefaultLogger()

	fs, err := filesystem.NewFileSystem(context.Background(), config.StorageConfig)
	if err != nil {
		slog.Error("failed to create filesystem", slog.Any("error", err))
		panic(fmt.Sprintf("failed to create filesystem: %v", err))
	}

	app.db = db
	app.adapter = adapter
	app.logger = logger
	app.cfg = config
	app.paymentClient = payment
	app.mailer = mailer
	app.fs = fs
	assembler := NewAssembler()
	assembler.AssembleApp(app)
	return app
}

func NewTestBaseApp(config *conf.EnvConfig, db database.Dbx) *BaseApp {
	app := new(BaseApp)
	adapter := stores.NewDbAdapterDecorators(db)

	payment := services.NewMockPaymentClient()

	mailer := mailer.NewTestMailer()

	logger := logger.GetDefaultLogger()

	app.logger = logger
	app.db = db
	app.adapter = adapter
	app.cfg = config
	app.paymentClient = payment
	app.mailer = mailer
	app.fs = filesystem.NewMockFileSystem(config.StorageConfig)
	assembler := NewAssembler()
	assembler.AssembleApp(app)
	return app
}
