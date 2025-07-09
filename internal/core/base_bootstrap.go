package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/jobs"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/logger"
	"github.com/tkahng/authgo/internal/tools/mailer"
	"github.com/tkahng/authgo/internal/tools/payment"
)

func (app *BaseApp) Bootstrap() error {
	event := &BootstrapEvent{}
	event.App = app

	err := app.Lifecycle().OnBootstrap().Trigger(event, func(e *BootstrapEvent) error {
		if err := app.ResetBootstrapState(); err != nil {
			return err
		}

		if err := app.Config(); err == nil {
			panic("config is nil")
		}
		if err := app.initLogger(); err != nil {
			panic(fmt.Errorf("error initializing logger: %w", err))
		}
		if err := app.initStore(); err != nil {
			panic(fmt.Errorf("error initializing store: %w", err))
		}
		if err := app.initMail(); err != nil {
			panic(fmt.Errorf("error initializing mailer: %w", err))
		}
		if err := app.initJobs(); err != nil {
			panic(fmt.Errorf("error initializing jobs: %w", err))
		}
		if err := app.initPayment(); err != nil {
			panic(fmt.Errorf("error initializing payment: %w", err))
		}

		if err := app.initNotifier(); err != nil {
			panic(fmt.Errorf("error initializing notifier: %w", err))
		}

		if err := app.initAuth(); err != nil {
			panic(fmt.Errorf("error initializing auth: %w", err))
		}
		if err := app.initTeams(); err != nil {
			panic(fmt.Errorf("error initializing teams: %w", err))
		}
		if err := app.initTasks(); err != nil {
			panic(fmt.Errorf("error initializing tasks: %w", err))
		}
		if err := app.initWorkers(); err != nil {
			panic(fmt.Errorf("error initializing workers: %w", err))
		}
		return nil
	})
	return err
}

func (app *BaseApp) ResetBootstrapState() error {
	return nil
}
func (app *BaseApp) initStore() error {

	queries := database.CreateQueries(app.cfg.Db.DatabaseUrl)

	if err := queries.Pool().Ping(context.Background()); err != nil {
		return err
	}

	app.db = queries

	app.adapter = stores.NewStorageAdapter(app.db)
	return nil
}

func (app *BaseApp) initMail() error {
	var funcs []func() mailer.Mailer = []func() mailer.Mailer{
		func() mailer.Mailer {
			if app.cfg.ResendApiKey != "" {
				return mailer.NewResendMailer(app.cfg.ResendConfig)
			}
			return &mailer.LogMailer{}
		},
	}
	funcs = append(funcs, func() mailer.Mailer {
		return &mailer.LogMailer{}
	})
	for _, fn := range funcs {
		mail := fn()
		if mail != nil {
			app.mail = mail
			break
		}
	}
	if app.mail == nil {
		app.mail = &mailer.LogMailer{}
	}
	mailServiece := services.NewOtpMailService(
		app.settings,
		app.mail,
		app.adapter,
	)
	app.mailService = mailServiece
	return nil
}

func (app *BaseApp) initLogger() error {
	app.logger = logger.GetDefaultLogger()
	return nil
}

func (app *BaseApp) initJobs() error {
	jobManager := jobs.NewDbJobManager(app.db)
	jobService := services.NewJobService(jobManager)

	app.jobManager = jobManager
	app.jobService = jobService
	return nil
}

func (app *BaseApp) initWorkers() error {
	if app.mailService == nil {
		return errors.New("mail service not initialized")
	}
	if app.payment == nil {
		return errors.New("payment service not initialized")
	}
	app.jobService.RegisterWorkers(app.mailService, app.payment)
	return nil
}
func (app *BaseApp) initAuth() error {
	if app.jobService == nil {
		return errors.New("job manager not initialized")
	}
	if app.adapter == nil {
		return errors.New("adapter not initialized")
	}
	authService := services.NewAuthService(
		app.settings,
		app.jobService,
		app.adapter,
	)
	app.auth = authService

	rbac := services.NewRBACService(app.adapter)
	app.rbac = rbac

	constraint := services.NewConstraintCheckerService(
		app.adapter,
	)

	app.checker = constraint
	return nil
}
func (app *BaseApp) initNotifier() error {
	return nil
}

func (app *BaseApp) initTeams() error {
	teamService := services.NewTeamService(app.adapter)
	teamInvitationService := services.NewInvitationService(
		app.adapter,
		*app.settings,
		app.jobService,
	)
	app.team = teamService
	app.teamInvitation = teamInvitationService
	return nil
}

func (app *BaseApp) initTasks() error {
	tasksService := services.NewTaskService(app.adapter)
	app.task = tasksService
	return nil
}
func (app *BaseApp) initPayment() error {
	paymentClient := payment.NewPaymentClient(app.cfg.StripeConfig)
	paymentService := services.NewPaymentService(
		paymentClient,
		app.adapter,
	)
	app.payment = paymentService
	return nil
}
