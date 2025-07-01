package jobs

import (
	"context"
	"time"

	"github.com/tkahng/authgo/internal/database"
)

type JobManager struct {
	store    JobStore
	poller   Poller
	enqueuer Enqueuer
}

// Enqueue implements JobManagerInterface.
func (j *JobManager) Enqueue(ctx context.Context, args JobArgs, uniqueKey *string, runAfter time.Time, maxAttempts int) error {
	return j.enqueuer.Enqueue(ctx, args, uniqueKey, runAfter, maxAttempts)
}

// EnqueueMany implements JobManagerInterface.
func (j *JobManager) EnqueueMany(ctx context.Context, jobs ...EnqueueParams) error {
	return j.enqueuer.EnqueueMany(ctx, jobs...)
}

// Run implements JobManagerInterface.
func (j *JobManager) Run(ctx context.Context) error {
	return j.poller.Run(ctx)
}

type JobManagerInterface interface {
	Enqueuer
	Poller
}

var _ JobManagerInterface = (*JobManager)(nil)

func NewJobManager(dbx database.Dbx) *JobManager {
	store := NewDbJobStore(dbx)
	dispatcher := NewDispatcher()
	poller := NewDbPoller(store, dispatcher)
	enqueuer := NewDBEnqueuer(dbx)
	return &JobManager{
		store:    store,
		poller:   poller,
		enqueuer: enqueuer,
	}
}

type DbJobManagerDecorator struct {
	Enqueuer        Enqueuer
	Store           JobStore
	Poller          Poller
	Delegate        *JobManager
	EnqueueFunc     func(ctx context.Context, args JobArgs, uniqueKey *string, runAfter time.Time, maxAttempts int) error
	EnqueueManyFunc func(ctx context.Context, jobs ...EnqueueParams) error
	RunFunc         func(ctx context.Context) error
}

func NewDbJobManagerDecorator(dbx database.Dbx) *DbJobManagerDecorator {
	delegate := NewJobManager(dbx)
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

var _ JobManagerInterface = (*DbJobManagerDecorator)(nil)
