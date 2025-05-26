package workers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/jobs"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/services"
)

type OtpEmailJobArgs struct {
	UserID uuid.UUID
	Type   services.EmailType
}

func (j OtpEmailJobArgs) Kind() string {
	return "otp_email"
}

type OtpEmailWorker struct {
	auth services.AuthService
}

// Work implements jobs.Worker.
func (w *OtpEmailWorker) Work(ctx context.Context, job *jobs.Job[OtpEmailJobArgs]) error {
	fmt.Println("sending security password reset email")
	user, err := w.auth.Store().FindUser(ctx, &models.User{ID: job.Args.UserID})
	if err != nil {
		slog.ErrorContext(
			ctx,
			"error getting user",
			slog.Any("error", err),
			slog.String("email", user.Email),
			slog.String("userId", user.ID.String()),
		)
		return err
	}
	err = w.auth.SendOtpEmail(job.Args.Type, ctx, user)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"error sending security password reset email",
			slog.Any("error", err),
			slog.String("email", user.Email),
			slog.String("userId", user.ID.String()),
		)
		return err
	}

	return nil
}

var _ jobs.Worker[OtpEmailJobArgs] = (*OtpEmailWorker)(nil)
