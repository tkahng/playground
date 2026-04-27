package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/workers"
)

// TaskNotificationScheduler periodically enqueues due-today and overdue notification jobs
// for tasks that were not caught by the on-demand triggers (e.g. tasks whose due date
// arrived without any API update occurring).
type TaskNotificationScheduler struct {
	taskStore  stores.DbTaskStoreInterface
	jobService JobService
	interval   time.Duration
}

func NewTaskNotificationScheduler(taskStore stores.DbTaskStoreInterface, jobService JobService) *TaskNotificationScheduler {
	return &TaskNotificationScheduler{
		taskStore:  taskStore,
		jobService: jobService,
		interval:   time.Hour,
	}
}

func (s *TaskNotificationScheduler) Run(ctx context.Context) {
	s.schedule(ctx)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.schedule(ctx)
		}
	}
}

// RunOnce executes one scheduling pass without starting the ticker loop. Used in tests.
func (s *TaskNotificationScheduler) RunOnce(ctx context.Context) {
	s.schedule(ctx)
}

func (s *TaskNotificationScheduler) schedule(ctx context.Context) {
	s.enqueueDueToday(ctx)
	s.enqueueOverdue(ctx)
}

func (s *TaskNotificationScheduler) enqueueDueToday(ctx context.Context) {
	tasks, err := s.taskStore.FindTasksDueToday(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "task_notification_scheduler: find due today", slog.Any("error", err))
		return
	}
	for _, task := range tasks {
		if task.EndAt == nil {
			continue
		}
		if err := s.jobService.EnqueTaskDueJob(ctx, &workers.TaskDueTodayJobArgs{
			TaskID:  task.ID,
			DueDate: *task.EndAt,
		}); err != nil {
			slog.ErrorContext(ctx, "task_notification_scheduler: enqueue due today", slog.Any("error", err), slog.String("task_id", task.ID.String()))
		}
	}
}

func (s *TaskNotificationScheduler) enqueueOverdue(ctx context.Context) {
	tasks, err := s.taskStore.FindTasksOverdue(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "task_notification_scheduler: find overdue", slog.Any("error", err))
		return
	}
	for _, task := range tasks {
		if task.EndAt == nil {
			continue
		}
		if err := s.jobService.EnqueueTaskOverdueJob(ctx, &workers.TaskOverdueJobArgs{
			TaskID:  task.ID,
			DueDate: *task.EndAt,
		}); err != nil {
			slog.ErrorContext(ctx, "task_notification_scheduler: enqueue overdue", slog.Any("error", err), slog.String("task_id", task.ID.String()))
		}
	}
}
