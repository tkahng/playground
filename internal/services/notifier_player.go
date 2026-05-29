package services

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/sse"
)

// PlayerNotifier persists a notification row and broadcasts via SSE for
// player-scoped (gaming) notifications.
type PlayerNotifier interface {
	// Notify stores the notification and sends it via SSE. playerID is the
	// gaming.players primary key; notifType is the event name used as the SSE
	// event type and notification.type column value.
	Notify(ctx context.Context, playerID uuid.UUID, notifType string, ssePayload any) error
}

var _ PlayerNotifier = (*DbPlayerNotifier)(nil)

type DbPlayerNotifier struct {
	sseManager sse.Manager
	adapter    stores.StorageAdapterInterface
}

func NewDbPlayerNotifier(sseManager sse.Manager, adapter stores.StorageAdapterInterface) *DbPlayerNotifier {
	return &DbPlayerNotifier{sseManager: sseManager, adapter: adapter}
}

func (n *DbPlayerNotifier) Notify(ctx context.Context, playerID uuid.UUID, notifType string, ssePayload any) error {
	payloadBytes, err := json.Marshal(ssePayload)
	if err != nil {
		return err
	}
	player, err := n.adapter.Gaming().FindPlayer(ctx, &stores.PlayersFilter{Ids: []uuid.UUID{playerID}})
	if err != nil {
		return err
	}
	if player == nil {
		slog.WarnContext(ctx, "player not found for notification, skipping persist", slog.String("player_id", playerID.String()))
	} else {
		_, persistErr := n.adapter.Notification().CreateNotification(ctx, &models.Notification{
			UserID:   player.UserID,
			Channel:  sse.PlayerChannel(playerID.String()),
			Type:     notifType,
			Payload:  payloadBytes,
			Metadata: map[string]any{},
		})
		if persistErr != nil {
			slog.WarnContext(ctx, "failed to persist player notification", slog.Any("error", persistErr))
		}
	}
	if sendErr := n.sseManager.Send(sse.PlayerChannel(playerID.String()), ssePayload); sendErr != nil {
		slog.WarnContext(ctx, "player SSE send failed", slog.String("player_id", playerID.String()), slog.Any("error", sendErr))
	}
	return nil
}
