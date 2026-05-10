package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/apierrors"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

const (
	HouseMaxBet      int64         = 500
	HouseCooldown    time.Duration = 5 * time.Minute
	HouseWinsMessage               = "House always wins."
)

type PlayerFindParams struct {
	Email  string
	UserID *uuid.UUID
}

type ChallengeHouseInput struct {
	RequestingPlayerID   uuid.UUID
	RequestingPlayerMove models.RpsParticipantMove
	BetAmount            *int64
	HostUserID           *uuid.UUID
}

type ChallengeHouseResult struct {
	Game           *RpsGameWithParticipants
	HouseMessage   *string
	CooldownEndsAt time.Time
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
	// ChallengeHouse creates and immediately resolves a game against the house bot.
	// The house picks a random move; the result is returned synchronously.
	ChallengeHouse(ctx context.Context, input *ChallengeHouseInput) (*ChallengeHouseResult, error)
	// ExpireGamesAndRefundBets finds all pending bet games whose expiry has passed,
	// marks each cancelled, and voids the host's pending escrow transfer.
	ExpireGamesAndRefundBets(ctx context.Context) (int, error)
}

type DbRpsGameService struct {
	adapter         stores.StorageAdapterInterface
	betting         BettingService
	houseThinkDelay time.Duration
}

func (d *DbRpsGameService) WithHouseThinkDelay(delay time.Duration) *DbRpsGameService {
	d.houseThinkDelay = delay
	return d
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

// PlayerCanPlayWithPlayer returns false when either player has blocked the other.
func (d *DbRpsGameService) PlayerCanPlayWithPlayer(ctx context.Context, requestingPlayerID uuid.UUID, invitedPlayerID uuid.UUID) (bool, error) {
	pair := [2]uuid.UUID{requestingPlayerID, invitedPlayerID}
	f, err := d.adapter.Gaming().FindFriendship(ctx, &stores.FriendshipFilter{
		PlayerPair: &pair,
		Statuses:   []models.FriendshipStatus{models.FriendshipStatusBlocked},
	})
	if err != nil {
		return false, err
	}
	return f == nil, nil
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

func (d *DbRpsGameService) playerHasActiveGame(ctx context.Context, playerID uuid.UUID) (bool, error) {
	count, err := d.adapter.Gaming().CountRpsGames(ctx, &stores.RpsGameFilter{
		ParticipantIds: []uuid.UUID{playerID},
		Statuses:       []models.RpsGameStatus{models.RpsGameStatusPending},
	})
	if err != nil {
		return false, fmt.Errorf("check active game for player %s: %w", playerID, err)
	}
	return count > 0, nil
}

func (d *DbRpsGameService) RequestGame(ctx context.Context, input *RpsGameRequestInput) (*RpsGameWithParticipants, error) {
	if input.RequestingPlayerID == input.InvitedPlayerID {
		return nil, errors.New("cannot challenge yourself")
	}
	// Enforce one active game per player.
	if active, err := d.playerHasActiveGame(ctx, input.RequestingPlayerID); err != nil {
		return nil, err
	} else if active {
		return nil, apierrors.Conflict("you already have an active game in progress")
	}
	if active, err := d.playerHasActiveGame(ctx, input.InvitedPlayerID); err != nil {
		return nil, err
	} else if active {
		return nil, apierrors.Conflict("invited player already has an active game in progress")
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
		// Verify HostUserID belongs to the requesting player.
		hostPlayer, err := d.adapter.Gaming().FindPlayer(ctx, &stores.PlayersFilter{
			Ids: []uuid.UUID{input.RequestingPlayerID},
		})
		if err != nil {
			return nil, fmt.Errorf("look up requesting player: %w", err)
		}
		if hostPlayer == nil || hostPlayer.UserID == nil || *hostPlayer.UserID != *input.HostUserID {
			return nil, errors.New("HostUserID does not match the requesting player's registered user")
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
		return nil, apierrors.NotFound("game not found")
	}

	gameWithParticipants, err := d.FindRpsGameWithParticipants(ctx, input.GameID)
	if err != nil {
		return nil, err
	}
	if gameWithParticipants.RpsGame.ExpiresAt.UTC().Before(time.Now().UTC()) {
		return nil, apierrors.Gone("game expired")
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
		gameWithParticipants.RequestingParticipant.Result, gameWithParticipants.InvitedParticipant.Result =
			determineRpsResult(hostMove, guestMove)

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
		return nil, apierrors.BadRequest("invalid status")
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
			if d.betting != nil && game.HostBetTransferID != nil {
				if game.GuestBetTransferID != nil {
					if err := d.betting.RefundBothBets(txCtx, *game.HostBetTransferID, *game.GuestBetTransferID); err != nil {
						return fmt.Errorf("refund both bets for expired game %s: %w", game.ID, err)
					}
				} else {
					if err := d.betting.RefundHostBet(txCtx, *game.HostBetTransferID); err != nil {
						return fmt.Errorf("refund host bet for expired game %s: %w", game.ID, err)
					}
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

var houseMoves = []models.RpsParticipantMove{
	models.RpsParticipantMoveRock,
	models.RpsParticipantMovePaper,
	models.RpsParticipantMoveScissors,
}

func randomHouseMove() models.RpsParticipantMove {
	return houseMoves[rand.Intn(len(houseMoves))]
}

// determineRpsResult returns (hostResult, guestResult).
func determineRpsResult(hostMove, guestMove models.RpsParticipantMove) (models.RpsParticipantResult, models.RpsParticipantResult) {
	if hostMove == guestMove {
		return models.RpsParticipantResultTie, models.RpsParticipantResultTie
	}
	if (hostMove == models.RpsParticipantMoveRock && guestMove == models.RpsParticipantMoveScissors) ||
		(hostMove == models.RpsParticipantMovePaper && guestMove == models.RpsParticipantMoveRock) ||
		(hostMove == models.RpsParticipantMoveScissors && guestMove == models.RpsParticipantMovePaper) {
		return models.RpsParticipantResultWin, models.RpsParticipantResultLose
	}
	return models.RpsParticipantResultLose, models.RpsParticipantResultWin
}

func (d *DbRpsGameService) ChallengeHouse(ctx context.Context, input *ChallengeHouseInput) (*ChallengeHouseResult, error) {
	// 1. Fetch house player.
	house, err := GetHousePlayer(ctx, d.adapter)
	if err != nil {
		return nil, fmt.Errorf("get house player: %w", err)
	}

	// 2. Requesting player must not have an active game.
	if active, err := d.playerHasActiveGame(ctx, input.RequestingPlayerID); err != nil {
		return nil, err
	} else if active {
		return nil, apierrors.Conflict("you already have an active game in progress")
	}

	// 3. Cooldown check.
	requester, err := d.adapter.Gaming().FindPlayer(ctx, &stores.PlayersFilter{
		Ids: []uuid.UUID{input.RequestingPlayerID},
	})
	if err != nil {
		return nil, fmt.Errorf("find requesting player: %w", err)
	}
	if requester == nil {
		return nil, apierrors.NotFound("player not found")
	}
	if requester.LastHouseGameAt != nil {
		elapsed := time.Since(*requester.LastHouseGameAt)
		if elapsed < HouseCooldown {
			remaining := HouseCooldown - elapsed
			return nil, apierrors.TooManyRequests(fmt.Sprintf(
				"house cooldown active — try again in %s", remaining.Truncate(time.Second),
			))
		}
	}

	// 4. Bet validation.
	if input.BetAmount != nil {
		if *input.BetAmount <= 0 {
			return nil, apierrors.BadRequest("bet amount must be positive")
		}
		if *input.BetAmount > HouseMaxBet {
			return nil, apierrors.BadRequest(fmt.Sprintf("bet amount cannot exceed %d pts vs the house", HouseMaxBet))
		}
		if input.HostUserID == nil {
			return nil, apierrors.BadRequest("bet requires a registered user (HostUserID)")
		}
		if requester.UserID == nil || *requester.UserID != *input.HostUserID {
			return nil, apierrors.BadRequest("HostUserID does not match the requesting player's registered user")
		}
	}

	// 5. Decide house move before touching the DB.
	houseMove := randomHouseMove()
	hostResult, guestResult := determineRpsResult(input.RequestingPlayerMove, houseMove)
	now := time.Now().UTC()

	var gameResult *RpsGameWithParticipants

	txErr := d.adapter.RunInTxCtx(ctx, func(txCtx context.Context) error {
		// Create game (completed immediately — no waiting for guest).
		game, err := d.adapter.Gaming().CreateRpsGame(txCtx, &models.RpsGame{
			ExpiresAt: now.Add(30 * time.Second), // short expiry; game is already done
			Status:    models.RpsGameStatusCompleted,
			BetAmount: input.BetAmount,
		})
		if err != nil {
			return fmt.Errorf("create game: %w", err)
		}

		completedAt := now
		game.CompletedAt = &completedAt

		participants, err := d.adapter.Gaming().CreateRpsParticipants(txCtx, []*models.RpsParticipant{
			{
				PlayerID: input.RequestingPlayerID,
				Move:     input.RequestingPlayerMove,
				GameID:   game.ID,
				Result:   hostResult,
				Status:   models.RpsParticipantStatusCompleted,
				Type:     models.RpsParticipantTypeHost,
			},
			{
				PlayerID: house.ID,
				Move:     houseMove,
				GameID:   game.ID,
				Result:   guestResult,
				Status:   models.RpsParticipantStatusCompleted,
				Type:     models.RpsParticipantTypeGuest,
			},
		})
		if err != nil {
			return fmt.Errorf("create participants: %w", err)
		}

		// Persist completedAt on game.
		game, err = d.adapter.Gaming().UpdateRpsGame(txCtx, game)
		if err != nil {
			return fmt.Errorf("update game: %w", err)
		}

		// Settle bet if one was placed.
		if input.BetAmount != nil && input.HostUserID != nil && d.betting != nil {
			if err := d.betting.EnsureGuestCanAffordBet(txCtx, *input.HostUserID, *input.BetAmount); err != nil {
				return fmt.Errorf("user cannot cover bet: %w", err)
			}
			pending, err := d.betting.PlaceHostBet(txCtx, game.ID, *input.HostUserID, *input.BetAmount)
			if err != nil {
				return fmt.Errorf("place bet: %w", err)
			}
			game.HostBetTransferID = &pending.ID
			if err = d.betting.SettleHouseGame(txCtx, SettleHouseGameInput{
				GameID:                game.ID,
				HostUserID:            *input.HostUserID,
				HostPendingTransferID: pending.ID,
				BetAmount:             *input.BetAmount,
				UserResult:            hostResult,
			}); err != nil {
				return fmt.Errorf("settle house bet: %w", err)
			}
			game, err = d.adapter.Gaming().UpdateRpsGame(txCtx, game)
			if err != nil {
				return fmt.Errorf("save bet transfer id: %w", err)
			}
		}

		// Update cooldown timestamp on the requesting player.
		requester.LastHouseGameAt = &now
		if _, err = d.adapter.Gaming().UpdatePlayer(txCtx, requester); err != nil {
			return fmt.Errorf("update player cooldown: %w", err)
		}

		hostParticipant, guestParticipant := participants[0], participants[1]
		hostParticipant.Player = requester
		guestParticipant.Player = house

		gameResult = &RpsGameWithParticipants{
			RpsGame:               game,
			RequestingParticipant: hostParticipant,
			InvitedParticipant:    guestParticipant,
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	// Simulate the house "thinking" (outside the transaction).
	if d.houseThinkDelay > 0 {
		time.Sleep(d.houseThinkDelay)
	}

	result := &ChallengeHouseResult{
		Game:           gameResult,
		CooldownEndsAt: now.Add(HouseCooldown),
	}
	if input.BetAmount != nil && guestResult == models.RpsParticipantResultWin {
		msg := HouseWinsMessage
		result.HouseMessage = &msg
	}
	return result, nil
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
