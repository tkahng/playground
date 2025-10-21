package jobs

import (
	"context"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
)

type DbJobManager struct {
	store      JobStore
	poller     Poller
	dispatcher Dispatcher
}

// WithTx implements JobManager.
func (j *DbJobManager) WithTx(db database.Dbx) JobManager {
	return NewDbJobManager(db)
}

// PollOnce implements JobManager.
func (j *DbJobManager) PollOnce(ctx context.Context) error {
	return j.poller.PollOnce(ctx)
}

// Dispatch implements JobManager.
func (j *DbJobManager) Dispatch(ctx context.Context, row *models.JobRow) error {
	return j.dispatcher.Dispatch(ctx, row)
}

// SetHandler implements JobManager.
func (j *DbJobManager) SetHandler(kind string, handler func(context.Context, *models.JobRow) error) {
	j.dispatcher.SetHandler(kind, handler)
}

// Enqueue implements JobManagerInterface.
func (j *DbJobManager) Enqueue(ctx context.Context, args *EnqueueParams) error {
	return j.store.SaveJob(ctx, args)
}

// EnqueueMany implements JobManagerInterface.
func (j *DbJobManager) EnqueueMany(ctx context.Context, jobs ...*EnqueueParams) error {
	return j.store.SaveManyJobs(ctx, jobs...)
}

// Run implements JobManagerInterface.
func (j *DbJobManager) Run(ctx context.Context) error {
	return j.poller.Run(ctx)
}

type JobManager interface {
	Dispatcher
	Enqueuer
	Poller
	WithTx(db database.Dbx) JobManager
}

var _ JobManager = (*DbJobManager)(nil)

func NewDbJobManager(dbx database.Dbx) *DbJobManager {
	store := NewDbJobStore(dbx)
	dispatcher := NewDispatcher()
	poller := NewDbPoller(store, dispatcher)
	return &DbJobManager{
		store:      store,
		poller:     poller,
		dispatcher: dispatcher,
	}
}
