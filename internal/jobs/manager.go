package jobs

import (
	"context"
	"time"

	"github.com/tkahng/authgo/internal/database"
)

type DbJobManager struct {
	store    JobStore
	poller   Poller
	enqueuer Enqueuer
}

// Enqueue implements JobManagerInterface.
func (j *DbJobManager) Enqueue(ctx context.Context, args JobArgs, uniqueKey *string, runAfter time.Time, maxAttempts int) error {
	return j.enqueuer.Enqueue(ctx, args, uniqueKey, runAfter, maxAttempts)
}

// EnqueueMany implements JobManagerInterface.
func (j *DbJobManager) EnqueueMany(ctx context.Context, jobs ...EnqueueParams) error {
	return j.enqueuer.EnqueueMany(ctx, jobs...)
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
	enqueuer := NewDBEnqueuer(dbx)
	return &DbJobManager{
		store:    store,
		poller:   poller,
		enqueuer: enqueuer,
	}
}

type DbJobManagerDecorator struct {
	Enqueuer        Enqueuer
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
