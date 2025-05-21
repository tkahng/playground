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
	User              *User         `db:"users" src:"user_id" dest:"id" table:"users" json:"user,omitempty"`
}

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

type Task struct {
	_         struct{}  `db:"tasks" json:"-"`
	ID        uuid.UUID `db:"id" json:"id"`
	CreatedBy uuid.UUID `db:"created_by" json:"created_by"`
	TeamID    uuid.UUID `db:"team_id" json:"team_id"`
	// UserID      uuid.UUID    `db:"user_id" json:"user_id"`
	ProjectID   uuid.UUID    `db:"project_id" json:"project_id"`
	Name        string       `db:"name" json:"name"`
	Description *string      `db:"description" json:"description"`
	Status      TaskStatus   `db:"status" json:"status" enum:"todo,in_progress,done"`
	Order       float64      `db:"order" json:"order"`
	ParentID    *uuid.UUID   `db:"parent_id" json:"parent_id"`
	CreatedAt   time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at" json:"updated_at"`
	Children    []*Task      `db:"children" src:"id" dest:"parent_id" table:"tasks" json:"children,omitempty"`
	User        *User        `db:"user" src:"user_id" dest:"id" table:"users" json:"user,omitempty"`
	Project     *TaskProject `db:"project" src:"project_id" dest:"id" table:"task_projects" json:"project,omitempty"`
}

type TaskProject struct {
	_         struct{}  `db:"task_projects" json:"-"`
	ID        uuid.UUID `db:"id" json:"id"`
	CreatedBy uuid.UUID `db:"created_by" json:"created_by"`
	TeamID    uuid.UUID `db:"team_id" json:"team_id"`
	// UserID      uuid.UUID         `db:"user_id" json:"user_id"`
	Name        string            `db:"name" json:"name"`
	Description *string           `db:"description" json:"description"`
	Status      TaskProjectStatus `db:"status" json:"status" enum:"todo,in_progress,done"`
	Order       float64           `db:"order" json:"order"`
	CreatedAt   time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `db:"updated_at" json:"updated_at"`
	User        *User             `db:"user" src:"user_id" dest:"id" table:"users" json:"user,omitempty"`
	Tasks       []*Task           `db:"tasks" src:"id" dest:"project_id" table:"tasks" json:"tasks,omitempty"`
}

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

type TaskStatus string

// Enum values for TaskProjectStatus
const (
	TaskProjectStatusTodo       TaskProjectStatus = "todo"
	TaskProjectStatusInProgress TaskProjectStatus = "in_progress"
	TaskProjectStatusDone       TaskProjectStatus = "done"
)

type TaskProjectStatus string

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

type Notification struct {
	_         struct{}       `db:"notifications" json:"-"`
	ID        uuid.UUID      `db:"id,pk" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	Channel   string         `db:"channel" json:"channel"`
	UserID    *uuid.UUID     `db:"user_id" json:"user_id"`
	Content   map[string]any `db:"content" json:"content"`
	Type      string         `db:"type" json:"type"`
	// User      *User          `db:"user" src:"user_id" dest:"id" table:"users" json:"user,omitempty"`
}
