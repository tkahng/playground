package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/auth/oauth"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mailer"
	"github.com/tkahng/playground/internal/tools/security"
	"github.com/tkahng/playground/internal/tools/types"
	"github.com/tkahng/playground/internal/workers"
	"golang.org/x/oauth2"
)

type (
	OAuth2SigninInput struct {
		Email             string           `json:"email" required:"true" format:"email" maxLength:"100"`
		Name              *string          `json:"name"`
		AvatarUrl         *string          `json:"avatar_url"`
		EmailVerifiedAt   *time.Time       `json:"email_verified_at"`
		Provider          models.Providers `json:"provider"`
		ProviderAccountID string           `json:"provider_account_id"`
		UserId            *uuid.UUID       `json:"user_id"`
		AccessToken       *string          `json:"access_token"`
		RefreshToken      *string          `json:"refresh_token"`
		RedirectTo        string           `json:"redirect_to"`
		Expiry            time.Time        `json:"expiry"`
	}
	OAuth2UrlInput struct {
		Provider models.Providers
	}
)

type Oauth2Authenticator interface {
	OAuth2Url(ctx context.Context, provider models.Providers, redirectUrl string, email string) (string, error)
	// OAuth2Signin user.
	// the callback handlers will call this method
	//
	// - if user with email does not exist, it will create a new user and a oauth account.
	//
	// - if user with email exists, and they have another oauth account, it will update the oauth account.
	OAuth2Signin(ctx context.Context, params *OAuth2SigninInput) (*models.UserInfoTokens, error)
	VerifyStateToken(ctx context.Context, token string) (*ProviderStateClaims, error)
	CreateAndPersistStateToken(ctx context.Context, payload *ProviderStatePayload) (string, error)
	FetchAuthUser(ctx context.Context, code string, parsedState *ProviderStateClaims) (*oauth.AuthUser, error)
}

// OAuth2Url implements AuthService.
func (a *AuthServiceImpl) OAuth2Url(ctx context.Context, providerInput models.Providers, redirectUrl string, email string) (string, error) {
	redirectTo := redirectUrl
	if redirectTo == "" {
		redirectTo = a.config.AppUrl
	}
	provider := oauth.NewProviderByName(string(providerInput))
	if provider == nil {
		return "", fmt.Errorf("provider %v not found", providerInput)
	}
	if !provider.Active() {
		return "", fmt.Errorf("provider %v is not enabled", providerInput)
	}
	urlOpts := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline,
	}
	info := &ProviderStatePayload{
		Type:       models.TokenTypesStateToken,
		Provider:   providerInput,
		RedirectTo: redirectTo,
		Email:      email,
		Token:      security.GenerateTokenKey(),
	}
	if provider.Pkce() {
		info.CodeVerifier = oauth2.GenerateVerifier()
		s256challengeOpt := oauth2.S256ChallengeOption(info.CodeVerifier)
		urlOpts = append(urlOpts,
			s256challengeOpt,
		)
	}
	if email != "" {
		info.Email = email
		urlOpts = append(urlOpts,
			oauth2.SetAuthURLParam("login_hint", email),
		)
	}
	state, err := a.CreateAndPersistStateToken(ctx, info)
	if err != nil {
		return "", err
	}
	encryptedState, err := a.encrypter.Encrypt([]byte(state))
	if err != nil {
		return "", err
	}

	res := provider.BuildAuthURL(encryptedState, urlOpts...)
	if res == "" {
		return "", fmt.Errorf("error at building auth url")
	}
	return res, nil
}

// OAuth2Signin implements AuthService.
func (a *AuthServiceImpl) OAuth2Signin(ctx context.Context, params *OAuth2SigninInput) (*models.UserInfoTokens, error) {
	var existingUser *models.User
	var resetPassword bool
	// check if user with email exists.
	existingUser, userErr := a.adapter.User().FindUser(ctx, &stores.UserFilter{
		Emails: []string{params.Email},
	})
	if userErr != nil {
		return nil, userErr
	}
	// if user exists:
	if existingUser != nil {
		// check if user has another oauth account of the same provider
		existingUserAccount, userAccountErr := a.adapter.UserAccount().FindUserAccount(ctx, &stores.UserAccountFilter{
			Providers: []models.Providers{params.Provider},
			UserIds:   []uuid.UUID{existingUser.ID},
		})
		if userAccountErr != nil {
			return nil, userAccountErr
		}
		// if there is another oauth account of the same provider
		if existingUserAccount != nil {
			// but if the provider account id is different then its another account trying to login
			if existingUserAccount.ProviderAccountID != params.ProviderAccountID {
				return nil, shared.ErrAccountProviderConflict
			}
			// if user has another oauth account of the same provider, update it
			existingUserAccount.AccessToken = params.AccessToken
			existingUserAccount.RefreshToken = params.RefreshToken

			err := a.adapter.UserAccount().UpdateUserAccount(ctx, existingUserAccount)
			if err != nil {
				return nil, err
			}
			tokens, err := a.GenerateAuthTokens(ctx, existingUser.Email)
			if err != nil {
				return nil, err
			}
			return tokens, nil
		}
		// check if user is unverified with credentials account
		if existingUser.EmailVerifiedAt == nil {
			resetPassword = true
		}
	}
	// create user and account. these should run inside a transaction.
	// if params.Verified is true, set email_verified_at and skip email verification
	txErr := a.adapter.RunInTxCtx(ctx, func(txCtx context.Context) error {
		var newUser *models.User
		if existingUser == nil {
			user, err := a.adapter.User().CreateUser(txCtx, &models.User{
				Name:            params.Name,
				Email:           params.Email,
				EmailVerifiedAt: params.EmailVerifiedAt,
				Image:           params.AvatarUrl,
			})
			if err != nil {
				return err
			}
			newUser = user
		} else {
			newUser = existingUser
		}
		if resetPassword {
			newUser.EmailVerifiedAt = types.Pointer(time.Now())
			err := a.adapter.User().UpdateUser(txCtx, newUser)
			if err != nil {
				return err
			}
			err = a.randomizeAccountPassword(txCtx, newUser)
			if err != nil {
				return err
			}
		}
		// create oauth account
		_, err := a.adapter.UserAccount().CreateUserAccount(txCtx, &models.UserAccount{
			UserID:            newUser.ID,
			Provider:          params.Provider,
			ProviderAccountID: params.ProviderAccountID,
			Type:              models.ProviderTypeOAuth,
			RefreshToken:      params.RefreshToken,
			AccessToken:       params.AccessToken,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}
	tokens, euserErr := a.GenerateAuthTokens(ctx, params.Email)
	if euserErr != nil {
		return nil, euserErr
	}
	return tokens, nil
}

func (a *AuthServiceImpl) randomizeAccountPassword(txCtx context.Context, newUser *models.User) error {
	credentialsAccount, credentialsAccountErr := a.adapter.UserAccount().FindUserAccount(txCtx, &stores.UserAccountFilter{
		Providers: []models.Providers{models.ProvidersCredentials},
		UserIds:   []uuid.UUID{newUser.ID},
	})
	if credentialsAccountErr != nil {
		return credentialsAccountErr
	}
	if credentialsAccount == nil {
		return fmt.Errorf("credentials account not found")
	}
	randomPassword := security.RandomString(20)
	hash, err := a.hash.Hash(randomPassword)
	if err != nil {
		return err
	}
	credentialsAccount.Password = &hash
	err = a.adapter.UserAccount().UpdateUserAccount(txCtx, credentialsAccount)
	if err != nil {
		return err
	}
	err = a.job.EnqueueOtpMailJob(txCtx, &workers.OtpEmailJobArgs{
		UserID: newUser.ID,
		Type:   mailer.EmailTypeSecurityPasswordReset,
	})
	if err != nil {
		return err
	}
	return nil
}

type ProviderStateClaims struct {
	jwt.RegisteredClaims
	ProviderStatePayload
}

type ProviderStatePayload struct {
	Token        string            `json:"token"`
	Type         models.TokenTypes `json:"type"`
	Provider     models.Providers  `json:"provider"`
	CodeVerifier string            `json:"code_verifier,omitempty"`
	RedirectTo   string            `json:"redirect_to,omitempty"`
	Email        string            `json:"email,omitempty"`
}

// CreateAndPersistStateToken implements AuthActions.
func (a *AuthServiceImpl) CreateAndPersistStateToken(ctx context.Context, payload *ProviderStatePayload) (string, error) {
	if payload == nil {
		return "", fmt.Errorf("payload is nil")
	}
	config := a.config.StateToken
	claims := ProviderStateClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: config.ExpiresAt(),
		},
		ProviderStatePayload: *payload,
	}
	token, err := a.jwt.CreateJwtToken(claims, config.Secret)
	if err != nil {
		return token, err
	}

	err = a.adapter.Token().SaveToken(ctx, &stores.CreateTokenDTO{
		Type:       models.TokenTypesStateToken,
		Identifier: payload.Token,
		Expires:    config.Expires(),
		Token:      payload.Token,
	})
	if err != nil {
		return token, err
	}
	return token, nil
}
func (a *AuthServiceImpl) VerifyStateToken(ctx context.Context, token string) (*ProviderStateClaims, error) {
	decryptedToken, err := a.encrypter.Decrypt(token)
	if err != nil {
		return nil, fmt.Errorf("error decrypting state token: %w", err)
	}

	opts := a.config.AuthOptions
	var claims ProviderStateClaims
	err = a.jwt.ParseToken(string(decryptedToken), opts.StateToken, &claims)
	if err != nil {
		return nil, fmt.Errorf("error verifying state token: %w", err)
	}
	_, err = a.adapter.Token().GetToken(ctx, claims.Token)
	if err != nil {
		return nil, err
	}
	err = a.adapter.Token().DeleteToken(ctx, claims.Token)
	if err != nil {
		return nil, fmt.Errorf("error deleting token: %w", err)
	}
	return &claims, nil
}
func (a *AuthServiceImpl) FetchAuthUser(ctx context.Context, code string, parsedState *ProviderStateClaims) (*oauth.AuthUser, error) {
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
