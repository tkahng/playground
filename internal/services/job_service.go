package services

import (
	"context"
	"time"

	"github.com/tkahng/authgo/internal/jobs"
	"github.com/tkahng/authgo/internal/workers"
)

type JobService interface {
	EnqueueOtpMailJob(ctx context.Context, args *workers.OtpEmailJobArgs) error
}

type DbJobService struct {
	manager jobs.JobManager
}

// EnqueueOtpMailJob implements JobService.
func (d *DbJobService) EnqueueOtpMailJob(ctx context.Context, job *workers.OtpEmailJobArgs) error {
	return d.manager.Enqueue(ctx, &jobs.EnqueueParams{
		Args:        job,
		RunAfter:    time.Now(),
		MaxAttempts: 3,
	})
}

func NewJobService(manager jobs.JobManager) JobService {
	return &DbJobService{
		manager: manager,
	}
}

type JobServiceDecorator struct {
	Delegate              JobService
	EnqueueOtpMailJobFunc func(ctx context.Context, job *workers.OtpEmailJobArgs) error
}

// EnqueueOtpMailJob implements JobService.
func (j *JobServiceDecorator) EnqueueOtpMailJob(ctx context.Context, job *workers.OtpEmailJobArgs) error {
	if j.EnqueueOtpMailJobFunc != nil {
		return j.EnqueueOtpMailJobFunc(ctx, job)
	}
	return j.Delegate.EnqueueOtpMailJob(ctx, job)
}

func NewJobServiceDecorator(enqueuer jobs.JobManager) *JobServiceDecorator {
	delegate := NewJobService(enqueuer)
	return &JobServiceDecorator{
		Delegate: delegate,
	}
}

var _ JobService = (*JobServiceDecorator)(nil)
