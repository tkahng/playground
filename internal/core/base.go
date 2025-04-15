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
	tokenStorage  *TokenStorage
	tokenVerifier *TokenVerifier
	cfg           *conf.EnvConfig
	db            *db.DBTx
	pool          *pgxpool.Pool
	settings      *AppOptions
	payment       *StripeService
	logger        *slog.Logger
	fs            *filesystem.FileSystem
	// onAfterRequestHandle  *hook.Hook[*BaseEvent]
	// onBeforeRequestHandle *hook.Hook[*BaseEvent]
}

func (app *BaseApp) Fs() *filesystem.FileSystem {
	return app.fs
}

func (app *BaseApp) Logger() *slog.Logger {
	return app.logger
}
func (app *BaseApp) Db() *db.DBTx {
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
	return &mailer.LogMailer{}
}

// InitHooks implements App.
// func (app *BaseApp) InitHooks() {
// 	app.onAfterRequestHandle = &hook.Hook[*BaseEvent]{}
// 	app.onBeforeRequestHandle = &hook.Hook[*BaseEvent]{}
// }

func InitBaseApp(ctx context.Context, cfg conf.EnvConfig) *BaseApp {
	pool := db.CreatePool(ctx, cfg.Db.DatabaseUrl)
	app := NewBaseApp(pool, cfg)
	app.Bootstrap()
	return app
}

func NewBaseApp(pool *pgxpool.Pool, cfg conf.EnvConfig) *BaseApp {
	oauth := OAuth2ConfigFromEnv(cfg)
	settings := NewDefaultSettings()
	settings.Auth.OAuth2Config = oauth
	fs, err := filesystem.NewFileSystem(cfg.StorageConfig)
	if err != nil {
		panic(err)
	}
	return &BaseApp{
		fs:       fs,
		pool:     pool,
		db:       db.NewDBTx(pool),
		settings: settings,
		logger:   logger.GetDefaultLogger(slog.LevelInfo),
		cfg:      &cfg,
		payment:  NewStripeService(payment.NewStripeClient(cfg.StripeConfig)),
	}
}

// // OnAfterRequestHandle implements App.
// func (app *BaseApp) OnAfterRequestHandle(tags ...string) *hook.TaggedHook[*BaseEvent] {
// 	return hook.NewTaggedHook(app.onAfterRequestHandle, tags...)
// }

// // OnBeforeRequestHandle implements App.
// func (app *BaseApp) OnBeforeRequestHandle(tags ...string) *hook.TaggedHook[*BaseEvent] {
// 	return hook.NewTaggedHook(app.onBeforeRequestHandle, tags...)
// }

func (app *BaseApp) Bootstrap() {
	ctx := context.Background()
	db := app.Db()
	repository.EnsureRoleAndPermissions(ctx, db, "superuser", "superuser")
	repository.EnsureRoleAndPermissions(ctx, db, "basic", "basic")
}
