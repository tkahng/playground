package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

type PlayerFindParams struct {
	Email  string
	UserID *uuid.UUID
}

type RpsGameService interface {
	// player
	FindPlayerByParams(ctx context.Context, params *PlayerFindParams) (*models.Player, error)
	CreatePlayerByParams(ctx context.Context, params *PlayerFindParams) (*models.Player, error)
	// friendship
	PlayerCanPlayWithPlayer(ctx context.Context, requestingPlayerID uuid.UUID, invitedPlayerID uuid.UUID) (bool, error)
	// rps game
	FindRpsGameWithParticipants(ctx context.Context, gameID uuid.UUID) (*RpsGameWithParticipants, error)
	RequestGame(ctx context.Context, input *RpsGameRequestInput) (*RpsGameWithParticipants, error)
	RespondToGameRequest(ctx context.Context, input *GameRequestResponse) (*RpsGameWithParticipants, error)
	// ExpireGamesAndRefundBets finds all pending bet games whose expiry has passed,
	// marks each cancelled, and voids the host's pending escrow transfer.
	ExpireGamesAndRefundBets(ctx context.Context) (int, error)
}

type DbRpsGameService struct {
	adapter stores.StorageAdapterInterface
	betting BettingService
}

// CreatePlayerByParams implements [RpsGameService].
func (d *DbRpsGameService) CreatePlayerByParams(ctx context.Context, params *PlayerFindParams) (*models.Player, error) {
	return d.adapter.Gaming().CreatePlayer(ctx, &models.Player{
		UserID: params.UserID,
		Email:  params.Email,
	})
}

// FindPlayerByParams implements [RpsGameService].
func (d *DbRpsGameService) FindPlayerByParams(ctx context.Context, params *PlayerFindParams) (*models.Player, error) {
	return d.adapter.Gaming().FindPlayer(ctx, &stores.PlayersFilter{
		Emails: []string{params.Email},
	})
}

// PlayerCanPlayWithPlayer implements [RpsGameService].
func (d *DbRpsGameService) PlayerCanPlayWithPlayer(ctx context.Context, requestingPlayerID uuid.UUID, invitedPlayerID uuid.UUID) (bool, error) {
	friendship, err := d.adapter.Gaming().FindFriendship(ctx, &stores.FriendshipFilter{
		RequestingPlayerIds: []uuid.UUID{requestingPlayerID},
		InvitedPlayerIds:    []uuid.UUID{invitedPlayerID},
	})
	if err != nil {
		return false, err
	}
	if friendship != nil {
		if friendship.Status == models.FriendshipStatusDeclined {
			return false, nil
		}
		return true, nil
	}

	return true, nil
}

var _ RpsGameService = (*DbRpsGameService)(nil)

func NewDbRpsGameService(adapter stores.StorageAdapterInterface, betting BettingService) *DbRpsGameService {
	return &DbRpsGameService{
		adapter: adapter,
		betting: betting,
	}
}

type RpsGameRequestInput struct {
	RequestingPlayerID   uuid.UUID
	InvitedPlayerID      uuid.UUID
	RequestingPlayerMove models.RpsParticipantMove
	DurationSeconds      int64
	// BetAmount is the number of points each player wagers.
	// nil means no bet. The requesting player (host) must have sufficient balance.
	BetAmount *int64
	// HostUserID is required when BetAmount is set; it identifies the wallet owner.
	HostUserID *uuid.UUID
}

func (d *DbRpsGameService) RequestGame(ctx context.Context, input *RpsGameRequestInput) (*RpsGameWithParticipants, error) {
	if input.RequestingPlayerID == input.InvitedPlayerID {
		return nil, errors.New("cannot challenge yourself")
	}
	// Validate betting prerequisites.
	if input.BetAmount != nil {
		if *input.BetAmount <= 0 {
			return nil, errors.New("bet amount must be positive")
		}
		if input.HostUserID == nil {
			return nil, errors.New("bet requires a registered host user (HostUserID)")
		}
		if d.betting == nil {
			return nil, errors.New("betting service is not available")
		}
	}

	gameWithParticipants := &RpsGameWithParticipants{}
	game, err := d.adapter.Gaming().CreateRpsGame(ctx, &models.RpsGame{
		ExpiresAt: time.Now().UTC().Add(time.Duration(input.DurationSeconds) * time.Second).UTC(),
		Status:    models.RpsGameStatusPending,
		BetAmount: input.BetAmount,
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

	// Place the host's bet escrow if a bet amount is set.
	if input.BetAmount != nil && input.HostUserID != nil {
		hostPending, err := d.betting.PlaceHostBet(ctx, game.ID, *input.HostUserID, *input.BetAmount)
		if err != nil {
			return nil, fmt.Errorf("place host bet: %w", err)
		}
		game.HostBetTransferID = &hostPending.ID
		updatedGame, err := d.adapter.Gaming().UpdateRpsGame(ctx, game)
		if err != nil {
			return nil, fmt.Errorf("save host bet transfer id: %w", err)
		}
		gameWithParticipants.RpsGame = updatedGame
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
	// Lock the game row first to prevent concurrent double-settlement.
	lockedGame, err := d.adapter.Gaming().FindRpsGameForUpdate(ctx, input.GameID)
	if err != nil {
		return nil, err
	}
	if lockedGame == nil {
		return nil, errors.New("game not found")
	}

	gameWithParticipants, err := d.FindRpsGameWithParticipants(ctx, input.GameID)
	if err != nil {
		return nil, err
	}
	if gameWithParticipants.RpsGame.ExpiresAt.UTC().Before(time.Now().UTC()) {
		return nil, errors.New("game expired")
	}
	if gameWithParticipants.RpsGame.Status != models.RpsGameStatusPending {
		return nil, errors.New("game is not pending")
	}
	if gameWithParticipants.InvitedParticipant.PlayerID != input.InvitedPlayerID {
		return nil, errors.New("invited player does not match")
	}

	game := gameWithParticipants.RpsGame
	hasBet := game.BetAmount != nil && *game.BetAmount > 0

	switch input.Status {
	case models.RpsGameStatusCancelled:
		game.Status = models.RpsGameStatusCancelled
		gameWithParticipants.InvitedParticipant.Status = models.RpsParticipantStatusDeclined

		// Refund the host's pending escrow if a bet was placed.
		if hasBet && game.HostBetTransferID != nil && d.betting != nil {
			if err := d.betting.RefundHostBet(ctx, *game.HostBetTransferID); err != nil {
				return nil, fmt.Errorf("refund host bet on cancel: %w", err)
			}
		}

	case models.RpsGameStatusCompleted:
		gameWithParticipants.InvitedParticipant.Move = input.Move
		game.Status = models.RpsGameStatusCompleted
		gameWithParticipants.RequestingParticipant.Status = models.RpsParticipantStatusCompleted
		gameWithParticipants.InvitedParticipant.Status = models.RpsParticipantStatusCompleted

		// Determine game result.
		hostMove := gameWithParticipants.RequestingParticipant.Move
		guestMove := gameWithParticipants.InvitedParticipant.Move
		if hostMove == guestMove {
			gameWithParticipants.RequestingParticipant.Result = models.RpsParticipantResultTie
			gameWithParticipants.InvitedParticipant.Result = models.RpsParticipantResultTie
		} else if (hostMove == models.RpsParticipantMoveRock && guestMove == models.RpsParticipantMoveScissors) ||
			(hostMove == models.RpsParticipantMovePaper && guestMove == models.RpsParticipantMoveRock) ||
			(hostMove == models.RpsParticipantMoveScissors && guestMove == models.RpsParticipantMovePaper) {
			gameWithParticipants.RequestingParticipant.Result = models.RpsParticipantResultWin
			gameWithParticipants.InvitedParticipant.Result = models.RpsParticipantResultLose
		} else {
			gameWithParticipants.RequestingParticipant.Result = models.RpsParticipantResultLose
			gameWithParticipants.InvitedParticipant.Result = models.RpsParticipantResultWin
		}

		// Settle the bet if one was placed.
		if hasBet && game.HostBetTransferID != nil && d.betting != nil {
			hostPlayer := gameWithParticipants.RequestingParticipant.Player
			if hostPlayer == nil || hostPlayer.UserID == nil {
				return nil, errors.New("host player must be a registered user to settle a bet")
			}
			guestPlayer := gameWithParticipants.InvitedParticipant.Player
			if guestPlayer == nil || guestPlayer.UserID == nil {
				return nil, errors.New("guest player must be a registered user to settle a bet")
			}
			// Server-side guest balance check before committing.
			if err := d.betting.EnsureGuestCanAffordBet(ctx, *guestPlayer.UserID, *game.BetAmount); err != nil {
				return nil, fmt.Errorf("guest cannot cover bet: %w", err)
			}
			guestTransferID, err := d.betting.PlaceGuestAndSettle(ctx, PlaceGuestAndSettleInput{
				GameID:                input.GameID,
				GuestUserID:           *guestPlayer.UserID,
				HostUserID:            *hostPlayer.UserID,
				BetAmount:             *game.BetAmount,
				HostPendingTransferID: *game.HostBetTransferID,
				HostResult:            gameWithParticipants.RequestingParticipant.Result,
				GuestResult:           gameWithParticipants.InvitedParticipant.Result,
			})
			if err != nil {
				return nil, fmt.Errorf("settle bet: %w", err)
			}
			game.GuestBetTransferID = &guestTransferID
		}

	default:
		return nil, errors.New("invalid status")
	}

	updatedGameWithParticipants, err := d.updateGame(ctx, gameWithParticipants)
	if err != nil {
		return nil, err
	}
	return updatedGameWithParticipants, nil
}

func (d *DbRpsGameService) updateGame(ctx context.Context, gameWithParticipants *RpsGameWithParticipants) (*RpsGameWithParticipants, error) {
	gameToUpdate := gameWithParticipants.RpsGame
	requestingParticipant := gameWithParticipants.RequestingParticipant
	invitedParticipant := gameWithParticipants.InvitedParticipant
	updatedGame, err := d.adapter.Gaming().UpdateRpsGame(ctx, gameToUpdate)
	if err != nil {
		return nil, err
	}

	updatedRequestingParticipant, err := d.adapter.Gaming().UpdateRpsParticipant(ctx, requestingParticipant)
	if err != nil {
		return nil, err
	}
	updatedRequestingParticipant.Player = requestingParticipant.Player
	updatedInvitedParticipant, err := d.adapter.Gaming().UpdateRpsParticipant(ctx, invitedParticipant)
	if err != nil {
		return nil, err
	}
	updatedInvitedParticipant.Player = invitedParticipant.Player
	updatedGameWithParticipants := &RpsGameWithParticipants{
		RpsGame:               updatedGame,
		RequestingParticipant: updatedRequestingParticipant,
		InvitedParticipant:    updatedInvitedParticipant,
	}
	return updatedGameWithParticipants, nil
}

type RpsGameWithParticipants struct {
	RpsGame               *models.RpsGame
	RequestingParticipant *models.RpsParticipant
	InvitedParticipant    *models.RpsParticipant
}

// ExpireGamesAndRefundBets finds all pending bet games whose expiry has passed,
// marks each cancelled, and voids the host's pending escrow transfer.
// Errors from individual games are logged and collected; the sweep continues
// for all remaining games and returns a joined error at the end.
func (d *DbRpsGameService) ExpireGamesAndRefundBets(ctx context.Context) (int, error) {
	expiredGames, err := d.adapter.Gaming().FindExpiredPendingBetGames(ctx)
	if err != nil {
		return 0, fmt.Errorf("find expired bet games: %w", err)
	}

	processed := 0
	var errs []error
	for _, game := range expiredGames {
		txErr := d.adapter.RunInTxCtx(ctx, func(txCtx context.Context) error {
			// Re-fetch with lock inside the transaction.
			locked, err := d.adapter.Gaming().FindRpsGameForUpdate(txCtx, game.ID)
			if err != nil {
				return err
			}
			if locked == nil || locked.Status != models.RpsGameStatusPending {
				return nil // already handled by another process
			}
			locked.Status = models.RpsGameStatusCancelled
			if _, err := d.adapter.Gaming().UpdateRpsGame(txCtx, locked); err != nil {
				return fmt.Errorf("cancel expired game %s: %w", game.ID, err)
			}
			if game.HostBetTransferID != nil && d.betting != nil {
				if err := d.betting.RefundHostBet(txCtx, *game.HostBetTransferID); err != nil {
					return fmt.Errorf("refund host bet for expired game %s: %w", game.ID, err)
				}
			}
			return nil
		})
		if txErr != nil {
			slog.ErrorContext(ctx, "expiry sweep: failed to expire game",
				slog.String("game_id", game.ID.String()),
				slog.Any("error", txErr),
			)
			errs = append(errs, txErr)
			continue
		}
		processed++
	}
	return processed, errors.Join(errs...)
}

func (d *DbRpsGameService) FindRpsGameWithParticipants(ctx context.Context, gameID uuid.UUID) (*RpsGameWithParticipants, error) {
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
		player, err := d.adapter.Gaming().FindPlayer(ctx, &stores.PlayersFilter{
			Ids: []uuid.UUID{p.PlayerID},
		})
		if err != nil {
			return nil, err
		}
		p.Player = player
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
