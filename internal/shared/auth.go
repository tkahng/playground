package shared

import (
	"time"

	"github.com/google/uuid"
)

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
	User        *User       `json:"user"`
	Permissions []string    `json:"permissions"`
	Roles       []string    `json:"roles"`
	Providers   []Providers `json:"providers"`
	Tokens      TokenDto    `json:"tokens"` //core.TokenDto `json:"tokens"`
}

type UserInfoTokens struct {
	UserInfo
	Tokens TokenDto `json:"tokens"`
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
