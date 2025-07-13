package core

import (
	"context"
	"log/slog"

	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/di"
	"github.com/tkahng/playground/internal/tools/filesystem"
	"github.com/tkahng/playground/internal/tools/logger"
	"github.com/tkahng/playground/internal/tools/mailer"
	"github.com/tkahng/playground/internal/tools/sse"
)

func NewAppDecorator(ctx context.Context, cfg conf.EnvConfig, pool database.Dbx) *BaseAppDecorator {
	settings := cfg.ToSettings()

	fs := filesystem.NewMockFileSystem(cfg.StorageConfig)
	adapter := stores.NewDbAdapterDecorators(pool)

	l := logger.GetDefaultLogger()
	mail := &mailer.LogMailer{}
	mailServiece := services.NewOtpMailService(
		settings,
		mail,
		adapter,
	)
	jobManager := jobs.NewDbJobManagerDecorator(pool)
	jobService := services.NewJobServiceDecorator(jobManager)
	rbacService := services.NewRBACService(adapter)
	taskService := services.NewTaskService(adapter, jobService)
	paymentClient := services.NewTestPaymentClient()
	paymentService := services.NewPaymentService(
		paymentClient,
		adapter,
	)

	jobService.RegisterWorkers(mailServiece, paymentService, nil)
	authService := services.NewAuthServiceDecorator(
		settings,
		adapter,
		jobService,
	)
	checker := services.NewConstraintCheckerService(
		adapter,
	)

	teamService := services.NewTeamService(adapter)
	invitation := services.NewInvitationService(
		adapter,
		*settings,
		jobService,
	)
	app := newApp(
		fs,
		pool,
		settings,
		l,
		cfg,
		authService,
		paymentService,
		checker, // pass as ConstraintChecker
		rbacService,
		taskService,
		teamService,
		adapter,
		invitation,
		jobManager,
		jobService,
	)
	return &BaseAppDecorator{app: app}
}

type BaseAppDecorator struct {
	app                       *BaseApp
	AuthFunc                  func() services.AuthService
	CfgFunc                   func() *conf.EnvConfig
	CheckerFunc               func() services.ConstraintChecker
	DbFunc                    func() database.Dbx
	FsFunc                    func() *filesystem.S3FileSystem
	MailerFunc                func() mailer.Mailer
	PaymentFunc               func() services.PaymentService
	RbacFunc                  func() services.RBACService
	TeamFunc                  func() services.TeamService
	TaskFunc                  func() services.TaskService
	AdapterFunc               func() stores.StorageAdapterInterface
	TeamInvitationFunc        func() services.TeamInvitationService
	JobManagerFunc            func() jobs.JobManager
	JobServiceFunc            func() services.JobService
	LifecycleFunc             func() Lifecycle
	LoggerFunc                func() *slog.Logger
	BootstrapFunc             func() error
	SseManagerFunc            func() sse.Manager
	NotificationPublisherFunc func() services.Notifier
	ContainerFunc             func() di.Container
}

// Container implements App.
func (b *BaseAppDecorator) Container() di.Container {
	if b.ContainerFunc != nil {
		return b.ContainerFunc()
	}
	return b.app.Container()
}

// NotificationPublisher implements App.
func (b *BaseAppDecorator) NotificationPublisher() services.Notifier {
	if b.NotificationPublisherFunc != nil {
		return b.NotificationPublisherFunc()
	}

	return b.app.NotificationPublisher()
}

// SseManager implements App.
func (b *BaseAppDecorator) SseManager() sse.Manager {
	if b.SseManagerFunc != nil {
		return b.SseManagerFunc()
	}
	return b.app.SseManager()
}

// Logger implements App.
func (b *BaseAppDecorator) Logger() *slog.Logger {
	if b.LoggerFunc != nil {
		return b.LoggerFunc()
	}
	return b.app.Logger()
}

// BootStrap implements App.
func (b *BaseAppDecorator) Bootstrap() error {
	if b.BootstrapFunc != nil {
		return b.BootstrapFunc()
	}
	return b.app.Bootstrap()
}

// RegisterBaseHooks implements App.
// Lifecycle implements App.
func (b *BaseAppDecorator) Lifecycle() Lifecycle {
	if b.LifecycleFunc != nil {
		return b.LifecycleFunc()
	}
	return b.app.Lifecycle()
}

// Notifier implements App.

// JobService implements App.
func (b *BaseAppDecorator) JobService() services.JobService {
	if b.JobServiceFunc != nil {
		return b.JobServiceFunc()
	}
	return b.app.JobService()
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

func (b *BaseAppDecorator) Config() *conf.EnvConfig {
	if b.CfgFunc != nil {
		return b.CfgFunc()
	}
	return b.app.Config()
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
