package services

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/auth/oauth"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mailer"
	"github.com/tkahng/authgo/internal/tools/routine"
	"github.com/tkahng/authgo/internal/tools/security"
)

type AuthService interface {
	HandlePasswordResetRequest(ctx context.Context, email string) error
	HandleAccessToken(ctx context.Context, token string) (*shared.UserInfo, error)
	HandleRefreshToken(ctx context.Context, token string) (*shared.UserInfoTokens, error)
	HandleVerificationToken(ctx context.Context, token string) error
	HandlePasswordResetToken(ctx context.Context, token, password string) error
	CheckResetPasswordToken(ctx context.Context, token string) error
	VerifyStateToken(ctx context.Context, token string) (*shared.ProviderStateClaims, error)
	CreateAndPersistStateToken(ctx context.Context, payload *shared.ProviderStatePayload) (string, error)
	FetchAuthUser(ctx context.Context, code string, parsedState *shared.ProviderStateClaims) (*oauth.AuthUser, error)
	VerifyAndParseOtpToken(ctx context.Context, emailType EmailType, token string) (*shared.OtpClaims, error)
	Authenticate(ctx context.Context, params *shared.AuthenticationInput) (*models.User, error)
	CreateAuthTokensFromEmail(ctx context.Context, email string) (*shared.UserInfoTokens, error)
	SendOtpEmail(emailType EmailType, ctx context.Context, user *models.User) error
	Signout(ctx context.Context, token string) error
	ResetPassword(ctx context.Context, userId uuid.UUID, oldPassword, newPassword string) error
}

type AuthAccountStore interface {
	FindUserAccountByUserIdAndProvider(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error)
	UpdateUserAccount(ctx context.Context, account *models.UserAccount) error
	LinkAccount(ctx context.Context, account *models.UserAccount) error
	UnlinkAccount(ctx context.Context, userId uuid.UUID, provider models.Providers) error
}

type AuthUserStore interface {
	GetUserInfo(ctx context.Context, email string) (*shared.UserInfo, error)
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type AuthTokenStore interface {
	GetToken(ctx context.Context, token string) (*models.Token, error)
	SaveToken(ctx context.Context, token *shared.CreateTokenDTO) error
	DeleteToken(ctx context.Context, token string) error
}

type AuthStore interface {
	// GetUserInfo(ctx context.Context, email string) (*shared.UserInfo, error)
	// CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	// AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error
	// FindUserByEmail(ctx context.Context, email string) (*models.User, error)
	// UpdateUser(ctx context.Context, user *models.User) error
	// DeleteUser(ctx context.Context, id uuid.UUID) error
	AuthUserStore
	GetToken(ctx context.Context, token string) (*models.Token, error)
	SaveToken(ctx context.Context, token *shared.CreateTokenDTO) error
	DeleteToken(ctx context.Context, token string) error
	FindUserAccountByUserIdAndProvider(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error)
	UpdateUserAccount(ctx context.Context, account *models.UserAccount) error
	LinkAccount(ctx context.Context, account *models.UserAccount) error
	UnlinkAccount(ctx context.Context, userId uuid.UUID, provider models.Providers) error
}

var _ AuthService = (*BaseAuthService)(nil)

type BaseAuthService struct {
	authStore AuthStore
	mail      mailer.Mailer
	token     JwtService
	password  PasswordService
	options   *conf.AppOptions
}

func NewAuthService(
	opts *conf.AppOptions,
	authStore AuthStore,
	mail mailer.Mailer,
	token JwtService,
	password PasswordService,
) AuthService {
	authService := &BaseAuthService{
		authStore: authStore,
		mail:      mail,
		token:     token,
		password:  password,
		options:   opts,
	}

	return authService
}

// FetchAuthUser implements Authenticator.
func (app *BaseAuthService) FetchAuthUser(ctx context.Context, code string, parsedState *shared.ProviderStateClaims) (*oauth.AuthUser, error) {
	var provider oauth.ProviderConfig
	switch parsedState.Provider {
	case shared.OAuthProvidersGithub:
		provider = oauth.NewProviderByName(oauth.NameGithub)
	case shared.OAuthProvidersGoogle:
		provider = oauth.NewProviderByName(oauth.NameGoogle)
	default:
		return nil, fmt.Errorf("invalid provider %v", parsedState.Provider)
	}
	if !provider.Active() {
		return nil, fmt.Errorf("provider %v is not enabled", parsedState.Provider)
	}
	opts := provider.FetchTokenOptions(parsedState.CodeVerifier)

	// fetch token
	token, err := provider.FetchToken(ctx, code, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OAuth2 token. %w", err)
	}

	// fetch external auth user
	authUser, err := provider.FetchAuthUser(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OAuth2 user. %w", err)
	}
	return authUser, nil
}

func (app *BaseAuthService) ResetPassword(ctx context.Context, userId uuid.UUID, oldPassword string, newPassword string) error {
	account, err := app.authStore.FindUserAccountByUserIdAndProvider(ctx, userId, models.ProvidersCredentials)
	if err != nil {
		return fmt.Errorf("error getting user account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("user account not found")
	}

	if match, err := app.password.VerifyPassword(*account.Password, oldPassword); err != nil {
		return fmt.Errorf("error at comparing password: %w", err)
	} else if !match {
		return fmt.Errorf("password is incorrect")
	}
	hash, err := app.password.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("error at hashing password: %w", err)
	}
	account.Password = &hash
	err = app.authStore.UpdateUserAccount(ctx, account)
	if err != nil {
		return fmt.Errorf("error updating user password: %w", err)
	}
	return nil
}

// Signout implements AuthActions.
func (app *BaseAuthService) Signout(ctx context.Context, token string) error {
	opts := app.options.Auth
	var claims shared.RefreshTokenClaims
	err := app.token.ParseToken(token, opts.RefreshToken, &claims)
	if err != nil {
		return fmt.Errorf("error verifying refresh token: %w", err)
	}
	_, err = app.authStore.GetToken(ctx, token) // corrected 'tokne' to 'token'
	if err != nil {
		return err
	}
	err = app.authStore.DeleteToken(ctx, token) // corrected to use 'app.token'
	if err != nil {
		return fmt.Errorf("error at deleting token: %w", err)
	}
	return nil
}

// HandlePasswordResetRequest implements AuthActions.
func (app *BaseAuthService) HandlePasswordResetRequest(ctx context.Context, email string) error {
	user, err := app.authStore.FindUserByEmail(
		ctx,
		email,
	)
	if err != nil {
		return fmt.Errorf("error getting user by email: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	account, err := app.authStore.FindUserAccountByUserIdAndProvider(ctx, user.ID, models.ProvidersCredentials)
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
func (app *BaseAuthService) CreateAndPersistStateToken(ctx context.Context, payload *shared.ProviderStatePayload) (string, error) {
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

	err = app.authStore.SaveToken(ctx, dto)
	if err != nil {
		return token, err
	}
	return token, nil
}

// CreateAuthTokensFromEmail implements AuthActions.
func (app *BaseAuthService) CreateAuthTokensFromEmail(ctx context.Context, email string) (*shared.UserInfoTokens, error) {
	user, err := app.authStore.GetUserInfo(ctx, email)
	if err != nil {
		return nil, err
	}
	return app.CreateAuthTokens(ctx, user)
}

func (app *BaseAuthService) CreateAuthTokens(ctx context.Context, payload *shared.UserInfo) (*shared.UserInfoTokens, error) {
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
		err = app.authStore.SaveToken(
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
func (app *BaseAuthService) CheckResetPasswordToken(ctx context.Context, tokenHash string) error {
	opts := app.options.Auth
	var claims shared.PasswordResetClaims
	err := app.token.ParseToken(tokenHash, opts.PasswordResetToken, &claims)
	if err != nil {
		return fmt.Errorf("error verifying password reset token: %w", err)
	}
	token, err := app.authStore.GetToken(ctx, claims.Token)
	if err != nil {
		return err
	}
	if token == nil {
		return fmt.Errorf("token not found")
	}
	return nil
}

// HandlePasswordResetToken implements AuthActions.
func (app *BaseAuthService) HandlePasswordResetToken(ctx context.Context, token, password string) error {
	opts := app.options.Auth
	var claims shared.PasswordResetClaims
	err := app.token.ParseToken(token, opts.PasswordResetToken, &claims)
	if err != nil {
		return fmt.Errorf("error verifying password reset token: %w", err)
	}
	_, err = app.authStore.GetToken(ctx, token) // corrected 'tokne' to 'token'
	if err != nil {
		return err
	}
	err = app.authStore.DeleteToken(ctx, token) // corrected to use 'app.token'
	if err != nil {
		return fmt.Errorf("error deleting token: %w", err)
	}
	user, err := app.authStore.FindUserByEmail(
		ctx,
		claims.Email,
	)
	if err != nil {
		return fmt.Errorf("error getting user by email: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	account, err := app.authStore.FindUserAccountByUserIdAndProvider(ctx, user.ID, models.ProvidersCredentials)
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
	err = app.authStore.UpdateUserAccount(ctx, account)
	if err != nil {
		return fmt.Errorf("error updating user password: %w", err)
	}
	return nil

}
func (app *BaseAuthService) VerifyStateToken(ctx context.Context, token string) (*shared.ProviderStateClaims, error) {
	opts := app.options.Auth
	var claims shared.ProviderStateClaims
	err := app.token.ParseToken(token, opts.StateToken, &claims)
	if err != nil {
		return nil, fmt.Errorf("error verifying state token: %w", err)
	}
	_, err = app.authStore.GetToken(ctx, token) // corrected 'tokne' to 'token'
	if err != nil {
		return nil, err
	}
	err = app.authStore.DeleteToken(ctx, token) // corrected to use 'app.token'
	if err != nil {
		return nil, fmt.Errorf("error deleting token: %w", err)
	}
	return &claims, nil
}
func (app *BaseAuthService) HandleAccessToken(ctx context.Context, token string) (*shared.UserInfo, error) {
	opts := app.options.Auth
	var claims shared.AuthenticationClaims
	err := app.token.ParseToken(token, opts.AccessToken, &claims)
	if err != nil {
		return nil, fmt.Errorf("error verifying access token: %w", err)
	}
	return app.authStore.GetUserInfo(ctx, claims.Email)
}

// HandleRefreshToken implements AuthActions.
func (app *BaseAuthService) HandleRefreshToken(ctx context.Context, token string) (*shared.UserInfoTokens, error) {
	opts := app.options.Auth
	var claims shared.RefreshTokenClaims
	err := app.token.ParseToken(token, opts.RefreshToken, &claims)
	if err != nil {
		return nil, fmt.Errorf("error verifying refresh token: %w", err)
	}
	_, err = app.authStore.GetToken(ctx, claims.Token)
	if err != nil {
		return nil, fmt.Errorf("error getting token: %w", err) // corrected to return nil before the error
	}
	err = app.authStore.DeleteToken(ctx, claims.Token)
	if err != nil {
		return nil, fmt.Errorf("error deleting token: %w", err)
	}
	info, err := app.authStore.GetUserInfo(ctx, claims.Email)
	if err != nil {
		return nil, err
	}

	return app.CreateAuthTokens(ctx, info)
}

func (app *BaseAuthService) HandleVerificationToken(ctx context.Context, token string) error {
	claims, err := app.VerifyAndParseOtpToken(ctx, EmailTypeVerify, token)
	if err != nil {
		return fmt.Errorf("error verifying verification token: %w", err)
	}
	_, err = app.authStore.GetToken(ctx, claims.Token)
	if err != nil {
		return fmt.Errorf("error getting token: %w", err)
	}
	err = app.authStore.DeleteToken(ctx, claims.Token)
	if err != nil {
		return fmt.Errorf("error deleting token: %w", err)
	}
	user, err := app.authStore.FindUserByEmail(ctx, claims.Email)
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
	err = app.authStore.UpdateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}
	return nil
}

// VerifyAndUseVerificationToken implements AuthActions.
func (app *BaseAuthService) VerifyAndParseOtpToken(ctx context.Context, emailType EmailType, token string) (*shared.OtpClaims, error) {
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

func (app *BaseAuthService) Authenticate(ctx context.Context, params *shared.AuthenticationInput) (*models.User, error) {
	var user *models.User
	var account *models.UserAccount
	var err error
	var isFirstLogin bool

	// get user by email
	user, err = app.authStore.FindUserByEmail(ctx, params.Email)
	if err != nil {
		return nil, fmt.Errorf("error at getting user by email: %w", err)
	}
	// if user exists, get user account
	if user != nil {
		account, err = app.authStore.FindUserAccountByUserIdAndProvider(ctx, user.ID, models.ProvidersCredentials)
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
		user, err = app.authStore.CreateUser(ctx, &models.User{
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
		err = app.authStore.AssignUserRoles(ctx, user.ID, shared.PermissionNameBasic)
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
			pw, err := app.password.HashPassword(*params.Password)
			if err != nil {
				return nil, fmt.Errorf("error at hashing password: %w", err)
			}
			params.HashPassword = &pw
		}
		// link account of requested type
		newVar := &models.UserAccount{
			UserID:            user.ID,
			Type:              models.ProviderTypes(params.Type),
			Provider:          models.Providers(params.Provider),
			ProviderAccountID: params.ProviderAccountID,
			Password:          params.HashPassword,
			AccessToken:       params.AccessToken,
			RefreshToken:      params.RefreshToken,
		}
		err = app.authStore.LinkAccount(ctx, newVar)
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
		if match, err := app.password.VerifyPassword(*account.Password, *params.Password); err != nil {
			return nil, fmt.Errorf("error at comparing password: %w", err)
		} else if !match {
			return nil, fmt.Errorf("password is incorrect")
		}
	}
	return user, nil
}

func (app *BaseAuthService) CheckUserCredentialsSecurity(ctx context.Context, user *models.User, params *shared.AuthenticationInput) error {

	if user == nil || params == nil {
		return fmt.Errorf("user not found")
	}
	// if user is not verified,
	if user.EmailVerifiedAt == nil {
		if params.EmailVerifiedAt != nil {
			// and if incoming request is oauth,
			if params.Type == shared.ProviderTypeOAuth {
				//  check if user has a credentials account
				account, err := app.authStore.FindUserAccountByUserIdAndProvider(ctx, user.ID, models.ProvidersCredentials)

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
					err = app.authStore.UpdateUserAccount(ctx, account)
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
			err := app.authStore.UpdateUser(ctx, user)
			if err != nil {
				return fmt.Errorf("error updating user email confirmation: %w", err)
			}
		}
	}
	return nil
}

type SendMailParams struct {
	Subject      string
	Type         string
	TemplatePath string
	Template     string
}

// SendOtpEmail creates and saves a new otp token and sends it to the user's email
func (app *BaseAuthService) SendOtpEmail(emailType EmailType, ctx context.Context, user *models.User) error {
	appOpts := app.options.Meta
	var tokenOpts conf.TokenOption
	switch emailType {
	case EmailTypeVerify:
		tokenOpts = app.options.Auth.VerificationToken
	case EmailTypeSecurityPasswordReset:
		tokenOpts = app.options.Auth.PasswordResetToken
	case EmailTypeConfirmPasswordReset:
		tokenOpts = app.options.Auth.PasswordResetToken
	default:
		return fmt.Errorf("invalid email type")
	}

	tokenKey := security.GenerateTokenKey()
	otp := security.GenerateOtp(6)
	expires := tokenOpts.ExpiresAt()
	userId := user.ID
	email := user.Email
	ttype := tokenOpts.Type

	payload := shared.OtpPayload{
		Type:       ttype,
		UserId:     userId,
		Email:      email,
		Token:      tokenKey,
		Otp:        otp,
		RedirectTo: app.options.Meta.AppUrl,
	}

	tokenHash, err := app.CreateOtpTokenHash(&payload, tokenOpts)
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

	err = app.authStore.SaveToken(ctx, dto)

	if err != nil {
		return fmt.Errorf("error at creating verification token: %w", err)
	}

	// err = app.mail.SendOtpEmail(emailType, tokenHash, &payload)
	// if payload == nil {
	// 	return fmt.Errorf("payload is nil")
	// }

	var params SendMailParams
	var ok bool
	if params, ok = EmailPathMap[emailType]; !ok {
		return fmt.Errorf("email type not found")
	}
	path, err := mailer.GetPath(params.TemplatePath, &mailer.EmailParams{
		Token:      tokenHash,
		Type:       string(payload.Type),
		RedirectTo: payload.RedirectTo,
	})
	if err != nil {
		return err
	}
	appUrl, err := url.Parse(appOpts.AppUrl)
	if err != nil {
		return err
	}
	param := &mailer.CommonParams{
		SiteURL:         appUrl.String(),
		ConfirmationURL: appUrl.ResolveReference(path).String(),
		Email:           payload.Email,
		Token:           payload.Otp,
		TokenHash:       tokenHash,
		RedirectTo:      payload.RedirectTo,
	}
	bodyStr := mailer.GetTemplate("body", params.Template, param)
	mailParams := &mailer.Message{
		From:    appOpts.SenderAddress,
		To:      payload.Email,
		Subject: fmt.Sprintf(params.Subject, appOpts.AppName),
		Body:    bodyStr,
	}
	return app.mail.Send(mailParams)

}

func (app *BaseAuthService) CreateOtpTokenHash(payload *shared.OtpPayload, config conf.TokenOption) (string, error) {
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
