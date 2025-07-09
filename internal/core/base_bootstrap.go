package core

import (
	"context"

	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/jobs"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/stores"
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
		if err := app.Config(); err != nil {
			panic(err)
		}
		if err := app.initDb(); err != nil {
			panic(err)
		}
		if err := app.initAdapter(); err != nil {
			panic(err)
		}
		if err := app.initPayment(); err != nil {
			panic(err)
		}
		if err := app.initJobs(); err != nil {
			panic(err)
		}
		if err := app.initMail(); err != nil {
			panic(err)
		}
		if err := app.initNotifier(); err != nil {
			panic(err)
		}
		if err := app.initAuth(); err != nil {
			panic(err)
		}
		if err := app.initTeams(); err != nil {
			panic(err)
		}
		if err := app.initTasks(); err != nil {
			panic(err)
		}
		if err := app.initWorkers(); err != nil {
			panic(err)
		}
		return nil
	})
	return err
}

func (app *BaseApp) ResetBootstrapState() error {
	return nil
}
func (app *BaseApp) initDb() error {
	if app.db == nil {
		queries := database.CreateQueries(app.cfg.Db.DatabaseUrl)
		if err := queries.Pool().Ping(context.Background()); err != nil {
			return err
		}
		app.db = queries
	}
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
			return nil
		}
	}
	app.mail = &mailer.LogMailer{}

	return nil
}

func (app *BaseApp) initLogger() error {
	return nil
}
func (app *BaseApp) initAdapter() error {
	adapter := stores.NewStorageAdapter(app.db)
	app.adapter = adapter
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
	mailServiece := services.NewOtpMailService(
		app.settings,
		app.mail,
		app.adapter,
	)
	app.jobService.RegisterWorkers(mailServiece, app.Payment())
	return nil
}
func (app *BaseApp) initAuth() error {
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

func (app *BaseApp) initTasks() any {
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
