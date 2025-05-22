package core

import (
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
	Settings() *conf.AppOptions
	NewMailClient() mailer.Mailer
	EncryptionEnv() string

	Rbac() services.RBACService
	User() services.UserService
	Payment() services.PaymentService

	Auth() services.AuthService

	Team() services.TeamService

	Checker() services.ConstraintChecker
}
