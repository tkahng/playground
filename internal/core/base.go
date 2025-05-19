package core

import (
	"context"
	"log/slog"

	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/queries"

	"github.com/tkahng/authgo/internal/modules/authmodule"
	"github.com/tkahng/authgo/internal/modules/paymentmodule"
	"github.com/tkahng/authgo/internal/modules/rbacmodule"
	"github.com/tkahng/authgo/internal/modules/teammodule"
	"github.com/tkahng/authgo/internal/tools/filesystem"
	"github.com/tkahng/authgo/internal/tools/logger"
	"github.com/tkahng/authgo/internal/tools/mailer"
)

var _ App = (*BaseApp)(nil)

type BaseApp struct {
	cfg      *conf.EnvConfig
	db       *database.Queries
	settings *conf.AppOptions
	payment  paymentmodule.PaymentService
	logger   *slog.Logger
	fs       *filesystem.FileSystem
	mail     mailer.Mailer
	auth     authmodule.AuthService
}

// Checker implements App.
func (a *BaseApp) NewChecker(ctx context.Context) ConstraintChecker {
	return NewConstraintCheckerService(ctx, a.db)
}

// Auth implements App.
func (a *BaseApp) Auth() authmodule.AuthService {
	return a.auth
}

func (app *BaseApp) Fs() *filesystem.FileSystem {
	return app.fs
}

func (app *BaseApp) Logger() *slog.Logger {
	return app.logger
}
func (app *BaseApp) Db() database.Dbx {
	return app.db
}

// Payment implements App.
func (a *BaseApp) Payment() paymentmodule.PaymentService {
	return a.payment
}

// Settings implements App.
func (a *BaseApp) Settings() *conf.AppOptions {
	return a.settings
}

// EncryptionEnv implements App.
func (app *BaseApp) EncryptionEnv() string {
	return app.cfg.EncryptionKey
}

func (app *BaseApp) Cfg() *conf.EnvConfig {
	return app.cfg
}

// NewMailClient implements App.
func (app *BaseApp) NewMailClient() mailer.Mailer {
	return app.mail
}

func InitBaseApp(ctx context.Context, cfg conf.EnvConfig) *BaseApp {
	pool := database.CreateQueries(ctx, cfg.Db.DatabaseUrl)
	fs, err := filesystem.NewFileSystem(cfg.StorageConfig)
	if err != nil {
		panic(err)
	}
	var mail mailer.Mailer
	if cfg.ResendConfig.ResendApiKey != "" {
		mail = mailer.NewResendMailer(cfg.ResendConfig)
	} else {
		mail = &mailer.LogMailer{}
	}
	paymentStore := paymentmodule.NewPostgresPaymentStore(pool)
	paymentClient := paymentmodule.NewPaymentClient(cfg.StripeConfig)

	rbacStore := rbacmodule.NewPostgresRBACStore(pool)
	teamStore := teammodule.NewPostgresTeamStore(pool)
	stripeService := paymentmodule.NewPaymentService(
		paymentClient,
		paymentStore,
		rbacStore,
		teamStore,
	)
	app := &BaseApp{
		fs:       fs,
		db:       pool,
		settings: cfg.ToSettings(),
		logger:   logger.GetDefaultLogger(slog.LevelInfo),
		cfg:      &cfg,
		mail:     mail,
		payment:  stripeService,
	}
	return app
}

func (app *BaseApp) Bootstrap() {
	ctx := context.Background()
	db := app.Db()
	queries.EnsureRoleAndPermissions(ctx, db, "superuser", "superuser")
	queries.EnsureRoleAndPermissions(ctx, db, "basic", "basic")
}
