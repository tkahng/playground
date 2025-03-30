package core

import (
	"context"
	"fmt"
	"net/url"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mailer"
)

func (app *BaseApp) VerifyAndUseVerificationToken(ctx context.Context, db bob.DB, verificationToken string) (*EmailVerificationClaims, error) {
	opts := app.Settings().Auth
	jsond, err := app.TokenVerifier().ParseVerificationToken(verificationToken, opts.VerificationToken)
	if err != nil {
		return nil, fmt.Errorf("error at parsing verification token: %w", err)
	}
	err = app.TokenStorage().UseVerificationTokenAndUpdateUser(ctx, db, jsond.Token)
	if err != nil {
		return nil, fmt.Errorf("error verifying refresh token: %w", err)
	}
	return jsond, nil
}

// SendVerificationEmail implements App.
func (app *BaseApp) SendVerificationEmail(ctx context.Context, db bob.DB, user *models.User, redirectTo string) error {
	opts := app.Settings().Auth
	config := app.Settings()
	client := app.NewMailClient()
	payload := app.TokenVerifier().CreateVerificationPayload(user, redirectTo)

	tokenHash, err := app.TokenVerifier().CreateVerificationToken(payload, opts.VerificationToken)
	if err != nil {
		return fmt.Errorf("error at creating verification token: %w", err)
	}

	err = app.TokenStorage().PersistVerificationToken(ctx, db, payload, opts.VerificationToken)
	if err != nil {
		return fmt.Errorf("error at storing verification token: %w", err)
	}
	mailParams, err := createVerificationMailParams(tokenHash, payload, config)
	client.Send(mailParams)
	if err != nil {
		return fmt.Errorf("error creating verification token: %w", err)
	}
	return nil
}

func createVerificationMailParams(tokenHash string, payload *OtpPayload, config *AppOptions) (*mailer.Message, error) {
	path, err := mailer.GetPath("/api/auth/confirm-verification", &mailer.EmailParams{
		Token:      tokenHash,
		Type:       string(shared.VerificationTokenType),
		RedirectTo: payload.RedirectTo,
	})
	appUrl, _ := url.Parse(config.Meta.AppURL)
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

// func SwitchTokenType(ctx context.Context, db bob.DB, tokenType shared.TokenType, token string) error {
// 	switch tokenType {
// 	case shared.VerificationTokenType:
// 		_, err := VerifyAndUseVerificationToken(ctx, db, token, TokenConfig{})
// 		if err != nil {
// 			return fmt.Errorf("error verifying verification token: %w", err)
// 		}
// 	case shared.AuthenticationTokenType:
// 		_, err := VerifyAuthenticationToken(token, TokenConfig{})
// 		if err != nil {
// 			return fmt.Errorf("error verifying authentication token: %w", err)
// 		}
// 	case shared.RefreshTokenType:
// 		_, err := VerifyRefreshToken(ctx, db, token, TokenConfig{})
// 		if err != nil {
// 			return fmt.Errorf("error verifying refresh token: %w", err)
// 		}
// 	}
// 	return nil
// }
// {
// 	"From": "support@example.com",
// 	"To": "tkahng+01@gmail.com",
// 	"Subject": "Acme - Verify your email address",
// 	"Body": "\u003ch2\u003eConfirm your email\u003c/h2\u003e\n\n\u003cp\u003eFollow this link to confirm your email:\u003c/p\u003e\n\u003cp\u003e\u003ca href=\"http://localhost:8080/api/auth/confirm-?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDM0MDEzNjksInR5cGUiOiJ2ZXJpZmljYXRpb25fdG9rZW4iLCJ1c2VyX2lkIjoiZTFiZTI1NTYtZDIzNS00ODIxLThiMWEtN2UxYjc4MDdhODlkIiwiZW1haWwiOiJ0a2FobmcrMDFAZ21haWwuY29tIiwidG9rZW4iOiJlMDljNWE5OC1lNTZkLTQ2M2ItODBlOS05YzAxNmQ1YjFjMmEiLCJvdHAiOiIxNTc0MTkiLCJyZWRpcmVjdF90byI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODA4MCJ9.UVanKM0McUD8fcBdMp3zWujy3jLDIohqtNtBXijUdsw\u0026amp;type=verification_token\u0026amp;redirect_to=http://localhost:8080\"\u003eConfirm your email address\u003c/a\u003e\u003c/p\u003e\n\u003cp\u003eAlternatively, enter the code: 157419\u003c/p\u003e\n"
//   }
