package auth

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/token"
	"github.com/tkahng/playground/internal/tools/mailer"
	"github.com/tkahng/playground/internal/workers"
)

type (
	SignupInput struct {
		Email    string  `json:"email" required:"true" format:"email" maxLength:"100"`
		Name     *string `json:"name"`
		Password string  `json:"password" required:"true" minLength:"8" maxLength:"100"`
	}
	SigninInput struct {
		Email    string `json:"email" required:"true" format:"email" maxLength:"100"`
		Password string `json:"password" required:"true" minLength:"8" maxLength:"100"`
	}
	SignoutInput struct {
		RefreshToken string `json:"refresh_token" required:"true"`
	}
	OAuth2SigninInput struct {
		Email             string
		Name              *string
		AvatarUrl         *string
		EmailVerifiedAt   *time.Time
		Provider          models.Providers
		ProviderAccountID string
		UserId            *uuid.UUID
		AccessToken       *string
		RefreshToken      *string
	}
)

type AuthService interface {
	// Signup credentials user.
	//
	// - if user with email does not exist, it will create a new user and a credentials account.
	//
	// - if user with email exists, and they have another credentials account or oauth account, it will return error.
	Signup(ctx context.Context, params *SignupInput) (*models.UserInfoTokens, error)
	// Signin credentials user.
	// If user with given email exists and has a credentials account, it will check password.
	Signin(ctx context.Context, params *SigninInput) (*models.UserInfoTokens, error)
	// Signout user.
	// if given refresh token is valid, it will delete the refresh token.
	Signout(ctx context.Context, refreshToken string) error
	// OAuth2Signin user.
	//
	// - if user with email does not exist, it will create a new user and a oauth account.
	//
	// - if user with email exists, and they have another oauth account, it will update the oauth account.
	OAuth2Signin(ctx context.Context, params *OAuth2SigninInput) (*models.UserInfoTokens, error)

	// GenerateAuthTokens
	GenerateAuthTokens(ctx context.Context, email string) (*models.UserInfoTokens, error)
}
type (
	PasswordService interface {
		HashPassword(password string) (string, error)
		VerifyPassword(hashedPassword, password string) (match bool, err error)
	}
	JwtService interface {
		ParseToken(token string, config conf.TokenOption, data any) error
		CreateJwtToken(payload jwt.Claims, signingKey string) (string, error)
	}
	JobService interface {
		EnqueueOtpMailJob(ctx context.Context, args *workers.OtpEmailJobArgs) error
	}
)

func NewAuthService(
	config *conf.EnvConfig,
	adapter stores.StorageAdapterInterface,
	password PasswordService,
	jwt JwtService,
	token token.TokenService,
	job JobService,
) AuthService {
	return &AuthServiceImpl{
		config:   config,
		adapter:  adapter,
		password: password,
		jwt:      jwt,
		token:    token,
		job:      job,
	}
}

type AuthServiceImpl struct {
	config   *conf.EnvConfig
	adapter  stores.StorageAdapterInterface
	password PasswordService
	jwt      JwtService
	token    token.TokenService
	job      JobService
}

// GenerateAuthTokens implements AuthService.
func (a *AuthServiceImpl) GenerateAuthTokens(ctx context.Context, email string) (*models.UserInfoTokens, error) {
	userInfo, err := a.adapter.User().GetUserInfo(ctx, email)
	if err != nil {
		return nil, err
	}
	opts := a.config.AuthOptions

	authToken, err := func() (string, error) {
		claims := shared.AuthenticationClaims{
			Type: models.TokenTypesAccessToken,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: opts.AccessToken.ExpiresAt(),
			},
			AuthenticationPayload: shared.AuthenticationPayload{
				UserId:      userInfo.User.ID,
				Email:       userInfo.User.Email,
				Roles:       userInfo.Roles,
				Permissions: userInfo.Permissions,
			},
		}
		token, err := a.jwt.CreateJwtToken(claims, opts.AccessToken.Secret)
		if err != nil {
			return token, err
		}
		return token, nil
	}()
	if err != nil {
		return nil, err
	}

	refreshToken, err := a.token.GenerateToken(ctx, userInfo.User.Email, models.TokenTypesRefreshToken)
	if err != nil {
		return nil, err
	}

	return &models.UserInfoTokens{
		UserInfo: *userInfo,
		Tokens: models.TokenDto{
			AccessToken:  authToken,
			RefreshToken: refreshToken,
			ExpiresIn:    opts.AccessToken.Duration,
			TokenType:    "Bearer",
		},
	}, nil
}

// OAuth2Signin implements AuthService.
func (a *AuthServiceImpl) OAuth2Signin(ctx context.Context, params *OAuth2SigninInput) (*models.UserInfoTokens, error) {
	panic("unimplemented")
}

// Signin implements AuthService.
func (a *AuthServiceImpl) Signin(ctx context.Context, params *SigninInput) (*models.UserInfoTokens, error) {
	panic("unimplemented")
}

// Signout implements AuthService.
func (a *AuthServiceImpl) Signout(ctx context.Context, refreshToken string) error {
	panic("unimplemented")
}

// Signup implements AuthService.
// Signup credentials user.
//
// - if user with email does not exist, it will create a new user and a credentials account.
//
// - if user with email exists, and they have another credentials account or oauth account, it will return error.
func (a *AuthServiceImpl) Signup(ctx context.Context, params *SignupInput) (*models.UserInfoTokens, error) {
	// check if user with email exists.
	existingUser, err := a.adapter.User().FindUser(ctx, &stores.UserFilter{
		Emails: []string{params.Email},
	})
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, shared.ErrUserExists
	}
	// create a new user and a credentials account.
	hashedPassword, err := a.password.HashPassword(params.Password)
	if err != nil {
		return nil, err
	}
	user, err := a.adapter.User().CreateUser(ctx, &models.User{
		Name:  params.Name,
		Email: params.Email,
	})
	if err != nil {
		return nil, err
	}
	_, err = a.adapter.UserAccount().CreateUserAccount(ctx, &models.UserAccount{
		UserID:   user.ID,
		Provider: models.ProvidersCredentials,
		Type:     models.ProviderTypeCredentials,
		Password: &hashedPassword,
	})
	if err != nil {
		return nil, err
	}
	err = a.job.EnqueueOtpMailJob(ctx, &workers.OtpEmailJobArgs{
		UserID: user.ID,
		Type:   mailer.EmailTypeConfirmPasswordReset,
	})
	if err != nil {
		return nil, err
	}
	tokens, err := a.GenerateAuthTokens(ctx, params.Email)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

var _ AuthService = (*AuthServiceImpl)(nil)
