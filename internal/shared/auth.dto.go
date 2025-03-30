package shared

import (
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db/models"
)

type AuthenticateUserParams struct {
	Email             string
	Name              *string
	AvatarUrl         *string
	EmailVerifiedAt   *time.Time
	Provider          models.Providers
	Password          *string
	HashPassword      *string
	Type              models.ProviderTypes
	ProviderAccountID string
	UserId            *uuid.UUID
	AccessToken       *string
	RefreshToken      *string
}

type AuthenticateUserState struct {
	User    *models.User
	Account *models.UserAccount
	error   error
}

type UserInfoDto struct {
	User        *models.User `db:"user" json:"user"`
	Roles       []string     `db:"roles" json:"roles"`
	Permissions []string     `db:"permissions" json:"permissions"`
}

func (r *AuthenticateUserState) Error() string {
	return r.error.Error()
}

type TokenType string

const (
	AccessTokenType        TokenType = "access_token"
	PasswordResetTokenType TokenType = "password_reset_token"
	VerificationTokenType  TokenType = "verification_token"
	RefreshTokenType       TokenType = "refresh_token"
	StateTokenType         TokenType = "state_token"
)

type OAuthProviders string

const (
	ProvidersGoogle OAuthProviders = "google"
	// ProvidersApple    OAuthProviders = "apple"
	// ProvidersFacebook OAuthProviders = "facebook"
	ProvidersGithub OAuthProviders = "github"
	// ProvidersCredentials OAuthProviders = "credentials"
)

// const

type TokenDto struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in" example:"3600"`
	TokenType   string `json:"token_type" example:"Bearer"`
	// Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
}

type AuthenticatedDTO struct {
	User   *models.User `json:"user"`
	Tokens TokenDto     `json:"tokens"` //core.TokenDto `json:"tokens"`
}

type RecordOAuth2LoginForm struct {
	// collection *core.Collection

	// Additional data that will be used for creating a new auth record
	// if an existing OAuth2 account doesn't exist.
	CreateData map[string]any `form:"createData" json:"createData"`

	// The name of the OAuth2 client provider (eg. "google")
	Provider string `form:"provider" json:"provider"`

	// The authorization code returned from the initial request.
	Code string `form:"code" json:"code"`

	// The optional PKCE code verifier as part of the code_challenge sent with the initial request.
	CodeVerifier string `form:"codeVerifier" json:"codeVerifier"`

	// The redirect url sent with the initial request.
	RedirectURL string `form:"redirectURL" json:"redirectURL"`

	// @todo
	// deprecated: use RedirectURL instead
	// RedirectUrl will be removed after dropping v0.22 support
	RedirectUrl string `form:"redirectUrl" json:"redirectUrl"`
}
