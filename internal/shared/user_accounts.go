package shared

import (
	"time"

	"github.com/aarondl/opt/null"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db/models"
)

// enum:oauth,credentials
type ProviderTypes string

const (
	ProviderTypeOAuth       ProviderTypes = "oauth"
	ProviderTypeCredentials ProviderTypes = "credentials"
)

func (p ProviderTypes) String() string {
	return string(p)
}

func ToProviderType(p models.ProviderTypes) ProviderTypes {
	switch p {
	case models.ProviderTypesOauth:
		return ProviderTypeOAuth
	case models.ProviderTypesCredentials:
		return ProviderTypeCredentials
	}
	return ProviderTypeCredentials
}

func ToModelProviderType(p ProviderTypes) models.ProviderTypes {
	switch p {
	case ProviderTypeOAuth:
		return models.ProviderTypesOauth
	case ProviderTypeCredentials:
		return models.ProviderTypesCredentials
	}
	return models.ProviderTypesCredentials
}

// enum:google,apple,facebook,github,credentials
type Providers string

const (
	ProvidersGoogle      Providers = "google"
	ProvidersApple       Providers = "apple"
	ProvidersFacebook    Providers = "facebook"
	ProvidersGithub      Providers = "github"
	ProvidersCredentials Providers = "credentials"
)

func (p Providers) String() string {
	return string(p)
}

func ToProvider(p models.Providers) Providers {
	switch p {
	case models.ProvidersGoogle:
		return ProvidersGoogle
	case models.ProvidersApple:
		return ProvidersApple
	case models.ProvidersFacebook:
		return ProvidersFacebook
	case models.ProvidersGithub:
		return ProvidersGithub
	case models.ProvidersCredentials:
		return ProvidersCredentials
	default:
		return ProvidersCredentials
	}
}

func ToModelProvider(p Providers) models.Providers {
	switch p {
	case ProvidersGoogle:
		return models.ProvidersGoogle
	case ProvidersApple:
		return models.ProvidersApple
	case ProvidersFacebook:
		return models.ProvidersFacebook
	case ProvidersGithub:
		return models.ProvidersGithub
	}
	return models.ProvidersCredentials
}

type UserAccount struct {
	ID                uuid.UUID     `db:"id,pk" json:"id"`
	UserID            uuid.UUID     `db:"user_id" json:"user_id"`
	Type              ProviderTypes `db:"type" json:"type"`
	Provider          Providers     `db:"provider" json:"provider"`
	ProviderAccountID string        `db:"provider_account_id" json:"provider_account_id"`
	Password          *string       `db:"password" json:"password"`
	RefreshToken      *string       `db:"refresh_token" json:"refresh_token"`
	AccessToken       *string       `db:"access_token" json:"access_token"`
	ExpiresAt         *int64        `db:"expires_at" json:"expires_at"`
	IDToken           *string       `db:"id_token" json:"id_token"`
	Scope             *string       `db:"scope" json:"scope"`
	SessionState      *string       `db:"session_state" json:"session_state"`
	TokenType         *string       `db:"token_type" json:"token_type"`
	CreatedAt         time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time     `db:"updated_at" json:"updated_at"`
}

type UserAccountOutput struct {
	ID                uuid.UUID     `db:"id,pk" json:"id"`
	UserID            uuid.UUID     `db:"user_id" json:"user_id"`
	Type              ProviderTypes `db:"type" json:"type" enum:"oauth,credentials"`
	Provider          Providers     `db:"provider" json:"provider" enum:"google,apple,facebook,github,credentials"`
	ProviderAccountID string        `db:"provider_account_id" json:"provider_account_id"`
	CreatedAt         time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time     `db:"updated_at" json:"updated_at"`
}

func ToUserAccountOutput(u *models.UserAccount) *UserAccountOutput {
	if u == nil {
		return nil
	}
	return &UserAccountOutput{
		ID:                u.ID,
		UserID:            u.UserID,
		Type:              ToProviderType(u.Type),
		Provider:          ToProvider(u.Provider),
		ProviderAccountID: u.ProviderAccountID,
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
	}
}

func ToUserAccount(u *models.UserAccount) *UserAccount {
	if u == nil {
		return nil
	}
	return &UserAccount{
		ID:                u.ID,
		UserID:            u.UserID,
		Type:              ToProviderType(u.Type),
		Provider:          ToProvider(u.Provider),
		ProviderAccountID: u.ProviderAccountID,
		Password:          u.Password.Ptr(),
		RefreshToken:      u.RefreshToken.Ptr(),
		AccessToken:       u.AccessToken.Ptr(),
		ExpiresAt:         u.ExpiresAt.Ptr(),
		IDToken:           u.IDToken.Ptr(),
		Scope:             u.Scope.Ptr(),
		SessionState:      u.SessionState.Ptr(),
		TokenType:         u.TokenType.Ptr(),
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
	}
}

func ToUserAccountModel(u *UserAccount) *models.UserAccount {
	if u == nil {
		return nil
	}
	return &models.UserAccount{
		ID:                u.ID,
		UserID:            u.UserID,
		Type:              ToModelProviderType(u.Type),
		Provider:          ToModelProvider(u.Provider),
		ProviderAccountID: u.ProviderAccountID,
		Password:          null.FromPtr(u.Password),
		RefreshToken:      null.FromPtr(u.RefreshToken),
		AccessToken:       null.FromPtr(u.AccessToken),
		ExpiresAt:         null.FromPtr(u.ExpiresAt),
		IDToken:           null.FromPtr(u.IDToken),
		Scope:             null.FromPtr(u.Scope),
		SessionState:      null.FromPtr(u.SessionState),
		TokenType:         null.FromPtr(u.TokenType),
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
	}
}

type UserAccountListFilter struct {
	Providers     []Providers     `query:"providers,omitempty" required:"false" uniqueItems:"true" minimum:"1" maximum:"100" enum:"google,apple,facebook,github,credentials"`
	ProviderTypes []ProviderTypes `query:"provider_types,omitempty" required:"false" uniqueItems:"true" minimum:"1" maximum:"100" enum:"oauth,credentials"`
	Q             string          `query:"q,omitempty" required:"false"`
	Ids           []string        `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	UserId        string          `query:"user_id,omitempty" required:"false" format:"uuid"`
}
type UserAccountListParams struct {
	PaginatedInput
	UserAccountListFilter
	SortParams
}
