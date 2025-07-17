package core

import (
	"log/slog"

	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/events"
	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/di"
	"github.com/tkahng/playground/internal/tools/filesystem"
	"github.com/tkahng/playground/internal/tools/sse"
)

type App interface {
	Bootstrap() error

	//  settings -------------------------------------------------------------------------------------
	Config() *conf.EnvConfig
	Settings() *conf.AppOptions

	// store -------------------------------------------------------------------------------------
	Db() database.Dbx
	Adapter() stores.StorageAdapterInterface

	// lifecycle
	Lifecycle() Lifecycle
	Logger() *slog.Logger

	// jobs -------------------------------------------------------------------------------------

	JobManager() jobs.JobManager

	JobService() services.JobService

	Fs() filesystem.FileSystem

	Rbac() services.RBACService

	Payment() services.PaymentService

	Auth() services.AuthService

	Team() services.TeamService

	TeamInvitation() services.TeamInvitationService

	Checker() services.ConstraintChecker

	Task() services.TaskService

	NotificationPublisher() services.Notifier

	SseManager() sse.Manager

	Container() di.Container

	EventManager() events.EventManager
}
