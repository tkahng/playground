package models

import (
	"time"

	"github.com/google/uuid"
)

type AuthAccount struct {
	_                 struct{}      `db:"auth.accounts" json:"-"`
	ID                uuid.UUID     `db:"id" json:"id"`
	UserID            uuid.UUID     `db:"user_id" json:"user_id"`
	Type              ProviderTypes `db:"type,quote" json:"type"`
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
	User              *User         `db:"user,quote" relation:"src=user_id,dest=id,table=auth.users" json:"user,omitempty"`
	// User              *User         `db:"user" src:"user_id" dest:"id" table:"users" json:"user,omitempty"`
}
