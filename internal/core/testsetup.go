package core

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/tkahng/playground/internal/auth"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/events"
	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/token"
	"github.com/tkahng/playground/internal/tools/logger"
	"github.com/tkahng/playground/internal/tools/mailer"
	"github.com/tkahng/playground/internal/tools/sse"
	"github.com/tkahng/playground/internal/userreaction"
)

func NewTestApp(ctx context.Context, cfg conf.EnvConfig, pool database.Dbx) *BaseApp {
	app := new(BaseApp)
	if err := TestingBootstrap(app, &cfg, pool); err != nil {
		panic(fmt.Errorf("failed to bootstrap app: %w", err))
	}
	return app
}

func TestingBootstrap(app *BaseApp, cfg *conf.EnvConfig, pool database.Dbx) error {
	app.cfg = cfg
	app.logger = logger.GetDefaultLogger()
	app.db = pool
	adapter := stores.NewStorageAdapter(app.db)
	app.adapter = adapter

	TestingSetBasicServices(app)
	TestingSetIntegrationServices(app)
	TestingRegisterWorkers(app)
	TestingAddEventHandlers(app)
	return nil
}

func TestingAddEventHandlers(app *BaseApp) {
	userReactionHandler := userreaction.NewUserReactionEventHandler(
		app.Logger(),
		app.Adapter().UserReaction(),
		app.SseManager(),
	)
	app.EventManager().AddHandlers(
		cqrs.NewEventHandler(
			"UserReactionCreated",
			userReactionHandler.OnUserReactionCreated,
		),
	)
}

func TestingInitializePrimitives(app *BaseApp) {
	opts := conf.ZeroEnvConfig()
	app.cfg = &opts
	app.logger = logger.GetDefaultLogger()
}

func TestingSetDb(app *BaseApp) {
	migrator := database.NewMigrator(&database.MigratorConfig{
		DatabaseUrl: app.cfg.Db.GetDatabaseUrl(),
	})
	app.migrator = migrator

	queries := database.CreateQueries(app.cfg.Db.GetDatabaseUrl())

	if err := queries.Pool().Ping(context.Background()); err != nil {
		panic(fmt.Errorf("failed to ping db: %w", err))
	}

	app.db = queries

	adapter := stores.NewStorageAdapter(app.db)
	app.adapter = adapter
}

func TestingSetBasicServices(app *BaseApp) {
	logger := app.Logger()
	adapter := app.Adapter()
	dbx := app.Db()
	cfg := app.Config()
	passWordService := services.NewPasswordService()
	app.password = passWordService
	jwtService := services.NewJwtService()
	app.jwt = jwtService
	app.rbac = services.NewRBACService(adapter)
	app.team = services.NewTeamService(adapter)
	app.checker = services.NewConstraintCheckerService(adapter)

	app.eventManager = events.NewEventManager(logger)
	app.sseManager = sse.NewManager(logger)

	app.jobManager = jobs.NewDbJobManager(dbx)
	app.jobService = services.NewJobService(app.jobManager)
	app.notifierPublisher = services.NewDbNotificationPublisher(
		app.sseManager,
		app.team,
		adapter,
	)
	app.task = services.NewTaskService(adapter, app.jobService)
	app.token = token.NewTokenService(cfg, adapter.Token())
}

func TestingSetIntegrationServices(app *BaseApp) {
	adapter := app.Adapter()
	cfg := app.Config()
	jobService := app.JobService()
	passwordService := app.Password()
	jwtService := app.Jwt()
	m := &mailer.TestMailer{
		Mailer: &mailer.LogMailer{},
		Wg:     nil,
	}
	app.mailer = m
	tokenService := app.Token()
	app.mailService = services.NewOtpMailService(
		cfg,
		adapter,
		m,
		tokenService,
		jwtService,
		passwordService,
	)

	client := services.NewTestPaymentClient()
	app.payment = services.NewPaymentService(client, adapter)
	app.teamInvitation = services.NewInvitationService(adapter, *cfg, jobService)
	app.auth = services.NewAuthService(
		cfg,
		jobService,
		adapter,
		tokenService,
		jwtService,
		passwordService,
	)
	auth2 := auth.NewAuthService(cfg, adapter, passwordService, jwtService, tokenService, jobService)
	app.auth2 = auth2
}

func TestingRegisterWorkers(app *BaseApp) {
	app.JobService().RegisterWorkers(app.mailService, app.Payment(), app.NotificationPublisher())
}
