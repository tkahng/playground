package core

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mailer"
	"github.com/tkahng/authgo/internal/tools/routine"
	"github.com/tkahng/authgo/internal/tools/security"
)

type Authenticator interface {
	HandlePasswordResetRequest(ctx context.Context, email string) error
	HandleAccessToken(ctx context.Context, token string) (*shared.UserInfo, error)
	HandleRefreshToken(ctx context.Context, token string) (*shared.UserInfoTokens, error)
	HandleVerificationToken(ctx context.Context, token string) error
	HandlePasswordResetToken(ctx context.Context, token, password string) error
	CheckResetPasswordToken(ctx context.Context, token string) error
	VerifyStateToken(ctx context.Context, token string) (*shared.ProviderStateClaims, error)
	VerifyAndParseOtpToken(ctx context.Context, emailType EmailType, token string) (*shared.OtpClaims, error)
	CreateAndPersistStateToken(ctx context.Context, payload *shared.ProviderStatePayload) (string, error)
	Authenticate(ctx context.Context, params *shared.AuthenticationInput) (*shared.User, error)
	// CreateAuthTokens(ctx context.Context, payload *shared.UserInfo) (*shared.UserInfoTokens, error)
	CreateAuthTokensFromEmail(ctx context.Context, email string) (*shared.UserInfoTokens, error)
	SendOtpEmail(emailType EmailType, ctx context.Context, user *shared.User) error
	Signout(ctx context.Context, token string) error
	ResetPassword(ctx context.Context, userId uuid.UUID, oldPassword, newPassword string) error
	// ParseTokenString(tokenString string, config TokenOption, data any) error
}

var _ Authenticator = (*BaseAuth)(nil)

type BaseAuth struct {
	storage  AuthStore
	mail     AuthMailer
	token    TokenManager
	password PasswordManager
	options  *conf.AppOptions
}

func NewAuthActions(dbx db.Dbx, mailer mailer.Mailer, settings *conf.AppOptions) Authenticator {
	actions := &BaseAuth{options: settings}
	storage := NewAuthStore(dbx)
	tokenManager := NewTokenManager()
	password := NewPasswordManager()
	mail := NewAuthMailer(mailer)
	actions.storage = storage
	actions.mail = mail
	actions.token = tokenManager
	actions.password = password

	return actions
}

func (app *BaseAuth) ResetPassword(ctx context.Context, userId uuid.UUID, oldPassword string, newPassword string) error {
	account, err := app.storage.FindUserAccountByUserIdAndProvider(ctx, userId, shared.ProvidersCredentials)
	if err != nil {
		return fmt.Errorf("error getting user account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("user account not found")
	}

	if match, err := app.password.VerifyPassword(oldPassword, *account.Password); err != nil {
		return fmt.Errorf("error at comparing password: %w", err)
	} else if !match {
		return fmt.Errorf("password is incorrect")
	}
	hash, err := app.password.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("error at hashing password: %w", err)
	}
	account.Password = &hash
	err = app.storage.UpdateUserAccount(ctx, account)
	if err != nil {
		return fmt.Errorf("error updating user password: %w", err)
	}
	return nil
}

// Signout implements AuthActions.
func (app *BaseAuth) Signout(ctx context.Context, token string) error {
	opts := app.options.Auth
	var claims shared.RefreshTokenClaims
	err := app.token.ParseToken(token, opts.RefreshToken, &claims)
	if err != nil {
		return fmt.Errorf("error verifying refresh token: %w", err)
	}
	err = app.storage.VerifyTokenStorage(ctx, claims.Token)
	if err != nil {
		return fmt.Errorf("error deleting token: %w", err)
	}
	return nil
}

// HandlePasswordResetRequest implements AuthActions.
func (app *BaseAuth) HandlePasswordResetRequest(ctx context.Context, email string) error {
	user, err := app.storage.FindUserByEmail(
		ctx,
		email,
	)
	if err != nil {
		return fmt.Errorf("error getting user by email: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	account, err := app.storage.FindUserAccountByUserIdAndProvider(ctx, user.ID, shared.ProvidersCredentials)
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
func (app *BaseAuth) CreateAndPersistStateToken(ctx context.Context, payload *shared.ProviderStatePayload) (string, error) {
	if payload == nil {
		return "", fmt.Errorf("payload is nil")
	}
	config := app.options.Auth.StateToken
	claims := shared.ProviderStateClaims{
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
	token, err := app.token.CreateJwtToken(claims, config.Secret)
	if err != nil {
		return token, err
	}

	err = app.storage.SaveToken(ctx, dto)
	if err != nil {
		return token, err
	}
	return token, nil
}

// CreateAuthTokensFromEmail implements AuthActions.
func (app *BaseAuth) CreateAuthTokensFromEmail(ctx context.Context, email string) (*shared.UserInfoTokens, error) {
	user, err := app.storage.GetUserInfo(ctx, email)
	if err != nil {
		return nil, err
	}
	return app.CreateAuthTokens(ctx, user)
}

func (app *BaseAuth) CreateAuthTokens(ctx context.Context, payload *shared.UserInfo) (*shared.UserInfoTokens, error) {
	if payload == nil {
		return nil, fmt.Errorf("payload is nil")
	}

	opts := app.options.Auth

	authToken, err := func() (string, error) {
		claims := shared.AuthenticationClaims{
			Type: shared.TokenTypesAccessToken,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: opts.AccessToken.ExpiresAt(),
			},
			AuthenticationPayload: shared.AuthenticationPayload{
				UserId:      payload.User.ID,
				Email:       payload.User.Email,
				Roles:       payload.Roles,
				Permissions: payload.Permissions,
			},
		}
		token, err := app.token.CreateJwtToken(claims, opts.AccessToken.Secret)
		if err != nil {
			return token, err
		}
		return token, nil
	}()
	if err != nil {
		return nil, err
	}

	tokenKey := security.GenerateTokenKey()

	refreshToken, err := func() (string, error) {
		payload := shared.RefreshTokenPayload{
			UserId: payload.User.ID,
			Email:  payload.User.Email,
			Token:  tokenKey,
		}

		claims := shared.RefreshTokenClaims{
			Type:                shared.TokenTypesRefreshToken,
			RegisteredClaims:    jwt.RegisteredClaims{ExpiresAt: opts.RefreshToken.ExpiresAt()},
			RefreshTokenPayload: payload,
		}

		token, err := app.token.CreateJwtToken(claims, opts.RefreshToken.Secret)
		if err != nil {
			return token, err
		}
		err = app.storage.SaveToken(
			ctx,
			&shared.CreateTokenDTO{
				Type:       shared.TokenTypesRefreshToken,
				Identifier: payload.Email,
				Expires:    opts.RefreshToken.Expires(),
				Token:      payload.Token,
				UserID:     &payload.UserId,
			},
		)
		if err != nil {
			return token, err
		}
		return token, nil
	}()

	if err != nil {
		return nil, err
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
func (app *BaseAuth) CheckResetPasswordToken(ctx context.Context, tokenHash string) error {
	opts := app.options.Auth
	var claims shared.PasswordResetClaims
	err := app.token.ParseToken(tokenHash, opts.PasswordResetToken, &claims)
	if err != nil {
		return fmt.Errorf("error verifying password reset token: %w", err)
	}
	token, err := app.storage.GetToken(ctx, claims.Token)
	if err != nil {
		return err
	}
	if token == nil {
		return fmt.Errorf("token not found")
	}
	return nil
}

// HandlePasswordResetToken implements AuthActions.
func (app *BaseAuth) HandlePasswordResetToken(ctx context.Context, token, password string) error {
	opts := app.options.Auth
	var claims shared.PasswordResetClaims
	err := app.token.ParseToken(token, opts.PasswordResetToken, &claims)
	if err != nil {
		return fmt.Errorf("error verifying password reset token: %w", err)
	}
	err = app.storage.VerifyTokenStorage(ctx, claims.Token)
	if err != nil {
		return fmt.Errorf("error deleting token: %w", err)
	}
	user, err := app.storage.FindUserByEmail(
		ctx,
		claims.Email,
	)
	if err != nil {
		return fmt.Errorf("error getting user by email: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	account, err := app.storage.FindUserAccountByUserIdAndProvider(ctx, user.ID, shared.ProvidersCredentials)
	if err != nil {
		return fmt.Errorf("error getting user account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("user account not found")
	}
	hash, err := app.password.HashPassword(password)
	if err != nil {
		return fmt.Errorf("error at hashing password: %w", err)
	}
	account.Password = &hash
	err = app.storage.UpdateUserAccount(ctx, account)
	if err != nil {
		return fmt.Errorf("error updating user password: %w", err)
	}
	return nil

}
func (app *BaseAuth) VerifyStateToken(ctx context.Context, token string) (*shared.ProviderStateClaims, error) {
	opts := app.options.Auth
	var claims shared.ProviderStateClaims
	err := app.token.ParseToken(token, opts.StateToken, &claims)
	if err != nil {
		return nil, fmt.Errorf("error verifying state token: %w", err)
	}
	err = app.storage.VerifyTokenStorage(ctx, claims.Token)
	if err != nil {
		return nil, fmt.Errorf("error deleting token: %w", err)
	}
	return &claims, nil
}
func (app *BaseAuth) HandleAccessToken(ctx context.Context, token string) (*shared.UserInfo, error) {
	opts := app.options.Auth
	var claims shared.AuthenticationClaims
	err := app.token.ParseToken(token, opts.AccessToken, &claims)
	if err != nil {
		return nil, fmt.Errorf("error verifying access token: %w", err)
	}
	return app.storage.GetUserInfo(ctx, claims.Email)
}

// HandleRefreshToken implements AuthActions.
func (app *BaseAuth) HandleRefreshToken(ctx context.Context, token string) (*shared.UserInfoTokens, error) {
	opts := app.options.Auth
	var claims shared.RefreshTokenClaims
	err := app.token.ParseToken(token, opts.RefreshToken, &claims)
	if err != nil {
		return nil, fmt.Errorf("error verifying refresh token: %w", err)
	}
	err = app.storage.VerifyTokenStorage(ctx, claims.Token)
	if err != nil {
		return nil, fmt.Errorf("error deleting token: %w", err)
	}
	info, err := app.storage.GetUserInfo(ctx, claims.Email)
	if err != nil {
		return nil, err
	}

	return app.CreateAuthTokens(ctx, info)
}

func (app *BaseAuth) HandleVerificationToken(ctx context.Context, token string) error {
	claims, err := app.VerifyAndParseOtpToken(ctx, EmailTypeVerify, token)
	if err != nil {
		return fmt.Errorf("error verifying verification token: %w", err)
	}
	err = app.storage.VerifyTokenStorage(ctx, claims.Token)
	if err != nil {
		return fmt.Errorf("error deleting token: %w", err)
	}
	user, err := app.storage.FindUserByEmail(ctx, claims.Email)
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
	err = app.storage.UpdateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}
	return nil
}

// VerifyAndUseVerificationToken implements AuthActions.
func (app *BaseAuth) VerifyAndParseOtpToken(ctx context.Context, emailType EmailType, token string) (*shared.OtpClaims, error) {
	var opt conf.TokenOption
	switch emailType {
	case EmailTypeVerify:
		opt = app.options.Auth.VerificationToken
	case EmailTypeConfirmPasswordReset:
		opt = app.options.Auth.PasswordResetToken
	case EmailTypeSecurityPasswordReset:
		opt = app.options.Auth.PasswordResetToken
	default:
		return nil, fmt.Errorf("invalid email type")
	}
	var err error
	var claims shared.OtpClaims
	err = app.token.ParseToken(token, opt, &claims)
	if err != nil {
		return nil, fmt.Errorf("error at parsing token: %w", err)
	}
	return &claims, nil
}

// methods

func (app *BaseAuth) Authenticate(ctx context.Context, params *shared.AuthenticationInput) (*shared.User, error) {
	var user *shared.User
	var account *shared.UserAccount
	var err error
	var isFirstLogin bool

	// get user by email
	user, err = app.storage.FindUserByEmail(ctx, params.Email)
	if err != nil {
		return nil, fmt.Errorf("error at getting user by email: %w", err)
	}
	// if user exists, get user account
	if user != nil {
		account, err = app.storage.FindUserAccountByUserIdAndProvider(ctx, user.ID, shared.ProvidersCredentials)
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
		user, err = app.storage.CreateUser(ctx, &shared.User{
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
		err = app.storage.AssignUserRoles(ctx, user.ID, shared.PermissionNameBasic)
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
		err = app.storage.LinkAccount(ctx, newVar)
		if err != nil {
			return nil, fmt.Errorf("error at linking account: %w", err)
		}
		// if user is first login, send verification email
		if isFirstLogin {
			fmt.Println("User is first login, sending verification email")
			routine.FireAndForget(
				func() {
					ctx := context.Background()
					err = app.SendOtpEmail(EmailTypeVerify, ctx, user)
					if err != nil {
						slog.Error(
							"error sending verification email",
							slog.Any("error", err),
							slog.String("email", user.Email),
							slog.String("userId", user.ID.String()),
						)
					}
				},
			)

		} else {
			fmt.Println("User is not first login, checking user credentials security")
			// if user is not first login, check if user credentials security
			routine.FireAndForget(
				func() {
					ctx := context.Background()
					err = app.CheckUserCredentialsSecurity(ctx, user, params)
					if err != nil {
						slog.Error(
							"error at checking user credentials security",
							slog.Any("error", err),
							slog.String("email", user.Email),
							slog.String("userId", user.ID.String()),
						)
					}
				},
			)
		}
		// return user
		return user, nil
	}
	// if user exists and account exists, check if password is correct  or check if provider key is correct ----------------------------------------------------------------------------------------------------
	if params.Type == shared.ProviderTypeCredentials {
		if params.Password == nil || account.Password == nil {
			return nil, fmt.Errorf("password or account password is nil")
		}
		if match, err := app.password.VerifyPassword(*params.Password, *account.Password); err != nil {
			return nil, fmt.Errorf("error at comparing password: %w", err)
		} else if !match {
			return nil, fmt.Errorf("password is incorrect")
		}
	}
	return user, nil
}

func (app *BaseAuth) CheckUserCredentialsSecurity(ctx context.Context, user *shared.User, params *shared.AuthenticationInput) error {

	if user == nil || params == nil {
		return fmt.Errorf("user not found")
	}
	// if user is not verified,
	if user.EmailVerifiedAt == nil {
		if params.EmailVerifiedAt != nil {
			// and if incoming request is oauth,
			if params.Type == shared.ProviderTypeOAuth {
				//  check if user has a credentials account
				account, err := app.storage.FindUserAccountByUserIdAndProvider(ctx, user.ID, shared.ProvidersCredentials)

				if err != nil {
					return fmt.Errorf("error loading user accounts: %w", err)
				}
				if account != nil {
					// if user has a credentials account, send security password reset email
					randomPassword := security.RandomString(20)
					hash, err := app.password.HashPassword(randomPassword)
					if err != nil {
						return fmt.Errorf("error at hashing password: %w", err)
					}
					account.Password = &hash
					err = app.storage.UpdateUserAccount(ctx, account)
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
			err := app.storage.UpdateUser(ctx, user)
			if err != nil {
				return fmt.Errorf("error updating user email confirmation: %w", err)
			}
		}
	}
	return nil
}

// SendOtpEmail creates and saves a new otp token and sends it to the user's email
func (app *BaseAuth) SendOtpEmail(emailType EmailType, ctx context.Context, user *shared.User) error {
	var opts conf.TokenOption
	switch emailType {
	case EmailTypeVerify:
		opts = app.options.Auth.VerificationToken
	case EmailTypeSecurityPasswordReset:
		opts = app.options.Auth.PasswordResetToken
	case EmailTypeConfirmPasswordReset:
		opts = app.options.Auth.PasswordResetToken
	default:
		return fmt.Errorf("invalid email type")
	}

	tokenKey := security.GenerateTokenKey()
	otp := security.GenerateOtp(6)
	expires := opts.ExpiresAt()
	userId := user.ID
	email := user.Email
	ttype := opts.Type

	payload := shared.OtpPayload{
		Type:       ttype,
		UserId:     userId,
		Email:      email,
		Token:      tokenKey,
		Otp:        otp,
		RedirectTo: app.options.Meta.AppUrl,
	}

	tokenHash, err := app.CreateOtpTokenHash(&payload, opts)
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

	err = app.storage.SaveToken(ctx, dto)

	if err != nil {
		return fmt.Errorf("error at creating verification token: %w", err)
	}

	err = app.mail.SendOtpEmail(emailType, tokenHash, &payload, app.options)
	if err != nil {
		return fmt.Errorf("error at sending verification email: %w", err)
	}
	return nil
}

func (app *BaseAuth) CreateOtpTokenHash(payload *shared.OtpPayload, config conf.TokenOption) (string, error) {
	if payload == nil {
		return "", fmt.Errorf("payload is nil")
	}
	claims := shared.OtpClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: config.ExpiresAt(),
		},
		OtpPayload: *payload,
	}
	token, err := app.token.CreateJwtToken(claims, config.Secret)
	if err != nil {
		return "", fmt.Errorf("error at creating verification token: %w", err)
	}
	return token, nil

}
