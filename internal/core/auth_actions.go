package core

import (
	"context"
	"fmt"

	"github.com/alexedwards/argon2id"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/security"
)

type AuthActions interface {
	// CreateAuthenticationToken(ctx context.Context, user *AuthenticationPayload) (string, error)

	VerifyAndParseOtpToken(ctx context.Context, emailType EmailType, token string) (*OtpClaims, error)
	// CreateAndSaveRefreshToken(ctx context.Context, user *RefreshTokenPayload) (string, error)
	// CreateAndSaveVerificationToken(ctx context.Context, user *OtpPayload) (string, error)
	// CreateAndSavePasswordResetToken(ctx context.Context, user *OtpPayload) (string, error)
	// CreateAndSaveStateToken(ctx context.Context, user *ProviderStatePayload) (string, error)
	// Signin(ctx context.Context, email string, password string) (*shared.AuthenticatedDTO, error)
	// Signup(ctx context.Context, email string, password string) (*shared.AuthenticatedDTO, error)
	// OAuth2Signin(ctx context.Context, code string, state string) (*shared.AuthenticatedDTO, error)
}

var _ AuthActions = (*AuthActionsBase)(nil)

type AuthActionsBase struct {
	authAdapter  *AuthAdapterBase
	authMailer   *AuthMailerBase
	tokenAdapter *TokenAdapterBase
	settings     *AppOptions
}

// VerifyAndUseVerificationToken implements AuthActions.
func (app *AuthActionsBase) VerifyAndParseOtpToken(ctx context.Context, emailType EmailType, token string) (*OtpClaims, error) {
	var opt TokenOption
	switch emailType {
	case EmailTypeVerify:
		opt = app.Settings().Auth.VerificationToken
	case EmailTypeConfirmPasswordReset:
		opt = app.Settings().Auth.PasswordResetToken
	case EmailTypeSecurityPasswordReset:
		opt = app.Settings().Auth.PasswordResetToken
	default:
		return nil, fmt.Errorf("invalid email type")
	}
	var err error
	var claims OtpClaims
	err = app.TokenAdapter().ParseTokenString(token, opt, &claims)
	if err != nil {
		return nil, fmt.Errorf("error at parsing token: %w", err)
	}
	err = app.TokenAdapter().DeleteToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("error at deleting token: %w", err)
	}
	return &claims, nil
}

// properties
// Settings implements App.
func (app *AuthActionsBase) Settings() *AppOptions {
	return app.settings
}

func (app *AuthActionsBase) AuthAdapter() AuthAdapter {
	return app.authAdapter
}

func (app *AuthActionsBase) AuthMailer() AuthMailer {
	return app.authMailer
}

func (app *AuthActionsBase) TokenAdapter() TokenAdapter {
	return app.tokenAdapter
}

// methods

func (app *AuthActionsBase) Authenticate(ctx context.Context, params *shared.AuthenticationInput) (*shared.User, error) {
	var user *shared.User
	var account *shared.UserAccount
	var err error
	var isFirstLogin bool

	// get user by email
	user, err = app.AuthAdapter().GetUserByEmail(ctx, params.Email)
	if err != nil {
		return nil, fmt.Errorf("error at getting user by email: %w", err)
	}
	// if user exists, get user account
	if user != nil {
		account, err = app.AuthAdapter().GetUserAccount(ctx, user.ID, params.Provider)
		if err != nil {
			return nil, fmt.Errorf("error at getting user account: %w", err)
		}
	}
	// if user does not exist, Create User and continue to create UserAccount ----------------------------------------------------------------------------------------------------
	if user == nil {
		// is first login
		isFirstLogin = true
		user, err = app.AuthAdapter().CreateUser(ctx, &shared.User{
			Email:           params.Email,
			Name:            params.Name,
			Image:           params.AvatarUrl,
			EmailVerifiedAt: params.EmailVerifiedAt,
		})
		if err != nil {
			return nil, fmt.Errorf("error at creating user: %w", err)
		}
		// assign user role
		err = app.AuthAdapter().AssignUserRoles(ctx, user.ID, shared.PermissionNameBasic)
		if err != nil {
			return nil, fmt.Errorf("error at assigning user role: %w", err)
		}
		// if all is good, we should have a user but no account
	}
	// if user exists, but requested account type does not exist, Create UserAccount  of requested type ----------------------------------------------------------------------------------------------------
	if account == nil {
		// from within this block, we should return and not continue to next block
		// if type is credentials, hash password and set params
		if params.Type == shared.ProviderTypeCredentials {
			pw, err := security.CreateHash(*params.Password, argon2id.DefaultParams)
			if err != nil {
				return nil, fmt.Errorf("error at hashing password: %w", err)
			}
			params.HashPassword = &pw
		}
		// link account of requested type
		err = app.AuthAdapter().LinkAccount(ctx, &shared.UserAccount{
			UserID:            user.ID,
			Type:              params.Type,
			Provider:          params.Provider,
			ProviderAccountID: params.ProviderAccountID,
			Password:          params.HashPassword,
			AccessToken:       params.AccessToken,
			RefreshToken:      params.RefreshToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error at linking account: %w", err)
		}
		// if user is first login, send verification email
		if isFirstLogin {
			err = app.SendOtpEmail(EmailTypeVerify, ctx, user)
			if err != nil {
				return nil, fmt.Errorf("error at sending verification email: %w", err)
			}
		} else {
			// if user is not first login, check if user credentials security
			err = app.CheckUserCredentialsSecurity(ctx, user, params)
			if err != nil {
				return nil, fmt.Errorf("error at checking user credentials security: %w", err)
			}
		}
		// return user
		return user, nil
	}
	// if user exists and account exists, check if password is correct  or check if provider key is correct ----------------------------------------------------------------------------------------------------
	if params.Type == shared.ProviderTypeCredentials {
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

func (app *AuthActionsBase) CheckUserCredentialsSecurity(ctx context.Context, user *shared.User, params *shared.AuthenticationInput) error {

	// err := user.LoadUserUserAccounts(ctx, db, models.SelectWhere.UserAccounts.UserID.EQ(user.ID))
	if user == nil || params == nil {
		return fmt.Errorf("user not found")
	}
	// if user is not verified,
	if user.EmailVerifiedAt == nil {
		if params.EmailVerifiedAt != nil {
			// and if incoming request is oauth,
			if params.Type == shared.ProviderTypeOAuth {
				//  check if user has a credentials account
				account, err := app.AuthAdapter().GetUserAccount(ctx, user.ID, shared.ProvidersCredentials)
				if err != nil {
					return fmt.Errorf("error loading user accounts: %w", err)
				}
				if account != nil {
					// if user has a credentials account, send security password reset email
					randomPassword := security.RandomString(20)
					account.Password = &randomPassword
					err = app.AuthAdapter().UpdateUserAccount(ctx, account)
					if err != nil {
						return fmt.Errorf("error updating user password: %w", err)
					}
					err = app.SendOtpEmail(EmailTypeSecurityPasswordReset, ctx, user)
					if err != nil {
						return fmt.Errorf("error sending password reset email: %w", err)
					}
				}
			}
			user.EmailVerifiedAt = params.EmailVerifiedAt
			err := app.AuthAdapter().UpdateUser(ctx, user)
			if err != nil {
				return fmt.Errorf("error updating user email confirmation: %w", err)
			}
		}
	}
	return nil
}

// SendOtpEmail creates and saves a new otp token and sends it to the user's email
func (app *AuthActionsBase) SendOtpEmail(emailType EmailType, ctx context.Context, user *shared.User) error {
	opts := app.Settings().Auth.VerificationToken

	tokenKey := security.GenerateTokenKey()
	otp := security.GenerateOtp(6)
	expires := opts.ExpiresAt()
	userId := user.ID
	email := user.Email
	ttype := opts.Type

	payload := OtpPayload{
		Type:       ttype,
		UserId:     userId,
		Email:      email,
		Token:      tokenKey,
		Otp:        otp,
		RedirectTo: app.Settings().Meta.AppURL,
	}

	tokenHash, err := app.TokenAdapter().CreateOtpTokenHash(&payload, opts)
	if err != nil {
		return fmt.Errorf("error at creating verification token: %w", err)
	}

	dto := &shared.CreateTokenDTO{
		Expires:    expires.Time,
		Token:      tokenKey,
		Type:       ttype,
		Identifier: email,
		UserID:     &userId,
	}

	err = app.TokenAdapter().SaveToken(ctx, dto)

	if err != nil {
		return fmt.Errorf("error at creating verification token: %w", err)
	}

	err = app.AuthMailer().SendOtpEmail(emailType, tokenHash, &payload, app.Settings())
	if err != nil {
		return fmt.Errorf("error at sending verification email: %w", err)
	}
	return nil
}
