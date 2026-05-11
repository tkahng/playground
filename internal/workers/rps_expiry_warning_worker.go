package workers

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/sse"
	"github.com/tkahng/playground/internal/tools/types"
)

const (
	RpsExpiryWarningSweepKind     = "rps_expiry_warning_sweep"
	RpsExpiryWarningSweepInterval = 30 * time.Minute
	// RpsExpiryWarningWindow is how far ahead of expiry the warning is sent.
	RpsExpiryWarningWindow = 24 * time.Hour
)

type RpsExpiryWarningJobArgs struct{}

func (j RpsExpiryWarningJobArgs) Kind() string { return RpsExpiryWarningSweepKind }

// RpsExpiryWarningSender is the minimal SSE interface the worker needs.
type RpsExpiryWarningSender interface {
	Send(channel string, payload interface{}) error
}

type RpsExpiryWarningWorker struct {
	adapter  stores.StorageAdapterInterface
	sse      RpsExpiryWarningSender
	enqueuer jobs.Enqueuer
}

func NewRpsExpiryWarningWorker(
	adapter stores.StorageAdapterInterface,
	sse RpsExpiryWarningSender,
	enqueuer jobs.Enqueuer,
) jobs.Worker[RpsExpiryWarningJobArgs] {
	return &RpsExpiryWarningWorker{adapter: adapter, sse: sse, enqueuer: enqueuer}
}

func SeedRpsExpiryWarningJob(ctx context.Context, enqueuer jobs.Enqueuer) error {
	return enqueuer.Enqueue(ctx, &jobs.EnqueueParams{
		Args:        &RpsExpiryWarningJobArgs{},
		RunAfter:    time.Now(),
		MaxAttempts: 1,
		UniqueKey:   types.Pointer(RpsExpiryWarningSweepKind),
	})
}

func (w *RpsExpiryWarningWorker) Work(ctx context.Context, job *jobs.Job[RpsExpiryWarningJobArgs]) error {
	games, err := w.adapter.Gaming().FindPendingGamesExpiringWithin(ctx, RpsExpiryWarningWindow)
	if err != nil {
		return err
	}

	for _, game := range games {
		w.sendWarningForGame(ctx, game)
		if err := w.adapter.Gaming().MarkRpsGameExpirySent(ctx, game); err != nil {
			slog.WarnContext(ctx, "rps expiry warning: failed to mark game as warned",
				slog.String("game_id", game.ID.String()), slog.Any("error", err))
		}
	}

	if len(games) > 0 {
		slog.InfoContext(ctx, "rps expiry warning: sent warnings", slog.Int("count", len(games)))
	}

	if err := w.enqueuer.Enqueue(context.Background(), &jobs.EnqueueParams{
		Args:        &RpsExpiryWarningJobArgs{},
		RunAfter:    time.Now().Add(RpsExpiryWarningSweepInterval),
		MaxAttempts: 1,
		UniqueKey:   types.Pointer(RpsExpiryWarningSweepKind),
	}); err != nil {
		slog.ErrorContext(ctx, "rps expiry warning: failed to schedule next run", slog.Any("error", err))
	}
	return nil
}

type rpsExpiryWarningPayload struct {
	Type      string `json:"type"`
	GameID    string `json:"game_id"`
	ExpiresAt string `json:"expires_at"`
}

func (w *RpsExpiryWarningWorker) sendWarningForGame(ctx context.Context, game *models.RpsGame) {
	participants, err := w.adapter.Gaming().FindRpsParticipants(ctx, &stores.RpsParticipantFilter{
		RpsGameIds: []uuid.UUID{game.ID},
	})
	if err != nil {
		slog.WarnContext(ctx, "rps expiry warning: failed to load participants",
			slog.String("game_id", game.ID.String()), slog.Any("error", err))
		return
	}
	payload := rpsExpiryWarningPayload{
		Type:      "rps_game_expiring_soon",
		GameID:    game.ID.String(),
		ExpiresAt: game.ExpiresAt.Format(time.RFC3339),
	}
	for _, p := range participants {
		if err := w.sse.Send(sse.PlayerChannel(p.PlayerID.String()), payload); err != nil {
			slog.WarnContext(ctx, "rps expiry warning: SSE send failed",
				slog.String("player_id", p.PlayerID.String()), slog.Any("error", err))
		}
	}
}
