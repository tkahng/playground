package core

import (
	"context"

	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/tools/filesystem"
	"github.com/tkahng/authgo/internal/tools/mailer"
)

type App interface {
	Cfg() *conf.EnvConfig
	Db() *db.Queries
	Fs() *filesystem.FileSystem
	NewChecker(ctx context.Context) ConstraintChecker
	Settings() *conf.AppOptions
	NewMailClient() mailer.Mailer
	EncryptionEnv() string

	Payment() *StripeService

	NewAuthActions() Authenticator
}
