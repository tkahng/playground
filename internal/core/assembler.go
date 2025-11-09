package core

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/tkahng/playground/internal/auth"
	"github.com/tkahng/playground/internal/events"
	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/token"
	"github.com/tkahng/playground/internal/tools/sse"
	"github.com/tkahng/playground/internal/userreaction"
)

type Assembler struct {
}

func NewAssembler() *Assembler {
	return &Assembler{}
}
func (a *Assembler) AssembleApp(app *BaseApp) {
	a.configure(app)
	a.setDatasource(app)
	a.setBasicServices(app)
	a.setIntegrationServices(app)
	a.registerWorkers(app)
	a.addEventHandlers(app)
}

// addEventHandlers implements Initiator.
func (a *Assembler) addEventHandlers(app *BaseApp) {
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

// configure implements Initiator.
func (a *Assembler) configure(app *BaseApp) {
	if app.cfg == nil {
		panic("config not initialized")
	}
	if app.db == nil {
		panic("db not initialized")
	}
	if app.mailer == nil {
		panic("mailer not initialized")
	}
	if app.logger == nil {
		panic("logger not initialized")
	}
	// if app.cfg == nil {
	// 	app.cfg = conf.AppConfigGetter()
	// }
	// if app.db == nil {
	// 	app.db = database.CreateSingletonQueriesContext(context.Background(), app.cfg.Db.GetDatabaseUrl())
	// }
	// if app.mailer == nil {
	// 	app.mailer = mailer.NewSmtpMailer(app.cfg.SmtpConfig)
	// }
	// if app.logger == nil {
	// 	app.logger = logger.GetDefaultLogger()
	// }
}

// registerWorkers implements Initiator.
func (a *Assembler) registerWorkers(app *BaseApp) {
	app.JobService().RegisterWorkers(app.mailService, app.Payment(), app.NotificationPublisher())
}

// setBasicServices implements Initiator.
func (a *Assembler) setBasicServices(app *BaseApp) {
	cfg := app.Config()
	logger := app.Logger()

	adapter := app.Adapter()
	dbx := app.Db()

	app.hash = services.NewHashService()

	app.jwt = services.NewJwtService()

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

// setDatasource implements Initiator.
func (a *Assembler) setDatasource(app *BaseApp) {
	// if app.db == nil {
	// 	queries := database.CreateSingletonQueriesContext(context.Background(), app.cfg.Db.GetDatabaseUrl())
	// 	if err := queries.Pool().Ping(context.Background()); err != nil {
	// 		panic(fmt.Errorf("failed to ping db: %w", err))
	// 	}
	// 	app.db = queries
	// }
	// if app.adapter == nil {
	// adapter := stores.NewStorageAdapter(app.db)
	// app.adapter = adapter
	// }
}

// setIntegrationServices implements Initiator.
func (a *Assembler) setIntegrationServices(app *BaseApp) {
	logger := app.Logger()
	adapter := app.Adapter()
	cfg := app.Config()
	jobService := app.JobService()
	tokenService := app.Token()
	hashService := app.Hash()
	jwtService := app.Jwt()

	app.mailService = services.NewOtpMailService(
		cfg,
		adapter,
		app.mailer,
		tokenService,
		jwtService,
		hashService,
	)

	client := app.paymentClient
	app.payment = services.NewPaymentService(client, adapter)
	app.teamInvitation = services.NewInvitationService(adapter, *cfg, jobService, app.payment)

	app.auth = auth.NewAuthService(cfg, logger, adapter, hashService, jwtService, tokenService, jobService)

}
