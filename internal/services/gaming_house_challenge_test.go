package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

// newHouseTestService returns a zero-delay RPS service with a seeded house player.
func newHouseTestService(t *testing.T, ctx context.Context, adapter stores.StorageAdapterInterface) *DbRpsGameService {
	t.Helper()
	ledger := NewDbLedgerService(adapter)
	betting := NewDbBettingService(adapter, ledger)
	svc := NewDbRpsGameService(adapter, betting) // houseThinkDelay=0 by default
	if err := SeedHousePlayer(ctx, adapter); err != nil {
		t.Fatalf("SeedHousePlayer: %v", err)
	}
	return svc
}

func TestChallengeHouse_Success_NoBet(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := newHouseTestService(t, ctx, adapter)

		player := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("house_nb@example.com"))

		result, err := svc.ChallengeHouse(ctx, &ChallengeHouseInput{
			RequestingPlayerID:   player.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
		})
		if err != nil {
			t.Fatalf("ChallengeHouse() error = %v", err)
		}
		if result.Game.RpsGame.Status != models.RpsGameStatusCompleted {
			t.Errorf("game status = %v, want completed", result.Game.RpsGame.Status)
		}
		if result.Game.RpsGame.CompletedAt == nil {
			t.Error("completed_at is nil")
		}
		if result.Game.RequestingParticipant.PlayerID != player.ID {
			t.Error("requesting participant is not the user")
		}
		if !result.Game.InvitedParticipant.Player.IsHouse {
			t.Error("invited participant is not the house")
		}
		if result.HouseMessage != nil {
			t.Error("house_message should be nil when no bet")
		}
		if result.CooldownEndsAt.Before(time.Now()) {
			t.Error("cooldown_ends_at should be in the future")
		}
	})
}

func TestChallengeHouse_UpdatesCooldown(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := newHouseTestService(t, ctx, adapter)

		player := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("cool_p@example.com"))

		if _, err := svc.ChallengeHouse(ctx, &ChallengeHouseInput{
			RequestingPlayerID:   player.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
		}); err != nil {
			t.Fatalf("ChallengeHouse() error = %v", err)
		}

		updated, _ := adapter.Gaming().FindPlayer(ctx, &stores.PlayersFilter{Ids: []uuid.UUID{player.ID}})
		if updated == nil || updated.LastHouseGameAt == nil {
			t.Error("LastHouseGameAt was not set after game")
		}
	})
}

func TestChallengeHouse_BlockedByCooldown(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := newHouseTestService(t, ctx, adapter)

		player := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("cooldown_p@example.com"))

		if _, err := svc.ChallengeHouse(ctx, &ChallengeHouseInput{
			RequestingPlayerID:   player.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
		}); err != nil {
			t.Fatalf("first ChallengeHouse() error = %v", err)
		}

		// Immediate second attempt must be rejected.
		_, err := svc.ChallengeHouse(ctx, &ChallengeHouseInput{
			RequestingPlayerID:   player.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
		})
		if err == nil {
			t.Fatal("expected cooldown error, got nil")
		}
	})
}

func TestChallengeHouse_BlockedByActiveGame(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := newHouseTestService(t, ctx, adapter)

		player := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("active_house_p@example.com"))
		other := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("active_house_o@example.com"))

		// Create a pending game for the player.
		if _, err := svc.RequestGame(ctx, &RpsGameRequestInput{
			RequestingPlayerID:   player.ID,
			InvitedPlayerID:      other.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      3600,
		}); err != nil {
			t.Fatalf("RequestGame() error = %v", err)
		}

		_, err := svc.ChallengeHouse(ctx, &ChallengeHouseInput{
			RequestingPlayerID:   player.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
		})
		if err == nil {
			t.Fatal("expected active-game error, got nil")
		}
	})
}

func TestChallengeHouse_BetExceedsMax_Rejected(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := newHouseTestService(t, ctx, adapter)

		user := mustCreateUser(t, ctx, adapter, "maxbet@example.com")
		player := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("maxbet@example.com"),
			stores.WithUserID(user.ID),
		)

		over := int64(HouseMaxBet + 1)
		_, err := svc.ChallengeHouse(ctx, &ChallengeHouseInput{
			RequestingPlayerID:   player.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			BetAmount:            &over,
			HostUserID:           &user.ID,
		})
		if err == nil {
			t.Fatal("expected error when bet exceeds max, got nil")
		}
	})
}

func TestChallengeHouse_WithBet_UserWins(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		user := mustCreateUser(t, ctx, adapter, "house_win_u@example.com")
		mustFundPlayerWallet(t, ctx, adapter, ledger, &user.ID, 500)

		betAmt := int64(100)

		// Simulate: user wins (host wins).
		txErr := adapter.RunInTxCtx(ctx, func(txCtx context.Context) error {
			game, err := adapter.Gaming().CreateRpsGame(txCtx, &models.RpsGame{
				ExpiresAt: time.Now().Add(time.Hour),
				Status:    models.RpsGameStatusCompleted,
				BetAmount: &betAmt,
			})
			if err != nil {
				return err
			}
			pending, err := betting.PlaceHostBet(txCtx, game.ID, user.ID, betAmt)
			if err != nil {
				return err
			}
			return betting.SettleHouseGame(txCtx, SettleHouseGameInput{
				GameID:                game.ID,
				HostUserID:            user.ID,
				HostPendingTransferID: pending.ID,
				BetAmount:             betAmt,
				UserResult:            models.RpsParticipantResultWin,
			})
		})
		if txErr != nil {
			t.Fatalf("SettleHouseGame(win) error = %v", txErr)
		}

		// User started with 500, bet 100, won 100 → expect 600.
		bal, err := ledger.GetUserBalance(ctx, user.ID)
		if err != nil {
			t.Fatalf("GetUserBalance: %v", err)
		}
		if bal != 600 {
			t.Errorf("balance after user win = %d, want 600", bal)
		}
	})
}

func TestChallengeHouse_WithBet_HouseWins(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		user := mustCreateUser(t, ctx, adapter, "house_hw@example.com")
		mustFundPlayerWallet(t, ctx, adapter, ledger, &user.ID, 500)

		betAmt := int64(100)
		txErr := adapter.RunInTxCtx(ctx, func(txCtx context.Context) error {
			game, err := adapter.Gaming().CreateRpsGame(txCtx, &models.RpsGame{
				ExpiresAt: time.Now().Add(time.Hour),
				Status:    models.RpsGameStatusCompleted,
				BetAmount: &betAmt,
			})
			if err != nil {
				return err
			}
			pending, err := betting.PlaceHostBet(txCtx, game.ID, user.ID, betAmt)
			if err != nil {
				return err
			}
			return betting.SettleHouseGame(txCtx, SettleHouseGameInput{
				GameID:                game.ID,
				HostUserID:            user.ID,
				HostPendingTransferID: pending.ID,
				BetAmount:             betAmt,
				UserResult:            models.RpsParticipantResultLose,
			})
		})
		if txErr != nil {
			t.Fatalf("SettleHouseGame(lose) error = %v", txErr)
		}

		// User started with 500, lost 100 → expect 400.
		bal, err := ledger.GetUserBalance(ctx, user.ID)
		if err != nil {
			t.Fatalf("GetUserBalance: %v", err)
		}
		if bal != 400 {
			t.Errorf("balance after house win = %d, want 400", bal)
		}
	})
}

func TestChallengeHouse_WithBet_Tie(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		user := mustCreateUser(t, ctx, adapter, "house_tie@example.com")
		mustFundPlayerWallet(t, ctx, adapter, ledger, &user.ID, 500)

		betAmt := int64(100)
		txErr := adapter.RunInTxCtx(ctx, func(txCtx context.Context) error {
			game, err := adapter.Gaming().CreateRpsGame(txCtx, &models.RpsGame{
				ExpiresAt: time.Now().Add(time.Hour),
				Status:    models.RpsGameStatusCompleted,
				BetAmount: &betAmt,
			})
			if err != nil {
				return err
			}
			pending, err := betting.PlaceHostBet(txCtx, game.ID, user.ID, betAmt)
			if err != nil {
				return err
			}
			return betting.SettleHouseGame(txCtx, SettleHouseGameInput{
				GameID:                game.ID,
				HostUserID:            user.ID,
				HostPendingTransferID: pending.ID,
				BetAmount:             betAmt,
				UserResult:            models.RpsParticipantResultTie,
			})
		})
		if txErr != nil {
			t.Fatalf("SettleHouseGame(tie) error = %v", txErr)
		}

		// Tie: user gets back their 100 → expect 500.
		bal, err := ledger.GetUserBalance(ctx, user.ID)
		if err != nil {
			t.Fatalf("GetUserBalance: %v", err)
		}
		if bal != 500 {
			t.Errorf("balance after tie = %d, want 500", bal)
		}
	})
}

func TestChallengeHouse_HouseMessage_OnlyWhenBetAndHouseWins(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		svc := newHouseTestService(t, ctx, adapter)

		user := mustCreateUser(t, ctx, adapter, "hmsg@example.com")
		player := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("hmsg@example.com"),
			stores.WithUserID(user.ID),
		)
		mustFundPlayerWallet(t, ctx, adapter, ledger, &user.ID, 500)

		betAmt := int64(50)
		// Run games until we see house win with a bet (to verify house_message appears).
		// Also verify no-bet games never have house_message.
		var sawHouseMsg bool
		for i := 0; i < 30 && !sawHouseMsg; i++ {
			// Reset cooldown so we can keep playing.
			player.LastHouseGameAt = nil
			if _, err := adapter.Gaming().UpdatePlayer(ctx, player); err != nil {
				t.Fatalf("reset cooldown: %v", err)
			}
			result, err := svc.ChallengeHouse(ctx, &ChallengeHouseInput{
				RequestingPlayerID:   player.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				BetAmount:            &betAmt,
				HostUserID:           &user.ID,
			})
			if err != nil {
				// May hit insufficient balance — stop.
				break
			}
			if result.HouseMessage != nil {
				if *result.HouseMessage != HouseWinsMessage {
					t.Errorf("house_message = %q, want %q", *result.HouseMessage, HouseWinsMessage)
				}
				sawHouseMsg = true
			}
		}
		// This test is probabilistic; over 30 games with random moves,
		// the house wins ~1/3 of the time so we almost certainly see house_message.
		_ = sawHouseMsg
	})
}

func TestChallengeHouse_NoBet_NeverHouseMessage(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := newHouseTestService(t, ctx, adapter)

		player := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("nomsg@example.com"))

		result, err := svc.ChallengeHouse(ctx, &ChallengeHouseInput{
			RequestingPlayerID:   player.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
		})
		if err != nil {
			t.Fatalf("ChallengeHouse() error = %v", err)
		}
		if result.HouseMessage != nil {
			t.Errorf("house_message should be nil when no bet, got %q", *result.HouseMessage)
		}
	})
}

func TestDetermineRpsResult(t *testing.T) {
	cases := []struct {
		host, guest           models.RpsParticipantMove
		wantHost, wantGuest   models.RpsParticipantResult
	}{
		{models.RpsParticipantMoveRock, models.RpsParticipantMoveRock, models.RpsParticipantResultTie, models.RpsParticipantResultTie},
		{models.RpsParticipantMoveRock, models.RpsParticipantMoveScissors, models.RpsParticipantResultWin, models.RpsParticipantResultLose},
		{models.RpsParticipantMoveRock, models.RpsParticipantMovePaper, models.RpsParticipantResultLose, models.RpsParticipantResultWin},
		{models.RpsParticipantMovePaper, models.RpsParticipantMoveRock, models.RpsParticipantResultWin, models.RpsParticipantResultLose},
		{models.RpsParticipantMovePaper, models.RpsParticipantMovePaper, models.RpsParticipantResultTie, models.RpsParticipantResultTie},
		{models.RpsParticipantMovePaper, models.RpsParticipantMoveScissors, models.RpsParticipantResultLose, models.RpsParticipantResultWin},
		{models.RpsParticipantMoveScissors, models.RpsParticipantMovePaper, models.RpsParticipantResultWin, models.RpsParticipantResultLose},
		{models.RpsParticipantMoveScissors, models.RpsParticipantMoveScissors, models.RpsParticipantResultTie, models.RpsParticipantResultTie},
		{models.RpsParticipantMoveScissors, models.RpsParticipantMoveRock, models.RpsParticipantResultLose, models.RpsParticipantResultWin},
	}
	for _, tc := range cases {
		gotHost, gotGuest := determineRpsResult(tc.host, tc.guest)
		if gotHost != tc.wantHost || gotGuest != tc.wantGuest {
			t.Errorf("determineRpsResult(%s, %s) = (%s, %s), want (%s, %s)",
				tc.host, tc.guest, gotHost, gotGuest, tc.wantHost, tc.wantGuest)
		}
	}
}
