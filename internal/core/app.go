package core

import (
	"context"

	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/tools/filesystem"
	"github.com/tkahng/authgo/internal/tools/mailer"
)

type App interface {
	Cfg() *conf.EnvConfig
	Db() database.Dbx
	Fs() *filesystem.FileSystem
	NewChecker(ctx context.Context) ConstraintChecker
	Settings() *conf.AppOptions
	NewMailClient() mailer.Mailer
	EncryptionEnv() string

	Payment() services.PaymentService

	Auth() services.AuthService

	Team() services.TeamService
}
