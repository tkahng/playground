package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/token"
)

type (
	SignupInput struct {
		Email    string  `json:"email" required:"true" format:"email" maxLength:"100"`
		Name     *string `json:"name"`
		Password string  `json:"password" required:"true" minLength:"8" maxLength:"100"`
		Verified bool    `json:"verified" required:"false" default:"false"`
	}
	SigninInput struct {
		Email    string `json:"email" required:"true" format:"email" maxLength:"100"`
		Password string `json:"password" required:"true" minLength:"8" maxLength:"100"`
	}
	SignoutInput struct {
		RefreshToken string `json:"refresh_token" required:"true"`
	}
)

type CredentialsAuthenticator interface {
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
}

// Signup implements AuthService.
// Signup credentials user.
//
// - if user with email does not exist, it will create a new user and a credentials account.
//
// - if user with email exists, and they have another credentials account or oauth account, it will return error.
func (a *AuthServiceImpl) Signup(ctx context.Context, params *SignupInput) (*models.UserInfoTokens, error) {
	// check if user with email exists.
	existingUser, err := a.adapter.User().FindUser(ctx, &stores.UserFilter{
		Emails: []string{params.Email},
	})
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, shared.ErrUserExists
	}
	// hash password
	hashedPassword, err := a.hash.Hash(params.Password)
	if err != nil {
		return nil, err
	}
	// create user and account. these should run inside a transaction.
	// if params.Verified is true, set email_verified_at and skip email verification
	txErr := a.adapter.RunInTxCtx(ctx, func(txCtx context.Context) error {
		u := &models.User{
			Name:  params.Name,
			Email: params.Email,
		}
		// if params.Verified is true, set email_verified_at
		if params.Verified {
			now := time.Now()
			u.EmailVerifiedAt = &now
		}
		user, err := a.adapter.User().CreateUser(txCtx, u)
		if err != nil {
			return err
		}
		_, err = a.adapter.UserAccount().CreateUserAccount(txCtx, &models.UserAccount{
			UserID:            user.ID,
			Provider:          models.ProvidersCredentials,
			ProviderAccountID: user.ID.String(),
			Type:              models.ProviderTypeCredentials,
			Password:          &hashedPassword,
		})
		if err != nil {
			return err
		}
		// if params.Verified is true, skip email verification
		if params.Verified {
			return nil
		}
		err = a.SendEmailVerification(txCtx, params.Email)
		if err != nil {
			return err
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}
	tokens, err := a.GenerateAuthTokens(ctx, params.Email)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// Signin implements AuthService.
// Signin credentials user.
// If user with given email exists and has a credentials account, it will check password.
func (a *AuthServiceImpl) Signin(ctx context.Context, params *SigninInput) (*models.UserInfoTokens, error) {
	user, err := a.adapter.User().FindUser(ctx, &stores.UserFilter{
		Emails: []string{params.Email},
	})
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, shared.ErrUserNotFound
	}
	account, err := a.adapter.UserAccount().FindUserAccount(ctx, &stores.UserAccountFilter{
		UserIds:   []uuid.UUID{user.ID},
		Providers: []models.Providers{models.ProvidersCredentials},
	})
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, shared.ErrAccountNotFound
	}
	match, err := a.hash.Verify(params.Password, *account.Password)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, shared.ErrPasswordIncorrect
	}
	tokens, err := a.GenerateAuthTokens(ctx, params.Email)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// Signout implements AuthService.
func (a *AuthServiceImpl) Signout(ctx context.Context, refreshToken string) error {
	_, err := a.token.ValidateToken(ctx, &token.ValidateTokenOptions{
		Token: refreshToken,
		Type:  models.TokenTypesRefreshToken,
	})
	if err != nil {
		return fmt.Errorf("error verifying refresh token: %w", err)
	}
	return nil
}
