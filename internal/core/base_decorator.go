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

func NewDecorator(ctx context.Context, cfg conf.EnvConfig, pool database.Dbx) *BaseAppDecorator {
	settings := cfg.ToSettings()

	fs := filesystem.NewMockFileSystem(cfg.StorageConfig)

	l := logger.GetDefaultLogger()
	enqueuer := jobs.NewDBEnqueuer(pool)
	var mail mailer.Mailer = &mailer.LogMailer{}
	authMailService := services.NewMailService(mail)
	adapter := stores.NewStorageAdapter(pool)
	userStore := stores.NewDbUserStore(pool)
	rbac := stores.NewDbRBACStore(pool)
	taskStore := stores.NewDbTaskStore(pool)
	paymentStore := stores.NewDbPaymentStore(pool)
	authStore := stores.NewDbAuthStore(pool)
	userAccountStore := stores.NewDbAccountStore(pool)
	userService := services.NewUserService(userStore)
	userAccountService := services.NewUserAccountService(userAccountStore)
	rbacService := services.NewRBACService(rbac)
	taskService := services.NewTaskService(taskStore)
	paymentClient := services.NewTestPaymentClient()
	paymentService := services.NewPaymentService(
		paymentClient,
		paymentStore,
	)

	tokenService := services.NewJwtServiceDecorator()
	passwordService := services.NewPasswordService()
	routine := services.NewRoutineServiceDecorator()

	authService := services.NewAuthServiceDecorator(
		settings,
		authStore,
		authMailService,
		tokenService,
		passwordService,
		routine,
		l,
		enqueuer,
	)
	checkerStore := stores.NewDbConstraintStore(pool)
	checker := services.NewConstraintCheckerService(
		checkerStore,
	)

	teamService := services.NewTeamService(adapter)

	app := NewApp(
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
		userService,
		userAccountService,
		taskService,
		teamService,
		adapter,
	)
	return &BaseAppDecorator{app: app}
}

type BaseAppDecorator struct {
	app             *BaseApp
	AuthFunc        func() services.AuthService
	CfgFunc         func() *conf.EnvConfig
	CheckerFunc     func() services.ConstraintChecker
	DbFunc          func() database.Dbx
	FsFunc          func() *filesystem.S3FileSystem
	MailerFunc      func() mailer.Mailer
	PaymentFunc     func() services.PaymentService
	RbacFunc        func() services.RBACService
	UserFunc        func() services.UserService
	UserAccountFunc func() services.UserAccountService
	TeamFunc        func() services.TeamService
	TaskFunc        func() services.TaskService
	AdapterFunc     func() stores.StorageAdapterInterface
}

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

func (b *BaseAppDecorator) Mailer() mailer.Mailer {
	if b.MailerFunc != nil {
		return b.MailerFunc()
	}
	return b.app.Mailer()
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

func (b *BaseAppDecorator) User() services.UserService {
	if b.UserFunc != nil {
		return b.UserFunc()
	}
	return b.app.User()
}

func (b *BaseAppDecorator) UserAccount() services.UserAccountService {
	if b.UserAccountFunc != nil {
		return b.UserAccountFunc()
	}
	return b.app.UserAccount()
}
