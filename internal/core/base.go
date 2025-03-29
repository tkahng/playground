package core

import (
	"context"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/conf"

	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/tools/hook"
	"github.com/tkahng/authgo/internal/tools/mailer"
)

var _ App = (*BaseApp)(nil)

type BaseApp struct {
	tokenStorage          *TokenStorage
	tokenVerifier         *TokenVerifier
	cfg                   *conf.EnvConfig
	db                    bob.DB
	settings              *Settings
	authOpt               *AuthOptions
	onAfterRequestHandle  *hook.Hook[*BaseEvent]
	onBeforeRequestHandle *hook.Hook[*BaseEvent]
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
func (a *BaseApp) Settings() *Settings {
	return a.settings
}

// // SendVerificationEmail implements App.
// func (a *BaseApp) SendVerificationEmail(ctx context.Context, db bob.DB, email string) error {
// 	panic("unimplemented")
// }

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
func (app *BaseApp) InitHooks() {
	app.onAfterRequestHandle = &hook.Hook[*BaseEvent]{}
	app.onBeforeRequestHandle = &hook.Hook[*BaseEvent]{}
}

// InitSettings implements App.
func (app *BaseApp) InitSettings(ctx context.Context, db bob.DB) {
	// a.InitAuthSettings(ctx, db)
}

func InitBaseApp(ctx context.Context, cfg conf.EnvConfig) *BaseApp {
	db := NewBobFromConf(ctx, cfg.Db)
	authOpt, err := GetOrSetEncryptedAppOptions(ctx, db, cfg.EncryptionKey)
	if err != nil {
		panic(err)
	}
	app := NewBaseApp(db, cfg, authOpt)
	app.Bootstrap()
	return app
}

func NewBaseApp(db bob.DB, cfg conf.EnvConfig, authOpt *Settings) *BaseApp {
	return &BaseApp{
		db:       db,
		authOpt:  &authOpt.Auth,
		settings: authOpt,
	}
}

func (app *BaseApp) Db() bob.DB {
	return app.db
}

// OnAfterRequestHandle implements App.
func (app *BaseApp) OnAfterRequestHandle(tags ...string) *hook.TaggedHook[*BaseEvent] {
	return hook.NewTaggedHook(app.onAfterRequestHandle, tags...)
}

// OnBeforeRequestHandle implements App.
func (app *BaseApp) OnBeforeRequestHandle(tags ...string) *hook.TaggedHook[*BaseEvent] {
	return hook.NewTaggedHook(app.onBeforeRequestHandle, tags...)
}

func (app *BaseApp) Bootstrap() {
	repository.InitRolesFromTree(context.Background(), app.db)
}
