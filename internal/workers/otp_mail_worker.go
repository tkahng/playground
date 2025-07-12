package workers

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/jobs"
	"github.com/tkahng/authgo/internal/tools/mailer"
)

type OtpEmailJobArgs struct {
	UserID uuid.UUID
	Type   mailer.EmailType
}

func (j OtpEmailJobArgs) Kind() string {
	return "otp_email"
}

type OtpEmailJobWorker jobs.Worker[OtpEmailJobArgs]

type otpMailWorker struct {
	mail OtpMailServiceInterface
}
type OtpMailServiceInterface interface {
	SendTeamInvitationEmail(ctx context.Context, params *TeamInvitationJobArgs) error
	SendOtpEmail(ctx context.Context, emailType mailer.EmailType, userId uuid.UUID) error
}

func NewOtpEmailWorker(otpMailService OtpMailServiceInterface) jobs.Worker[OtpEmailJobArgs] {
	return &otpMailWorker{
		mail: otpMailService,
	}
}

// Work implements jobs.Worker.
func (w *otpMailWorker) Work(ctx context.Context, job *jobs.Job[OtpEmailJobArgs]) error {
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

type OtpMailWorkerDecorator struct {
	Delegate jobs.Worker[OtpEmailJobArgs]
	WorkFunc func(ctx context.Context, job *jobs.Job[OtpEmailJobArgs]) error
}

// Work implements jobs.Worker.
func (o *OtpMailWorkerDecorator) Work(ctx context.Context, job *jobs.Job[OtpEmailJobArgs]) error {
	if o.WorkFunc != nil {
		return o.WorkFunc(ctx, job)
	}
	return o.Delegate.Work(ctx, job)
}

var _ jobs.Worker[OtpEmailJobArgs] = (*OtpMailWorkerDecorator)(nil)
