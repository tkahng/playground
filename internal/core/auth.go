package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/alexedwards/argon2id"
	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/security"
)

var (
	ErrPasswordIncorrect  = shared.AppError{Status: 400, Message: "password incorrect"}
	ErrInvalidProviderKey = shared.AppError{Status: 400, Message: "invalid provider key"}
	ErrBadRequest         = shared.AppError{Status: 400, Message: "input is missing"}
)

func (a *BaseApp) CreateAuthTokens(ctx context.Context, db bob.Executor, payload *shared.UserInfoDto) (*shared.TokenDto, error) {
	if payload == nil || payload.User == nil {
		return nil, fmt.Errorf("payload is nil")
	}

	opts := a.Settings().Auth
	authToken, err := CreateAuthenticationToken(&AuthenticationPayload{
		UserId:      payload.User.ID,
		Email:       payload.User.Email,
		Roles:       payload.Roles,
		Permissions: payload.Permissions,
	}, opts.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("error creating auth token: %w", err)
	}

	tokenKey := security.GenerateTokenKey()

	refreshToken, err := CreateAndPersistRefreshToken(ctx, db, &RefreshTokenPayload{
		UserId: payload.User.ID,
		Email:  payload.User.Email,
		Token:  tokenKey,
	}, opts.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("error creating refresh token: %w", err)
	}
	return &shared.TokenDto{
		RefreshToken: refreshToken,
		AccessToken:  authToken,
		ExpiresIn:    opts.AccessToken.Duration,
		TokenType:    "Bearer",
	}, nil
}

func (app *BaseApp) AuthenticateUser(ctx context.Context, db bob.Executor, params *shared.AuthenticateUserParams, autoCreateUser bool) (*shared.AuthenticateUserState, error) {

	// Query User and UserAccount by email and provider ----------------------------------------------------------------------------------------------------
	result, err := repository.FindUserAccountByProviderAndEmail(ctx, db, params.Email, params.Provider)
	if err != nil {
		return nil, err
	}
	// if user does not exist, Create User and continue to create UserAccount ----------------------------------------------------------------------------------------------------
	if result.User == nil {
		if !autoCreateUser {
			return nil, fmt.Errorf("user not found")
		}
		user, err := repository.CreateUser(ctx, db, params)

		if err != nil {
			return nil, fmt.Errorf("error creating user: %w", err)
		}
		roles, err := repository.FindRolesByNames(ctx, db, []string{"basic"})
		if err != nil {
			return nil, fmt.Errorf("error finding user role: %w", err)
		}
		if len(roles) > 0 {
			err = repository.AssignRoles(ctx, db, user, roles...)
			if err != nil {
				return nil, fmt.Errorf("error assigning user role: %w", err)
			}
		}
		result.User = user
		result.Account = nil
		app.SendVerificationEmail(ctx, db, user, "http://localhost:8080")
	}
	// if user exists, but account does not exist, Create UserAccount ----------------------------------------------------------------------------------------------------
	if result.Account == nil {
		// if type is credentials, hash password and set params
		if params.Type == "credentials" {
			pw, err := security.CreateHash(*params.Password, argon2id.DefaultParams)
			if err != nil {
				return nil, fmt.Errorf("error at hashing password: %w", err)
			}
			params.HashPassword = &pw
		}
		// else just create account and return
		account, err := repository.CreateAccount(ctx, db, result.User, params)
		if err != nil {
			return nil, fmt.Errorf("error creating user account: %w", err)
		}
		result.Account = account

		return result, nil
	}
	// if user exists and account exists, check if password is correct  or check if provider key is correct ----------------------------------------------------------------------------------------------------
	if result.Account != nil {
		if params.Type == models.ProviderTypesCredentials {
			if params.Password == nil || result.Account.Password.IsNull() {
				return nil, ErrBadRequest
			}
			if match, err := security.ComparePasswordAndHash(*params.Password, *result.Account.Password.Ptr()); err != nil {
				return nil, fmt.Errorf("error comparing password: %w", err)
			} else if !match {
				return nil, ErrPasswordIncorrect
			} else {
				return result, nil
			}
		} else if params.Type == models.ProviderTypesOauth {
			if result.Account.ProviderAccountID == params.ProviderAccountID {
				return result, nil
			}
			return nil, ErrInvalidProviderKey
		} else {
			return nil, ErrBadRequest
		}
	}
	return nil, errors.New("unknown error")
}
