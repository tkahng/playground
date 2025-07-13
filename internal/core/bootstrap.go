package core

import (
	"context"
	"errors"
	"log/slog"

	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/jobs"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/di"
	"github.com/tkahng/authgo/internal/tools/logger"
	"github.com/tkahng/authgo/internal/tools/mailer"
	"github.com/tkahng/authgo/internal/tools/payment"
	"github.com/tkahng/authgo/internal/tools/sse"
)

type diContextKey string

const (
	diContextKeyLogger            = "logger"
	diContextKeyConfig            = "config"
	diContextKeySettings          = "settings"
	diContextKeyDB                = "db"
	diContextKeyStoreAdapter      = "store_adapter"
	diContextKeyMail              = "mail"
	diContextKeyRbacService       = "rbac_service"
	diContextKeyJobManager        = "job_manager"
	diContextKeyJobService        = "job_service"
	diContextKeyTaskService       = "task_service"
	diContextKeyPaymentClient     = "payment_client"
	diContextKeyPaymentService    = "payment_service"
	diContextKeySseManager        = "sse_manager"
	diContextKeyTeamService       = "team_service"
	diContextKeyNotifierPublisher = "notifier_publisher"
	diContextKeyAuthService       = "auth_service"
	diContextKeyChecker           = "constraint_checker"
	diContextKeyInvitationService = "team_invitation_service"
)

func (s diContextKey) String() string {
	return string(s)
}

func (app *BaseApp) Bootstrap() error {
	event := &BootstrapEvent{}
	event.App = app

	err := app.Lifecycle().OnBootstrap().Trigger(event, func(e *BootstrapEvent) error {
		if err := app.ResetBootstrapState(); err != nil {
			return err
		}
		container := di.New()
		if err := register(container); err != nil {
			return err
		}
		app.UseContainer(container)
		return nil
	})
	return err
}

func register(container di.Container) error {
	// add logger
	container.AddSingleton(diContextKeyLogger, func(c di.Container) (any, error) {
		return logger.GetDefaultLogger(), nil
	})
	// add config
	container.AddSingleton(diContextKeyConfig, func(c di.Container) (any, error) {
		opts := conf.AppConfigGetter()
		return &opts, nil
	})
	//  add settings
	container.AddSingleton(diContextKeySettings, func(c di.Container) (any, error) {
		opts, err := getConfig(c)
		if err != nil {
			return nil, err
		}
		return opts.ToSettings(), nil
	})
	// add db
	container.AddSingleton(diContextKeyDB, func(c di.Container) (any, error) {
		opts, err := getConfig(c)
		if err != nil {
			return nil, err
		}
		queries := database.CreateQueries(opts.Db.DatabaseUrl)

		if err := queries.Pool().Ping(context.Background()); err != nil {
			return nil, err
		}
		return queries, nil
	})
	// 	add adapter
	container.AddSingleton(diContextKeyStoreAdapter, func(c di.Container) (any, error) {
		dbx, err := getDbx(c)
		if err != nil {
			return nil, err
		}
		adapter := stores.NewStorageAdapter(dbx)
		return adapter, nil
	})
	// add mail service
	container.AddSingleton(diContextKeyMail, func(c di.Container) (any, error) {
		cfg, err := getConfig(c)
		if err != nil {
			return nil, err
		}
		adapter, err := getAdapter(c)
		if err != nil {
			return nil, err
		}
		var m mailer.Mailer
		if cfg.ResendApiKey != "" {
			m = mailer.NewResendMailer(cfg.ResendConfig)
		} else {
			m = &mailer.LogMailer{}
		}
		mailServiece := services.NewOtpMailService(
			cfg.ToSettings(),
			m,
			adapter,
		)
		return mailServiece, nil
	})
	// rbac service
	container.AddSingleton(diContextKeyRbacService, func(c di.Container) (any, error) {
		adapter, err := getAdapter(c)
		if err != nil {
			return nil, err
		}
		rbac := services.NewRBACService(adapter)
		return rbac, nil
	})

	// add job manager
	container.AddSingleton(diContextKeyJobManager, func(c di.Container) (any, error) {
		dbx, err := getDbx(c)
		if err != nil {
			return nil, err
		}
		jobManager := jobs.NewDbJobManager(dbx)
		return jobManager, nil
	})
	// add job service
	container.AddSingleton(diContextKeyJobService, func(c di.Container) (any, error) {
		jobManager, ok := c.Get(diContextKeyJobManager).(jobs.JobManager)
		if !ok {
			return nil, errors.New("failed to get job manager")
		}
		jobService := services.NewJobService(jobManager)
		return jobService, nil
	})
	// payment client
	container.AddSingleton(diContextKeyPaymentClient, func(c di.Container) (any, error) {
		config, err := getConfig(c)
		if err != nil {
			return nil, err
		}
		paymentClient := payment.NewPaymentClient(config.StripeConfig)
		return paymentClient, nil
	})
	// task service
	container.AddSingleton(diContextKeyTaskService, func(c di.Container) (any, error) {
		adapter, err := getAdapter(c)
		if err != nil {
			return nil, err
		}
		jobService, ok := c.Get(diContextKeyJobService).(services.JobService)
		if !ok {
			return nil, errors.New("failed to get job manager")
		}
		taskService := services.NewTaskService(adapter, jobService)
		return taskService, nil
	})
	// payment service
	container.AddSingleton(diContextKeyPaymentService, func(c di.Container) (any, error) {
		paymentClient, ok := c.Get(diContextKeyPaymentClient).(services.PaymentClient)
		if !ok {
			return nil, errors.New("failed to get payment client")
		}
		adapter, err := getAdapter(c)
		if err != nil {
			return nil, err
		}
		paymentService := services.NewPaymentService(paymentClient, adapter)
		return paymentService, nil
	})
	// sse
	container.AddSingleton(diContextKeySseManager, func(c di.Container) (any, error) {
		logger, err := getLogger(c)
		if err != nil {
			return nil, err
		}
		sseManager := sse.NewManager(logger)
		return sseManager, nil
	})
	// team
	container.AddSingleton(diContextKeyTeamService, func(c di.Container) (any, error) {
		adapter, err := getAdapter(c)
		if err != nil {
			return nil, err
		}
		teamService := services.NewTeamService(adapter)
		return teamService, nil
	})

	container.AddSingleton(diContextKeyNotifierPublisher, func(c di.Container) (any, error) {
		sseManager, ok := c.Get(diContextKeySseManager).(sse.Manager)
		if !ok {
			return nil, errors.New("failed to get sse manager")
		}
		teamService, ok := c.Get(diContextKeyTeamService).(services.TeamService)
		if !ok {
			return nil, errors.New("failed to get team service")
		}
		adapter, err := getAdapter(c)
		if err != nil {
			return nil, err
		}
		notifierPublisher := services.NewDbNotificationPublisher(
			sseManager,
			teamService,
			adapter,
		)
		return notifierPublisher, nil
	})

	if err := registerJobWorkers(container); err != nil {
		return err
	}

	// register auth service
	container.AddSingleton(diContextKeyAuthService, func(c di.Container) (any, error) {
		adapter, err := getAdapter(c)
		if err != nil {
			return nil, err
		}
		settings, err := getSettings(c)
		if err != nil {
			return nil, err
		}
		jobService, ok := c.Get(diContextKeyJobService).(services.JobService)
		if !ok {
			return nil, errors.New("failed to get job manager")
		}
		authService := services.NewAuthService(settings, jobService, adapter)
		return authService, nil
	})

	// checker
	container.AddSingleton(diContextKeyChecker, func(c di.Container) (any, error) {
		adapter, err := getAdapter(c)
		if err != nil {
			return nil, err
		}
		checker := services.NewConstraintCheckerService(adapter)
		return checker, nil
	})

	// invitation
	container.AddSingleton(diContextKeyInvitationService, func(c di.Container) (any, error) {
		adapter, err := getAdapter(c)
		if err != nil {
			return nil, err
		}
		settings, err := getSettings(c)
		if err != nil {
			return nil, err
		}
		jobService, ok := c.Get(diContextKeyJobService).(services.JobService)
		if !ok {
			return nil, errors.New("failed to get job service")
		}
		invitationService := services.NewInvitationService(adapter, *settings, jobService)
		return invitationService, nil
	})

	return nil
}

func registerJobWorkers(c di.Container) error {
	jobService, ok := c.Get(diContextKeyJobService).(services.JobService)
	if !ok {
		return errors.New("failed to get job manager")
	}
	mailServiece, ok := c.Get(diContextKeyMail).(services.OtpMailService)
	if !ok {
		return errors.New("failed to get mail service")
	}
	paymentService, ok := c.Get(diContextKeyPaymentService).(services.PaymentService)
	if !ok {
		return errors.New("failed to get payment service")
	}
	notifier, ok := c.Get(diContextKeyNotifierPublisher).(services.Notifier)
	if !ok {
		return errors.New("failed to get notifier publisher")
	}
	jobService.RegisterWorkers(mailServiece, paymentService, notifier)
	return nil
}

func (app *BaseApp) ResetBootstrapState() error {
	app.container = nil
	return nil
}
func (app *BaseApp) UseContainer(c di.Container) {
	app.logger = c.Get(diContextKeyLogger).(*slog.Logger)
	app.cfg = c.Get(diContextKeyConfig).(*conf.EnvConfig)
	app.settings = c.Get(diContextKeySettings).(*conf.AppOptions)
	app.db = c.Get(diContextKeyDB).(database.Dbx)
	app.adapter = c.Get(diContextKeyStoreAdapter).(stores.StorageAdapterInterface)
	app.mailService = c.Get(diContextKeyMail).(services.OtpMailService)
	app.rbac = c.Get(diContextKeyRbacService).(services.RBACService)
	app.jobManager = c.Get(diContextKeyJobManager).(jobs.JobManager)
	app.jobService = c.Get(diContextKeyJobService).(services.JobService)
	app.task = c.Get(diContextKeyTaskService).(services.TaskService)
	app.payment = c.Get(diContextKeyPaymentService).(services.PaymentService)
	app.sseManager = c.Get(diContextKeySseManager).(sse.Manager)
	app.team = c.Get(diContextKeyTeamService).(services.TeamService)
	app.notifierPublisher = c.Get(diContextKeyNotifierPublisher).(services.Notifier)
	app.auth = c.Get(diContextKeyAuthService).(services.AuthService)
	app.checker = c.Get(diContextKeyChecker).(services.ConstraintChecker)
	app.teamInvitation = c.Get(diContextKeyInvitationService).(services.TeamInvitationService)
	app.container = c
}
