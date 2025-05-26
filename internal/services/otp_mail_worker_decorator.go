package services

import (
	"context"

	"github.com/tkahng/authgo/internal/jobs"
	"github.com/tkahng/authgo/internal/workers"
)

type WorkerDecorator[T jobs.JobArgs] struct {
	Job      *jobs.Job[T]
	Delegate jobs.Worker[T]
	WorkFunc func(ctx context.Context, job *jobs.Job[T]) error
}

func NewWorkerDecorator(authService AuthService) *WorkerDecorator[workers.OtpEmailJobArgs] {
	worker := workers.NewOtpEmailWorker(authService.Store(), authService)
	return &WorkerDecorator[workers.OtpEmailJobArgs]{
		Delegate: worker,
	}
}

// Work implements jobs.Worker.
func (w *WorkerDecorator[T]) Work(ctx context.Context, job *jobs.Job[T]) error {
	w.Job = job
	if w.WorkFunc == nil {
		return w.Delegate.Work(ctx, job)
	}
	return w.WorkFunc(ctx, job)
}

var _ jobs.Worker[workers.OtpEmailJobArgs] = (*WorkerDecorator[workers.OtpEmailJobArgs])(nil)
