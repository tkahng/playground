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

type DbJobManagerDecorator struct {
	Store           JobStore
	Poller          Poller
	Dispatcher      Dispatcher
	Delegate        *DbJobManager
	pollOnceFunc    func(ctx context.Context) error
	EnqueueFunc     func(ctx context.Context, args *EnqueueParams) error
	EnqueueManyFunc func(ctx context.Context, jobs ...*EnqueueParams) error
	RunFunc         func(ctx context.Context) error
	DispatchFunc    func(ctx context.Context, row *models.JobRow) error
	WithTxFunc      func(db database.Dbx) JobManager
}

// WithTx implements JobManager.
func (d *DbJobManagerDecorator) WithTx(db database.Dbx) JobManager {
	if d.WithTxFunc != nil {
		return d.WithTxFunc(db)
	}
	return d.Delegate.WithTx(db)
}

// PollOnce implements JobManager.
func (d *DbJobManagerDecorator) PollOnce(ctx context.Context) error {
	if d.Poller != nil {
		return d.Poller.PollOnce(ctx)
	}
	return d.Delegate.PollOnce(ctx)
}

// Dispatch implements JobManager.
func (d *DbJobManagerDecorator) Dispatch(ctx context.Context, row *models.JobRow) error {
	if d.Dispatcher != nil {
		return d.Dispatcher.Dispatch(ctx, row)
	}
	return d.Delegate.Dispatch(ctx, row)
}

// SetHandler implements JobManager.
func (d *DbJobManagerDecorator) SetHandler(kind string, handler func(context.Context, *models.JobRow) error) {
	if d.Dispatcher != nil {
		d.Dispatcher.SetHandler(kind, handler)
	}
	d.Delegate.SetHandler(kind, handler)
}

func NewDbJobManagerDecorator(dbx database.Dbx) *DbJobManagerDecorator {
	delegate := NewDbJobManager(dbx)
	return &DbJobManagerDecorator{Delegate: delegate}
}

// Enqueue implements JobManagerInterface.
func (d *DbJobManagerDecorator) Enqueue(ctx context.Context, args *EnqueueParams) error {
	if d.EnqueueFunc != nil {
		return d.EnqueueFunc(ctx, args)
	}
	return d.Delegate.Enqueue(ctx, args)
}

// EnqueueMany implements JobManagerInterface.
func (d *DbJobManagerDecorator) EnqueueMany(ctx context.Context, jobs ...*EnqueueParams) error {
	if d.EnqueueManyFunc != nil {
		return d.EnqueueManyFunc(ctx, jobs...)
	}
	return d.Delegate.EnqueueMany(ctx, jobs...)
}

// Run implements JobManagerInterface.
func (d *DbJobManagerDecorator) Run(ctx context.Context) error {
	if d.RunFunc != nil {
		return d.RunFunc(ctx)
	}
	return d.Delegate.Run(ctx)
}

var _ JobManager = (*DbJobManagerDecorator)(nil)
