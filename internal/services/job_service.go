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
