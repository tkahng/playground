package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
)

type SignupInput struct {
	Email    string  `json:"email" required:"true" format:"email" maxLength:"100"`
	Name     *string `json:"name"`
	Password string  `json:"password" required:"true" minLength:"8" maxLength:"100"`
}
type SigninInput struct {
	Email    string `json:"email" required:"true" format:"email" maxLength:"100"`
	Password string `json:"password" required:"true" minLength:"8" maxLength:"100"`
}

type SignoutInput struct {
	RefreshToken string `json:"refresh_token" required:"true"`
}
type OAuth2SigninInput struct {
	Email             string
	Name              *string
	AvatarUrl         *string
	EmailVerifiedAt   *time.Time
	Provider          models.Providers
	ProviderAccountID string
	UserId            *uuid.UUID
	AccessToken       *string
	RefreshToken      *string
}

type AuthService interface {
	// Signup credentials user.
	//
	// - if user with email does not exist, it will create a new user and a credentials account.
	//
	// - if user with email exists, and they have another credentials account or oauth account, it will return error.
	Signup(ctx context.Context, params *SignupInput) (*models.UserInfoTokens, error)
	// Signin credentials user.
	// If user with given email exists and has a credentials account, it will check password.
	Signin(ctx context.Context, params *SigninInput) (*models.UserInfoTokens, error)
	// Signout user.
	// if given refresh token is valid, it will delete the refresh token.
	Signout(ctx context.Context, refreshToken string) error
	// OAuth2Signin user.
	//
	// - if user with email does not exist, it will create a new user and a oauth account.
	//
	// - if user with email exists, and they have another oauth account, it will update the oauth account.
	OAuth2Signin(ctx context.Context, params *OAuth2SigninInput) (*models.UserInfoTokens, error)
}
