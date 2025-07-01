package jobs

import (
	"context"
	"time"

	"github.com/tkahng/authgo/internal/database"
)

type DbJobManager struct {
	store      JobStore
	poller     Poller
	dispatcher Dispatcher
}

// Enqueue implements JobManagerInterface.
func (j *DbJobManager) Enqueue(ctx context.Context, args JobArgs, uniqueKey *string, runAfter time.Time, maxAttempts int) error {
	return j.store.SaveJob(ctx, &EnqueueParams{Args: args, UniqueKey: uniqueKey, RunAfter: runAfter, MaxAttempts: maxAttempts})
}

// EnqueueMany implements JobManagerInterface.
func (j *DbJobManager) EnqueueMany(ctx context.Context, jobs ...EnqueueParams) error {
	return j.store.SaveManyJobs(ctx, jobs...)
}

// Run implements JobManagerInterface.
func (j *DbJobManager) Run(ctx context.Context) error {
	return j.poller.Run(ctx)
}

type JobManager interface {
	Enqueuer
	Poller
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
	Delegate        *DbJobManager
	EnqueueFunc     func(ctx context.Context, args JobArgs, uniqueKey *string, runAfter time.Time, maxAttempts int) error
	EnqueueManyFunc func(ctx context.Context, jobs ...EnqueueParams) error
	RunFunc         func(ctx context.Context) error
}

func NewDbJobManagerDecorator(dbx database.Dbx) *DbJobManagerDecorator {
	delegate := NewDbJobManager(dbx)
	return &DbJobManagerDecorator{Delegate: delegate}
}

// Enqueue implements JobManagerInterface.
func (d *DbJobManagerDecorator) Enqueue(ctx context.Context, args JobArgs, uniqueKey *string, runAfter time.Time, maxAttempts int) error {
	if d.EnqueueFunc != nil {
		return d.EnqueueFunc(ctx, args, uniqueKey, runAfter, maxAttempts)
	}
	return d.Delegate.Enqueue(ctx, args, uniqueKey, runAfter, maxAttempts)
}

// EnqueueMany implements JobManagerInterface.
func (d *DbJobManagerDecorator) EnqueueMany(ctx context.Context, jobs ...EnqueueParams) error {
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
