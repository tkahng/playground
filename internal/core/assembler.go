package core

import (
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/tkahng/playground/internal/auth"
	"github.com/tkahng/playground/internal/events"
	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/token"
	"github.com/tkahng/playground/internal/tools/sse"
	"github.com/tkahng/playground/internal/tools/ticket"
	"github.com/tkahng/playground/internal/userreaction"
)

type Assembler struct{}

func NewAssembler() *Assembler {
	return &Assembler{}
}

func (a *Assembler) AssembleApp(app *BaseApp) {
	a.configure(app)
	a.setBasicServices(app)
	a.setIntegrationServices(app)
	a.registerWorkers(app)
	a.addEventHandlers(app)
	a.validate(app)
}

func (a *Assembler) validate(app *BaseApp) {
	type check struct {
		name string
		val  any
	}
	checks := []check{
		{"fs", app.fs},
		{"jwt", app.jwt},
		{"hash", app.hash},
		{"encrypt", app.encrypt},
		{"auth", app.auth},
		{"rbac", app.rbac},
		{"checker", app.checker},
		{"task", app.task},
		{"aiUsage", app.aiUsage},
		{"token", app.token},
		{"team", app.team},
		{"teamInvitation", app.teamInvitation},
		{"notifierPublisher", app.notifierPublisher},
		{"sseManager", app.sseManager},
		{"sseTickets", app.sseTickets},
		{"eventManager", app.eventManager},
		{"jobManager", app.jobManager},
		{"jobService", app.jobService},
		{"payment", app.payment},
		{"paymentClient", app.paymentClient},
		{"mailService", app.mailService},
		{"rpsGame", app.rpsGame},
		{"ledger", app.ledger},
		{"betting", app.betting},
	}
	for _, c := range checks {
		if c.val == nil {
			panic(fmt.Sprintf("service %q not initialized at startup", c.name))
		}
	}
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
}

// registerWorkers implements Initiator.
func (a *Assembler) registerWorkers(app *BaseApp) {
	app.JobService().RegisterWorkers(
		app.mailService,
		app.Payment(),
		app.NotificationPublisher(),
		app.RpsGame(),
		app.Adapter(),
		app.SseManager(),
	)
}

// setBasicServices implements Initiator.
func (a *Assembler) setBasicServices(app *BaseApp) {
	cfg := app.Config()
	logger := app.Logger()

	adapter := app.Adapter()
	dbx := app.Db()
	app.ledger = services.NewDbLedgerService(adapter)
	app.betting = services.NewDbBettingService(adapter, app.ledger)
	app.rpsGame = services.NewDbRpsGameService(adapter, app.betting).
		WithHouseThinkDelay(2 * time.Second)
	app.hash = services.NewHashService()
	app.encrypt = services.NewCrypto(cfg.EncryptionKey)
	app.jwt = services.NewJwtService()

	app.rbac = services.NewRBACService(adapter)
	app.team = services.NewTeamService(adapter)
	app.checker = services.NewConstraintCheckerService(adapter)

	app.eventManager = events.NewEventManager(logger)
	app.sseManager = sse.NewManager(logger)
	app.sseTickets = ticket.New(60 * time.Second)

	app.jobManager = jobs.NewDbJobManager(dbx)
	app.jobService = services.NewJobService(app.jobManager)
	app.notifierPublisher = services.NewDbNotificationPublisher(
		app.sseManager,
		app.team,
		adapter,
	)
	app.task = services.NewTaskService(adapter, app.jobService)
	app.aiUsage = services.NewAiUsageService(adapter)
	app.token = token.NewTokenService(cfg, adapter.Token())
}

// setIntegrationServices implements Initiator.
func (a *Assembler) setIntegrationServices(app *BaseApp) {
	logger := app.Logger()
	adapter := app.Adapter()
	cfg := app.Config()
	jobService := app.JobService()
	tokenService := app.Token()
	hashService := app.Hash()
	enc := app.Encrypt()
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

	app.auth = auth.NewAuthService(cfg, logger, adapter, hashService, jwtService, tokenService, jobService, enc)
}
