package services

import (
	"context"
	"time"

	"github.com/tkahng/authgo/internal/jobs"
	"github.com/tkahng/authgo/internal/workers"
)

type JobService interface {
	EnqueueOtpMailJob(ctx context.Context, args *workers.OtpEmailJobArgs) error
	EnqueueTeamInvitationJob(ctx context.Context, args *workers.TeamInvitationJobArgs) error
	RegisterWorkers(mail OtpMailService)
}

type DbJobService struct {
	manager jobs.JobManager
}

// EnqueueTeamInvitationJob implements JobService.
func (d *DbJobService) EnqueueTeamInvitationJob(ctx context.Context, args *workers.TeamInvitationJobArgs) error {
	return d.manager.Enqueue(ctx, &jobs.EnqueueParams{
		Args:        args,
		RunAfter:    time.Now(),
		MaxAttempts: 3,
	})
}

// RegisterWorkers implements JobService.
func (d *DbJobService) RegisterWorkers(mail OtpMailService) {
	jobs.RegisterWorker(d.manager, workers.NewOtpEmailWorker(mail))
	jobs.RegisterWorker(d.manager, workers.NewTeamInvitationWorker(mail))
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
	Delegate                  JobService
	EnqueueOtpMailJobFunc     func(ctx context.Context, job *workers.OtpEmailJobArgs) error
	EnqueueTeamInvitationFunc func(ctx context.Context, job *workers.TeamInvitationJobArgs) error
	RegisterWorkersFunc       func(mail OtpMailService)
}

// EnqueueTeamInvitationJob implements JobService.
func (j *JobServiceDecorator) EnqueueTeamInvitationJob(ctx context.Context, args *workers.TeamInvitationJobArgs) error {
	if j.EnqueueTeamInvitationFunc != nil {
		return j.EnqueueTeamInvitationFunc(ctx, args)
	}
	return j.Delegate.EnqueueTeamInvitationJob(ctx, args)
}

// RegisterWorkers implements JobService.
func (j *JobServiceDecorator) RegisterWorkers(mail OtpMailService) {
	if j.RegisterWorkersFunc != nil {
		j.RegisterWorkersFunc(mail)
	}
	j.Delegate.RegisterWorkers(mail)
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
