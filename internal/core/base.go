package core

import (
	"context"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/repository"

	"github.com/tkahng/authgo/internal/tools/mailer"
)

var _ App = (*BaseApp)(nil)

type BaseApp struct {
	tokenStorage  *TokenStorage
	tokenVerifier *TokenVerifier
	cfg           *conf.EnvConfig
	db            bob.DB
	settings      *AppOptions
	// onAfterRequestHandle  *hook.Hook[*BaseEvent]
	// onBeforeRequestHandle *hook.Hook[*BaseEvent]
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
	db := NewBobFromConf(ctx, cfg.Db)
	app := NewBaseApp(db, cfg)
	app.Bootstrap()
	return app
}

func NewBaseApp(db bob.DB, cfg conf.EnvConfig) *BaseApp {
	oauth := OAuth2ConfigFromEnv(cfg)
	settings := NewDefaultSettings()
	settings.Auth.OAuth2Config = oauth
	return &BaseApp{
		db:       db,
		settings: settings,
		cfg:      &cfg,
	}
}

func (app *BaseApp) Db() bob.DB {
	return app.db
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
	repository.EnsureRoleAndPermissions(ctx, app.db, "superuser", "superuser")
	repository.EnsureRoleAndPermissions(ctx, app.db, "basic", "basic")
}
