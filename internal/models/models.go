package models

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	_          struct{}   `db:"tokens" json:"-"`
	ID         uuid.UUID  `db:"id" json:"id"`
	Type       TokenTypes `db:"type" json:"type"`
	UserID     *uuid.UUID `db:"user_id" json:"user_id"`
	Otp        *string    `db:"otp" json:"otp"`
	Identifier string     `db:"identifier" json:"identifier"`
	Expires    time.Time  `db:"expires" json:"expires"`
	Token      string     `db:"token" json:"token"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updated_at"`
	User       *User      `db:"users" src:"user_id" dest:"id" table:"users" json:"user,omitempty"`
}

type TokenTypes string

const (
	TokenTypesAccessToken           TokenTypes = "access_token"
	TokenTypesRecoveryToken         TokenTypes = "recovery_token"
	TokenTypesInviteToken           TokenTypes = "invite_token"
	TokenTypesReauthenticationToken TokenTypes = "reauthentication_token"
	TokenTypesRefreshToken          TokenTypes = "refresh_token"
	TokenTypesVerificationToken     TokenTypes = "verification_token"
	TokenTypesPasswordResetToken    TokenTypes = "password_reset_token"
	TokenTypesStateToken            TokenTypes = "state_token"
)

type Medium struct {
	_                struct{}   `db:"media" json:"-"`
	ID               uuid.UUID  `db:"id" json:"id"`
	UserID           *uuid.UUID `db:"user_id" json:"user_id"`
	Disk             string     `db:"disk" json:"disk"`
	Directory        string     `db:"directory" json:"directory"`
	Filename         string     `db:"filename" json:"filename"`
	OriginalFilename string     `db:"original_filename" json:"original_filename"`
	Extension        string     `db:"extension" json:"extension"`
	MimeType         string     `db:"mime_type" json:"mime_type"`
	Size             int64      `db:"size" json:"size"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
	User             *User      `db:"users" src:"user_id" dest:"id" table:"users" json:"user,omitempty"`
}

type AiUsage struct {
	_                struct{}  `db:"ai_usages" json:"-"`
	ID               uuid.UUID `db:"id,pk" json:"id"`
	UserID           uuid.UUID `db:"user_id" json:"user_id"`
	PromptTokens     int64     `db:"prompt_tokens" json:"prompt_tokens"`
	CompletionTokens int64     `db:"completion_tokens" json:"completion_tokens"`
	TotalTokens      int64     `db:"total_tokens" json:"total_tokens"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
	User             *User     `db:"user" src:"user_id" dest:"id" table:"users" json:"user,omitempty"`
}
