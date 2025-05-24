package core

import (
	"context"
	"log/slog"

	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/filesystem"
	"github.com/tkahng/authgo/internal/tools/logger"
	"github.com/tkahng/authgo/internal/tools/mailer"
	"github.com/tkahng/authgo/internal/tools/payment"
)

func NewDecorator(ctx context.Context, cfg conf.EnvConfig, pool *database.Queries) *BaseAppDecorator {
	settings := cfg.ToSettings()

	fs, err := filesystem.NewFileSystem(cfg.StorageConfig)
	if err != nil {
		panic(err)
	}
	l := logger.GetDefaultLogger(slog.LevelInfo)

	var mail mailer.Mailer = &mailer.LogMailer{}
	userStore := stores.NewPostgresUserStore(pool)
	userService := services.NewUserService(userStore)
	userAccountStore := stores.NewPostgresUserAccountStore(pool)
	userAccountService := services.NewUserAccountService(userAccountStore)
	rbac := stores.NewPostgresRBACStore(pool)
	rbacService := services.NewRBACService(rbac)
	taskStore := stores.NewTaskStore(pool)
	taskService := services.NewTaskService(taskStore)
	paymentStore := stores.NewPostgresPaymentStore(pool)
	paymentClient := payment.NewPaymentClient(cfg.StripeConfig)
	paymentService := services.NewPaymentService(
		paymentClient,
		paymentStore,
	)

	tokenService := services.NewJwtService()
	passwordService := services.NewPasswordService()
	authStore := stores.NewPostgresAuthStore(pool)
	workerService := services.NewRoutineService()
	authMailService := services.NewMailService(mail)
	authService := services.NewAuthService(
		settings,
		authStore,
		authMailService,
		tokenService,
		passwordService,
		workerService,
		l,
	)
	checkerStore := stores.NewPostgresConstraintStore(pool)
	checker := services.NewConstraintCheckerService(
		checkerStore,
	)

	teamStore := stores.NewPostgresTeamServiceStore(pool)
	teamService := services.NewTeamService(teamStore)

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
