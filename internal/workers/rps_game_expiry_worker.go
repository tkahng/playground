package workers

import (
	"context"
	"log/slog"
	"time"

	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/tools/types"
)

const RpsGameExpirySweepKind = "rps_game_expiry_sweep"
const RpsGameExpirySweepInterval = time.Minute

type RpsGameExpiryJobArgs struct{}

func (j RpsGameExpiryJobArgs) Kind() string {
	return RpsGameExpirySweepKind
}

// RpsGameExpiryServiceInterface is the subset of RpsGameService needed by the expiry workers.
type RpsGameExpiryServiceInterface interface {
	ExpireGamesAndRefundBets(ctx context.Context) (int, error)
	ExpireRematches(ctx context.Context) (int, error)
}

type RpsGameExpiryWorker struct {
	rpsGame  RpsGameExpiryServiceInterface
	enqueuer jobs.Enqueuer
}

func NewRpsGameExpiryWorker(rpsGame RpsGameExpiryServiceInterface, enqueuer jobs.Enqueuer) jobs.Worker[RpsGameExpiryJobArgs] {
	return &RpsGameExpiryWorker{rpsGame: rpsGame, enqueuer: enqueuer}
}

// SeedRpsGameExpiryJob enqueues the first sweep job. Idempotent — the unique_key
// prevents a duplicate if a pending or processing job already exists.
func SeedRpsGameExpiryJob(ctx context.Context, enqueuer jobs.Enqueuer) error {
	return enqueuer.Enqueue(ctx, &jobs.EnqueueParams{
		Args:        &RpsGameExpiryJobArgs{},
		RunAfter:    time.Now(),
		MaxAttempts: 1,
		UniqueKey:   types.Pointer(RpsGameExpirySweepKind),
	})
}

// Work voids pending escrow for any expired bet games, then schedules the next run.
func (w *RpsGameExpiryWorker) Work(ctx context.Context, job *jobs.Job[RpsGameExpiryJobArgs]) error {
	processed, err := w.rpsGame.ExpireGamesAndRefundBets(ctx)
	if err != nil {
		return err
	}
	if processed > 0 {
		slog.InfoContext(ctx, "rps expiry sweep: refunded expired bet games", slog.Int("count", processed))
	}

	// Self-schedule the next run. Use context.Background so a job-level timeout
	// on the current execution does not cancel the re-enqueue write.
	if err := w.enqueuer.Enqueue(context.Background(), &jobs.EnqueueParams{
		Args:        &RpsGameExpiryJobArgs{},
		RunAfter:    time.Now().Add(RpsGameExpirySweepInterval),
		MaxAttempts: 1,
		UniqueKey:   types.Pointer(RpsGameExpirySweepKind),
	}); err != nil {
		slog.ErrorContext(ctx, "rps expiry sweep: failed to schedule next run", slog.Any("error", err))
	}

	return nil
}
