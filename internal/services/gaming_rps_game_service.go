package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

type RpsGameRequestInput struct {
	RequestingPlayerEmail string
	InvitedPlayerEmail    string
	RequestingPlayerMove  models.RpsParticipantMove
	DurationSeconds       int64
}

type RpsGameInvitedPlayerResposne struct {
}
type RpsGameService interface {
	// FindOrCreatePlayerByEmail(ctx context.Context, email string) (*models.Player, error)
	// RequestGame(ctx context.Context, input *RpsGameRequestInput) (*models.RpsGame, error)
}

type DbRpsGameService struct {
	jobService JobService
	adapter    stores.StorageAdapterInterface
}

// FindOrCreatePlayerByEmail implements [RpsGameService].
func (d *DbRpsGameService) FindOrCreatePlayerByEmail(ctx context.Context, email string) (*models.Player, error) {
	player, err := d.adapter.Gaming().FindPlayer(ctx, &stores.PlayersFilter{
		Emails: []string{email},
	})
	if err != nil {
		return nil, err
	}
	if player == nil {
		player, err = d.adapter.Gaming().CreatePlayer(ctx, &models.Player{
			Email: email,
		})
		if err != nil {
			return nil, err
		}
	}
	return player, nil
}

// RequestGame implements [RpsGameService].
func (d *DbRpsGameService) RequestGame(ctx context.Context, input *RpsGameRequestInput) (*models.RpsGame, error) {
	var (
		requestingPlayer, invitedPlayer *models.Player
		err                             error
	)
	requestingPlayer, err = d.adapter.Gaming().FindPlayer(ctx, &stores.PlayersFilter{
		Emails: []string{input.RequestingPlayerEmail},
	})
	if err != nil {
		return nil, err
	}
	if requestingPlayer == nil {
		return nil, fmt.Errorf("requesting player not found")
	}
	invitedPlayer, err = d.adapter.Gaming().FindPlayer(ctx, &stores.PlayersFilter{
		Emails: []string{input.InvitedPlayerEmail},
	})
	if err != nil {
		return nil, err
	}
	if invitedPlayer == nil {
		invitedPlayer, err = d.adapter.Gaming().CreatePlayer(ctx, &models.Player{
			Email: input.InvitedPlayerEmail,
		})
		if err != nil {
			return nil, err
		}
	}

	return d.CreateRpsGameWithParticipants(ctx, &RpsGameCreateInput{
		RequestingPlayerID:   requestingPlayer.ID,
		InvitedPlayerID:      invitedPlayer.ID,
		DurationSeconds:      input.DurationSeconds,
		RequestingPlayerMove: input.RequestingPlayerMove,
	})
}

type RpsGameCreateInput struct {
	RequestingPlayerID   uuid.UUID
	InvitedPlayerID      uuid.UUID
	DurationSeconds      int64
	RequestingPlayerMove models.RpsParticipantMove
}

func (d *DbRpsGameService) CreateRpsGameWithParticipants(ctx context.Context, input *RpsGameCreateInput) (*models.RpsGame, error) {
	game, err := d.adapter.Gaming().CreateRpsGame(ctx, &models.RpsGame{
		ExpiresAt: time.Now().UTC().Add(time.Duration(input.DurationSeconds) * time.Second).UTC(),
		Status:    models.RpsGameStatusPending,
	})
	if err != nil {
		return nil, err
	}
	participants := []*models.RpsParticipant{
		{
			PlayerID: input.RequestingPlayerID,
			Move:     input.RequestingPlayerMove,
			GameID:   game.ID,
			Type:     models.RpsParticipantTypeHost,
		},
		{
			PlayerID: input.InvitedPlayerID,
			Move:     models.RpsParticipantMoveRock,
			GameID:   game.ID,
			Type:     models.RpsParticipantTypeGuest,
		},
	}
	participants, err = d.adapter.Gaming().CreateRpsParticipants(ctx, participants)
	if err != nil {
		return nil, err
	}
	game.Participants = participants
	return game, nil
}

var _ RpsGameService = (*DbRpsGameService)(nil)

func NewDbRpsGameService(adapter stores.StorageAdapterInterface, jobService JobService) *DbRpsGameService {
	return &DbRpsGameService{
		adapter:    adapter,
		jobService: jobService,
	}
}
