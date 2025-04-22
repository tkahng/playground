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

	// jwt
	// CreateAuthTokens(ctx context.Context, db bob.Executor, payload *shared.UserInfoDto) (*shared.TokenDto, error)
	// CreateAuthDto(ctx context.Context, email string) (*shared.AuthenticatedDTO, error)
	// HandleAuthToken(ctx context.Context, token string) (*shared.UserInfoDto, error)
	// RefreshTokens(ctx context.Context, db bob.Executor, refreshToken string) (*shared.AuthenticatedDTO, error)
	// verification
	// SendVerificationEmail(ctx context.Context, db bob.Executor, user *models.User, redirectTo string) error
	// VerifyAndUsePasswordResetToken(ctx context.Context, db bob.Executor, token string) (*PasswordResetClaims, error)
	// SendPasswordResetEmail(ctx context.Context, db bob.Executor, user *models.User, redirectTo string) error

	// SendSecurityPasswordResetEmail(ctx context.Context, db bob.Executor, user *models.User, redirectTo string) error
	// CheckUserCredentialsSecurity(ctx context.Context, db bob.Executor, user *models.User, params *shared.AuthenticateUserParams) error
	// stripe
	Payment() *StripeService

	NewAuthActions(db bob.Executor) AuthActions
}
