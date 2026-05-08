package services

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/apierrors"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/token"
	"github.com/tkahng/playground/internal/tools/mailer"
	"github.com/tkahng/playground/internal/tools/security"
	"github.com/tkahng/playground/internal/workers"
)

type OtpMailService interface {
	SendRpsGameInvitationEmail(ctx context.Context, params *workers.RpsGameInvitationJobArgs) error
	SendOtpEmail(ctx context.Context, emailType mailer.EmailType, userId uuid.UUID) error
	SendTeamInvitationEmail(ctx context.Context, params *workers.TeamInvitationJobArgs) error
}

var _ OtpMailService = (*DbOtpMailService)(nil)

type DbOtpMailService struct {
	options *conf.EnvConfig
	adapter stores.StorageAdapterInterface
	mail    mailer.Mailer
	token   token.TokenService
	jwt     JwtService
	hash    HashService
}

func NewOtpMailService(
	opts *conf.EnvConfig,
	adapter stores.StorageAdapterInterface,
	mailer mailer.Mailer,
	token token.TokenService,
	jwt JwtService,
	hash HashService,
) OtpMailService {
	return &DbOtpMailService{
		options: opts,
		adapter: adapter,
		mail:    mailer,
		jwt:     jwt,
		hash:    hash,
		token:   token,
	}
}

type OtpClaims struct {
	jwt.RegisteredClaims
	OtpPayload
}

type OtpPayload struct {
	UserId     uuid.UUID         `json:"user_id,omitempty"`
	Email      string            `json:"email,omitempty"`
	Token      string            `json:"token"`
	Type       models.TokenTypes `json:"type"`
	Otp        string            `json:"otp,omitempty"`
	RedirectTo string            `json:"redirect_to,omitempty"`
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
	if app.jwt == nil {
		return fmt.Errorf("token service is nil")
	}
	if app.mail == nil {
		return fmt.Errorf("mail service is nil")
	}
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	var tokenOpts conf.TokenOption
	switch emailType {
	case mailer.EmailTypeVerify:
		tokenOpts = app.options.VerificationToken
	case mailer.EmailTypeSecurityPasswordReset:
		tokenOpts = app.options.PasswordResetToken
	case mailer.EmailTypeConfirmPasswordReset:
		tokenOpts = app.options.PasswordResetToken
	default:
		return apierrors.BadRequest("invalid email type")
	}
	otp := security.GenerateOtp(6)
	tokenHash, err := app.token.GenerateToken(ctx, &token.GenerateTokenOptions{
		Email: user.Email,
		Type:  tokenOpts.Type,
		Otp:   otp,
	})
	if err != nil {
		return fmt.Errorf("error at creating verification token: %w", err)
	}

	claims := OtpClaims{}
	claims.ExpiresAt = tokenOpts.ExpiresAt()
	claims.Type = tokenOpts.Type
	claims.UserId = user.ID
	claims.Email = user.Email
	claims.Token = tokenHash
	claims.Otp = otp

	sendMailParams, err := app.getSendMailParams(emailType, tokenHash, claims)
	if err != nil {
		return fmt.Errorf("error at getting send mail params: %w", err)
	}

	return app.mail.Send(sendMailParams)
}

func (app *DbOtpMailService) getSendMailParams(emailType mailer.EmailType, tokenHash string, claims OtpClaims) (*mailer.Message, error) {
	appOpts := app.options.AppConfig
	var sendMailParams mailer.SendMailParams
	var ok bool
	if sendMailParams, ok = mailer.EmailPathMap[emailType]; !ok {
		return nil, apierrors.NotFound("email type not found")
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
		To:      claims.Email,
		Subject: fmt.Sprintf(sendMailParams.Subject, appOpts.AppName),
		Body:    mailer.GenerateBody("body", sendMailParams.Template, common),
	}

	return message, nil
}

func (i *DbOtpMailService) CreateTeamConfirmationUrl(tokenhash string) (string, error) {
	path, err := mailer.GetPathParams(
		"/team-invitation",
		tokenhash,
		"",
		"",
	)
	if err != nil {
		return "", err
	}
	appUrl, err := url.Parse(i.options.AppUrl)
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

	confUrl, err := i.CreateTeamConfirmationUrl(params.TokenHash)
	if err != nil {
		return err
	}
	params.ConfirmationURL = confUrl
	body := mailer.GenerateBody("body", string(mailer.DefaultTeamInviteMail), params)
	message := &mailer.Message{
		From:    i.options.SenderAddress,
		To:      params.Email,
		Subject: fmt.Sprintf("Invitation to join %s", params.TeamName),
		Body:    body,
	}
	return i.mail.Send(message)
}
func (i *DbOtpMailService) CreateRpsGameUrl(tokenhash string) (string, error) {
	path, err := mailer.GetPathParams(
		"/rock-paper-scissors",
		tokenhash,
		"",
		"",
	)
	if err != nil {
		return "", err
	}
	appUrl, err := url.Parse(i.options.AppUrl)
	if err != nil {
		return "", err
	}
	return appUrl.ResolveReference(path).String(), nil
}

// SendRpsGameInvitationEmail implements [OtpMailService].
func (i *DbOtpMailService) SendRpsGameInvitationEmail(ctx context.Context, params *workers.RpsGameInvitationJobArgs) error {
	if params == nil {
		return fmt.Errorf("params is nil")
	}
	if params.Email == "" {
		return fmt.Errorf("email is empty")
	}

	confUrl, err := i.CreateRpsGameUrl(params.TokenHash)
	if err != nil {
		return err
	}
	params.ConfirmationURL = confUrl
	body := mailer.GenerateBody("body", string(mailer.DefaultRpsGameInviteMail), params)
	message := &mailer.Message{
		From:    i.options.SenderAddress,
		To:      params.Email,
		Subject: "Invitation to play a game of Rock Paper Scissors",
		Body:    body,
	}
	return i.mail.Send(message)
}
