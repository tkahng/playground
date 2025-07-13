package workers

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/jobs"
)

type RefreshSubscriptionQuantityJobArgs struct {
	TeamID uuid.UUID `json:"team_id"`
}

func (j RefreshSubscriptionQuantityJobArgs) Kind() string {
	return "team_member_added_job"
}

type refreshSubscriptionQuantityWorker struct {
	service RefreshSubscriptionQuantityInterface
}
type RefreshSubscriptionQuantityInterface interface {
	VerifyAndUpdateTeamSubscriptionQuantity(ctx context.Context, teamId uuid.UUID) error
}

func NewRefreshSubscriptionQuantityWorker(teamMemberAddedService RefreshSubscriptionQuantityInterface) jobs.Worker[RefreshSubscriptionQuantityJobArgs] {
	return &refreshSubscriptionQuantityWorker{
		service: teamMemberAddedService,
	}
}

// Work implements jobs.Worker.
func (w *refreshSubscriptionQuantityWorker) Work(ctx context.Context, job *jobs.Job[RefreshSubscriptionQuantityJobArgs]) error {
	err := w.service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, job.Args.TeamID)
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

var _ jobs.Worker[RefreshSubscriptionQuantityJobArgs] = (*refreshSubscriptionQuantityWorker)(nil)

type TeamMemberAddedWorkerDecorator struct {
	Delegate jobs.Worker[RefreshSubscriptionQuantityJobArgs]
	WorkFunc func(ctx context.Context, job *jobs.Job[RefreshSubscriptionQuantityJobArgs]) error
}

// Work implements jobs.Worker.
func (o *TeamMemberAddedWorkerDecorator) Work(ctx context.Context, job *jobs.Job[RefreshSubscriptionQuantityJobArgs]) error {
	if o.WorkFunc != nil {
		return o.WorkFunc(ctx, job)
	}
	return o.Delegate.Work(ctx, job)
}

var _ jobs.Worker[RefreshSubscriptionQuantityJobArgs] = (*TeamMemberAddedWorkerDecorator)(nil)
