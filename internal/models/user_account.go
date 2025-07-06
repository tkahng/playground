package models

import (
	"time"

	"github.com/google/uuid"
)

type ProviderTypes string

const (
	ProviderTypeOAuth       ProviderTypes = "oauth"
	ProviderTypeCredentials ProviderTypes = "credentials"
)

func (p ProviderTypes) String() string {
	return string(p)
}

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

type UserAccount struct {
	_                 struct{}      `db:"user_accounts" json:"-"`
	ID                uuid.UUID     `db:"id" json:"id"`
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
	User              *User         `db:"user" src:"user_id" dest:"id" table:"users" json:"user,omitempty"`
}

type userAccountTable struct {
	Columns           []string
	ID                string
	UserID            string
	Type              string
	Provider          string
	ProviderAccountID string
	Password          string
	RefreshToken      string
	AccessToken       string
	ExpiresAt         string
	IDToken           string
	Scope             string
	SessionState      string
	TokenType         string
	CreatedAt         string
	UpdatedAt         string
	User              string
}

var UserAccountTable = userAccountTable{
	Columns: []string{
		"id",
		"user_id",
		"type",
		"provider",
		"provider_account_id",
		"password",
		"refresh_token",
		"access_token",
		"expires_at",
		"id_token",
		"scope",
		"session_state",
		"token_type",
		"created_at",
		"updated_at",
		"user",
	},
	ID:                "id",
	UserID:            "user_id",
	Type:              "type",
	Provider:          "provider",
	ProviderAccountID: "provider_account_id",
	Password:          "password",
	RefreshToken:      "refresh_token",
	AccessToken:       "access_token",
	ExpiresAt:         "expires_at",
	IDToken:           "id_token",
	Scope:             "scope",
	SessionState:      "session_state",
	TokenType:         "token_type",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
	User:              "user",
}
