package workers

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/jobs"
)

type NewMemberNotificationJobArgs struct {
	TeamMemberID uuid.UUID
	TeamID       uuid.UUID
}

func (j NewMemberNotificationJobArgs) Kind() string {
	return "new_member_notification_job"
}

type NewMemberNotificationService interface {
	NotifyMembersOfNewMember(ctx context.Context, teamMemberID uuid.UUID, teamID uuid.UUID) error
}

type NewMemberNotificationWorker struct {
	service NewMemberNotificationService
}

func (w *NewMemberNotificationWorker) Kind() string {
	return "new_member_notification_worker"
}

func (w *NewMemberNotificationWorker) Work(ctx context.Context, args *jobs.Job[NewMemberNotificationJobArgs]) error {
	err := w.service.NotifyMembersOfNewMember(ctx, args.Args.TeamMemberID, args.Args.TeamID)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"failed to notify members of new member",
			slog.Any("error", err),
			slog.Any("args", args.Args),
		)
	}
	return nil
}

func NewNewMemberNotificationWorker(service NewMemberNotificationService) *NewMemberNotificationWorker {
	return &NewMemberNotificationWorker{
		service: service,
	}
}
