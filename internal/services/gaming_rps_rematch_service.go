package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/apierrors"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

const RematchTTL = 45 * time.Second

type RematchRequestInput struct {
	OriginalGameID     uuid.UUID
	RequestingPlayerID uuid.UUID
	InvitedPlayerID    uuid.UUID
}

func (d *DbRpsGameService) RequestRematch(ctx context.Context, input *RematchRequestInput) (*models.RpsRematchRequest, error) {
	game, err := d.adapter.Gaming().FindRpsGame(ctx, &stores.RpsGameFilter{
		Ids: []uuid.UUID{input.OriginalGameID},
	})
	if err != nil {
		return nil, err
	}
	if game == nil {
		return nil, apierrors.NotFound("original game not found")
	}
	if game.Status != models.RpsGameStatusCompleted {
		return nil, apierrors.BadRequest("can only rematch a completed game")
	}

	// Check no pending rematch already exists for this game.
	existing, err := d.adapter.Gaming().FindRpsRematchRequest(ctx, &stores.RpsRematchFilter{
		OriginalGameIDs: []uuid.UUID{input.OriginalGameID},
		Statuses:        []models.RpsRematchStatus{models.RpsRematchStatusPending},
	})
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apierrors.Conflict("a pending rematch request already exists for this game")
	}

	req := &models.RpsRematchRequest{
		OriginalGameID:     input.OriginalGameID,
		RequestingPlayerID: input.RequestingPlayerID,
		InvitedPlayerID:    input.InvitedPlayerID,
		Status:             models.RpsRematchStatusPending,
		ExpiresAt:          time.Now().UTC().Add(RematchTTL),
	}
	return d.adapter.Gaming().CreateRpsRematchRequest(ctx, req)
}

func (d *DbRpsGameService) AcceptRematch(ctx context.Context, rematchID uuid.UUID, invitedPlayerID uuid.UUID) (*models.RpsRematchRequest, error) {
	rematch, err := d.adapter.Gaming().FindRpsRematchRequest(ctx, &stores.RpsRematchFilter{
		Ids: []uuid.UUID{rematchID},
	})
	if err != nil {
		return nil, err
	}
	if rematch == nil {
		return nil, apierrors.NotFound("rematch request not found")
	}
	if rematch.InvitedPlayerID != invitedPlayerID {
		return nil, apierrors.Forbidden("not the invited player")
	}
	if rematch.Status != models.RpsRematchStatusPending {
		return nil, apierrors.BadRequest(fmt.Sprintf("rematch is already %s", rematch.Status))
	}
	if time.Now().UTC().After(rematch.ExpiresAt) {
		rematch.Status = models.RpsRematchStatusExpired
		if _, err := d.adapter.Gaming().UpdateRpsRematchRequest(ctx, rematch); err != nil {
			slog.ErrorContext(ctx, "failed to expire rematch on accept", "rematch_id", rematchID, "error", err)
		}
		return nil, apierrors.Gone("rematch request has expired")
	}

	// Create new game with roles swapped (previous guest is now host, no move yet).
	// The new host needs to supply their move via the normal challenge flow.
	// We just create a pending game here and both players complete it via submit-move.
	newGame, err := d.adapter.Gaming().CreateRpsGame(ctx, &models.RpsGame{
		ExpiresAt: time.Now().UTC().Add(3 * 24 * time.Hour),
		Status:    models.RpsGameStatusPending,
	})
	if err != nil {
		return nil, fmt.Errorf("create rematch game: %w", err)
	}

	_, err = d.adapter.Gaming().CreateRpsParticipants(ctx, []*models.RpsParticipant{
		{
			PlayerID: invitedPlayerID,
			Move:     models.RpsParticipantMoveRock,
			GameID:   newGame.ID,
			Result:   models.RpsParticipantResultTie,
			Status:   models.RpsParticipantStatusPending,
			Type:     models.RpsParticipantTypeHost,
		},
		{
			PlayerID: rematch.RequestingPlayerID,
			Move:     models.RpsParticipantMoveRock,
			GameID:   newGame.ID,
			Result:   models.RpsParticipantResultTie,
			Status:   models.RpsParticipantStatusPending,
			Type:     models.RpsParticipantTypeGuest,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create rematch participants: %w", err)
	}

	rematch.Status = models.RpsRematchStatusAccepted
	rematch.NewGameID = &newGame.ID
	updated, err := d.adapter.Gaming().UpdateRpsRematchRequest(ctx, rematch)
	if err != nil {
		return nil, fmt.Errorf("update rematch status: %w", err)
	}
	return updated, nil
}

func (d *DbRpsGameService) DeclineRematch(ctx context.Context, rematchID uuid.UUID, invitedPlayerID uuid.UUID) (*models.RpsRematchRequest, error) {
	rematch, err := d.adapter.Gaming().FindRpsRematchRequest(ctx, &stores.RpsRematchFilter{
		Ids: []uuid.UUID{rematchID},
	})
	if err != nil {
		return nil, err
	}
	if rematch == nil {
		return nil, apierrors.NotFound("rematch request not found")
	}
	if rematch.InvitedPlayerID != invitedPlayerID {
		return nil, apierrors.Forbidden("not the invited player")
	}
	if rematch.Status != models.RpsRematchStatusPending {
		return nil, apierrors.BadRequest(fmt.Sprintf("rematch is already %s", rematch.Status))
	}
	rematch.Status = models.RpsRematchStatusDeclined
	return d.adapter.Gaming().UpdateRpsRematchRequest(ctx, rematch)
}

func (d *DbRpsGameService) ExpireRematches(ctx context.Context) (int, error) {
	expired, err := d.adapter.Gaming().FindExpiredPendingRpsRematches(ctx)
	if err != nil {
		return 0, fmt.Errorf("find expired rematches: %w", err)
	}
	count := 0
	var errs []error
	for _, r := range expired {
		r.Status = models.RpsRematchStatusExpired
		if _, err := d.adapter.Gaming().UpdateRpsRematchRequest(ctx, r); err != nil {
			errs = append(errs, fmt.Errorf("expire rematch %s: %w", r.ID, err))
			continue
		}
		count++
	}
	return count, errors.Join(errs...)
}
