package core

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stephenafamo/bob"
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
	tokenStorage  *TokenStorage
	tokenVerifier *TokenVerifier
	cfg           *conf.EnvConfig
	db            *db.Queries
	pool          *pgxpool.Pool
	settings      *AppOptions
	payment       *StripeService
	logger        *slog.Logger
	fs            *filesystem.FileSystem
	mail          mailer.Mailer
	authAdapter   *AuthAdapterBase
	authMailer    *AuthMailerBase
	tokenAdapter  *TokenAdapterBase
	// onAfterRequestHandle  *hook.Hook[*BaseEvent]
	// onBeforeRequestHandle *hook.Hook[*BaseEvent]
}

// NewAuthActions implements App.
func (a *BaseApp) NewAuthActions(db bob.Executor) AuthActions {
	return NewAuthActions(db, a.mail, a.settings)
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

// TokenVerifier implements App.
func (a *BaseApp) TokenVerifier() *TokenVerifier {
	return a.tokenVerifier
}

// TokenStorage implements App.
func (app *BaseApp) TokenStorage() *TokenStorage {
	return app.tokenStorage
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
	settings := NewSettingsFromConf(&cfg)
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
	authAdapter := NewAuthAdapter(db)
	authMailer := NewAuthMailer(mail)
	tokenAdapter := NewTokenAdapter(db)
	return &BaseApp{
		fs:           fs,
		pool:         pool,
		db:           db,
		settings:     settings,
		logger:       logger.GetDefaultLogger(slog.LevelInfo),
		cfg:          &cfg,
		mail:         mail,
		payment:      NewStripeService(payment.NewStripeClient(cfg.StripeConfig)),
		authAdapter:  authAdapter,
		authMailer:   authMailer,
		tokenAdapter: tokenAdapter,
	}
}

func (app *BaseApp) Bootstrap() {
	ctx := context.Background()
	db := app.Db()
	repository.EnsureRoleAndPermissions(ctx, db, "superuser", "superuser")
	repository.EnsureRoleAndPermissions(ctx, db, "basic", "basic")
}
