package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/auth/oauth"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/token"
	"github.com/tkahng/playground/internal/tools/mailer"
	"github.com/tkahng/playground/internal/workers"
)

type AuthService interface {
	CredentialsAuthenticator
	PasswordManager
	Oauth2Authenticator

	VerifyAccessToken(ctx context.Context, token string) (*models.UserInfo, error)
	RefreshToken(ctx context.Context, refreshToken string) (*models.UserInfoTokens, error)

	SendEmailVerification(ctx context.Context, email string) error
	ValidateEmailVerification(ctx context.Context, code string) error

	// GenerateAuthTokens
	GenerateAuthTokens(ctx context.Context, email string) (*models.UserInfoTokens, error)
}

type (
	Encrypter interface {
		Encrypt(data []byte) (string, error)
		Decrypt(cipherText string) ([]byte, error)
	}

	HashService interface {
		Hash(input string) (string, error)
		Verify(value, hash string) (match bool, err error)
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
	logger *slog.Logger,
	adapter stores.StorageAdapterInterface,
	hash HashService,
	jwt JwtService,
	token token.TokenService,
	job JobService,
	encrypter Encrypter,
) AuthService {
	oauth.OAuth2ConfigFromEnv(*config)
	return &AuthServiceImpl{
		config:    config,
		adapter:   adapter,
		hash:      hash,
		jwt:       jwt,
		token:     token,
		job:       job,
		logger:    logger,
		encrypter: encrypter,
	}
}

type AuthServiceImpl struct {
	config    *conf.EnvConfig
	adapter   stores.StorageAdapterInterface
	hash      HashService
	jwt       JwtService
	token     token.TokenService
	job       JobService
	logger    *slog.Logger
	encrypter Encrypter
}

var _ AuthService = (*AuthServiceImpl)(nil)

// SendEmailVerification implements AuthService.
func (a *AuthServiceImpl) SendEmailVerification(ctx context.Context, email string) error {
	user, err := a.adapter.User().FindUser(ctx, &stores.UserFilter{
		Emails: []string{email},
	})
	if err != nil {
		return err
	}
	if user == nil {
		return shared.ErrUserNotFound
	}
	err = a.job.EnqueueOtpMailJob(ctx, &workers.OtpEmailJobArgs{
		UserID: user.ID,
		Type:   mailer.EmailTypeVerify,
	})
	if err != nil {
		return err
	}
	return nil
}

// ValidateEmailVerification implements AuthService.
func (a *AuthServiceImpl) ValidateEmailVerification(ctx context.Context, code string) error {
	email, err := a.token.ValidateToken(ctx, &token.ValidateTokenOptions{
		Token: code,
		Type:  models.TokenTypesVerificationToken,
	})
	if err != nil {
		return err
	}
	user, err := a.adapter.User().FindUser(ctx, &stores.UserFilter{
		Emails: []string{email},
	})
	if err != nil {
		return err
	}
	if user == nil {
		return shared.ErrUserNotFound
	}
	if user.EmailVerifiedAt != nil {
		return errors.New("user email already verified")
	}
	now := time.Now()
	user.EmailVerifiedAt = &now
	err = a.adapter.User().UpdateUser(ctx, user)
	if err != nil {
		return err
	}
	return nil
}

type authenticationClaims struct {
	jwt.RegisteredClaims
	Type models.TokenTypes `json:"type"`
	authenticationPayload
}

type authenticationPayload struct {
	UserId      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
}

// GenerateAuthTokens implements AuthService.
func (a *AuthServiceImpl) GenerateAuthTokens(ctx context.Context, email string) (*models.UserInfoTokens, error) {
	userInfo, err := a.adapter.User().GetUserInfo(ctx, email)
	if err != nil {
		return nil, err
	}
	opts := a.config.AuthOptions

	authToken, err := func() (string, error) {
		claims := authenticationClaims{
			Type: models.TokenTypesAccessToken,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: opts.AccessToken.ExpiresAt(),
			},
			authenticationPayload: authenticationPayload{
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

	refreshToken, err := a.token.GenerateToken(ctx, &token.GenerateTokenOptions{
		Email: userInfo.User.Email,
		Type:  models.TokenTypesRefreshToken,
	})
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

func (a *AuthServiceImpl) VerifyAccessToken(ctx context.Context, token string) (*models.UserInfo, error) {
	opts := a.config.AuthOptions
	var claims authenticationClaims
	err := a.jwt.ParseToken(token, opts.AccessToken, &claims)
	if err != nil {
		return nil, fmt.Errorf("error verifying access token: %w", err)
	}
	return a.adapter.User().GetUserInfo(ctx, claims.Email)
}

// RefreshToken implements AuthService.
func (a *AuthServiceImpl) RefreshToken(ctx context.Context, refreshToken string) (*models.UserInfoTokens, error) {
	email, err := a.token.ValidateToken(ctx, &token.ValidateTokenOptions{
		Token: refreshToken,
		Type:  models.TokenTypesRefreshToken,
	})
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid refresh token")
	}
	claims, err := a.GenerateAuthTokens(ctx, email)
	if err != nil {
		return nil, err
	}
	return claims, nil
}
