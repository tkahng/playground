package core

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/tkahng/playground/internal/auth"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/events"
	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/token"
	"github.com/tkahng/playground/internal/tools/logger"
	"github.com/tkahng/playground/internal/tools/sse"
	"github.com/tkahng/playground/internal/userreaction"
)

type Assembler struct {
	provider Provider
}

func NewInitiatorBase(provider Provider) *Assembler {
	return &Assembler{
		provider: provider,
	}
}

func (a *Assembler) AssembleNewApp() *BaseApp {
	app := new(BaseApp)
	a.Configure(app)
	a.SetDatasource(app)
	a.SetBasicServices(app)
	a.SetIntegrationServices(app)
	a.RegisterWorkers(app)
	a.AddEventHandlers(app)
	return app
}

// AddEventHandlers implements Initiator.
func (a *Assembler) AddEventHandlers(app *BaseApp) {
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

// Configure implements Initiator.
func (a *Assembler) Configure(app *BaseApp) {
	app.cfg = a.provider.Config()
	app.db = a.provider.Db()
	app.mailer = a.provider.Mailer()
	app.logger = logger.GetDefaultLogger()
}

// RegisterWorkers implements Initiator.
func (a *Assembler) RegisterWorkers(app *BaseApp) {
	app.JobService().RegisterWorkers(app.mailService, app.Payment(), app.NotificationPublisher())
}

// SetBasicServices implements Initiator.
func (a *Assembler) SetBasicServices(app *BaseApp) {
	cfg := app.Config()
	logger := app.Logger()

	adapter := app.Adapter()
	dbx := app.Db()

	app.password = services.NewPasswordService()

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

// SetDatasource implements Initiator.
func (a *Assembler) SetDatasource(app *BaseApp) {
	if app.db == nil {
		queries := database.CreateSingletonQueriesContext(context.Background(), app.cfg.Db.GetDatabaseUrl())
		if err := queries.Pool().Ping(context.Background()); err != nil {
			panic(fmt.Errorf("failed to ping db: %w", err))
		}
		app.db = queries
	}
	if app.adapter == nil {
		adapter := stores.NewStorageAdapter(app.db)
		app.adapter = adapter
	}
}

// SetIntegrationServices implements Initiator.
func (a *Assembler) SetIntegrationServices(app *BaseApp) {
	adapter := app.Adapter()
	cfg := app.Config()
	jobService := app.JobService()
	tokenService := app.Token()
	passwordService := app.Password()
	jwtService := app.Jwt()

	app.mailer = a.provider.Mailer()

	app.mailService = services.NewOtpMailService(
		cfg,
		adapter,
		app.mailer,
		tokenService,
		jwtService,
		passwordService,
	)

	client := a.provider.Payment()
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
