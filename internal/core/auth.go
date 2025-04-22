package core

import (
	"context"
	"fmt"
	"net/url"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mailer"
	"github.com/tkahng/authgo/internal/tools/security"
)

var (
	ErrPasswordIncorrect  = shared.AppError{Status: 400, Message: "password incorrect"}
	ErrInvalidProviderKey = shared.AppError{Status: 400, Message: "invalid provider key"}
	ErrBadRequest         = shared.AppError{Status: 400, Message: "input is missing"}
)

func (a *BaseApp) CreateAuthTokens(ctx context.Context, db bob.Executor, payload *shared.UserInfoDto) (*shared.TokenDto, error) {
	if payload == nil {
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

func (app *BaseApp) CheckUserCredentialsSecurity(ctx context.Context, db bob.Executor, user *models.User, params *shared.AuthenticateUserParams) error {

	// err := user.LoadUserUserAccounts(ctx, db, models.SelectWhere.UserAccounts.UserID.EQ(user.ID))
	if user == nil || params == nil {
		return fmt.Errorf("user not found")
	}
	// if user is not verified,
	if user.EmailVerifiedAt.IsNull() {
		if params.EmailVerifiedAt != nil {
			// and if incoming request is oauth,
			if params.Type == models.ProviderTypesOauth {
				//  check if user has a credentials account
				account, err := repository.FindUserAccountByUserIdAndProvider(ctx, db, user.ID, models.ProvidersCredentials)
				if err != nil {
					return fmt.Errorf("error loading user accounts: %w", err)
				}
				if account != nil {
					// if user has a credentials account, send security password reset email
					randomPassword := security.RandomString(20)
					err = repository.UpdateUserPassword(ctx, db, user.ID, randomPassword)
					if err != nil {
						return fmt.Errorf("error updating user password: %w", err)
					}
					err = app.SendSecurityPasswordResetEmail(ctx, db, user, app.Settings().Meta.AppURL)
					if err != nil {
						return fmt.Errorf("error sending password reset email: %w", err)
					}
				}
			}
			_, err := repository.UpdateUserEmailConfirm(ctx, db, user.ID, *params.EmailVerifiedAt)
			if err != nil {
				return fmt.Errorf("error updating user email confirmation: %w", err)
			}
		}
	}
	return nil
}

func createVerificationMailParams(tokenHash string, payload *OtpPayload, config *AppOptions) (*mailer.Message, error) {
	path, err := mailer.GetPath("/api/auth/verify", &mailer.EmailParams{
		Token:      tokenHash,
		Type:       string(shared.VerificationTokenType),
		RedirectTo: payload.RedirectTo,
	})
	if err != nil {
		return nil, err
	}
	appUrl, err := url.Parse(config.Meta.AppURL)
	if err != nil {
		return nil, err
	}
	param := &mailer.CommonParams{
		SiteURL:         appUrl.String(),
		ConfirmationURL: appUrl.ResolveReference(path).String(),
		Email:           payload.Email,
		Token:           payload.Otp,
		TokenHash:       tokenHash,
		RedirectTo:      payload.RedirectTo,
	}
	bodyStr := mailer.GetTemplate("body", mailer.DefaultConfirmationMail, param)
	mailParams := &mailer.Message{
		From:    config.Meta.SenderAddress,
		To:      payload.Email,
		Subject: fmt.Sprintf("%s - Verify your email address", config.Meta.AppName),
		Body:    bodyStr,
	}
	return mailParams, err
}

func createPasswordResetMailParams(tokenHash string, payload *OtpPayload, config *AppOptions) (*mailer.Message, error) {
	path, err := mailer.GetPath("/api/auth/confirm-password-reset", &mailer.EmailParams{
		Token:      tokenHash,
		Type:       string(shared.PasswordResetTokenType),
		RedirectTo: payload.RedirectTo,
	})
	if err != nil {
		return nil, err
	}
	appUrl, err := url.Parse(config.Meta.AppURL)
	if err != nil {
		return nil, err
	}
	param := &mailer.CommonParams{
		SiteURL:         appUrl.String(),
		ConfirmationURL: appUrl.ResolveReference(path).String(),
		Email:           payload.Email,
		Token:           payload.Otp,
		TokenHash:       tokenHash,
		RedirectTo:      payload.RedirectTo,
	}
	bodyStr := mailer.GetTemplate("body", mailer.DefaultRecoveryMail, param)
	mailParams := &mailer.Message{
		From:    config.Meta.SenderAddress,
		To:      payload.Email,
		Subject: fmt.Sprintf("%s - Verify your email address", config.Meta.AppName),
		Body:    bodyStr,
	}
	return mailParams, err
}

func createSecurityPasswordResetMailParams(tokenHash string, payload *OtpPayload, config *AppOptions) (*mailer.Message, error) {
	path, err := mailer.GetPath("/api/auth/confirm-password-reset", &mailer.EmailParams{
		Token:      tokenHash,
		Type:       string(shared.PasswordResetTokenType),
		RedirectTo: payload.RedirectTo,
	})
	if err != nil {
		return nil, err
	}
	appUrl, err := url.Parse(config.Meta.AppURL)
	if err != nil {
		return nil, err
	}
	param := &mailer.CommonParams{
		SiteURL:         appUrl.String(),
		ConfirmationURL: appUrl.ResolveReference(path).String(),
		Email:           payload.Email,
		Token:           payload.Otp,
		TokenHash:       tokenHash,
		RedirectTo:      payload.RedirectTo,
	}
	bodyStr := mailer.GetTemplate("body", mailer.DefaultSecurityPasswordResetMail, param)
	mailParams := &mailer.Message{
		From:    config.Meta.SenderAddress,
		To:      payload.Email,
		Subject: fmt.Sprintf("%s - Reset your password", config.Meta.AppName),
		Body:    bodyStr,
	}
	return mailParams, err
}
