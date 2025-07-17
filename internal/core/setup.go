package core

import (
	"context"
	"fmt"

	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/events"
	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/di"
	"github.com/tkahng/playground/internal/tools/logger"
	"github.com/tkahng/playground/internal/tools/sse"
)

func (app *BaseApp) Bootstrap() error {
	event := &BootstrapEvent{}
	container := di.New()
	event.App = app
	event.Container = container
	err := app.Lifecycle().OnBootstrap().Trigger(event, func(e *BootstrapEvent) error {
		e.App.InitializePrimitives()
		e.App.SetDb()
		e.App.SetBasicServices()
		e.App.SetIntegrationServices()
		e.App.RegisterWorkers()
		return nil
	})
	return err
}
func (app *BaseApp) InitializePrimitives() {
	opts := conf.AppConfigGetter()
	settings := opts.ToSettings()
	app.cfg = &opts
	app.settings = settings
	app.logger = logger.GetDefaultLogger()
}

func (app *BaseApp) SetDb() {

	queries := database.CreateQueries(app.cfg.Db.DatabaseUrl)

	if err := queries.Pool().Ping(context.Background()); err != nil {
		panic(fmt.Errorf("failed to ping db: %w", err))
	}

	app.db = queries

	adapter := stores.NewDbAdapterDecorators(app.db)
	app.adapter = adapter
}

func (app *BaseApp) SetBasicServices() {
	logger := app.Logger()
	cfg := app.Config()
	adapter := app.Adapter()
	dbx := app.Db()

	app.rbac = services.NewRBACService(adapter)
	app.team = services.NewTeamService(adapter)
	app.checker = services.NewConstraintCheckerService(adapter)

	app.eventManager = events.NewEventManager(logger)
	app.sseManager = sse.NewManager(logger)

	app.mailService = services.NewOtpMailService(
		cfg,
		adapter,
	)

	app.jobManager = jobs.NewDbJobManager(dbx)
	app.jobService = services.NewJobService(app.jobManager)
	app.notifierPublisher = services.NewDbNotificationPublisher(
		app.sseManager,
		app.team,
		adapter,
	)
}
func (app *BaseApp) SetIntegrationServices() {
	adapter := app.Adapter()
	cfg := app.Config()
	jobService := app.JobService()
	app.mailService = services.NewOtpMailService(
		cfg,
		adapter,
	)

	client := services.NewPaymentClient(cfg.StripeConfig)
	app.payment = services.NewPaymentService(client, adapter)
	app.teamInvitation = services.NewInvitationService(adapter, *cfg, jobService)
}

func (app *BaseApp) RegisterWorkers() {
	app.JobService().RegisterWorkers(app.mailService, app.Payment(), app.NotificationPublisher())
}
