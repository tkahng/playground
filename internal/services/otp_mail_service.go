package services

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mailer"
	"github.com/tkahng/playground/internal/tools/security"
	"github.com/tkahng/playground/internal/workers"
)

type OtpMailService interface {
	SendOtpEmail(ctx context.Context, emailType mailer.EmailType, userId uuid.UUID) error
	SendTeamInvitationEmail(ctx context.Context, params *workers.TeamInvitationJobArgs) error
}

var _ OtpMailService = (*DbOtpMailService)(nil)

type DbOtpMailService struct {
	options  *conf.EnvConfig
	adapter  stores.StorageAdapterInterface
	mail     mailer.Mailer
	token    JwtService
	password PasswordService
}

func NewOtpMailService(
	opts *conf.EnvConfig,
	adapter stores.StorageAdapterInterface,
) OtpMailService {
	var m mailer.Mailer
	if opts.ResendApiKey != "" {
		m = mailer.NewResendMailer(opts.ResendConfig)
	} else {
		m = &mailer.LogMailer{}
	}
	return &DbOtpMailService{
		options:  opts,
		adapter:  adapter,
		mail:     m,
		token:    NewJwtService(),
		password: NewPasswordService(),
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

	appOpts := app.options.AppConfig
	var tokenOpts conf.TokenOption
	switch emailType {
	case mailer.EmailTypeVerify:
		tokenOpts = app.options.AuthOptions.VerificationToken
	case mailer.EmailTypeSecurityPasswordReset:
		tokenOpts = app.options.AuthOptions.PasswordResetToken
	case mailer.EmailTypeConfirmPasswordReset:
		tokenOpts = app.options.AuthOptions.PasswordResetToken
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
	appOpts := app.options.AppConfig
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

func (i *DbOtpMailService) CreateConfirmationUrl(tokenhash string) (string, error) {
	path, err := mailer.GetPathParams(
		"/team-invitation",
		tokenhash,
		string(models.TokenTypesInviteToken),
		i.options.AppConfig.AppUrl,
	)
	if err != nil {
		return "", err
	}
	appUrl, err := url.Parse(i.options.AppConfig.AppUrl)
	if err != nil {
		return "", err
	}
	return appUrl.ResolveReference(path).String(), nil
}

// SendInvitationEmail implements TeamInvitationService.
func (i *DbOtpMailService) SendTeamInvitationEmail(ctx context.Context, params *workers.TeamInvitationJobArgs) error {
	if params == nil {
		return fmt.Errorf("params is nil")
	}
	if params.Email == "" {
		return fmt.Errorf("email is empty")
	}
	if params.TeamName == "" {
		return fmt.Errorf("team name is empty")
	}

	confUrl, err := i.CreateConfirmationUrl(params.TokenHash)
	if err != nil {
		return err
	}
	params.ConfirmationURL = confUrl
	body := mailer.GenerateBody("body", string(mailer.DefaultTeamInviteMail), params)
	param := &mailer.AllEmailParams{}
	// param.SendMailParams = &mailer.SendMailParams{
	// 	Template: string(mailer.DefaultTeamInviteMail),
	// }
	// param.CommonParams = &mailer.CommonParams{
	// 	ConfirmationURL: params.ConfirmationURL,
	// 	Email:           params.Email,
	// 	SiteURL:         i.options.Meta.AppUrl,
	// 	Token:           params.TokenHash,
	// }
	param.Message = &mailer.Message{
		From:    i.options.AppConfig.SenderAddress,
		To:      params.Email,
		Subject: fmt.Sprintf("Invitation to join %s", params.TeamName),
		Body:    body,
	}
	return i.mail.Send(param.Message)
}

type OtpMailDecorator struct {
	Delegate                DbOtpMailService
	SendOtpEmailFunc        func(ctx context.Context, emailType mailer.EmailType, userId uuid.UUID) error
	SendInvitationEmailFunc func(ctx context.Context, params *workers.TeamInvitationJobArgs) error
}

// SendTeamInvitationEmail implements OtpMailService.
func (o *OtpMailDecorator) SendTeamInvitationEmail(ctx context.Context, params *workers.TeamInvitationJobArgs) error {
	if o.SendInvitationEmailFunc != nil {
		return o.SendInvitationEmailFunc(ctx, params)
	}
	return o.Delegate.SendTeamInvitationEmail(ctx, params)
}

// SendOtpEmail implements OtpMailService.
func (o *OtpMailDecorator) SendOtpEmail(ctx context.Context, emailType mailer.EmailType, userId uuid.UUID) error {
	if o.SendOtpEmailFunc != nil {
		return o.SendOtpEmailFunc(ctx, emailType, userId)
	}
	return o.Delegate.SendOtpEmail(ctx, emailType, userId)
}

var _ OtpMailService = (*OtpMailDecorator)(nil)
