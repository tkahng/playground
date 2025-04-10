package core

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/pool"
	"github.com/tkahng/authgo/internal/repository"

	"github.com/tkahng/authgo/internal/tools/mailer"
	"github.com/tkahng/authgo/internal/tools/payment"
)

var _ App = (*BaseApp)(nil)

type BaseApp struct {
	tokenStorage  *TokenStorage
	tokenVerifier *TokenVerifier
	cfg           *conf.EnvConfig
	db            bob.DB
	pool          *pgxpool.Pool
	settings      *AppOptions
	payment       *StripeService
	// onAfterRequestHandle  *hook.Hook[*BaseEvent]
	// onBeforeRequestHandle *hook.Hook[*BaseEvent]
}

func (app *BaseApp) Db() bob.DB {
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
	pool := pool.CreatePool(ctx, cfg.Db.DatabaseUrl)
	app := NewBaseApp(pool, cfg)
	app.Bootstrap()
	return app
}

func NewBaseApp(pool *pgxpool.Pool, cfg conf.EnvConfig) *BaseApp {
	oauth := OAuth2ConfigFromEnv(cfg)
	settings := NewDefaultSettings()
	settings.Auth.OAuth2Config = oauth
	return &BaseApp{
		pool:     pool,
		db:       NewBobFromPool(pool),
		settings: settings,
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
