package services

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/mailer"
	"github.com/tkahng/authgo/internal/tools/security"
)

type TokenSaver interface {
	SaveToken(ctx context.Context, token *stores.CreateTokenDTO) error
	SaveOtpToken(ctx context.Context, token *stores.CreateTokenDTO) error
}

type TokenCreator interface {
	CreateJwtToken(claims jwt.Claims, secret string) (string, error)
}

type OtpMailer struct {
	mail      MailService
	token     TokenCreator
	authStore TokenSaver
	options   conf.AppOptions
}

// SendOtpEmail creates and saves a new otp token and sends it to the user's email
func (app *OtpMailer) SendOtpEmail(emailType mailer.EmailType, ctx context.Context, user *models.User) error {
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

	claims := shared.OtpClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: tokenOpts.ExpiresAt(),
		},
		OtpPayload: shared.OtpPayload{
			Type:       tokenOpts.Type,
			UserId:     user.ID,
			Email:      user.Email,
			Token:      security.GenerateTokenKey(),
			Otp:        security.GenerateOtp(6),
			RedirectTo: appOpts.AppUrl,
		},
	}
	tokenHash, err := app.token.CreateJwtToken(claims, tokenOpts.Secret)
	if err != nil {
		return fmt.Errorf("error at creating verification token: %w", err)
	}

	dto := &stores.CreateTokenDTO{
		Expires:    claims.ExpiresAt.Time,
		Token:      claims.Token,
		Type:       models.TokenTypes(claims.Type),
		Identifier: claims.Email,
		UserID:     &claims.UserId,
	}

	err = app.authStore.SaveToken(ctx, dto)

	if err != nil {
		return fmt.Errorf("error at creating verification token: %w", err)
	}

	sendMailParams, err := app.GetSendMailParams(emailType, tokenHash, claims)
	if err != nil {
		return fmt.Errorf("error at getting send mail params: %w", err)
	}

	return app.mail.SendMail(sendMailParams)
}

func (app *OtpMailer) GetSendMailParams(emailType mailer.EmailType, tokenHash string, claims shared.OtpClaims) (*mailer.AllEmailParams, error) {
	appOpts := app.options.Meta
	var sendMailParams mailer.SendMailParams
	var ok bool
	if sendMailParams, ok = mailer.EmailPathMap[emailType]; !ok {
		return nil, fmt.Errorf("email type not found")
	}
	appUrl, err := url.Parse(appOpts.AppUrl)
	if err != nil {
		return nil, err
	}

	confirmUrl, err := sendMailParams.GeneratePath(appUrl, tokenHash, string(claims.Type), claims.RedirectTo)
	if err != nil {
		return nil, err
	}
	common := &mailer.CommonParams{
		SiteURL:         appUrl.String(),
		ConfirmationURL: confirmUrl,
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
