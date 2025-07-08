package core

import (
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/jobs"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/filesystem"
)

type App interface {
	Cfg() *conf.EnvConfig

	Settings() *conf.AppOptions

	Db() database.Dbx

	Fs() filesystem.FileSystem

	Rbac() services.RBACService

	Payment() services.PaymentService

	Auth() services.AuthService

	Team() services.TeamService

	TeamInvitation() services.TeamInvitationService

	Checker() services.ConstraintChecker

	Task() services.TaskService

	Adapter() stores.StorageAdapterInterface

	JobManager() jobs.JobManager

	JobService() services.JobService

	Notifier() services.NotifierService

	Lifecycle() Lifecycle
}
