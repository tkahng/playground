package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/auth/oauth"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/jobs"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/mailer"
	"github.com/tkahng/authgo/internal/tools/security"
	"golang.org/x/oauth2"
)

type AuthService interface {
	// properties -----------------------------------------------------------------------------------------------------------

	Password() PasswordService
	Token() JwtService
	Mail() MailService

	// handlers -----------------------------------------------------------------------------------------------------------

	HandlePasswordResetRequest(ctx context.Context, email string) error
	HandleAccessToken(ctx context.Context, token string) (*shared.UserInfo, error)
	HandleRefreshToken(ctx context.Context, token string) (*shared.UserInfoTokens, error)
	HandleVerificationToken(ctx context.Context, token string) error
	HandlePasswordResetToken(ctx context.Context, token, password string) error
	HandleCheckResetPasswordToken(ctx context.Context, token string) error
	Signout(ctx context.Context, token string) error
	ResetPassword(ctx context.Context, userId uuid.UUID, oldPassword, newPassword string) error

	// methods -----------------------------------------------------------------------------------------------------------

	VerifyStateToken(ctx context.Context, token string) (*shared.ProviderStateClaims, error)
	CreateOAuthUrl(ctx context.Context, provider shared.Providers, redirectUrl string) (string, error)
	CreateAndPersistStateToken(ctx context.Context, payload *shared.ProviderStatePayload) (string, error)
	FetchAuthUser(ctx context.Context, code string, parsedState *shared.ProviderStateClaims) (*oauth.AuthUser, error)
	VerifyAndParseOtpToken(ctx context.Context, emailType mailer.EmailType, token string) (*shared.OtpClaims, error)
	Authenticate(ctx context.Context, params *shared.AuthenticationInput) (*models.User, error)
	CreateAuthTokensFromEmail(ctx context.Context, email string) (*shared.UserInfoTokens, error)
	SendOtpEmail(emailType mailer.EmailType, ctx context.Context, user *models.User, adapter stores.StorageAdapterInterface) error
}

var _ AuthService = (*BaseAuthService)(nil)

type BaseAuthService struct {
	routine RoutineService
	// authStore AuthStore
	mail     MailService
	token    JwtService
	password PasswordService
	options  *conf.AppOptions
	logger   *slog.Logger
	enqueuer jobs.Enqueuer
	adapter  stores.StorageAdapterInterface
}

func NewAuthService(
	opts *conf.AppOptions,
	mail MailService,
	token JwtService,
	password PasswordService,
	workerService RoutineService,
	logger *slog.Logger,
	enqueuer jobs.Enqueuer,
	adapter stores.StorageAdapterInterface,
) AuthService {
	authService := &BaseAuthService{
		routine:  workerService,
		mail:     mail,
		token:    token,
		password: password,
		options:  opts,
		logger:   logger,
		enqueuer: enqueuer,
		adapter:  adapter,
	}

	return authService
}

// Mail implements AuthService.
func (app *BaseAuthService) Mail() MailService {
	return app.mail
}

// Password implements AuthService.
func (app *BaseAuthService) Password() PasswordService {
	return app.password
}

// Token implements AuthService.
func (app *BaseAuthService) Token() JwtService {
	return app.token
}

// CreateOAuthUrl implements AuthService.
func (app *BaseAuthService) CreateOAuthUrl(ctx context.Context, providerName shared.Providers, redirectUrl string) (string, error) {
	redirectTo := redirectUrl
	if redirectTo == "" {
		redirectTo = app.options.Meta.AppUrl
	}
	provider := oauth.NewProviderByName(string(providerName))
	if provider == nil {
		return "", fmt.Errorf("provider %v not found", providerName)
	}
	if !provider.Active() {
		return "", fmt.Errorf("provider %v is not enabled", providerName)
	}
	urlOpts := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline,
	}
	info := &shared.ProviderStatePayload{
		Type:       shared.TokenTypesStateToken,
		Provider:   providerName,
		RedirectTo: redirectTo,
		Token:      security.GenerateTokenKey(),
	}
	if provider.Pkce() {

		info.CodeVerifier = security.RandomString(43)
		info.CodeChallenge = security.S256Challenge(info.CodeVerifier)
		info.CodeChallengeMethod = "S256"
		urlOpts = append(urlOpts,
			oauth2.SetAuthURLParam("code_challenge", info.CodeChallenge),
			oauth2.SetAuthURLParam("code_challenge_method", info.CodeChallengeMethod),
		)
	}
	state, err := app.CreateAndPersistStateToken(ctx, info)
	if err != nil {
		return "", err
	}
	res := provider.BuildAuthURL(state, urlOpts...)
	if res == "" {
		return "", fmt.Errorf("error at building auth url")
	}
	return res, nil
}

// FireAndForget implements AuthService.
// Subtle: this method shadows the method (WorkerService).FireAndForget of BaseAuthService.WorkerService.
// SendOtpEmail creates and saves a new otp token and sends it to the user's email
func (app *BaseAuthService) SendOtpEmail(emailType mailer.EmailType, ctx context.Context, user *models.User, adapter stores.StorageAdapterInterface) error {
	if adapter == nil {
		adapter = app.adapter
	}
	if app.options == nil {
		return fmt.Errorf("app options is nil")
	}
	if app.token == nil {
		return fmt.Errorf("token service is nil")
	}
	if app.mail == nil {
		return fmt.Errorf("mail service is nil")
	}
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	appOpts := app.options.Meta
	var tokenOpts conf.TokenOption
	switch emailType {
	case mailer.EmailTypeVerify:
		tokenOpts = app.options.Auth.VerificationToken
	case mailer.EmailTypeSecurityPasswordReset:
		tokenOpts = app.options.Auth.PasswordResetToken
	case mailer.EmailTypeConfirmPasswordReset:
		tokenOpts = app.options.Auth.PasswordResetToken
	default:
		return fmt.Errorf("invalid email type")
	}

	claims := shared.OtpClaims{}
	claims.ExpiresAt = tokenOpts.ExpiresAt()
	claims.Type = tokenOpts.Type
	claims.UserId = user.ID
	claims.Email = user.Email
	claims.Token = security.GenerateTokenKey()
	claims.Otp = security.GenerateOtp(6)
	claims.RedirectTo = appOpts.AppUrl

	tokenHash, err := app.token.CreateJwtToken(claims, tokenOpts.Secret)
	if err != nil {
		return fmt.Errorf("error at creating verification token: %w", err)
	}
	dto := &shared.CreateTokenDTO{
		Expires:    claims.ExpiresAt.Time,
		Token:      claims.Token,
		Type:       claims.Type,
		Identifier: claims.Email,
		UserID:     &claims.UserId,
	}
	err = adapter.Token().SaveToken(ctx, dto)
	// err = app.authStore.SaveToken(ctx, dto)
	if err != nil {
		return fmt.Errorf("error at creating verification token: %w", err)
	}

	sendMailParams, err := app.GetSendMailParams(emailType, tokenHash, claims)
	if err != nil {
		return fmt.Errorf("error at getting send mail params: %w", err)
	}

	return app.mail.SendMail(sendMailParams)
}

func (app *BaseAuthService) GetSendMailParams(emailType mailer.EmailType, tokenHash string, claims shared.OtpClaims) (*mailer.AllEmailParams, error) {
	appOpts := app.options.Meta
	var sendMailParams mailer.SendMailParams
	var ok bool
	if sendMailParams, ok = mailer.EmailPathMap[emailType]; !ok {
		return nil, fmt.Errorf("email type not found")
	}
	path, err := mailer.GetPathParams(sendMailParams.TemplatePath, tokenHash, string(claims.Type), claims.RedirectTo)
	if err != nil {
		return nil, err
	}
	appUrl, err := url.Parse(appOpts.AppUrl)
	if err != nil {
		return nil, err
	}
	common := &mailer.CommonParams{
		SiteURL:         appUrl.String(),
		ConfirmationURL: appUrl.ResolveReference(path).String(),
		Email:           claims.Email,
		Token:           claims.Otp,
		TokenHash:       tokenHash,
		RedirectTo:      claims.RedirectTo,
	}
	message := &mailer.Message{
		From:    appOpts.SenderAddress,
		To:      common.Email,
		Subject: fmt.Sprintf(sendMailParams.Subject, appOpts.AppName),
		Body:    mailer.GenerateBody("body", sendMailParams.Template, common),
	}
	allEmailParams := &mailer.AllEmailParams{
		SendMailParams: &sendMailParams,
		CommonParams:   common,
		Message:        message,
	}
	return allEmailParams, nil
}

// FetchAuthUser implements Authenticator.
func (app *BaseAuthService) FetchAuthUser(ctx context.Context, code string, parsedState *shared.ProviderStateClaims) (*oauth.AuthUser, error) {
	var provider = oauth.NewProviderByName(parsedState.Provider.String())
	if provider == nil {
		return nil, fmt.Errorf("provider %v not found", parsedState.Provider)
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
	account, err := app.adapter.UserAccount().FindUserAccount(
		ctx,
		&stores.UserAccountFilter{
			UserIds:   []uuid.UUID{userId},
			Providers: []models.Providers{models.ProvidersCredentials},
		},
	)
	// account, err := app.authStore.FindUserAccountByUserIdAndProvider(ctx, userId, models.ProvidersCredentials)
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

	err = app.adapter.UserAccount().UpdateUserAccount(ctx, account)
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
	_, err = app.adapter.Token().GetToken(ctx, token) // corrected 'tokne' to 'token'
	// _, err = app.authStore.GetToken(ctx, token) // corrected 'tokne' to 'token'
	if err != nil {
		return err
	}
	err = app.adapter.Token().DeleteToken(ctx, token) // corrected to use 'app.token'
	if err != nil {
		return fmt.Errorf("error at deleting token: %w", err)
	}
	return nil
}

// HandlePasswordResetRequest implements AuthActions.
func (app *BaseAuthService) HandlePasswordResetRequest(ctx context.Context, email string) error {

	user, err := app.adapter.User().FindUser(
		ctx,
		&stores.UserFilter{
			Emails: []string{email},
		},
	)
	if err != nil {
		return fmt.Errorf("error getting user by email: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	account, err := app.adapter.UserAccount().FindUserAccount(ctx, &stores.UserAccountFilter{
		UserIds:   []uuid.UUID{user.ID},
		Providers: []models.Providers{models.ProvidersCredentials},
	})
	if err != nil {
		return fmt.Errorf("error getting user account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("user account not found")
	}

	err = app.SendOtpEmail(mailer.EmailTypeConfirmPasswordReset, ctx, user, app.adapter)
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
	token, err := app.token.CreateJwtToken(claims, config.Secret)
	if err != nil {
		return token, err
	}

	err = app.adapter.Token().SaveToken(ctx, &shared.CreateTokenDTO{
		Type:       shared.TokenTypesStateToken,
		Identifier: payload.Token,
		Expires:    config.Expires(),
		Token:      payload.Token,
	})
	if err != nil {
		return token, err
	}
	return token, nil
}

// CreateAuthTokensFromEmail implements AuthActions.
func (app *BaseAuthService) CreateAuthTokensFromEmail(ctx context.Context, email string) (*shared.UserInfoTokens, error) {
	user, err := app.adapter.User().GetUserInfo(ctx, email)
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

		claims := shared.RefreshTokenClaims{
			Type:             shared.TokenTypesRefreshToken,
			RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: opts.RefreshToken.ExpiresAt()},
			RefreshTokenPayload: shared.RefreshTokenPayload{
				UserId: payload.User.ID,
				Email:  payload.User.Email,
				Token:  tokenKey,
			},
		}

		token, err := app.token.CreateJwtToken(claims, opts.RefreshToken.Secret)
		if err != nil {
			return token, err
		}
		err = app.adapter.Token().SaveToken(
			ctx,
			&shared.CreateTokenDTO{
				Type:       shared.TokenTypesRefreshToken,
				Identifier: claims.Email,
				Expires:    opts.RefreshToken.Expires(),
				Token:      claims.Token,
				UserID:     &claims.UserId,
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

// HandleCheckResetPasswordToken implements AuthActions.
func (app *BaseAuthService) HandleCheckResetPasswordToken(ctx context.Context, tokenHash string) error {
	opts := app.options.Auth
	var claims shared.PasswordResetClaims
	err := app.token.ParseToken(tokenHash, opts.PasswordResetToken, &claims)
	if err != nil {
		return fmt.Errorf("error verifying password reset token: %w", err)
	}
	token, err := app.adapter.Token().GetToken(ctx, claims.Token)
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
	_, err = app.adapter.Token().GetToken(ctx, token) // corrected 'tokne' to 'token'
	if err != nil {
		return err
	}
	err = app.adapter.Token().DeleteToken(ctx, token) // corrected to use 'app.token'
	if err != nil {
		return fmt.Errorf("error deleting token: %w", err)
	}

	user, err := app.adapter.User().FindUser(
		ctx,
		&stores.UserFilter{
			Emails: []string{claims.Email},
		},
	)
	if err != nil {
		return fmt.Errorf("error getting user by email: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	account, err := app.adapter.UserAccount().FindUserAccount(ctx, &stores.UserAccountFilter{
		UserIds:   []uuid.UUID{user.ID},
		Providers: []models.Providers{models.ProvidersCredentials},
	})
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
	err = app.adapter.UserAccount().UpdateUserAccount(ctx, account)
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
	_, err = app.adapter.Token().GetToken(ctx, token) // corrected 'tokne' to 'token'
	if err != nil {
		return nil, err
	}
	err = app.adapter.Token().DeleteToken(ctx, token) // corrected to use 'app.token'
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
	return app.adapter.User().GetUserInfo(ctx, claims.Email)
}

// HandleRefreshToken implements AuthActions.
func (app *BaseAuthService) HandleRefreshToken(ctx context.Context, token string) (*shared.UserInfoTokens, error) {
	opts := app.options.Auth
	var claims shared.RefreshTokenClaims
	err := app.token.ParseToken(token, opts.RefreshToken, &claims)
	if err != nil {
		return nil, fmt.Errorf("error verifying refresh token: %w", err)
	}
	_, err = app.adapter.Token().GetToken(ctx, claims.Token)
	if err != nil {
		return nil, fmt.Errorf("error getting token: %w", err) // corrected to return nil before the error
	}
	err = app.adapter.Token().DeleteToken(ctx, claims.Token)
	if err != nil {
		return nil, fmt.Errorf("error deleting token: %w", err)
	}
	info, err := app.adapter.User().GetUserInfo(ctx, claims.Email)
	if err != nil {
		return nil, err
	}

	return app.CreateAuthTokens(ctx, info)
}

func (app *BaseAuthService) HandleVerificationToken(ctx context.Context, token string) error {
	claims, err := app.VerifyAndParseOtpToken(ctx, mailer.EmailTypeVerify, token)
	if err != nil {
		return fmt.Errorf("error verifying verification token: %w", err)
	}
	_, err = app.adapter.Token().GetToken(ctx, claims.Token)
	if err != nil {
		return fmt.Errorf("error getting token: %w", err)
	}
	err = app.adapter.Token().DeleteToken(ctx, claims.Token)
	if err != nil {
		return fmt.Errorf("error deleting token: %w", err)
	}
	user, err := app.adapter.User().FindUser(
		ctx,
		&stores.UserFilter{
			Emails: []string{claims.Email},
		},
	)
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
	err = app.adapter.User().UpdateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}
	return nil
}

// VerifyAndUseVerificationToken implements AuthActions.
func (app *BaseAuthService) VerifyAndParseOtpToken(ctx context.Context, emailType mailer.EmailType, token string) (*shared.OtpClaims, error) {
	var opt conf.TokenOption
	switch emailType {
	case mailer.EmailTypeVerify:
		opt = app.options.Auth.VerificationToken
	case mailer.EmailTypeConfirmPasswordReset:
		opt = app.options.Auth.PasswordResetToken
	case mailer.EmailTypeSecurityPasswordReset:
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

func (app *BaseAuthService) CreateUserAndAccount(ctx context.Context, params *shared.AuthenticationInput, adapter stores.StorageAdapterInterface) (*models.User, error) {
	user, err := adapter.User().CreateUser(
		ctx,
		&models.User{
			Email:           params.Email,
			Name:            params.Name,
			Image:           params.AvatarUrl,
			EmailVerifiedAt: params.EmailVerifiedAt,
		},
	)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not created")
	}
	err1 := app.HashInputPassword(params)
	if err1 != nil {
		return nil, err1
	}
	account, err := adapter.UserAccount().CreateUserAccount(
		ctx,
		&models.UserAccount{
			UserID:            user.ID,
			Type:              models.ProviderTypes(params.Type),
			Provider:          models.Providers(params.Provider),
			ProviderAccountID: params.ProviderAccountID,
			Password:          params.HashPassword,
			AccessToken:       params.AccessToken,
			RefreshToken:      params.RefreshToken,
		},
	)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("user account not created")
	}
	return user, nil
}

func (app *BaseAuthService) CreateAccountFromUser(ctx context.Context, params *shared.AuthenticationInput, adapter stores.StorageAdapterInterface) (*models.UserAccount, error) {
	if params.UserId == nil {
		return nil, fmt.Errorf("user id is required")
	}
	err := app.HashInputPassword(params)
	if err != nil {
		return nil, err
	}
	account, err := adapter.UserAccount().CreateUserAccount(
		ctx,
		&models.UserAccount{
			UserID:            *params.UserId,
			Type:              models.ProviderTypes(params.Type),
			Provider:          models.Providers(params.Provider),
			ProviderAccountID: params.ProviderAccountID,
			Password:          params.HashPassword,
			AccessToken:       params.AccessToken,
			RefreshToken:      params.RefreshToken,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error at creating user account: %w", err)
	}
	if account == nil {
		return nil, fmt.Errorf("user account not created")
	}
	return account, nil
}

func (app *BaseAuthService) Authenticate(ctx context.Context, params *shared.AuthenticationInput) (*models.User, error) {
	var user *models.User
	var account *models.UserAccount
	var err error

	// find user by email ----------------------------------------------------------------------------------------------------
	user, err = app.adapter.User().FindUser(
		ctx,
		&stores.UserFilter{
			Emails: []string{params.Email},
		},
	)
	if err != nil {
		return nil, err
	}
	// if user is not found, create user and account, then send verification email ----------------------------------------------------------------------------------------------------
	if user == nil {
		return app.authenticateNewUser(ctx, params)
	}

	params.UserId = &user.ID

	account, err = app.adapter.UserAccount().FindUserAccount(ctx, &stores.UserAccountFilter{
		UserIds:   []uuid.UUID{user.ID},
		Providers: []models.Providers{models.Providers(params.Provider)},
	})

	if err != nil {
		return nil, err
	}
	// if user exists, but requested account type does not exist, Create UserAccount  of requested type ----------------------------------------------------------------------------------------------------
	if account == nil {
		return app.authenticateNewAccount(ctx, user, params)
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
func (app *BaseAuthService) authenticateNewAccount(ctx context.Context, user *models.User, params *shared.AuthenticationInput) (*models.User, error) {
	if user == nil {
		return nil, fmt.Errorf("user is nil")
	}
	var resetPassword bool
	err := app.adapter.RunInTx(func(tx stores.StorageAdapterInterface) error {
		var err error
		resetPassword, err = app.CheckAndResetCredentialsPassword(ctx, user, params, app.adapter)
		if err != nil {
			return err
		}

		_, err = app.CreateAccountFromUser(ctx, params, tx)
		if err != nil {
			return err
		}
		err = app.UpdateUserEmailVerifiedAt(ctx, user, params, tx)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if resetPassword {
		app.routine.FireAndForget(
			func() {
				ctx := context.Background()
				fmt.Println("User is first login, sending reset password email")
				err := app.SendOtpEmail(mailer.EmailTypeSecurityPasswordReset, ctx, user, nil)
				if err != nil {
					app.logger.Error(
						"error sending reset password email",
						slog.Any("error", err),
						slog.String("email", user.Email),
						slog.String("userId", user.ID.String()),
					)
				}
			},
		)
	}
	if user.EmailVerifiedAt == nil {
		app.routine.FireAndForget(
			func() {
				ctx := context.Background()
				fmt.Println("User is first login, sending verification email")
				err := app.SendOtpEmail(mailer.EmailTypeVerify, ctx, user, nil)
				if err != nil {
					app.logger.Error(
						"error sending verification email",
						slog.Any("error", err),
						slog.String("email", user.Email),
						slog.String("userId", user.ID.String()),
					)
				}
			},
		)
	}
	return user, nil
}
func (app *BaseAuthService) authenticateNewUser(ctx context.Context, params *shared.AuthenticationInput) (*models.User, error) {
	var user *models.User
	err := app.adapter.RunInTx(func(tx stores.StorageAdapterInterface) error {
		newUser, err := app.CreateUserAndAccount(ctx, params, app.adapter)
		user = newUser
		return err
	})
	if err != nil {
		return nil, err
	}
	app.routine.FireAndForget(
		func() {
			ctx := context.Background()
			fmt.Println("User is first login, sending verification email")
			err := app.SendOtpEmail(mailer.EmailTypeVerify, ctx, user, app.adapter)
			if err != nil {
				app.logger.Error(
					"error sending verification email",
					slog.Any("error", err),
					slog.String("email", user.Email),
					slog.String("userId", user.ID.String()),
				)
			}
		},
	)
	return user, nil
}

// HashInputPassword hashes the password and sets the HashPassword field in the params if it is not already set.
func (app *BaseAuthService) HashInputPassword(params *shared.AuthenticationInput) error {
	if params.Type == shared.ProviderTypeCredentials {
		if params.HashPassword == nil {
			if params.Password == nil {
				return fmt.Errorf("password is nil")
			}
			if pw, err := app.password.HashPassword(*params.Password); err != nil {
				return fmt.Errorf("error at hashing password: %w", err)
			} else {
				params.HashPassword = &pw
			}
			if params.HashPassword == nil {
				return fmt.Errorf("password is nil")
			}
		}
	}
	return nil
}

func (app *BaseAuthService) UpdateUserEmailVerifiedAt(ctx context.Context, user *models.User, params *shared.AuthenticationInput, adapter stores.StorageAdapterInterface) error {
	if user == nil {
		return errors.New("user is nil")
	}
	if params == nil {
		return errors.New("params is nil")
	}
	if params.EmailVerifiedAt == nil {
		return nil
	}
	if user.EmailVerifiedAt != nil {
		return nil
	}
	user.EmailVerifiedAt = params.EmailVerifiedAt
	err := adapter.User().UpdateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("error updating user email verified at: %w", err)
	}
	return nil
}

// if incoming request is a oauth type,
// and if user already has a credentials account,
// and if user email is not verified,
// then reset the password of the credentials account.
// if there was a reset, return true, else return false
func (app *BaseAuthService) CheckAndResetCredentialsPassword(ctx context.Context, user *models.User, params *shared.AuthenticationInput, adapter stores.StorageAdapterInterface) (bool, error) {
	if user == nil {
		return false, fmt.Errorf("user is nil")
	}
	if params == nil {
		return false, fmt.Errorf("params is nil")
	}
	// if incoming request is not a oauth type, return false
	if params.Type == shared.ProviderTypeCredentials {
		return false, nil
	}
	// if incoming request does not have email verified at, return false
	if params.EmailVerifiedAt == nil {
		return false, nil
	}
	// if user email is verified, return false
	if user.EmailVerifiedAt != nil {
		return false, nil
	}
	account, err := adapter.UserAccount().FindUserAccount(ctx, &stores.UserAccountFilter{
		UserIds:   []uuid.UUID{user.ID},
		Providers: []models.Providers{models.ProvidersCredentials},
	})
	if err != nil {
		return false, fmt.Errorf("error finding credentials account: %w", err)
	}
	if account == nil {
		return false, nil
	}
	randomPassword := security.RandomString(20)
	hash, err := app.password.HashPassword(randomPassword)
	if err != nil {
		return false, fmt.Errorf("error at hashing password: %w", err)
	}
	account.Password = &hash
	err = adapter.UserAccount().UpdateUserAccount(ctx, account)
	if err != nil {
		return false, fmt.Errorf("error updating user password: %w", err)
	}
	return true, nil
}
