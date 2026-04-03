package workers

import (
	"context"
	"log/slog"

	"github.com/tkahng/playground/internal/jobs"
)

type RpsGameExpiryJobArgs struct{}

func (j RpsGameExpiryJobArgs) Kind() string {
	return "rps_game_expiry_sweep"
}

// RpsGameExpiryServiceInterface is the subset of RpsGameService needed by the expiry worker.
type RpsGameExpiryServiceInterface interface {
	ExpireGamesAndRefundBets(ctx context.Context) (int, error)
}

type RpsGameExpiryWorker struct {
	rpsGame RpsGameExpiryServiceInterface
}

func NewRpsGameExpiryWorker(rpsGame RpsGameExpiryServiceInterface) jobs.Worker[RpsGameExpiryJobArgs] {
	return &RpsGameExpiryWorker{rpsGame: rpsGame}
}

// Work voids pending escrow for any expired bet games.
func (w *RpsGameExpiryWorker) Work(ctx context.Context, job *jobs.Job[RpsGameExpiryJobArgs]) error {
	processed, err := w.rpsGame.ExpireGamesAndRefundBets(ctx)
	if err != nil {
		return err
	}
	if processed > 0 {
		slog.InfoContext(ctx, "rps expiry sweep: refunded expired bet games", slog.Int("count", processed))
	}
	return nil
}
