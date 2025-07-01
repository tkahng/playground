package jobs

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
)

type DispatchDecorator struct {
	Delegate       Dispatcher
	SetHandlerFunc func(kind string, handler func(context.Context, *models.JobRow) error)
	DispatchFunc   func(ctx context.Context, row *models.JobRow) error
}

func (d *DispatchDecorator) Dispatch(ctx context.Context, row *models.JobRow) error {
	if d.DispatchFunc != nil {
		d.DispatchFunc(ctx, row)
	}
	return d.Delegate.Dispatch(ctx, row)
}

func (d *DispatchDecorator) SetHandler(kind string, handler func(context.Context, *models.JobRow) error) {
	if d.SetHandlerFunc != nil {
		d.SetHandlerFunc(kind, handler)
	}
	d.Delegate.SetHandler(kind, handler)
}

func NewDispatchDecorator() *DispatchDecorator {
	return &DispatchDecorator{Delegate: NewDispatcher()}
}

type JobStoreDecorator struct {
	Job                  *models.JobRow
	Delegate             JobStore
	RunInTxFunc          func(ctx context.Context, fn func(JobStore) error) error
	ClaimPendingJobsFunc func(ctx context.Context, limit int) ([]*models.JobRow, error)
	MarkDoneFunc         func(ctx context.Context, id uuid.UUID) error
	MarkFailedFunc       func(ctx context.Context, id uuid.UUID, reason string) error
	RescheduleJobFunc    func(ctx context.Context, id uuid.UUID, delay time.Duration) error
}

// RunInTx implements JobStore.
func (d *JobStoreDecorator) RunInTx(ctx context.Context, fn func(JobStore) error) error {
	if d.RunInTxFunc != nil {
		return d.RunInTxFunc(ctx, fn)
	}
	return d.Delegate.RunInTx(ctx, fn)
}

func NewJobStoreDecorator() *JobStoreDecorator {
	return &JobStoreDecorator{}
}

var _ JobStore = (*JobStoreDecorator)(nil)

func (d *JobStoreDecorator) ClaimPendingJobs(ctx context.Context, limit int) ([]*models.JobRow, error) {
	if d.ClaimPendingJobsFunc != nil {
		return d.ClaimPendingJobsFunc(ctx, limit)
	}
	return d.Delegate.ClaimPendingJobs(ctx, limit)
}

func (d *JobStoreDecorator) MarkDone(ctx context.Context, id uuid.UUID) error {
	if d.MarkDoneFunc != nil {
		return d.MarkDoneFunc(ctx, id)
	}
	return d.Delegate.MarkDone(ctx, id)
}

func (d *JobStoreDecorator) MarkFailed(ctx context.Context, id uuid.UUID, reason string) error {
	if d.MarkFailedFunc != nil {
		return d.MarkFailedFunc(ctx, id, reason)
	}
	return d.Delegate.MarkFailed(ctx, id, reason)
}

func (d *JobStoreDecorator) RescheduleJob(ctx context.Context, id uuid.UUID, delay time.Duration) error {
	if d.RescheduleJobFunc != nil {
		return d.RescheduleJobFunc(ctx, id, delay)
	}
	return d.Delegate.RescheduleJob(ctx, id, delay)
}

type DBEnqueuerDecorator struct {
	Jobs            []*models.JobRow
	Delegate        Enqueuer
	EnqueueFunc     func(ctx context.Context, args JobArgs, uniqueKey *string, runAfter time.Time, maxAttempts int) error
	EnqueueManyFunc func(ctx context.Context, jobs ...EnqueueParams) error
}

// Enqueue implements Enqueuer.
func (d *DBEnqueuerDecorator) Enqueue(ctx context.Context, args JobArgs, uniqueKey *string, runAfter time.Time, maxAttempts int) error {
	if d.EnqueueFunc != nil {
		return d.EnqueueFunc(ctx, args, uniqueKey, runAfter, maxAttempts)
	}
	return d.Delegate.Enqueue(ctx, args, uniqueKey, runAfter, maxAttempts)
}

// EnqueueMany implements Enqueuer.
func (d *DBEnqueuerDecorator) EnqueueMany(ctx context.Context, jobs ...EnqueueParams) error {

	if d.EnqueueManyFunc != nil {
		return d.EnqueueManyFunc(ctx, jobs...)
	}
	return d.Delegate.EnqueueMany(ctx, jobs...)
}

var _ Enqueuer = &DBEnqueuerDecorator{}

type PollerDecorator struct {
	Delegate     *Poller
	RunFunc      func(ctx context.Context) error
	PollOnceFunc func(ctx context.Context) error
}

func (d *PollerDecorator) Run(ctx context.Context) error {
	if d.RunFunc != nil {
		return d.RunFunc(ctx)
	}
	return d.Delegate.Run(ctx)
}

func (d *PollerDecorator) PollOnce(ctx context.Context) error {
	if d.PollOnceFunc != nil {
		return d.PollOnceFunc(ctx)
	}
	return d.Delegate.pollOnce(ctx)
}

func NewPollerDecorator(store JobStore, dispatcher Dispatcher, opts ...PollerOptsFunc) *PollerDecorator {
	return &PollerDecorator{
		Delegate: NewPoller(store, dispatcher, opts...),
	}
}
