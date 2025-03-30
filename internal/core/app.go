package core

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mailer"
)

type AppDbx interface {

	// Dbx() DBX
	Db() bob.DB
	Pool() *pgxpool.Pool
	AuthConfig() *AuthOptions
}

type App interface {
	TokenStorage() *TokenStorage
	TokenVerifier() *TokenVerifier

	Db() bob.DB
	SetSettings(settings *AppOptions)
	Settings() *AppOptions
	NewMailClient() mailer.Mailer
	EncryptionEnv() string
	// Signup(ctx context.Context, params *shared.AuthenticateUserParams) (*shared.AuthenticatedDTO, error)
	AuthenticateUser(ctx context.Context, db bob.DB, params *shared.AuthenticateUserParams, autoCreateUser bool) (*shared.AuthenticateUserState, error)

	// jwt
	CreateAuthTokens(ctx context.Context, db bob.DB, payload *shared.UserInfoDto) (*shared.TokenDto, error)
	CreateAuthDto(ctx context.Context, email string) (*shared.AuthenticatedDTO, error)
	HandleAuthToken(ctx context.Context, token string) (*shared.UserInfoDto, error)
	RefreshTokens(ctx context.Context, db bob.DB, refreshToken string) (*shared.AuthenticatedDTO, error)
	// verification
	VerifyAndUseVerificationToken(ctx context.Context, db bob.DB, token string) (*EmailVerificationClaims, error)
	SendVerificationEmail(ctx context.Context, db bob.DB, user *models.User, redirectTo string) error
	VerifyAndUsePasswordResetToken(ctx context.Context, db bob.DB, token string) (*PasswordResetClaims, error)
	SendPasswordResetEmail(ctx context.Context, db bob.DB, user *models.User, redirectTo string) error
}
