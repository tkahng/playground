package core

import (
	"context"
	"fmt"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mailer"
	"github.com/tkahng/authgo/internal/tools/security"
)

var _ AuthActions = (*AuthActionsBase)(nil)

type AuthActionsBase struct {
	authAdapter  *AuthAdapterBase
	authMailer   *AuthMailerBase
	tokenAdapter *TokenAdapterBase
	settings     *AppOptions
}

// ResetPassword implements AuthActions.
func (app *AuthActionsBase) ResetPassword(ctx context.Context, userId uuid.UUID, oldPassword string, newPassword string) error {
	account, err := app.authAdapter.FindUserAccount(ctx, &map[string]any{
		"user_id": map[string]any{
			"_eq": userId.String(),
		},
		"provider": map[string]any{
			"_eq": shared.ProvidersCredentials.String(),
		},
	})
	if err != nil {
		return fmt.Errorf("error getting user account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("user account not found")
	}

	if match, err := security.ComparePasswordAndHash(oldPassword, *account.Password); err != nil {
		return fmt.Errorf("error at comparing password: %w", err)
	} else if !match {
		return fmt.Errorf("password is incorrect")
	}
	hash, err := security.CreateHash(newPassword, argon2id.DefaultParams)
	if err != nil {
		return fmt.Errorf("error at hashing password: %w", err)
	}
	account.Password = &hash
	err = app.authAdapter.UpdateUserAccount(ctx, account)
	if err != nil {
		return fmt.Errorf("error updating user password: %w", err)
	}
	return nil
}

// Signout implements AuthActions.
func (app *AuthActionsBase) Signout(ctx context.Context, token string) error {
	opts := app.settings.Auth
	var claims RefreshTokenClaims
	err := app.tokenAdapter.ParseTokenString(token, opts.RefreshToken, &claims)
	if err != nil {
		return fmt.Errorf("error verifying refresh token: %w", err)
	}
	err = app.tokenAdapter.VerifyTokenStorage(ctx, claims.Token)
	if err != nil {
		return fmt.Errorf("error deleting token: %w", err)
	}
	return nil
}

// HandlePasswordResetRequest implements AuthActions.
func (app *AuthActionsBase) HandlePasswordResetRequest(ctx context.Context, email string) error {
	user, err := app.authAdapter.FindUser(
		ctx,
		&map[string]any{
			"email": map[string]any{
				"_eq": email,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("error getting user by email: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	account, err := app.authAdapter.FindUserAccount(
		ctx,
		&map[string]any{
			"user_id": map[string]any{
				"_eq": user.ID.String(),
			},
			"provider": map[string]any{
				"_eq": shared.ProvidersCredentials.String(),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("error getting user account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("user account not found")
	}

	err = app.SendOtpEmail(EmailTypeConfirmPasswordReset, ctx, user)
	if err != nil {
		return fmt.Errorf("error sending password reset email: %w", err)
	}
	return nil
}

// CreateAndPersistStateToken implements AuthActions.
func (app *AuthActionsBase) CreateAndPersistStateToken(ctx context.Context, payload *ProviderStatePayload) (string, error) {
	if payload == nil {
		return "", fmt.Errorf("payload is nil")
	}
	config := app.settings.Auth.StateToken
	claims := ProviderStateClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: config.ExpiresAt(),
		},
		ProviderStatePayload: *payload,
	}
	dto := &shared.CreateTokenDTO{
		Type:       shared.TokenTypesStateToken,
		Identifier: payload.Token,
		Expires:    config.Expires(),
		Token:      payload.Token,
	}
	token, err := security.NewJWTWithClaims(claims, config.Secret)
	if err != nil {
		return token, fmt.Errorf("error at error: %w", err)
	}

	err = app.tokenAdapter.SaveToken(ctx, dto)
	if err != nil {
		return token, fmt.Errorf("error at error: %w", err)
	}
	return token, nil
}

// CreateAuthTokensFromEmail implements AuthActions.
func (app *AuthActionsBase) CreateAuthTokensFromEmail(ctx context.Context, email string) (*shared.UserInfoTokens, error) {
	user, err := app.authAdapter.GetUserInfo(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("error at getting user by email: %w", err)
	}
	return app.CreateAuthTokens(ctx, user)
}

func NewAuthActions(dbx *pgxpool.Pool, mailer mailer.Mailer, settings *AppOptions) AuthActions {
	authAdapter := NewAuthAdapter(dbx)
	authMailer := NewAuthMailer(mailer)
	tokenAdapter := NewTokenAdapter(dbx)
	return &AuthActionsBase{
		authAdapter:  authAdapter,
		authMailer:   authMailer,
		tokenAdapter: tokenAdapter,
		settings:     settings,
	}
}
func (app *AuthActionsBase) createAuthenticationToken(payload *AuthenticationPayload, config TokenOption) (string, error) {
	if payload == nil {
		return "", fmt.Errorf("payload is nil")
	}
	claims := AuthenticationClaims{
		Type: shared.TokenTypesAccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: config.ExpiresAt(),
		},
		AuthenticationPayload: *payload,
	}
	token, err := security.NewJWTWithClaims(claims, config.Secret)
	if err != nil {
		return token, fmt.Errorf("error at error: %w", err)
	}

	return token, nil
}
func (app *AuthActionsBase) createRefreshToken(ctx context.Context, payload *RefreshTokenPayload, config TokenOption) (string, error) {
	if payload == nil {
		return "", fmt.Errorf("payload is nil")
	}
	claims := RefreshTokenClaims{
		Type: shared.TokenTypesRefreshToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: config.ExpiresAt(),
		},
		RefreshTokenPayload: *payload,
	}
	dto := &shared.CreateTokenDTO{
		Type:       shared.TokenTypesRefreshToken,
		Identifier: payload.Email,
		Expires:    config.Expires(),
		Token:      payload.Token,
		UserID:     &payload.UserId,
	}
	token, err := security.NewJWTWithClaims(claims, config.Secret)
	if err != nil {
		return token, fmt.Errorf("error at error: %w", err)
	}

	err = app.tokenAdapter.SaveToken(ctx, dto)
	if err != nil {
		return token, fmt.Errorf("error at error: %w", err)
	}
	return token, nil
}

func (app *AuthActionsBase) CreateAuthTokens(ctx context.Context, payload *shared.UserInfo) (*shared.UserInfoTokens, error) {
	if payload == nil {
		return nil, fmt.Errorf("payload is nil")
	}

	opts := app.settings.Auth

	authToken, err := app.createAuthenticationToken(&AuthenticationPayload{
		UserId:      payload.User.ID,
		Email:       payload.User.Email,
		Roles:       payload.Roles,
		Permissions: payload.Permissions,
	}, opts.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("error creating auth token: %w", err)
	}

	tokenKey := security.GenerateTokenKey()

	refreshToken, err := app.createRefreshToken(ctx, &RefreshTokenPayload{
		UserId: payload.User.ID,
		Email:  payload.User.Email,
		Token:  tokenKey,
	}, opts.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("error creating refresh token: %w", err)
	}
	return &shared.UserInfoTokens{
		UserInfo: *payload,
		Tokens: shared.TokenDto{
			AccessToken:  authToken,
			RefreshToken: refreshToken,
			ExpiresIn:    opts.AccessToken.Duration,
			TokenType:    "Bearer",
		},
	}, nil
}

// CheckResetPasswordToken implements AuthActions.
func (app *AuthActionsBase) CheckResetPasswordToken(ctx context.Context, tokenHash string) error {
	opts := app.settings.Auth
	var claims PasswordResetClaims
	err := app.tokenAdapter.ParseTokenString(tokenHash, opts.PasswordResetToken, &claims)
	if err != nil {
		return fmt.Errorf("error verifying password reset token: %w", err)
	}
	token, err := app.tokenAdapter.GetToken(ctx, claims.Token)
	if err != nil {
		return fmt.Errorf("error getting token: %w", err)
	}
	if token == nil {
		return fmt.Errorf("token not found")
	}
	return nil
}

// HandlePasswordResetToken implements AuthActions.
func (app *AuthActionsBase) HandlePasswordResetToken(ctx context.Context, token, password string) error {
	opts := app.settings.Auth
	var claims PasswordResetClaims
	err := app.tokenAdapter.ParseTokenString(token, opts.PasswordResetToken, &claims)
	if err != nil {
		return fmt.Errorf("error verifying password reset token: %w", err)
	}
	err = app.tokenAdapter.VerifyTokenStorage(ctx, claims.Token)
	if err != nil {
		return fmt.Errorf("error deleting token: %w", err)
	}
	user, err := app.authAdapter.FindUser(
		ctx,
		&map[string]any{
			"email": map[string]any{
				"_eq": claims.Email,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("error getting user by email: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	account, err := app.authAdapter.FindUserAccount(ctx, &map[string]any{
		"user_id": map[string]any{
			"_eq": user.ID.String(),
		},
		"provider": map[string]any{
			"_eq": shared.ProvidersCredentials.String(),
		},
	})
	if err != nil {
		return fmt.Errorf("error getting user account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("user account not found")
	}
	hash, err := security.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return fmt.Errorf("error at hashing password: %w", err)
	}
	account.Password = &hash
	err = app.authAdapter.UpdateUserAccount(ctx, account)
	if err != nil {
		return fmt.Errorf("error updating user password: %w", err)
	}
	return nil

}
func (app *AuthActionsBase) VerifyStateToken(ctx context.Context, token string) (*ProviderStateClaims, error) {
	opts := app.settings.Auth
	var claims ProviderStateClaims
	err := app.tokenAdapter.ParseTokenString(token, opts.StateToken, &claims)
	if err != nil {
		return nil, fmt.Errorf("error verifying state token: %w", err)
	}
	err = app.tokenAdapter.VerifyTokenStorage(ctx, claims.Token)
	if err != nil {
		return nil, fmt.Errorf("error deleting token: %w", err)
	}
	return &claims, nil
}
func (app *AuthActionsBase) HandleAccessToken(ctx context.Context, token string) (*shared.UserInfo, error) {
	opts := app.settings.Auth
	var claims AuthenticationClaims
	err := app.tokenAdapter.ParseTokenString(token, opts.AccessToken, &claims)
	if err != nil {
		return nil, fmt.Errorf("error verifying access token: %w", err)
	}
	return app.authAdapter.GetUserInfo(ctx, claims.Email)
}

// HandleRefreshToken implements AuthActions.
func (app *AuthActionsBase) HandleRefreshToken(ctx context.Context, token string) (*shared.UserInfoTokens, error) {
	opts := app.settings.Auth
	var claims RefreshTokenClaims
	err := app.tokenAdapter.ParseTokenString(token, opts.RefreshToken, &claims)
	if err != nil {
		return nil, fmt.Errorf("error verifying refresh token: %w", err)
	}
	err = app.tokenAdapter.VerifyTokenStorage(ctx, claims.Token)
	if err != nil {
		return nil, fmt.Errorf("error deleting token: %w", err)
	}
	info, err := app.authAdapter.GetUserInfo(ctx, claims.Email)
	if err != nil {
		return nil, fmt.Errorf("error getting user info: %w", err)
	}

	return app.CreateAuthTokens(ctx, info)
}

func (app *AuthActionsBase) HandleVerificationToken(ctx context.Context, token string) error {
	claims, err := app.VerifyAndParseOtpToken(ctx, EmailTypeVerify, token)
	if err != nil {
		return fmt.Errorf("error verifying verification token: %w", err)
	}
	err = app.tokenAdapter.VerifyTokenStorage(ctx, claims.Token)
	if err != nil {
		return fmt.Errorf("error deleting token: %w", err)
	}
	user, err := app.authAdapter.FindUser(ctx, &map[string]any{
		"email": map[string]any{
			"_eq": claims.Email,
		}})
	if err != nil {
		return fmt.Errorf("error getting user info: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	if user.EmailVerifiedAt != nil {
		return fmt.Errorf("user already verified")
	}
	now := time.Now()
	user.EmailVerifiedAt = &now
	err = app.authAdapter.UpdateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}
	return nil
}

// VerifyAndUseVerificationToken implements AuthActions.
func (app *AuthActionsBase) VerifyAndParseOtpToken(ctx context.Context, emailType EmailType, token string) (*OtpClaims, error) {
	var opt TokenOption
	switch emailType {
	case EmailTypeVerify:
		opt = app.settings.Auth.VerificationToken
	case EmailTypeConfirmPasswordReset:
		opt = app.settings.Auth.PasswordResetToken
	case EmailTypeSecurityPasswordReset:
		opt = app.settings.Auth.PasswordResetToken
	default:
		return nil, fmt.Errorf("invalid email type")
	}
	var err error
	var claims OtpClaims
	err = app.tokenAdapter.ParseTokenString(token, opt, &claims)
	if err != nil {
		return nil, fmt.Errorf("error at parsing token: %w", err)
	}
	return &claims, nil
}

// methods

func (app *AuthActionsBase) Authenticate(ctx context.Context, params *shared.AuthenticationInput) (*shared.User, error) {
	var user *shared.User
	var account *shared.UserAccount
	var err error
	var isFirstLogin bool
	fmt.Println("Authenticate")
	// get user by email
	user, err = app.authAdapter.FindUser(ctx, &map[string]any{
		"email": map[string]any{
			"_eq": params.Email,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error at getting user by email: %w", err)
	}
	// if user exists, get user account
	if user != nil {
		account, err = app.authAdapter.FindUserAccount(ctx, &map[string]any{
			"user_id": map[string]any{
				"_eq": user.ID.String(),
			},
			"provider": map[string]any{
				"_eq": shared.ProvidersCredentials.String(),
			},
		})
		if err != nil {
			return nil, fmt.Errorf("error at getting user account: %w", err)
		}
	}
	// if user does not exist, Create User and continue to create UserAccount ----------------------------------------------------------------------------------------------------
	if user == nil {
		fmt.Println("User does not exist, creating user")
		// is first login
		if params.EmailVerifiedAt != nil {
			isFirstLogin = false
		} else {
			isFirstLogin = true
		}
		user, err = app.authAdapter.CreateUser(ctx, &shared.User{
			Email:           params.Email,
			Name:            params.Name,
			Image:           params.AvatarUrl,
			EmailVerifiedAt: params.EmailVerifiedAt,
		})

		if err != nil {
			return nil, fmt.Errorf("error at creating user: %w", err)
		}
		if user == nil {
			return nil, fmt.Errorf("user not created")
		}
		// assign user role
		err = app.authAdapter.AssignUserRoles(ctx, user.ID, shared.PermissionNameBasic)
		if err != nil {
			return nil, fmt.Errorf("error at assigning user role: %w", err)
		}
		// if all is good, we should have a user but no account
	}
	// if user exists, but requested account type does not exist, Create UserAccount  of requested type ----------------------------------------------------------------------------------------------------
	if account == nil {
		fmt.Println("Account does not exist, creating account")
		// from within this block, we should return and not continue to next block
		// if type is credentials, hash password and set params
		if params.Type == shared.ProviderTypeCredentials {
			if params.Password == nil {
				return nil, fmt.Errorf("password is nil")
			}
			pw, err := security.CreateHash(*params.Password, argon2id.DefaultParams)
			if err != nil {
				return nil, fmt.Errorf("error at hashing password: %w", err)
			}
			params.HashPassword = &pw
		}
		// link account of requested type
		newVar := &shared.UserAccount{
			UserID:            user.ID,
			Type:              params.Type,
			Provider:          params.Provider,
			ProviderAccountID: params.ProviderAccountID,
			Password:          params.HashPassword,
			AccessToken:       params.AccessToken,
			RefreshToken:      params.RefreshToken,
		}
		err = app.authAdapter.LinkAccount(ctx, newVar)
		if err != nil {
			return nil, fmt.Errorf("error at linking account: %w", err)
		}
		// if user is first login, send verification email
		if isFirstLogin {
			fmt.Println("User is first login, sending verification email")
			err = app.SendOtpEmail(EmailTypeVerify, ctx, user)
			if err != nil {
				return nil, fmt.Errorf("error at sending verification email: %w", err)
			}
		} else {
			fmt.Println("User is not first login, checking user credentials security")
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
				account, err := app.authAdapter.FindUserAccount(ctx, &map[string]any{
					"user_id":  user.ID.String(),
					"provider": shared.ProvidersCredentials.String(),
				})

				if err != nil {
					return fmt.Errorf("error loading user accounts: %w", err)
				}
				if account != nil {
					// if user has a credentials account, send security password reset email
					randomPassword := security.RandomString(20)
					hash, err := security.CreateHash(randomPassword, argon2id.DefaultParams)
					if err != nil {
						return fmt.Errorf("error at hashing password: %w", err)
					}
					account.Password = &hash
					err = app.authAdapter.UpdateUserAccount(ctx, account)
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
			err := app.authAdapter.UpdateUser(ctx, user)
			if err != nil {
				return fmt.Errorf("error updating user email confirmation: %w", err)
			}
		}
	}
	return nil
}

// SendOtpEmail creates and saves a new otp token and sends it to the user's email
func (app *AuthActionsBase) SendOtpEmail(emailType EmailType, ctx context.Context, user *shared.User) error {
	var opts TokenOption
	switch emailType {
	case EmailTypeVerify:
		opts = app.settings.Auth.VerificationToken
	case EmailTypeSecurityPasswordReset:
		opts = app.settings.Auth.PasswordResetToken
	case EmailTypeConfirmPasswordReset:
		opts = app.settings.Auth.PasswordResetToken
	default:
		return fmt.Errorf("invalid email type")
	}

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
		RedirectTo: app.settings.Meta.AppURL,
	}

	tokenHash, err := app.tokenAdapter.CreateOtpTokenHash(&payload, opts)
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

	err = app.tokenAdapter.SaveToken(ctx, dto)

	if err != nil {
		return fmt.Errorf("error at creating verification token: %w", err)
	}

	err = app.authMailer.SendOtpEmail(emailType, tokenHash, &payload, app.settings)
	if err != nil {
		return fmt.Errorf("error at sending verification email: %w", err)
	}
	return nil
}
