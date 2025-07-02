package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/auth/oauth"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/mailer"
)

type AuthServiceDecorator struct {
	Delegate                       *BaseAuthService
	MailFunc                       func() MailService
	PasswordFunc                   func() PasswordService
	TokenFunc                      func() JwtService
	CreateOAuthUrlFunc             func(ctx context.Context, provider models.Providers, redirectUrl string) (string, error)
	AuthenticateFunc               func(ctx context.Context, params *AuthenticationInput) (*models.User, error)
	CheckResetPasswordTokenFunc    func(ctx context.Context, token string) error
	HandleAccessTokenFunc          func(ctx context.Context, token string) (*models.UserInfo, error)
	HandleRefreshTokenFunc         func(ctx context.Context, token string) (*models.UserInfoTokens, error)
	HandlePasswordResetRequestFunc func(ctx context.Context, email string) error
	HandlePasswordResetTokenFunc   func(ctx context.Context, token string, password string) error
	HandleVerificationTokenFunc    func(ctx context.Context, token string) error
	SignoutFunc                    func(ctx context.Context, token string) error
	ResetPasswordFunc              func(ctx context.Context, userId uuid.UUID, oldPassword string, newPassword string) error
	SendOtpEmailFunc               func(emailType mailer.EmailType, ctx context.Context, user *models.User, adapter stores.StorageAdapterInterface) error
	VerifyAndParseOtpTokenFunc     func(ctx context.Context, emailType mailer.EmailType, token string) (*shared.OtpClaims, error)
	VerifyStateTokenFunc           func(ctx context.Context, token string) (*shared.ProviderStateClaims, error)
	CreateAndPersistStateTokenFunc func(ctx context.Context, payload *shared.ProviderStatePayload) (string, error)
	CreateAuthTokensFromEmailFunc  func(ctx context.Context, email string) (*models.UserInfoTokens, error)
	FetchAuthUserFunc              func(ctx context.Context, code string, parsedState *shared.ProviderStateClaims) (*oauth.AuthUser, error)
}

func NewAuthServiceDecorator(
	opts *conf.AppOptions,
	mail MailService,
	adapter stores.StorageAdapterInterface,
	jobService JobService,
) AuthService {
	tokenService := NewJwtServiceDecorator()
	passwordService := NewPasswordServiceDecorator()
	routine := NewRoutineServiceDecorator()
	authService := &AuthServiceDecorator{}
	authService.Delegate = &BaseAuthService{
		routine:  routine,
		mail:     mail,
		token:    tokenService,
		password: passwordService,
		options:  opts,
		adapter:  adapter,
		jobService: &JobServiceDecorator{
			Delegate: jobService,
		},
	}
	return authService
}

var _ AuthService = (*AuthServiceDecorator)(nil)

// Mail implements AuthService.
func (a *AuthServiceDecorator) Mail() MailService {
	if a.MailFunc != nil {
		return a.MailFunc()
	}
	return a.Delegate.Mail()
}

// Password implements AuthService.
func (a *AuthServiceDecorator) Password() PasswordService {
	if a.PasswordFunc != nil {
		return a.PasswordFunc()
	}
	return a.Delegate.Password()
}

// Token implements AuthService.
func (a *AuthServiceDecorator) Token() JwtService {
	if a.TokenFunc != nil {
		return a.TokenFunc()
	}
	return a.Delegate.Token()
}

// CreateOAuthUrl implements AuthService.
func (a *AuthServiceDecorator) CreateOAuthUrl(ctx context.Context, provider models.Providers, redirectUrl string) (string, error) {
	if a.CreateOAuthUrlFunc != nil {
		return a.CreateOAuthUrlFunc(ctx, provider, redirectUrl)
	}
	return a.Delegate.CreateOAuthUrl(ctx, provider, redirectUrl)
}

// Authenticate implements AuthService.
func (a *AuthServiceDecorator) Authenticate(ctx context.Context, params *AuthenticationInput) (*models.User, error) {
	if a.AuthenticateFunc != nil {
		return a.AuthenticateFunc(ctx, params)
	}
	return a.Delegate.Authenticate(ctx, params)
}

// HandleCheckResetPasswordToken implements AuthService.
func (a *AuthServiceDecorator) HandleCheckResetPasswordToken(ctx context.Context, token string) error {
	if a.CheckResetPasswordTokenFunc != nil {
		return a.CheckResetPasswordTokenFunc(ctx, token)
	}
	return a.Delegate.HandleCheckResetPasswordToken(ctx, token)
}

// CreateAndPersistStateToken implements AuthService.
func (a *AuthServiceDecorator) CreateAndPersistStateToken(ctx context.Context, payload *shared.ProviderStatePayload) (string, error) {
	if a.CreateAndPersistStateTokenFunc != nil {
		return a.CreateAndPersistStateTokenFunc(ctx, payload)
	}
	return a.Delegate.CreateAndPersistStateToken(ctx, payload)
}

// CreateAuthTokensFromEmail implements AuthService.
func (a *AuthServiceDecorator) CreateAuthTokensFromEmail(ctx context.Context, email string) (*models.UserInfoTokens, error) {
	if a.CreateAuthTokensFromEmailFunc != nil {
		return a.CreateAuthTokensFromEmailFunc(ctx, email)
	}
	return a.Delegate.CreateAuthTokensFromEmail(ctx, email)
}

// FetchAuthUser implements AuthService.
func (a *AuthServiceDecorator) FetchAuthUser(ctx context.Context, code string, parsedState *shared.ProviderStateClaims) (*oauth.AuthUser, error) {
	if a.FetchAuthUserFunc != nil {
		return a.FetchAuthUserFunc(ctx, code, parsedState)
	}
	return a.Delegate.FetchAuthUser(ctx, code, parsedState)
}

// HandleAccessToken implements AuthService.
func (a *AuthServiceDecorator) HandleAccessToken(ctx context.Context, token string) (*models.UserInfo, error) {
	if a.HandleAccessTokenFunc != nil {
		return a.HandleAccessTokenFunc(ctx, token)
	}
	return a.Delegate.HandleAccessToken(ctx, token)
}

// HandlePasswordResetRequest implements AuthService.
func (a *AuthServiceDecorator) HandlePasswordResetRequest(ctx context.Context, email string) error {
	if a.HandlePasswordResetRequestFunc != nil {
		return a.HandlePasswordResetRequestFunc(ctx, email)
	}
	return a.Delegate.HandlePasswordResetRequest(ctx, email)
}

// HandlePasswordResetToken implements AuthService.
func (a *AuthServiceDecorator) HandlePasswordResetToken(ctx context.Context, token string, password string) error {
	if a.HandlePasswordResetTokenFunc != nil {
		return a.HandlePasswordResetTokenFunc(ctx, token, password)
	}
	return a.Delegate.HandlePasswordResetToken(ctx, token, password)
}

// HandleRefreshToken implements AuthService.
func (a *AuthServiceDecorator) HandleRefreshToken(ctx context.Context, token string) (*models.UserInfoTokens, error) {
	if a.HandleRefreshTokenFunc != nil {
		return a.HandleRefreshTokenFunc(ctx, token)
	}
	return a.Delegate.HandleRefreshToken(ctx, token)
}

// HandleVerificationToken implements AuthService.
func (a *AuthServiceDecorator) HandleVerificationToken(ctx context.Context, token string) error {
	if a.HandleVerificationTokenFunc != nil {
		return a.HandleVerificationTokenFunc(ctx, token)
	}
	return a.Delegate.HandleVerificationToken(ctx, token)
}

// ResetPassword implements AuthService.
func (a *AuthServiceDecorator) ResetPassword(ctx context.Context, userId uuid.UUID, oldPassword string, newPassword string) error {
	if a.ResetPasswordFunc != nil {
		return a.ResetPasswordFunc(ctx, userId, oldPassword, newPassword)
	}
	return a.Delegate.ResetPassword(ctx, userId, oldPassword, newPassword)
}

// SendOtpEmail implements AuthService.
// func (a *AuthServiceDecorator) SendOtpEmail(emailType mailer.EmailType, ctx context.Context, user *models.User, adapter stores.StorageAdapterInterface) error {
// 	if a.SendOtpEmailFunc != nil {
// 		return a.SendOtpEmailFunc(emailType, ctx, user, adapter)
// 	}
// 	return a.Delegate.SendOtpEmail(emailType, ctx, user, adapter)
// }

// Signout implements AuthService.
func (a *AuthServiceDecorator) Signout(ctx context.Context, token string) error {
	if a.SignoutFunc != nil {
		return a.SignoutFunc(ctx, token)
	}
	return a.Delegate.Signout(ctx, token)
}

// VerifyAndParseOtpToken implements AuthService.
func (a *AuthServiceDecorator) VerifyAndParseOtpToken(ctx context.Context, emailType mailer.EmailType, token string) (*shared.OtpClaims, error) {
	if a.VerifyAndParseOtpTokenFunc != nil {
		return a.VerifyAndParseOtpTokenFunc(ctx, emailType, token)
	}
	return a.Delegate.VerifyAndParseOtpToken(ctx, emailType, token)
}

// VerifyStateToken implements AuthService.
func (a *AuthServiceDecorator) VerifyStateToken(ctx context.Context, token string) (*shared.ProviderStateClaims, error) {
	if a.VerifyStateTokenFunc != nil {
		return a.VerifyStateTokenFunc(ctx, token)
	}
	return a.Delegate.VerifyStateToken(ctx, token)
}
