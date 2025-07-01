package workers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/jobs"
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

func RegisterMailWorker(
	dispatcher jobs.Dispatcher,
	mailService OtpMailServiceInterface,

) {
	worker := NewOtpEmailWorker(mailService)
	jobs.RegisterWorker(dispatcher, worker)
}

type otpMailWorker struct {
	mail OtpMailServiceInterface
}
type OtpMailServiceInterface interface {
	SendOtpEmail(ctx context.Context, emailType mailer.EmailType, userId uuid.UUID) error
}

func NewOtpEmailWorker(otpMailService OtpMailServiceInterface) jobs.Worker[OtpEmailJobArgs] {
	return &otpMailWorker{
		mail: otpMailService,
	}
}

// Work implements jobs.Worker.
func (w *otpMailWorker) Work(ctx context.Context, job *jobs.Job[OtpEmailJobArgs]) error {
	fmt.Println("otp mail")
	utils.PrettyPrintJSON(job)
	err := w.mail.SendOtpEmail(ctx, job.Args.Type, job.Args.UserID)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"failed to send otp email",
			slog.Any("error", err),
			slog.Any("args", job.Args),
		)
	}
	return err
}

var _ jobs.Worker[OtpEmailJobArgs] = (*otpMailWorker)(nil)
