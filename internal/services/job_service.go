package services

import (
	"context"

	"github.com/tkahng/authgo/internal/jobs"
)

type JobService interface {
	EnqueueOtpMailJob(ctx context.Context, job jobs.EnqueueParams) error
}

type DbJobService struct {
	enqueuer jobs.Enqueuer
}

// EnqueueOtpMailJob implements JobService.
func (d *DbJobService) EnqueueOtpMailJob(ctx context.Context, job jobs.EnqueueParams) error {
	return d.enqueuer.Enqueue(ctx, &job)
}

func NewJobService(enqueuer jobs.Enqueuer) JobService {
	return &DbJobService{
		enqueuer: enqueuer,
	}
}

type JobServiceDecorator struct {
	Delegate              JobService
	EnqueueOtpMailJobFunc func(ctx context.Context, job jobs.EnqueueParams) error
}

// EnqueueOtpMailJob implements JobService.
func (j *JobServiceDecorator) EnqueueOtpMailJob(ctx context.Context, job jobs.EnqueueParams) error {
	if j.EnqueueOtpMailJobFunc != nil {
		return j.EnqueueOtpMailJobFunc(ctx, job)
	}
	return j.Delegate.EnqueueOtpMailJob(ctx, job)
}

func NewJobServiceDecorator(enqueuer jobs.Enqueuer) *JobServiceDecorator {
	delegate := NewJobService(enqueuer)
	return &JobServiceDecorator{
		Delegate: delegate,
	}
}

var _ JobService = (*JobServiceDecorator)(nil)
