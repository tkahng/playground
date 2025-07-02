package services

import (
	"context"

	"github.com/tkahng/authgo/internal/jobs"
	"github.com/tkahng/authgo/internal/workers"
)

type JobService interface {
	EnqueueOtpMailJob(ctx context.Context, job jobs.Job[workers.OtpEmailJobArgs]) error
}
