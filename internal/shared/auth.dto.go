package shared

import (
	"time"

	"github.com/aarondl/opt/null"
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

type AuthenticateUserState struct {
	User    *models.User
	Account *models.UserAccount
}

type UserInfoDto struct {
	User        models.User        `db:"user" json:"user"`
	Roles       []string           `db:"roles" json:"roles"`
	Permissions []string           `db:"permissions" json:"permissions"`
	Providers   []models.Providers `db:"providers" json:"providers"`
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
		result = append(result, ToProvider(provider))
	}
	return result
}

func ToModelProvidersArray(providers []Providers) []models.Providers {
	var result []models.Providers
	for _, provider := range providers {
		result = append(result, ToModelProvider(provider))
	}
	return result
}
func ToUserInfo(user *UserInfoDto) *UserInfo {
	if user == nil {
		return nil
	}
	return &UserInfo{
		User:        *ToUser(&user.User),
		Roles:       user.Roles,
		Permissions: user.Permissions,
		Providers:   ToProvidersArray(user.Providers),
	}
}

func ToUserInfoDto(user *UserInfo) *UserInfoDto {
	if user == nil {
		return nil
	}
	return &UserInfoDto{
		User:        *ToModelUser(&user.User),
		Roles:       user.Roles,
		Permissions: user.Permissions,
		Providers:   ToModelProvidersArray(user.Providers),
	}
}

type TokenType string

const (
	AccessTokenType        TokenType = "access_token"
	PasswordResetTokenType TokenType = "password_reset_token"
	VerificationTokenType  TokenType = "verification_token"
	RefreshTokenType       TokenType = "refresh_token"
	StateTokenType         TokenType = "state_token"
)

func (t TokenType) String() string {
	return string(t)
}

func ToTokenType(t models.TokenTypes) TokenType {
	switch t {
	case models.TokenTypesReauthenticationToken:
		return AccessTokenType
	case models.TokenTypesRefreshToken:
		return RefreshTokenType
	case models.TokenTypesVerificationToken:
		return VerificationTokenType
	case models.TokenTypesPasswordResetToken:
		return PasswordResetTokenType
	case models.TokenTypesStateToken:
		return StateTokenType
	}
	return AccessTokenType
}

func ToModelTokenType(t TokenType) models.TokenTypes {
	switch t {
	// case AccessTokenType:
	// 	return models.TokenTypesReauthenticationToken
	case PasswordResetTokenType:
		return models.TokenTypesPasswordResetToken
	case VerificationTokenType:
		return models.TokenTypesVerificationToken
	case RefreshTokenType:
		return models.TokenTypesRefreshToken
	case StateTokenType:
		return models.TokenTypesStateToken
	default:
		return models.TokenTypesReauthenticationToken
	}
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

func ToModelToken(dto *Token) *models.Token {
	if dto == nil {
		return nil
	}
	return &models.Token{
		ID:         dto.ID,
		Type:       ToModelTokenType(dto.Type),
		UserID:     null.FromPtr(dto.UserID),
		Otp:        null.FromPtr(dto.Otp),
		Identifier: dto.Identifier,
		Expires:    dto.Expires,
		Token:      dto.Token,
		CreatedAt:  dto.CreatedAt,
		UpdatedAt:  dto.UpdatedAt,
	}
}

func ToToken(model *models.Token) *Token {
	if model == nil {
		return nil
	}
	return &Token{
		ID:         model.ID,
		Type:       ToTokenType(model.Type),
		UserID:     model.UserID.Ptr(),
		Otp:        model.Otp.Ptr(),
		Identifier: model.Identifier,
		Expires:    model.Expires,
		Token:      model.Token,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}
}
