package core

import (
	"context"
	"log/slog"

	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/events"
	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/filesystem"
	"github.com/tkahng/playground/internal/tools/sse"
)

type App interface {
	AppContainer
	Bootstrap() error

	//  settings -------------------------------------------------------------------------------------
	Config() *conf.EnvConfig

	// store -------------------------------------------------------------------------------------
	Db() database.Dbx
	Adapter() stores.StorageAdapterInterface

	// lifecycle
	Lifecycle() Lifecycle
	Logger() *slog.Logger

	// jobs -------------------------------------------------------------------------------------

	JobManager() jobs.JobManager

	JobService() services.JobService
	// fs -------------------------------------------------------------------------------------

	Fs() filesystem.FileSystem
	//
	MailService() services.OtpMailService

	Rbac() services.RBACService

	Payment() services.PaymentService

	Auth() services.AuthService

	Team() services.TeamService

	TeamInvitation() services.TeamInvitationService

	Checker() services.ConstraintChecker

	Task() services.TaskService

	NotificationPublisher() services.Notifier

	SseManager() sse.Manager

	EventManager() events.EventManager

	RunBackgroundProcesses(ctx context.Context)
}

type AppContainer interface {
	InitializePrimitives()
	SetDb()
	SetBasicServices()
	SetIntegrationServices()
	RegisterWorkers()
}
