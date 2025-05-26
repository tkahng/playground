package services

import (
	"context"

	"github.com/tkahng/authgo/internal/jobs"
	"github.com/tkahng/authgo/internal/workers"
)

type WorkerDecorator[T jobs.JobArgs] struct {
	Delegate jobs.Worker[T]
	WorkFunc func(ctx context.Context, job *jobs.Job[T]) error
}

// Work implements jobs.Worker.
func (w *WorkerDecorator[T]) Work(ctx context.Context, job *jobs.Job[T]) error {

	return w.WorkFunc(ctx, job)
}

var _ jobs.Worker[workers.OtpEmailJobArgs] = (*WorkerDecorator[workers.OtpEmailJobArgs])(nil)

// Work implements jobs.Worker.
func RegisterMailWorker(
	dispatcher jobs.Dispatcher,
	authService AuthService,
) {
	worker := workers.NewOtpEmailWorker(authService.Store(), authService)
	jobs.RegisterWorker(dispatcher, worker)
}
