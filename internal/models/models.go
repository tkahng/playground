package models

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	_          struct{}   `db:"tokens" schema:"auth" json:"-"`
	ID         uuid.UUID  `db:"id" json:"id"`
	Type       TokenTypes `db:"type" json:"type"`
	UserID     *uuid.UUID `db:"user_id" json:"user_id"`
	Otp        *string    `db:"otp" json:"otp"`
	Identifier string     `db:"identifier" json:"identifier"`
	Expires    time.Time  `db:"expires" json:"expires"`
	Token      string     `db:"token" json:"token"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updated_at"`
	User       *User      `db:"users" src:"user_id" dest:"id" table:"auth.users" json:"user,omitempty"`
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
	_                struct{}   `db:"media" schema:"storage" json:"-"`
	ID               uuid.UUID  `db:"id" json:"id"`
	StorageKey       string     `db:"storage_key" json:"storage_key"`
	PublicURL        *string    `db:"public_url" json:"public_url"`
	MimeType         string     `db:"mime_type" json:"mime_type"`
	Size             int64      `db:"size" json:"size"`
	OriginalFilename string     `db:"original_filename" json:"original_filename"`
	AltText          *string    `db:"alt_text" json:"alt_text"`
	Width            *int       `db:"width" json:"width"`
	Height           *int       `db:"height" json:"height"`
	UserID           *uuid.UUID `db:"user_id" json:"user_id"`
	// Legacy columns retained for DB compatibility — use StorageKey instead.
	Disk      string `db:"disk" json:"disk"`
	Directory string `db:"directory" json:"directory"`
	Filename  string `db:"filename" json:"filename"`
	Extension string `db:"extension" json:"extension"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	User      *User     `db:"users" src:"user_id" dest:"id" table:"auth.users" json:"user,omitempty"`
}

type MediaAttachment struct {
	_          struct{}  `db:"media_attachments" schema:"storage" json:"-"`
	ID         uuid.UUID `db:"id" json:"id"`
	MediaID    uuid.UUID `db:"media_id" json:"media_id"`
	EntityType string    `db:"entity_type" json:"entity_type"`
	EntityID   uuid.UUID `db:"entity_id" json:"entity_id"`
	Slot       string    `db:"slot" json:"slot"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type AiUsage struct {
	_                struct{}   `db:"ai_usages" schema:"app" json:"-"`
	ID               uuid.UUID  `db:"id,pk" json:"id"`
	UserID           uuid.UUID  `db:"user_id" json:"user_id"`
	TeamMemberID     *uuid.UUID `db:"team_member_id" json:"team_member_id"`
	TeamID           *uuid.UUID `db:"team_id" json:"team_id"`
	PromptTokens     int64      `db:"prompt_tokens" json:"prompt_tokens"`
	CompletionTokens int64      `db:"completion_tokens" json:"completion_tokens"`
	TotalTokens      int64      `db:"total_tokens" json:"total_tokens"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
	User             *User      `db:"user" src:"user_id" dest:"id" table:"auth.users" json:"user,omitempty"`
}
