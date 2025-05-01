package core

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
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
	NewChecker(ctx context.Context) ConstraintChecker
	Settings() *AppOptions
	NewMailClient() mailer.Mailer
	EncryptionEnv() string

	Payment() *StripeService

	NewAuthActions() AuthActions
}
