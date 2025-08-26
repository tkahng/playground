package core

import (
	"log/slog"

	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/events"
	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/filesystem"
	"github.com/tkahng/playground/internal/tools/mailer"
	"github.com/tkahng/playground/internal/tools/sse"
)

type Container interface {
	// Start starts the container
	Config() *conf.EnvConfig
	SetConfig(config *conf.EnvConfig)

	// store -------------------------------------------------------------------------------------
	Db() database.Dbx
	SetDb(db database.Dbx)
	Adapter() stores.StorageAdapterInterface
	SetAdapter(adapter stores.StorageAdapterInterface)

	Logger() *slog.Logger
	SetLogger(logger *slog.Logger)

	// jobs -------------------------------------------------------------------------------------

	JobManager() jobs.JobManager
	SetJobManager(jobManager jobs.JobManager)

	JobService() services.JobService
	SetJobService(jobService services.JobService)
	// fs -------------------------------------------------------------------------------------

	Fs() filesystem.FileSystem
	SetFs(db filesystem.FileSystem)
	//
	Mailer() mailer.Mailer
	SetMailer(db mailer.Mailer)
	MailService() services.OtpMailService
	SetMailService(db services.OtpMailService)

	Rbac() services.RBACService
	SetRbac(db services.RBACService)

	Payment() services.PaymentService
	SetPayment(db services.PaymentService)

	Auth() services.AuthService
	SetAuth(db services.AuthService)

	Team() services.TeamService
	SetTeam(db services.TeamService)

	TeamInvitation() services.TeamInvitationService
	SetTeamInvitation(db services.TeamInvitationService)

	Checker() services.ConstraintChecker
	SetChecker(db services.ConstraintChecker)

	Task() services.TaskService
	SetTask(db services.TaskService)

	NotificationPublisher() services.Notifier
	SetNotificationPublisher(db services.Notifier)

	SseManager() sse.Manager
	SetSseManager(db sse.Manager)

	EventManager() events.EventManager
	SetEventManager(db events.EventManager)
}

var _ Container = (*container)(nil)

type container struct {
	mailer mailer.Mailer
	cfg    *conf.EnvConfig

	db      database.Dbx
	adapter stores.StorageAdapterInterface

	logger      *slog.Logger
	mailService services.OtpMailService

	jobManager jobs.JobManager
	jobService services.JobService

	payment services.PaymentService

	auth    services.AuthService
	rbac    services.RBACService
	checker services.ConstraintChecker

	task services.TaskService

	team           services.TeamService
	teamInvitation services.TeamInvitationService

	notifierPublisher services.Notifier

	fs filesystem.FileSystem

	sseManager sse.Manager

	eventManager events.EventManager
}

func NewContainer() Container {
	return &container{}
}

// Adapter implements Container.
func (c *container) Adapter() stores.StorageAdapterInterface {
	return c.adapter
}

// Auth implements Container.
func (c *container) Auth() services.AuthService {
	return c.auth
}

// Checker implements Container.
func (c *container) Checker() services.ConstraintChecker {
	return c.checker
}

// Config implements Container.
func (c *container) Config() *conf.EnvConfig {
	return c.cfg
}

// Db implements Container.
func (c *container) Db() database.Dbx {
	return c.db
}

// EventManager implements Container.
func (c *container) EventManager() events.EventManager {
	return c.eventManager
}

// Fs implements Container.
func (c *container) Fs() filesystem.FileSystem {
	return c.fs
}

// JobManager implements Container.
func (c *container) JobManager() jobs.JobManager {
	return c.jobManager
}

// JobService implements Container.
func (c *container) JobService() services.JobService {
	return c.jobService
}

// Logger implements Container.
func (c *container) Logger() *slog.Logger {
	return c.logger
}

// MailService implements Container.
func (c *container) MailService() services.OtpMailService {
	return c.mailService
}

// Mailer implements Container.
func (c *container) Mailer() mailer.Mailer {
	return c.mailer
}

// NotificationPublisher implements Container.
func (c *container) NotificationPublisher() services.Notifier {
	return c.notifierPublisher
}

// Payment implements Container.
func (c *container) Payment() services.PaymentService {
	return c.payment
}

// Rbac implements Container.
func (c *container) Rbac() services.RBACService {
	return c.rbac
}

// SetAdapter implements Container.
func (c *container) SetAdapter(adapter stores.StorageAdapterInterface) {
	c.adapter = adapter
}

// SetAuth implements Container.
func (c *container) SetAuth(db services.AuthService) {
	c.auth = db
}

// SetChecker implements Container.
func (c *container) SetChecker(db services.ConstraintChecker) {
	c.checker = db
}

// SetConfig implements Container.
func (c *container) SetConfig(config *conf.EnvConfig) {
	c.cfg = config
}

// SetDb implements Container.
func (c *container) SetDb(db database.Dbx) {
	c.db = db
}

// SetEventManager implements Container.
func (c *container) SetEventManager(db events.EventManager) {
	c.eventManager = db
}

// SetFs implements Container.
func (c *container) SetFs(db filesystem.FileSystem) {
	c.fs = db
}

// SetJobManager implements Container.
func (c *container) SetJobManager(db jobs.JobManager) {
	c.jobManager = db
}

// SetJobService implements Container.
func (c *container) SetJobService(db services.JobService) {
	c.jobService = db
}

// SetLogger implements Container.
func (c *container) SetLogger(logger *slog.Logger) {
	c.logger = logger
}

// SetMailService implements Container.
func (c *container) SetMailService(db services.OtpMailService) {
	c.mailService = db
}

// SetMailer implements Container.
func (c *container) SetMailer(db mailer.Mailer) {
	c.mailer = db
}

// SetNotificationPublisher implements Container.
func (c *container) SetNotificationPublisher(db services.Notifier) {
	c.notifierPublisher = db
}

// SetPayment implements Container.
func (c *container) SetPayment(db services.PaymentService) {
	c.payment = db
}

// SetRbac implements Container.
func (c *container) SetRbac(db services.RBACService) {
	c.rbac = db
}

// SetSseManager implements Container.
func (c *container) SetSseManager(db sse.Manager) {
	c.sseManager = db
}

// SetTask implements Container.
func (c *container) SetTask(db services.TaskService) {
	c.task = db
}

// SetTeam implements Container.
func (c *container) SetTeam(db services.TeamService) {
	c.team = db
}

// SetTeamInvitation implements Container.
func (c *container) SetTeamInvitation(db services.TeamInvitationService) {
	c.teamInvitation = db
}

// SseManager implements Container.
func (c *container) SseManager() sse.Manager {
	return c.sseManager
}

// Task implements Container.
func (c *container) Task() services.TaskService {
	return c.task
}

// Team implements Container.
func (c *container) Team() services.TeamService {
	return c.team
}

// TeamInvitation implements Container.
func (c *container) TeamInvitation() services.TeamInvitationService {
	return c.teamInvitation
}
