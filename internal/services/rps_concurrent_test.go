package services

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

// TestRpsGame_ConcurrentGuestResponses_OnlyOneSucceeds verifies that when two goroutines
// simultaneously call RespondToGameRequest for the same game, exactly one succeeds and
// one fails. This relies on the FindRpsGameForUpdate row-level lock.
func TestRpsGame_ConcurrentGuestResponses_OnlyOneSucceeds(t *testing.T) {
	database.WithNewDatabase2(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		// Setup: create players and a game in a committed transaction.
		var guestPlayerID uuid.UUID
		var gameID uuid.UUID

		if err := adapter.RunInTxCtx(ctx, func(txCtx context.Context) error {
			hostUser := mustCreateUser(t, txCtx, adapter, "host_concurrent@example.com")
			guestUser := mustCreateUser(t, txCtx, adapter, "guest_concurrent@example.com")

			player1 := stores.MustCreatePlayer(t, txCtx, adapter.Gaming(),
				stores.WithPlayerEmail("host_concurrent@example.com"),
				stores.WithUserID(hostUser.ID),
			)
			player2 := stores.MustCreatePlayer(t, txCtx, adapter.Gaming(),
				stores.WithPlayerEmail("guest_concurrent@example.com"),
				stores.WithUserID(guestUser.ID),
			)
			guestPlayerID = player2.ID

			game, err := rpsService.RequestGame(txCtx, &RpsGameRequestInput{
				RequestingPlayerID:   player1.ID,
				InvitedPlayerID:      player2.ID,
				RequestingPlayerMove: models.RpsParticipantMovePaper,
				DurationSeconds:      60 * 60 * 24, // 1 day
			})
			if err != nil {
				return err
			}
			gameID = game.RpsGame.ID
			return nil
		}); err != nil {
			t.Fatalf("setup: %v", err)
		}

		// Race: two goroutines respond simultaneously.
		const numGoroutines = 2
		errCh := make(chan error, numGoroutines)
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := adapter.RunInTxCtx(ctx, func(txCtx context.Context) error {
					_, err := rpsService.RespondToGameRequest(txCtx, &GameRequestResponse{
						InvitedPlayerID: guestPlayerID,
						GameID:          gameID,
						Status:          models.RpsGameStatusCompleted,
						Move:            models.RpsParticipantMoveRock,
					})
					return err
				})
				errCh <- err
			}()
		}

		wg.Wait()
		close(errCh)

		var successes, failures int
		for err := range errCh {
			if err == nil {
				successes++
			} else {
				failures++
			}
		}

		if successes != 1 {
			t.Errorf("successes = %d, want exactly 1", successes)
		}
		if failures != 1 {
			t.Errorf("failures = %d, want exactly 1", failures)
		}

		// Verify final game state: exactly one completion.
		var finalGame *models.RpsGame
		if err := adapter.RunInTxCtx(ctx, func(txCtx context.Context) error {
			game, err := adapter.Gaming().FindRpsGame(txCtx, &stores.RpsGameFilter{
				Ids: []uuid.UUID{gameID},
			})
			if err != nil {
				return err
			}
			finalGame = game
			return nil
		}); err != nil {
			t.Fatalf("fetch final game: %v", err)
		}
		if finalGame == nil {
			t.Fatal("final game not found")
		}
		if finalGame.Status != models.RpsGameStatusCompleted {
			t.Errorf("final game status = %q, want %q", finalGame.Status, models.RpsGameStatusCompleted)
		}
	})
}

// TestRpsGame_ConcurrentExpiry_OnlyOneRefunds verifies that when two goroutines
// simultaneously call ExpireGamesAndRefundBets, the host bet is refunded exactly once
// (no double-refund). This relies on the re-fetch-with-lock inside ExpireGamesAndRefundBets.
func TestRpsGame_ConcurrentExpiry_OnlyOneRefunds(t *testing.T) {
	database.WithNewDatabase2(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		var hostUserID uuid.UUID
		const betAmount int64 = 100

		// Setup: fund host wallet and create an expired game with a bet.
		if err := adapter.RunInTxCtx(ctx, func(txCtx context.Context) error {
			hostUser := mustCreateUser(t, txCtx, adapter, "host_expiry@example.com")
			hostUserID = hostUser.ID
			mustFundWallet(t, txCtx, adapter, ledger, hostUserID, 500)

			player1 := stores.MustCreatePlayer(t, txCtx, adapter.Gaming(),
				stores.WithPlayerEmail("host_expiry@example.com"),
				stores.WithUserID(hostUserID),
			)
			player2 := stores.MustCreatePlayer(t, txCtx, adapter.Gaming(),
				stores.WithPlayerEmail("guest_expiry@example.com"),
			)

			betAmt := betAmount
			game, err := rpsService.RequestGame(txCtx, &RpsGameRequestInput{
				RequestingPlayerID:   player1.ID,
				InvitedPlayerID:      player2.ID,
				RequestingPlayerMove: models.RpsParticipantMovePaper,
				DurationSeconds:      1, // 1 second: will expire very soon
				BetAmount:            &betAmt,
				HostUserID:           &hostUserID,
			})
			if err != nil {
				return err
			}
			// Backdate the expiry so the game is already expired.
			game.RpsGame.ExpiresAt = time.Now().UTC().Add(-5 * time.Second)
			if _, err = adapter.Gaming().UpdateRpsGame(txCtx, game.RpsGame); err != nil {
				return err
			}
			return nil
		}); err != nil {
			t.Fatalf("setup: %v", err)
		}

		// Record host balance before expiry.
		hostBalBefore, err := ledger.GetUserBalance(ctx, hostUserID)
		if err != nil {
			t.Fatalf("GetUserBalance before: %v", err)
		}

		// Race: two goroutines call ExpireGamesAndRefundBets simultaneously.
		const numGoroutines = 2
		errCh := make(chan error, numGoroutines)
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := rpsService.ExpireGamesAndRefundBets(ctx)
				errCh <- err
			}()
		}

		wg.Wait()
		close(errCh)

		for err := range errCh {
			if err != nil {
				t.Errorf("ExpireGamesAndRefundBets: unexpected error: %v", err)
			}
		}

		// Host balance must be fully restored (refunded exactly once).
		hostBalAfter, err := ledger.GetUserBalance(ctx, hostUserID)
		if err != nil {
			t.Fatalf("GetUserBalance after: %v", err)
		}
		if hostBalAfter != hostBalBefore {
			t.Errorf("host balance after concurrent expiry = %d, want %d (refunded exactly once)", hostBalAfter, hostBalBefore)
		}
	})
}

// TestLedger_ConcurrentPendingTransfers_BalanceConsistency verifies that under concurrent
// load, the available-balance constraint is correctly enforced. With a 100-point wallet and
// 10 goroutines each requesting a 20-point hold, exactly 5 succeed and 5 fail.
func TestLedger_ConcurrentPendingTransfers_BalanceConsistency(t *testing.T) {
	database.WithNewDatabase2(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)

		userID := uuid.New()

		// Fund wallet in a committed transaction.
		if err := adapter.RunInTxCtx(ctx, func(txCtx context.Context) error {
			mustFundWallet(t, txCtx, adapter, ledger, userID, 100)
			return nil
		}); err != nil {
			t.Fatalf("fund wallet: %v", err)
		}

		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}

		// Launch 10 goroutines, each trying to place a 20-point hold.
		const numGoroutines = 10
		const holdAmount int64 = 20
		errCh := make(chan error, numGoroutines)
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := adapter.RunInTxCtx(ctx, func(txCtx context.Context) error {
					_, err := ledger.CreatePendingTransfer(txCtx, PostTransferInput{
						LedgerCode:      "POINTS",
						DebitAccountID:  wallet.ID,
						CreditAccountID: escrow.ID,
						Amount:          holdAmount,
						TransferCode:    models.TransferCodeBetEscrow,
					})
					return err
				})
				errCh <- err
			}()
		}

		wg.Wait()
		close(errCh)

		var successes, failures int
		for err := range errCh {
			if err == nil {
				successes++
			} else {
				failures++
			}
		}

		// 100 pts / 20 pts each = exactly 5 should succeed.
		if successes != 5 {
			t.Errorf("successes = %d, want 5", successes)
		}
		if failures != 5 {
			t.Errorf("failures = %d, want 5", failures)
		}

		// Re-fetch wallet and verify final state.
		wallet, err = ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("re-fetch wallet: %v", err)
		}
		if wallet.DebitsPending != 100 {
			t.Errorf("DebitsPending = %d, want 100", wallet.DebitsPending)
		}
		if wallet.AvailableBalance() != 0 {
			t.Errorf("AvailableBalance = %d, want 0", wallet.AvailableBalance())
		}
		if wallet.Balance() != 100 {
			t.Errorf("Balance = %d, want 100 (no posted debits yet)", wallet.Balance())
		}
	})
}
