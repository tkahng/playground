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
type AuthenticationInput struct {
	Email             string
	Name              *string
	AvatarUrl         *string
	EmailVerifiedAt   *time.Time
	Provider          Providers
	Password          *string
	HashPassword      *string
	Type              ProviderTypes
	ProviderAccountID string
	UserId            *uuid.UUID
	AccessToken       *string
	RefreshToken      *string
}

type UserInfo struct {
	User        User        `db:"user" json:"user"`
	Roles       []string    `db:"roles" json:"roles"`
	Permissions []string    `db:"permissions" json:"permissions"`
	Providers   []Providers `db:"providers" json:"providers" enum:"google,apple,facebook,github,credentials"`
}

func ToProvidersArray(providers []models.Providers) []Providers {
	var result []Providers
	for _, provider := range providers {
		result = append(result, Providers(provider))
	}
	return result
}

func ToModelProvidersArray(providers []Providers) []models.Providers {
	var result []models.Providers
	for _, provider := range providers {
		result = append(result, models.Providers(provider))
	}
	return result
}

type TokenType string

const (
	TokenTypesAccessToken           TokenType = "access_token"
	TokenTypesRecoveryToken         TokenType = "recovery_token"
	TokenTypesInviteToken           TokenType = "invite_token"
	TokenTypesReauthenticationToken TokenType = "reauthentication_token"
	TokenTypesRefreshToken          TokenType = "refresh_token"
	TokenTypesVerificationToken     TokenType = "verification_token"
	TokenTypesPasswordResetToken    TokenType = "password_reset_token"
	TokenTypesStateToken            TokenType = "state_token"
)

func (t TokenType) String() string {
	return string(t)
}

type OAuthProviders string

const (
	OAuthProvidersGoogle OAuthProviders = "google"
	// ProvidersApple    OAuthProviders = "apple"
	// ProvidersFacebook OAuthProviders = "facebook"
	OAuthProvidersGithub OAuthProviders = "github"
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
	User        *User              `json:"user"`
	Permissions []string           `json:"permissions"`
	Roles       []string           `json:"roles"`
	Providers   []models.Providers `json:"providers"`
	Tokens      TokenDto           `json:"tokens"` //core.TokenDto `json:"tokens"`
}

type UserInfoTokens struct {
	UserInfo
	Tokens TokenDto `json:"tokens"`
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

type CreateTokenDTO struct {
	Type       TokenType  `db:"type" json:"type"`
	Identifier string     `db:"identifier" json:"identifier"`
	Expires    time.Time  `db:"expires" json:"expires"`
	Token      string     `db:"token" json:"token"`
	ID         *uuid.UUID `db:"id" json:"id"`
	UserID     *uuid.UUID `db:"user_id" json:"user_id"`
	Otp        *string    `db:"otp" json:"otp"`
}

type Token struct {
	ID         uuid.UUID  `db:"id,pk" json:"id"`
	Type       TokenType  `db:"type" json:"type"`
	UserID     *uuid.UUID `db:"user_id" json:"user_id"`
	Otp        *string    `db:"otp" json:"otp"`
	Identifier string     `db:"identifier" json:"identifier"`
	Expires    time.Time  `db:"expires" json:"expires"`
	Token      string     `db:"token" json:"token"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updated_at"`
}

type PasswordResetInput struct {
	PreviousPassword string `form:"previous_password" json:"previous_password" required:"true" minimum:"8" maximum:"64"`
	NewPassword      string `form:"new_password" json:"new_password" required:"true" minimum:"8" maximum:"64"`
}
