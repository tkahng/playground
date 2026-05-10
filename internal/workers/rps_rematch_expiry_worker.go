package workers

import (
	"context"
	"log/slog"
	"time"

	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/tools/types"
)

const RpsRematchExpirySweepKind = "rps_rematch_expiry_sweep"
const RpsRematchExpirySweepInterval = 30 * time.Second

type RpsRematchExpiryJobArgs struct{}

func (j RpsRematchExpiryJobArgs) Kind() string {
	return RpsRematchExpirySweepKind
}

type RpsRematchExpiryWorker struct {
	svc      RpsGameExpiryServiceInterface
	enqueuer jobs.Enqueuer
}

func NewRpsRematchExpiryWorker(svc RpsGameExpiryServiceInterface, enqueuer jobs.Enqueuer) jobs.Worker[RpsRematchExpiryJobArgs] {
	return &RpsRematchExpiryWorker{svc: svc, enqueuer: enqueuer}
}

func SeedRpsRematchExpiryJob(ctx context.Context, enqueuer jobs.Enqueuer) error {
	return enqueuer.Enqueue(ctx, &jobs.EnqueueParams{
		Args:        &RpsRematchExpiryJobArgs{},
		RunAfter:    time.Now(),
		MaxAttempts: 1,
		UniqueKey:   types.Pointer(RpsRematchExpirySweepKind),
	})
}

func (w *RpsRematchExpiryWorker) Work(ctx context.Context, job *jobs.Job[RpsRematchExpiryJobArgs]) error {
	count, err := w.svc.ExpireRematches(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		slog.InfoContext(ctx, "rps rematch expiry sweep: expired pending rematches", slog.Int("count", count))
	}
	if err := w.enqueuer.Enqueue(context.Background(), &jobs.EnqueueParams{
		Args:        &RpsRematchExpiryJobArgs{},
		RunAfter:    time.Now().Add(RpsRematchExpirySweepInterval),
		MaxAttempts: 1,
		UniqueKey:   types.Pointer(RpsRematchExpirySweepKind),
	}); err != nil {
		slog.ErrorContext(ctx, "rps rematch expiry sweep: failed to schedule next run", slog.Any("error", err))
	}
	return nil
}
