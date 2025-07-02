package core

import (
	"context"

	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/jobs"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/filesystem"
	"github.com/tkahng/authgo/internal/tools/logger"
	"github.com/tkahng/authgo/internal/tools/mailer"
)

func NewAppDecorator(ctx context.Context, cfg conf.EnvConfig, pool database.Dbx) *BaseAppDecorator {
	settings := cfg.ToSettings()

	fs := filesystem.NewMockFileSystem(cfg.StorageConfig)

	l := logger.GetDefaultLogger()
	jobManager := jobs.NewDbJobManagerDecorator(pool)
	adapter := stores.NewAdapterDecorators()
	adapter.Delegate = stores.NewStorageAdapter(pool)
	mail := &mailer.LogMailer{}
	authMailService := services.NewMailService(mail)
	rbacService := services.NewRBACService(adapter)
	taskService := services.NewTaskService(adapter)
	paymentClient := services.NewTestPaymentClient()
	paymentService := services.NewPaymentService(
		paymentClient,
		adapter,
	)

	tokenService := services.NewJwtServiceDecorator()
	passwordService := services.NewPasswordService()
	routine := services.NewRoutineServiceDecorator()

	authService := services.NewAuthServiceDecorator(
		settings,
		authMailService,
		tokenService,
		passwordService,
		routine,
		jobManager,
		adapter,
	)
	checker := services.NewConstraintCheckerService(
		adapter,
	)

	teamService := services.NewTeamService(adapter)
	invitation := services.NewInvitationService(
		adapter,
		authMailService,
		*settings,
		routine,
	)
	app := newApp(
		fs,
		pool,
		settings,
		l,
		cfg,
		mail,
		authService,
		paymentService,
		checker, // pass as ConstraintChecker
		rbacService,
		taskService,
		teamService,
		adapter,
		invitation,
	)
	return &BaseAppDecorator{app: app}
}

type BaseAppDecorator struct {
	app                *BaseApp
	AuthFunc           func() services.AuthService
	CfgFunc            func() *conf.EnvConfig
	CheckerFunc        func() services.ConstraintChecker
	DbFunc             func() database.Dbx
	FsFunc             func() *filesystem.S3FileSystem
	MailerFunc         func() mailer.Mailer
	PaymentFunc        func() services.PaymentService
	RbacFunc           func() services.RBACService
	TeamFunc           func() services.TeamService
	TaskFunc           func() services.TaskService
	AdapterFunc        func() stores.StorageAdapterInterface
	TeamInvitationFunc func() services.TeamInvitationService
	JobManagerFunc     func() jobs.JobManager
}

// JobManager implements App.
func (b *BaseAppDecorator) JobManager() jobs.JobManager {
	if b.JobManagerFunc != nil {
		return b.JobManagerFunc()
	}
	return b.app.JobManager()
}

// TeamInvitation implements App.
func (b *BaseAppDecorator) TeamInvitation() services.TeamInvitationService {
	if b.TeamInvitationFunc != nil {
		return b.TeamInvitationFunc()
	}
	return b.app.TeamInvitation()
}

var _ App = (*BaseAppDecorator)(nil)

func (b *BaseAppDecorator) Adapter() stores.StorageAdapterInterface {
	if b.AdapterFunc != nil {
		return b.AdapterFunc()
	}
	return b.app.Adapter()
}
func (b *BaseAppDecorator) Auth() services.AuthService {
	if b.AuthFunc != nil {
		return b.AuthFunc()
	}
	return b.app.Auth()
}

func (b *BaseAppDecorator) Cfg() *conf.EnvConfig {
	if b.CfgFunc != nil {
		return b.CfgFunc()
	}
	return b.app.Cfg()
}

func (b *BaseAppDecorator) Checker() services.ConstraintChecker {
	if b.CheckerFunc != nil {
		return b.CheckerFunc()
	}
	return b.app.Checker()
}

func (b *BaseAppDecorator) Db() database.Dbx {
	if b.DbFunc != nil {
		return b.DbFunc()
	}
	return b.app.Db()
}

func (b *BaseAppDecorator) Fs() filesystem.FileSystem {
	if b.FsFunc != nil {
		return b.FsFunc()
	}
	return b.app.Fs()
}
func (b *BaseAppDecorator) Payment() services.PaymentService {
	if b.PaymentFunc != nil {
		return b.PaymentFunc()
	}
	return b.app.Payment()
}

func (b *BaseAppDecorator) Rbac() services.RBACService {
	if b.RbacFunc != nil {
		return b.RbacFunc()
	}
	return b.app.Rbac()
}

func (b *BaseAppDecorator) Settings() *conf.AppOptions {
	return b.app.Settings()
}

func (b *BaseAppDecorator) Task() services.TaskService {
	if b.TaskFunc != nil {
		return b.TaskFunc()
	}
	return b.app.Task()
}

func (b *BaseAppDecorator) Team() services.TeamService {
	if b.TeamFunc != nil {
		return b.TeamFunc()
	}
	return b.app.Team()
}
