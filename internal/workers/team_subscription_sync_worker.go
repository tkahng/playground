package workers

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/jobs"
)

type TeamMemberAddedJobArgs struct {
	TeamMemberID uuid.UUID `json:"team_member_id"`
}

func (j TeamMemberAddedJobArgs) Kind() string {
	return "team_member_added_job"
}

type teamMemberAddedWorker struct {
	service TeamMemberAddedServiceInterface
}
type TeamMemberAddedServiceInterface interface {
	VerifyAndUpdateTeamSubscriptionQuantity(ctx context.Context, teamId uuid.UUID) error
}

func NewTeamMemberAddedWorker(teamMemberAddedService TeamMemberAddedServiceInterface) jobs.Worker[TeamMemberAddedJobArgs] {
	return &teamMemberAddedWorker{
		service: teamMemberAddedService,
	}
}

// Work implements jobs.Worker.
func (w *teamMemberAddedWorker) Work(ctx context.Context, job *jobs.Job[TeamMemberAddedJobArgs]) error {
	err := w.service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, job.Args.TeamMemberID)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"failed to verify and update team subscription quantity",
			slog.Any("error", err),
			slog.Any("args", job.Args),
		)
	}
	return err
}

var _ jobs.Worker[TeamMemberAddedJobArgs] = (*teamMemberAddedWorker)(nil)

type TeamMemberAddedWorkerDecorator struct {
	Delegate jobs.Worker[TeamMemberAddedJobArgs]
	WorkFunc func(ctx context.Context, job *jobs.Job[TeamMemberAddedJobArgs]) error
}

// Work implements jobs.Worker.
func (o *TeamMemberAddedWorkerDecorator) Work(ctx context.Context, job *jobs.Job[TeamMemberAddedJobArgs]) error {
	if o.WorkFunc != nil {
		return o.WorkFunc(ctx, job)
	}
	return o.Delegate.Work(ctx, job)
}

var _ jobs.Worker[TeamMemberAddedJobArgs] = (*TeamMemberAddedWorkerDecorator)(nil)
