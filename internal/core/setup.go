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

func NewApp(cfg conf.EnvConfig) *BaseApp {
	app := new(BaseApp)
	if err := Bootstrap(app); err != nil {
		panic(fmt.Errorf("failed to bootstrap app: %w", err))
	}
	return app
}
func Bootstrap(app *BaseApp) error {
	InitializePrimitives(app)
	SetDb(app)
	SetBasicServices(app)
	SetIntegrationServices(app)
	RegisterWorkers(app)
	AddEventHandlers(app)
	return nil
}

func AddEventHandlers(app *BaseApp) {
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

func InitializePrimitives(app *BaseApp) {
	opts := conf.AppConfigGetter()
	app.cfg = &opts
	app.logger = logger.GetDefaultLogger()
}

func SetDb(app *BaseApp) {
	queries := database.CreateQueries(app.cfg.Db.DatabaseUrl)

	if err := queries.Pool().Ping(context.Background()); err != nil {
		panic(fmt.Errorf("failed to ping db: %w", err))
	}

	app.db = queries

	adapter := stores.NewStorageAdapter(app.db)
	app.adapter = adapter
}

func SetBasicServices(app *BaseApp) {
	logger := app.Logger()
	adapter := app.Adapter()
	dbx := app.Db()
	cfg := app.Config()
	passWordService := services.NewPasswordService()
	app.password = passWordService
	app.rbac = services.NewRBACService(adapter)
	app.team = services.NewTeamService(adapter)
	app.checker = services.NewConstraintCheckerService(adapter)
	app.jwt = services.NewJwtService()
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

func SetIntegrationServices(app *BaseApp) {
	adapter := app.Adapter()
	cfg := app.Config()
	jobService := app.JobService()
	tokenService := app.Token()
	passwordService := app.Password()
	jwtService := app.Jwt()

	m := mailer.NewResendMailer(cfg.ResendConfig)

	app.mailer = m
	app.mailService = services.NewOtpMailService(
		cfg,
		adapter,
		m,
		tokenService,
	)

	client := services.NewPaymentClient(cfg.StripeConfig)
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
	auth2 := auth.NewAuthService(adapter)
	app.auth2 = auth2
}

func RegisterWorkers(app *BaseApp) {
	app.JobService().RegisterWorkers(app.mailService, app.Payment(), app.NotificationPublisher())
}
