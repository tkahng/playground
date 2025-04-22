package core

import (
	"context"
	"fmt"

	"github.com/alexedwards/argon2id"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/security"
)

type AuthActions interface {
	// CreateAuthenticationToken(ctx context.Context, user *AuthenticationPayload) (string, error)
	// CreateAndSaveRefreshToken(ctx context.Context, user *RefreshTokenPayload) (string, error)
	// CreateAndSaveVerificationToken(ctx context.Context, user *OtpPayload) (string, error)
	// CreateAndSavePasswordResetToken(ctx context.Context, user *OtpPayload) (string, error)
	// CreateAndSaveStateToken(ctx context.Context, user *ProviderStatePayload) (string, error)
	Signin(ctx context.Context, email string, password string) (*shared.AuthenticatedDTO, error)
	// Signup(ctx context.Context, email string, password string) (*shared.AuthenticatedDTO, error)
	// OAuth2Signin(ctx context.Context, code string, state string) (*shared.AuthenticatedDTO, error)
}

type AuthActionsBase struct {
	adapter AuthAdapter
}

// func ()

func (a *AuthActionsBase) AuthenticateUser(ctx context.Context, params *shared.AuthenticateUserParams) (*shared.User, error) {
	var user *shared.User
	var account *shared.UserAccount
	var err error
	user, err = a.adapter.GetUserByEmail(ctx, params.Email)
	if err != nil {
		return nil, fmt.Errorf("error at getting user by email: %w", err)
	}
	if user != nil {
		provider := shared.ToProvider(params.Provider)
		account, err = a.adapter.GetUserAccount(ctx, user.ID, provider)
		if err != nil {
			return nil, fmt.Errorf("error at getting user account: %w", err)
		}
	}
	// if user does not exist, Create User and continue to create UserAccount ----------------------------------------------------------------------------------------------------
	if user == nil {
		user, err = a.adapter.CreateUser(ctx, &shared.User{
			Email:           params.Email,
			Name:            params.Name,
			Image:           params.AvatarUrl,
			EmailVerifiedAt: params.EmailVerifiedAt,
		})
		if err != nil {
			return nil, fmt.Errorf("error at creating user: %w", err)
		}
	}
	// if user exists, but account does not exist, Create UserAccount ----------------------------------------------------------------------------------------------------
	if account == nil {
		// if type is credentials, hash password and set params
		if params.Type == models.ProviderTypesCredentials {
			pw, err := security.CreateHash(*params.Password, argon2id.DefaultParams)
			if err != nil {
				return nil, fmt.Errorf("error at hashing password: %w", err)
			}
			params.HashPassword = &pw
		}
		err = a.adapter.LinkAccount(ctx, &shared.UserAccount{
			UserID:            user.ID,
			Type:              shared.ToProviderType(params.Type),
			Provider:          shared.ToProvider(params.Provider),
			ProviderAccountID: params.ProviderAccountID,
			Password:          params.HashPassword,
			AccessToken:       params.AccessToken,
			RefreshToken:      params.RefreshToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error at linking account: %w", err)
		}
		return user, nil
	}
	// if user exists and account exists, check if password is correct  or check if provider key is correct ----------------------------------------------------------------------------------------------------
	if params.Type == models.ProviderTypesCredentials {
		if params.Password == nil || account.Password == nil {
			return nil, fmt.Errorf("password or account password is nil")
		}
		if match, err := security.ComparePasswordAndHash(*params.Password, *account.Password); err != nil {
			return nil, fmt.Errorf("error at comparing password: %w", err)
		} else if !match {
			return nil, fmt.Errorf("password is incorrect")
		}
	}
	return user, nil
}
