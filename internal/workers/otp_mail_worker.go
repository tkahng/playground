package workers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/jobs"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mailer"
	"github.com/tkahng/authgo/internal/tools/utils"
)

type OtpEmailJobArgs struct {
	UserID uuid.UUID
	Type   mailer.EmailType
}

func (j OtpEmailJobArgs) Kind() string {
	return "otp_email"
}

type UserFinder interface {
	FindUser(ctx context.Context, user *models.User) (*models.User, error)
}

type TokenSaver interface {
	SaveToken(ctx context.Context, token *shared.CreateTokenDTO) error
	SaveOtpToken(ctx context.Context, token *shared.CreateTokenDTO) error
}

type TokenCreator interface {
	CreateJwtToken(claims jwt.Claims, secret string) (string, error)
}

type MailService interface {
	SendMail(params *mailer.AllEmailParams) error
}

type OtpMail interface {
	SendOtpEmail(emailType mailer.EmailType, ctx context.Context, user *models.User) error
}

func RegisterMailWorker(
	dispatcher jobs.Dispatcher,
	userFinder UserFinder,
	otpMailer OtpMail,

) {
	worker := NewOtpEmailWorker(userFinder, otpMailer)
	jobs.RegisterWorker(dispatcher, worker)
}

type otpMailWorker struct {
	mail OtpMail
	user UserFinder
}

func NewOtpEmailWorker(user UserFinder, mail OtpMail) jobs.Worker[OtpEmailJobArgs] {
	return &otpMailWorker{
		mail: mail,
		user: user,
	}
}

// Work implements jobs.Worker.
func (w *otpMailWorker) Work(ctx context.Context, job *jobs.Job[OtpEmailJobArgs]) error {
	fmt.Println("otp mail")
	utils.PrettyPrintJSON(job)
	user, err := w.user.FindUser(ctx, &models.User{ID: job.Args.UserID})
	if err != nil {
		slog.ErrorContext(
			ctx,
			"error getting user",
			slog.Any("error", err),
			slog.String("email", user.Email),
			slog.String("emailType", job.Args.Type),
			slog.String("userId", user.ID.String()),
		)
		return err
	}
	err = w.mail.SendOtpEmail(job.Args.Type, ctx, user)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"error sending email",
			slog.Any("error", err),
			slog.String("email", user.Email),
			slog.String("emailType", job.Args.Type),
			slog.String("userId", user.ID.String()),
		)
		return err
	}

	return nil
}

var _ jobs.Worker[OtpEmailJobArgs] = (*otpMailWorker)(nil)
