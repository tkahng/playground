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

// Providers is the type of authentication providers. enum:access_token,recovery_token,invite_token,reauthentication_token,refresh_token,verification_token,password_reset_token
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
