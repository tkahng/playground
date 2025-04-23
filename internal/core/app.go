package core

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/tools/filesystem"
	"github.com/tkahng/authgo/internal/tools/mailer"
)

type App interface {
	Cfg() *conf.EnvConfig
	Pool() *pgxpool.Pool
	Db() *db.Queries
	Fs() *filesystem.FileSystem

	Settings() *AppOptions
	NewMailClient() mailer.Mailer
	EncryptionEnv() string

	Payment() *StripeService

	NewAuthActions(db bob.Executor) AuthActions
}
