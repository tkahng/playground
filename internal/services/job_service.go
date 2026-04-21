package services

import (
	"context"
	"errors"
	"time"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/tools/types"
	"github.com/tkahng/playground/internal/workers"
)

type JobService interface {
	WithTx(db database.Dbx) JobService
	// En
	EnqueueRpsGameInviteJob(ctx context.Context, job *workers.RpsGameInvitationJobArgs) error
	EnqueueTaskCompletedJob(ctx context.Context, job *workers.TaskCompletedJobArgs) error
	EnqueTaskDueJob(ctx context.Context, job *workers.TaskDueTodayJobArgs) error
	EnqueAssignedToTaskJob(ctx context.Context, job *workers.AssignedToTasJobArgs) error
	EnqueueTeamMemberAddedJob(ctx context.Context, job *workers.NewMemberNotificationJobArgs) error
	EnqueueRefreshSubscriptionQuantityJob(ctx context.Context, job *workers.RefreshSubscriptionQuantityJobArgs) error
	EnqueueOtpMailJob(ctx context.Context, args *workers.OtpEmailJobArgs) error
	EnqueueTeamInvitationJob(ctx context.Context, args *workers.TeamInvitationJobArgs) error
	EnqueueTaskOverdueJob(ctx context.Context, job *workers.TaskOverdueJobArgs) error
	EnqueueTaskStatusChangedJob(ctx context.Context, job *workers.TaskStatusChangedJobArgs) error
	EnqueueProjectStatusChangedJob(ctx context.Context, job *workers.ProjectStatusChangedJobArgs) error
	RegisterWorkers(mail OtpMailService, paymentService PaymentService, notification Notifier, rpsGame workers.RpsGameExpiryServiceInterface)
}

type DbJobService struct {
	manager jobs.JobManager
}

// EnqueueRpsGameInviteJob implements [JobService].
func (d *DbJobService) EnqueueRpsGameInviteJob(ctx context.Context, job *workers.RpsGameInvitationJobArgs) error {
	return d.manager.Enqueue(ctx, &jobs.EnqueueParams{
		Args:        job,
		RunAfter:    time.Now(),
		MaxAttempts: 3,
	})
}

// EnqueueTaskCompletedJob implements JobService.
func (d *DbJobService) EnqueueTaskCompletedJob(ctx context.Context, job *workers.TaskCompletedJobArgs) error {
	return d.manager.Enqueue(ctx, &jobs.EnqueueParams{
		Args:        job,
		RunAfter:    time.Now().Add(time.Second * 10),
		MaxAttempts: 3,
		UniqueKey:   types.Pointer(`task_completed:` + job.TaskID.String()),
	})
}

// EnqueTaskDueJob implements JobService.
func (d *DbJobService) EnqueTaskDueJob(ctx context.Context, job *workers.TaskDueTodayJobArgs) error {
	uniqueKey := "task_due_today:" + job.TaskID.String()
	return d.manager.Enqueue(ctx, &jobs.EnqueueParams{
		Args:        job,
		UniqueKey:   &uniqueKey,
		RunAfter:    job.DueDate.Add(time.Second * 10),
		MaxAttempts: 3,
	})
}

// EnqueAssignedToTaskJob implements JobService.
func (d *DbJobService) EnqueAssignedToTaskJob(ctx context.Context, job *workers.AssignedToTasJobArgs) error {
	return d.manager.Enqueue(ctx, &jobs.EnqueueParams{
		Args:        job,
		RunAfter:    time.Now(),
		MaxAttempts: 3,
	})
}

// EnqueueRefreshSubscriptionQuantityJob implements JobService.
func (d *DbJobService) EnqueueRefreshSubscriptionQuantityJob(ctx context.Context, job *workers.RefreshSubscriptionQuantityJobArgs) error {
	return d.manager.Enqueue(ctx, &jobs.EnqueueParams{
		Args:        job,
		RunAfter:    time.Now(),
		MaxAttempts: 3,
	})
}

// WithTx implements JobService.
func (d *DbJobService) WithTx(db database.Dbx) JobService {
	return &DbJobService{
		manager: d.manager.WithTx(db),
	}
}

func (d *DbJobService) EnqueueTeamMemberAddedJob(ctx context.Context, job *workers.NewMemberNotificationJobArgs) error {
	return d.manager.Enqueue(ctx, &jobs.EnqueueParams{
		Args:        job,
		RunAfter:    time.Now(),
		MaxAttempts: 3,
	})
}

// EnqueueTeamInvitationJob implements JobService.
func (d *DbJobService) EnqueueTeamInvitationJob(ctx context.Context, args *workers.TeamInvitationJobArgs) error {
	return d.manager.Enqueue(ctx, &jobs.EnqueueParams{
		Args:        args,
		RunAfter:    time.Now(),
		MaxAttempts: 3,
	})
}

// EnqueueTaskOverdueJob implements JobService.
func (d *DbJobService) EnqueueTaskOverdueJob(ctx context.Context, job *workers.TaskOverdueJobArgs) error {
	uniqueKey := "task_overdue:" + job.TaskID.String()
	return d.manager.Enqueue(ctx, &jobs.EnqueueParams{
		Args:        job,
		UniqueKey:   &uniqueKey,
		RunAfter:    job.DueDate.Add(24 * time.Hour),
		MaxAttempts: 3,
	})
}

// EnqueueTaskStatusChangedJob implements JobService.
func (d *DbJobService) EnqueueTaskStatusChangedJob(ctx context.Context, job *workers.TaskStatusChangedJobArgs) error {
	uniqueKey := "task_status_changed:" + job.TaskID.String()
	return d.manager.Enqueue(ctx, &jobs.EnqueueParams{
		Args:        job,
		UniqueKey:   &uniqueKey,
		RunAfter:    time.Now(),
		MaxAttempts: 3,
	})
}

// EnqueueProjectStatusChangedJob implements JobService.
func (d *DbJobService) EnqueueProjectStatusChangedJob(ctx context.Context, job *workers.ProjectStatusChangedJobArgs) error {
	uniqueKey := "project_status_changed:" + job.ProjectID.String()
	return d.manager.Enqueue(ctx, &jobs.EnqueueParams{
		Args:        job,
		UniqueKey:   &uniqueKey,
		RunAfter:    time.Now(),
		MaxAttempts: 3,
	})
}

// RegisterWorkers implements JobService.
func (d *DbJobService) RegisterWorkers(mail OtpMailService, paymentService PaymentService, notification Notifier, rpsGame workers.RpsGameExpiryServiceInterface) {
	jobs.RegisterWorker(d.manager, workers.NewOtpEmailWorker(mail))
	jobs.RegisterWorker(d.manager, workers.NewTeamInvitationWorker(mail))
	jobs.RegisterWorker(d.manager, workers.NewRpsGameInvitationWorker(mail))
	jobs.RegisterWorker(d.manager, workers.NewRefreshSubscriptionQuantityWorker(paymentService))
	jobs.RegisterWorker(d.manager, workers.NewNewMemberNotificationWorker(notification))
	jobs.RegisterWorker(d.manager, NewAssignedToTaskWorker(notification))
	jobs.RegisterWorker(d.manager, NewTaskDueTodayWorker(notification))
	jobs.RegisterWorker(d.manager, NewTaskCompletedWorker(notification))
	jobs.RegisterWorker(d.manager, NewTaskOverdueWorker(notification))
	jobs.RegisterWorker(d.manager, NewTaskStatusChangedWorker(notification))
	jobs.RegisterWorker(d.manager, NewProjectStatusChangedWorker(notification))
	jobs.RegisterWorker(d.manager, workers.NewRpsGameExpiryWorker(rpsGame, d.manager))
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
	Delegate                                  JobService
	EnqueueOtpMailJobFunc                     func(ctx context.Context, job *workers.OtpEmailJobArgs) error
	EnqueueTeamInvitationFunc                 func(ctx context.Context, job *workers.TeamInvitationJobArgs) error
	RegisterWorkersFunc                       func(mail OtpMailService, paymentService PaymentService, notification Notifier, rpsGame workers.RpsGameExpiryServiceInterface)
	EnqueueTeamMemberAddedJobFunc             func(ctx context.Context, job *workers.NewMemberNotificationJobArgs) error
	WithTxFunc                                func(db database.Dbx) JobService
	EnqueueRefreshSubscriptionQuantityJobFunc func(ctx context.Context, job *workers.RefreshSubscriptionQuantityJobArgs) error
	EnqueAssignedToTaskJobFunc                func(ctx context.Context, job *workers.AssignedToTasJobArgs) error
	EnqueTaskDueJobFunc                       func(ctx context.Context, job *workers.TaskDueTodayJobArgs) error
	EnqueueTaskCompletedJobFunc               func(ctx context.Context, job *workers.TaskCompletedJobArgs) error
	EnqueueRpsGameInviteJobFunc               func(ctx context.Context, job *workers.RpsGameInvitationJobArgs) error
	EnqueueTaskOverdueJobFunc                 func(ctx context.Context, job *workers.TaskOverdueJobArgs) error
	EnqueueTaskStatusChangedJobFunc           func(ctx context.Context, job *workers.TaskStatusChangedJobArgs) error
	EnqueueProjectStatusChangedJobFunc        func(ctx context.Context, job *workers.ProjectStatusChangedJobArgs) error
}

// EnqueueRpsGameInviteJob implements [JobService].
func (j *JobServiceDecorator) EnqueueRpsGameInviteJob(ctx context.Context, job *workers.RpsGameInvitationJobArgs) error {
	if j.EnqueueRpsGameInviteJobFunc != nil {
		return j.EnqueueRpsGameInviteJobFunc(ctx, job)
	}
	if j.Delegate == nil {
		return errors.New("delegate for EnqueueRpsGameInviteJob in JobService is nil")
	}
	return j.Delegate.EnqueueRpsGameInviteJob(ctx, job)
}

// EnqueueTaskCompletedJob implements JobService.
func (j *JobServiceDecorator) EnqueueTaskCompletedJob(ctx context.Context, job *workers.TaskCompletedJobArgs) error {
	if j.EnqueueTaskCompletedJobFunc != nil {
		return j.EnqueueTaskCompletedJobFunc(ctx, job)
	}
	if j.Delegate == nil {
		return errors.New("delegate for EnqueueTaskCompletedJob in JobService is nil")
	}
	return j.Delegate.EnqueueTaskCompletedJob(ctx, job)
}

// EnqueTaskDueJob implements JobService.
func (j *JobServiceDecorator) EnqueTaskDueJob(ctx context.Context, job *workers.TaskDueTodayJobArgs) error {
	if j.EnqueTaskDueJobFunc != nil {
		return j.EnqueTaskDueJobFunc(ctx, job)
	}
	if j.Delegate == nil {
		return errors.New("delegate for EnqueTaskDueJob in JobService is nil")
	}
	return j.Delegate.EnqueTaskDueJob(ctx, job)
}

// EnqueAssignedToTaskJob implements JobService.
func (j *JobServiceDecorator) EnqueAssignedToTaskJob(ctx context.Context, job *workers.AssignedToTasJobArgs) error {
	if j.EnqueAssignedToTaskJobFunc != nil {
		return j.EnqueAssignedToTaskJobFunc(ctx, job)
	}
	if j.Delegate == nil {
		return errors.New("delegate for EnqueAssignedToTaskJob in JobService is nil")
	}
	return j.Delegate.EnqueAssignedToTaskJob(ctx, job)
}

// EnqueueRefreshSubscriptionQuantityJob implements JobService.
func (j *JobServiceDecorator) EnqueueRefreshSubscriptionQuantityJob(ctx context.Context, job *workers.RefreshSubscriptionQuantityJobArgs) error {
	if j.EnqueueRefreshSubscriptionQuantityJobFunc != nil {
		return j.EnqueueRefreshSubscriptionQuantityJobFunc(ctx, job)
	}
	if j.Delegate == nil {
		return errors.New("delegate for EnqueueRefreshSubscriptionQuantityJob in JobService is nil")
	}
	return j.Delegate.EnqueueRefreshSubscriptionQuantityJob(ctx, job)
}

// WithTx implements JobService.
func (j *JobServiceDecorator) WithTx(db database.Dbx) JobService {
	if j.WithTxFunc != nil {
		return j.WithTxFunc(db)
	}
	return j.Delegate.WithTx(db)
}

// EnqueueTeamMemberAddedJob implements JobService.
func (j *JobServiceDecorator) EnqueueTeamMemberAddedJob(ctx context.Context, job *workers.NewMemberNotificationJobArgs) error {
	if j.EnqueueTeamMemberAddedJobFunc != nil {
		return j.EnqueueTeamMemberAddedJobFunc(ctx, job)
	}
	return j.Delegate.EnqueueTeamMemberAddedJob(ctx, job)
}

// EnqueueTeamInvitationJob implements JobService.
func (j *JobServiceDecorator) EnqueueTeamInvitationJob(ctx context.Context, args *workers.TeamInvitationJobArgs) error {
	if j.EnqueueTeamInvitationFunc != nil {
		return j.EnqueueTeamInvitationFunc(ctx, args)
	}
	return j.Delegate.EnqueueTeamInvitationJob(ctx, args)
}

// RegisterWorkers implements JobService.
func (j *JobServiceDecorator) RegisterWorkers(mail OtpMailService, paymentService PaymentService, notification Notifier, rpsGame workers.RpsGameExpiryServiceInterface) {
	if j.RegisterWorkersFunc != nil {
		j.RegisterWorkersFunc(mail, paymentService, notification, rpsGame)
	}
	j.Delegate.RegisterWorkers(mail, paymentService, notification, rpsGame)
}

// EnqueueOtpMailJob implements JobService.
func (j *JobServiceDecorator) EnqueueOtpMailJob(ctx context.Context, job *workers.OtpEmailJobArgs) error {
	if j.EnqueueOtpMailJobFunc != nil {
		return j.EnqueueOtpMailJobFunc(ctx, job)
	}
	return j.Delegate.EnqueueOtpMailJob(ctx, job)
}

// EnqueueTaskOverdueJob implements JobService.
func (j *JobServiceDecorator) EnqueueTaskOverdueJob(ctx context.Context, job *workers.TaskOverdueJobArgs) error {
	if j.EnqueueTaskOverdueJobFunc != nil {
		return j.EnqueueTaskOverdueJobFunc(ctx, job)
	}
	if j.Delegate == nil {
		return errors.New("delegate for EnqueueTaskOverdueJob in JobService is nil")
	}
	return j.Delegate.EnqueueTaskOverdueJob(ctx, job)
}

// EnqueueTaskStatusChangedJob implements JobService.
func (j *JobServiceDecorator) EnqueueTaskStatusChangedJob(ctx context.Context, job *workers.TaskStatusChangedJobArgs) error {
	if j.EnqueueTaskStatusChangedJobFunc != nil {
		return j.EnqueueTaskStatusChangedJobFunc(ctx, job)
	}
	if j.Delegate == nil {
		return errors.New("delegate for EnqueueTaskStatusChangedJob in JobService is nil")
	}
	return j.Delegate.EnqueueTaskStatusChangedJob(ctx, job)
}

// EnqueueProjectStatusChangedJob implements JobService.
func (j *JobServiceDecorator) EnqueueProjectStatusChangedJob(ctx context.Context, job *workers.ProjectStatusChangedJobArgs) error {
	if j.EnqueueProjectStatusChangedJobFunc != nil {
		return j.EnqueueProjectStatusChangedJobFunc(ctx, job)
	}
	if j.Delegate == nil {
		return errors.New("delegate for EnqueueProjectStatusChangedJob in JobService is nil")
	}
	return j.Delegate.EnqueueProjectStatusChangedJob(ctx, job)
}

func NewJobServiceDecorator(enqueuer jobs.JobManager) *JobServiceDecorator {
	var delegate JobService
	if enqueuer != nil {
		delegate = NewJobService(enqueuer)
	}
	return &JobServiceDecorator{
		Delegate: delegate,
	}
}

var _ JobService = (*JobServiceDecorator)(nil)
