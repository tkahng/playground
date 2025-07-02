package services

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/mailer"
	"github.com/tkahng/authgo/internal/tools/security"
)

type OtpMailService interface {
	SendOtpEmail(ctx context.Context, emailType mailer.EmailType, userId uuid.UUID) error
}

var _ OtpMailService = &DbOtpMailService{}

type DbOtpMailService struct {
	options  *conf.AppOptions
	adapter  stores.StorageAdapterInterface
	mail     mailer.Mailer
	token    JwtService
	password PasswordService
}

func NewOtpMailService(
	opts *conf.AppOptions,
	mail mailer.Mailer,
	adapter stores.StorageAdapterInterface,
) OtpMailService {
	return &DbOtpMailService{
		options:  opts,
		adapter:  adapter,
		mail:     mail,
		token:    NewJwtService(),
		password: NewPasswordService(),
	}
}

func NewDbOtpMailService(opts *conf.AppOptions, mail mailer.Mailer, token JwtService, password PasswordService, adapter stores.StorageAdapterInterface) DbOtpMailService {
	return DbOtpMailService{
		options:  opts,
		adapter:  adapter,
		mail:     mail,
		token:    token,
		password: password,
	}
}

func (app *DbOtpMailService) SendOtpEmail(ctx context.Context, emailType mailer.EmailType, userId uuid.UUID) error {
	adapter := app.adapter
	user, err := adapter.User().FindUserByID(ctx, userId)
	if err != nil {
		return err
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
	dto := &stores.CreateTokenDTO{
		Expires:    claims.ExpiresAt.Time,
		Token:      claims.Token,
		Type:       models.TokenTypes(claims.Type),
		Identifier: claims.Email,
		UserID:     &claims.UserId,
	}
	err = adapter.Token().SaveToken(ctx, dto)
	// err = app.authStore.SaveToken(ctx, dto)
	if err != nil {
		return fmt.Errorf("error at creating verification token: %w", err)
	}

	sendMailParams, err := app.getSendMailParams(emailType, tokenHash, claims)
	if err != nil {
		return fmt.Errorf("error at getting send mail params: %w", err)
	}

	return app.mail.Send(sendMailParams.Message)
}

func (app *DbOtpMailService) getSendMailParams(emailType mailer.EmailType, tokenHash string, claims shared.OtpClaims) (*mailer.AllEmailParams, error) {
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

type OtpMailDecorator struct {
	Delegate         DbOtpMailService
	SendOtpEmailFunc func(ctx context.Context, emailType mailer.EmailType, userId uuid.UUID) error
}

// SendOtpEmail implements OtpMailService.
func (o *OtpMailDecorator) SendOtpEmail(ctx context.Context, emailType mailer.EmailType, userId uuid.UUID) error {
	if o.SendOtpEmailFunc != nil {
		return o.SendOtpEmailFunc(ctx, emailType, userId)
	}
	return o.Delegate.SendOtpEmail(ctx, emailType, userId)
}

var _ OtpMailService = (*OtpMailDecorator)(nil)
