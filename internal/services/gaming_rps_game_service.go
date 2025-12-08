package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

type RpsGameService interface {
	RequestGame(ctx context.Context, input *RpsGameRequestInput) (*RpsGameWithParticipants, error)
	RespondToGameRequest(ctx context.Context, input *GameRequestResponse) (*RpsGameWithParticipants, error)
}

type DbRpsGameService struct {
	adapter stores.StorageAdapterInterface
}

var _ RpsGameService = (*DbRpsGameService)(nil)

func NewDbRpsGameService(adapter stores.StorageAdapterInterface) *DbRpsGameService {
	return &DbRpsGameService{
		adapter: adapter,
	}
}

type RpsGameRequestInput struct {
	RequestingPlayerID   uuid.UUID
	InvitedPlayerID      uuid.UUID
	RequestingPlayerMove models.RpsParticipantMove
	DurationSeconds      int64
}

func (d *DbRpsGameService) RequestGame(ctx context.Context, input *RpsGameRequestInput) (*RpsGameWithParticipants, error) {
	gameWithParticipants := &RpsGameWithParticipants{}
	game, err := d.adapter.Gaming().CreateRpsGame(ctx, &models.RpsGame{
		ExpiresAt: time.Now().UTC().Add(time.Duration(input.DurationSeconds) * time.Second).UTC(),
		Status:    models.RpsGameStatusPending,
	})
	if err != nil {
		return nil, err
	}
	gameWithParticipants.RpsGame = game
	participantsInput := []*models.RpsParticipant{
		{
			PlayerID: input.RequestingPlayerID,
			Move:     input.RequestingPlayerMove,
			GameID:   game.ID,
			Result:   models.RpsParticipantResultTie,
			Status:   models.RpsParticipantStatusCompleted,
			Type:     models.RpsParticipantTypeHost,
		},
		{
			PlayerID: input.InvitedPlayerID,
			Move:     models.RpsParticipantMoveRock,
			Status:   models.RpsParticipantStatusPending,
			Result:   models.RpsParticipantResultTie,
			GameID:   game.ID,
			Type:     models.RpsParticipantTypeGuest,
		},
	}
	participants, err := d.adapter.Gaming().CreateRpsParticipants(ctx, participantsInput)
	if err != nil {
		return nil, err
	}
	gameWithParticipants.RpsGame.Participants = participants
	for _, p := range participants {
		if p.Type == models.RpsParticipantTypeGuest {
			gameWithParticipants.InvitedParticipant = p
		}
		if p.Type == models.RpsParticipantTypeHost {
			gameWithParticipants.RequestingParticipant = p
		}
	}
	return gameWithParticipants, nil
}

type GameRequestResponse struct {
	InvitedPlayerID uuid.UUID
	GameID          uuid.UUID
	Status          models.RpsGameStatus
	Move            models.RpsParticipantMove
}

func (d *DbRpsGameService) RespondToGameRequest(ctx context.Context, input *GameRequestResponse) (*RpsGameWithParticipants, error) {
	gameWithParticipants, err := d.FindRpsWithParticipants(ctx, input.GameID)
	if err != nil {
		return nil, err
	}
	switch input.Status {
	case models.RpsGameStatusCancelled:
		cancelGame(gameWithParticipants)
	case models.RpsGameStatusCompleted:
		gameWithParticipants.InvitedParticipant.Move = input.Move
		gameWithParticipants.Status = models.RpsGameStatusCompleted
		playGame(gameWithParticipants)
	default:
		return nil, errors.New("invalid status")
	}

	updatedGameWithParticipants, err := d.updateGame(ctx, gameWithParticipants)
	if err != nil {
		return nil, err
	}
	return updatedGameWithParticipants, nil
}

func playGame(game *RpsGameWithParticipants) {
	if game.RequestingParticipant.Move == game.InvitedParticipant.Move {
		game.RequestingParticipant.Result = models.RpsParticipantResultTie
		game.InvitedParticipant.Result = models.RpsParticipantResultTie
	}
	if (game.RequestingParticipant.Move == models.RpsParticipantMoveRock && game.InvitedParticipant.Move == models.RpsParticipantMoveScissors) ||
		(game.RequestingParticipant.Move == models.RpsParticipantMovePaper && game.InvitedParticipant.Move == models.RpsParticipantMoveRock) ||
		(game.RequestingParticipant.Move == models.RpsParticipantMoveScissors && game.InvitedParticipant.Move == models.RpsParticipantMovePaper) {
		game.RequestingParticipant.Result = models.RpsParticipantResultWin
		game.InvitedParticipant.Result = models.RpsParticipantResultLose
	} else {
		game.RequestingParticipant.Result = models.RpsParticipantResultLose
		game.InvitedParticipant.Result = models.RpsParticipantResultWin
	}
}

func cancelGame(game *RpsGameWithParticipants) {
	game.RpsGame.Status = models.RpsGameStatusCancelled
	game.InvitedParticipant.Status = models.RpsParticipantStatusDeclined
}

func (d *DbRpsGameService) updateGame(ctx context.Context, gameWithParticipants *RpsGameWithParticipants) (*RpsGameWithParticipants, error) {
	gameToUpdate := gameWithParticipants.RpsGame
	requestingPlayer := gameWithParticipants.RequestingParticipant
	invitedParticipant := gameWithParticipants.InvitedParticipant
	updatedGame, err := d.adapter.Gaming().UpdateRpsGame(ctx, gameToUpdate)
	if err != nil {
		return nil, err
	}

	updatedRequestingPlayer, err := d.adapter.Gaming().UpdateRpsParticipant(ctx, requestingPlayer)
	if err != nil {
		return nil, err
	}
	updatedInvitedParticipant, err := d.adapter.Gaming().UpdateRpsParticipant(ctx, invitedParticipant)
	if err != nil {
		return nil, err
	}
	updatedGameWithParticipants := &RpsGameWithParticipants{
		RpsGame:               updatedGame,
		RequestingParticipant: updatedRequestingPlayer,
		InvitedParticipant:    updatedInvitedParticipant,
	}
	return updatedGameWithParticipants, nil
}

type RpsGameWithParticipants struct {
	*models.RpsGame
	RequestingParticipant *models.RpsParticipant
	InvitedParticipant    *models.RpsParticipant
}

func (d *DbRpsGameService) FindRpsWithParticipants(ctx context.Context, gameID uuid.UUID) (*RpsGameWithParticipants, error) {
	game, err := d.adapter.Gaming().FindRpsGame(ctx, &stores.RpsGameFilter{
		Ids: []uuid.UUID{gameID},
	})
	if err != nil {
		return nil, err
	}
	participants, err := d.adapter.Gaming().FindRpsParticipants(ctx, &stores.RpsParticipantFilter{
		RpsGameIds: []uuid.UUID{gameID},
	})
	if err != nil {
		return nil, err
	}
	var requestingPlayer, invitedPlayer *models.RpsParticipant
	for _, p := range participants {
		if p.Type == models.RpsParticipantTypeHost {
			requestingPlayer = p
		}
		if p.Type == models.RpsParticipantTypeGuest {
			invitedPlayer = p
		}
	}
	return &RpsGameWithParticipants{
		RpsGame:               game,
		RequestingParticipant: requestingPlayer,
		InvitedParticipant:    invitedPlayer,
	}, nil
}
