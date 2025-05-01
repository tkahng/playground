package core

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/repository"

	"github.com/tkahng/authgo/internal/tools/filesystem"
	"github.com/tkahng/authgo/internal/tools/logger"
	"github.com/tkahng/authgo/internal/tools/mailer"
	"github.com/tkahng/authgo/internal/tools/payment"
)

var _ App = (*BaseApp)(nil)

type BaseApp struct {
	cfg      *conf.EnvConfig
	db       *db.Queries
	pool     *pgxpool.Pool
	settings *AppOptions
	payment  *StripeService
	logger   *slog.Logger
	fs       *filesystem.FileSystem
	mail     mailer.Mailer
	checker  ConstraintChecker
}

// Checker implements App.
func (a *BaseApp) NewChecker(ctx context.Context) ConstraintChecker {
	return NewConstraintCheckerService(ctx, a.db)
}

// NewAuthActions implements App.
func (a *BaseApp) NewAuthActions() AuthActions {
	return NewAuthActions(a.Pool(), a.mail, a.settings)
}

func (app *BaseApp) Fs() *filesystem.FileSystem {
	return app.fs
}

func (app *BaseApp) Logger() *slog.Logger {
	return app.logger
}
func (app *BaseApp) Db() *db.Queries {
	return app.db
}
func (a *BaseApp) Pool() *pgxpool.Pool {
	return a.pool
}

// Payment implements App.
func (a *BaseApp) Payment() *StripeService {
	return a.payment
}

// Settings implements App.
func (a *BaseApp) Settings() *AppOptions {
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
	pool := db.CreatePool(ctx, cfg.Db.DatabaseUrl)
	app := NewBaseApp(pool, cfg)
	app.Bootstrap()
	return app
}

func NewBaseApp(pool *pgxpool.Pool, cfg conf.EnvConfig) *BaseApp {
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
	db := db.NewQueries(pool)
	return &BaseApp{
		fs:       fs,
		pool:     pool,
		db:       db,
		settings: NewSettingsFromConf(&cfg),
		logger:   logger.GetDefaultLogger(slog.LevelInfo),
		cfg:      &cfg,
		mail:     mail,
		payment:  NewStripeService(payment.NewStripeClient(cfg.StripeConfig)),
	}
}

func (app *BaseApp) Bootstrap() {
	ctx := context.Background()
	db := app.Db()
	repository.EnsureRoleAndPermissions(ctx, db, "superuser", "superuser")
	repository.EnsureRoleAndPermissions(ctx, db, "basic", "basic")
}
