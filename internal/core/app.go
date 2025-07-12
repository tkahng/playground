package core

import (
	"log/slog"

	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/jobs"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/filesystem"
	"github.com/tkahng/authgo/internal/tools/sse"
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
}
