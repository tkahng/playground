package crudModels

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	_               struct{}      `db:"users" json:"-"`
	ID              uuid.UUID     `db:"id" json:"id"`
	Email           string        `db:"email" json:"email"`
	EmailVerifiedAt *time.Time    `db:"email_verified_at" json:"email_verified_at"`
	Name            *string       `db:"name" json:"name"`
	Image           *string       `db:"image" json:"image"`
	CreatedAt       time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time     `db:"updated_at" json:"updated_at"`
	Accounts        []UserAccount `db:"accounts" src:"id" dest:"user_id" table:"user_accounts" json:"-"`
	Roles           []Role        `db:"roles" src:"id" dest:"user_id" table:"roles" through:"user_roles,role_id,id" json:"-"`
	Permissions     []Permission  `db:"permissions" src:"id" dest:"user_id" table:"permissions" through:"user_permissions,permission_id,id" json:"-"`
}
type UserRole struct {
	_      struct{}  `db:"user_roles" json:"-"`
	UserID uuid.UUID `db:"user_id,pk" json:"user_id"`
	RoleID uuid.UUID `db:"role_id,pk" json:"role_id"`
}
type ProductRole struct {
	_         struct{}  `db:"product_roles" json:"-"`
	ProductID string    `db:"product_id,pk" json:"product_id"`
	RoleID    uuid.UUID `db:"role_id,pk" json:"role_id"`
}
type UserPermission struct {
	_            struct{}  `db:"user_permissions" json:"-"`
	UserID       uuid.UUID `db:"user_id,pk" json:"user_id"`
	PermissionID uuid.UUID `db:"permission_id,pk" json:"permission_id"`
}
type Role struct {
	_           struct{}     `db:"roles" json:"-"`
	ID          uuid.UUID    `db:"id" json:"id"`
	Name        string       `db:"name" json:"name"`
	Description *string      `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at" json:"updated_at"`
	Permissions []Permission `db:"permissions" src:"id" dest:"role_id" table:"permissions" through:"role_permissions,permission_id,id" json:"-"`
	Users       []User       `db:"users" src:"id" dest:"role_id" table:"users" through:"user_roles,user_id,id" json:"-"`
}

type RolePermission struct {
	_            struct{}  `db:"role_permissions" json:"-"`
	RoleID       uuid.UUID `db:"role_id,pk" json:"role_id"`
	PermissionID uuid.UUID `db:"permission_id,pk" json:"permission_id"`
}

type Permission struct {
	_           struct{}  `db:"permissions" json:"-"`
	ID          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type ProviderTypes string

const (
	ProviderTypeOAuth       ProviderTypes = "oauth"
	ProviderTypeCredentials ProviderTypes = "credentials"
)

type Providers string

const (
	ProvidersGoogle      Providers = "google"
	ProvidersApple       Providers = "apple"
	ProvidersFacebook    Providers = "facebook"
	ProvidersGithub      Providers = "github"
	ProvidersCredentials Providers = "credentials"
)

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
	User              User          `db:"users" src:"user_id" dest:"id" table:"users" json:"-"`
}

type Token struct {
	_          struct{}   `db:"tokens" json:"-"`
	ID         uuid.UUID  `db:"id,pk" json:"id"`
	Type       TokenTypes `db:"type" json:"type"`
	UserID     *uuid.UUID `db:"user_id" json:"user_id"`
	Otp        *string    `db:"otp" json:"otp"`
	Identifier string     `db:"identifier" json:"identifier"`
	Expires    time.Time  `db:"expires" json:"expires"`
	Token      string     `db:"token" json:"token"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updated_at"`
	User       *User      `db:"users" src:"user_id" dest:"id" table:"users" json:"-"`
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
	_           struct{}   `db:"tasks" json:"-"`
	ID          uuid.UUID  `db:"id,pk" json:"id"`
	UserID      uuid.UUID  `db:"user_id" json:"user_id"`
	ProjectID   uuid.UUID  `db:"project_id" json:"project_id"`
	Name        string     `db:"name" json:"name"`
	Description *string    `db:"description" json:"description"`
	Status      TaskStatus `db:"status" json:"status" enum:"todo,in_progress,done"`
	Order       float64    `db:"order" json:"order"`
	ParentID    *uuid.UUID `db:"parent_id" json:"parent_id"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

type TaskProject struct {
	_           struct{}          `db:"task_projects" json:"-"`
	ID          uuid.UUID         `db:"id,pk" json:"id"`
	UserID      uuid.UUID         `db:"user_id" json:"user_id"`
	Name        string            `db:"name" json:"name"`
	Description *string           `db:"description" json:"description"`
	Status      TaskProjectStatus `db:"status" json:"status" enum:"todo,in_progress,done"`
	Order       float64           `db:"order" json:"order"`
	CreatedAt   time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `db:"updated_at" json:"updated_at"`
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
