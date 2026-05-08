package core

import (
	"context"
	"log/slog"

	"github.com/tkahng/playground/internal/auth"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/events"
	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/token"
	"github.com/tkahng/playground/internal/tools/filesystem"
	"github.com/tkahng/playground/internal/tools/mailer"
	"github.com/tkahng/playground/internal/tools/sse"
)

type App interface {
	Close()

	//  settings -------------------------------------------------------------------------------------
	Config() *conf.EnvConfig

	// store -------------------------------------------------------------------------------------
	Db() database.Dbx
	Adapter() stores.StorageAdapterInterface

	Logger() *slog.Logger

	// gaming
	RpsGame() services.RpsGameService

	// ledger / betting
	Ledger() services.LedgerService
	Betting() services.BettingService
	// jobs -------------------------------------------------------------------------------------

	JobManager() jobs.JobManager

	JobService() services.JobService
	// fs -------------------------------------------------------------------------------------

	Fs() filesystem.FileSystem
	//
	Mailer() mailer.Mailer
	MailService() services.OtpMailService

	Rbac() services.RBACService

	PaymentClient() services.PaymentClient
	Payment() services.PaymentService
	Hash() services.HashService

	Auth() auth.AuthService

	Token() token.TokenService

	Team() services.TeamService

	TeamInvitation() services.TeamInvitationService

	Checker() services.ConstraintChecker

	Task() services.TaskService

	AiUsage() services.AiUsageService

	NotificationPublisher() services.Notifier

	SseManager() sse.Manager

	EventManager() events.EventManager

	RunBackgroundProcesses(ctx context.Context)
}
