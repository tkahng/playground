package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

// TestBettingInvariant_MultiGame_TotalSystemConservation runs six games covering
// every terminal path (host-win × 2, guest-win × 2, tie × 1, expire × 1) and
// asserts that the total of all user balances plus the escrow account equals the
// total points ever issued. No money is created or destroyed.
func TestBettingInvariant_MultiGame_TotalSystemConservation(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		const startBalance int64 = 1000
		const numPairs = 6
		const totalIssued int64 = startBalance * numPairs * 2 // 12 000

		type playerPair struct {
			host  *models.Player
			guest *models.Player
		}

		makePair := func(i int) playerPair {
			hEmail := fmt.Sprintf("inv_host%d@example.com", i)
			gEmail := fmt.Sprintf("inv_guest%d@example.com", i)
			hUser := mustCreateUser(t, ctx, adapter, hEmail)
			gUser := mustCreateUser(t, ctx, adapter, gEmail)
			h := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
				stores.WithPlayerEmail(hEmail),
				stores.WithUserID(hUser.ID),
			)
			g := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
				stores.WithPlayerEmail(gEmail),
				stores.WithUserID(gUser.ID),
			)
			mustFundPlayerWallet(t, ctx, adapter, ledger, h.UserID, startBalance)
			mustFundPlayerWallet(t, ctx, adapter, ledger, g.UserID, startBalance)
			return playerPair{h, g}
		}

		pairs := make([]playerPair, numPairs)
		for i := range pairs {
			pairs[i] = makePair(i)
		}

		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount escrow: %v", err)
		}
		escrowStart := escrow.Balance()
		if escrowStart != 0 {
			t.Fatalf("escrow account has unexpected pre-existing balance %d; test assumes a clean starting state", escrowStart)
		}

		betPtr := func(v int64) *int64 { return &v }

		// Game 0: host wins (100 pt bet) — Rock beats Scissors
		{
			p := pairs[0]
			g, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   p.host.ID,
				InvitedPlayerID:      p.guest.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      3600,
				BetAmount:            betPtr(100),
				HostUserID:           p.host.UserID,
			})
			if err != nil {
				t.Fatalf("game0 RequestGame: %v", err)
			}
			if _, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
				InvitedPlayerID: p.guest.ID,
				GameID:          g.RpsGame.ID,
				Status:          models.RpsGameStatusCompleted,
				Move:            models.RpsParticipantMoveScissors,
			}); err != nil {
				t.Fatalf("game0 Respond: %v", err)
			}
		}

		// Game 1: guest wins (100 pt bet) — Paper beats Rock
		{
			p := pairs[1]
			g, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   p.host.ID,
				InvitedPlayerID:      p.guest.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      3600,
				BetAmount:            betPtr(100),
				HostUserID:           p.host.UserID,
			})
			if err != nil {
				t.Fatalf("game1 RequestGame: %v", err)
			}
			if _, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
				InvitedPlayerID: p.guest.ID,
				GameID:          g.RpsGame.ID,
				Status:          models.RpsGameStatusCompleted,
				Move:            models.RpsParticipantMovePaper,
			}); err != nil {
				t.Fatalf("game1 Respond: %v", err)
			}
		}

		// Game 2: host wins (200 pt bet) — Rock beats Scissors
		{
			p := pairs[2]
			g, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   p.host.ID,
				InvitedPlayerID:      p.guest.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      3600,
				BetAmount:            betPtr(200),
				HostUserID:           p.host.UserID,
			})
			if err != nil {
				t.Fatalf("game2 RequestGame: %v", err)
			}
			if _, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
				InvitedPlayerID: p.guest.ID,
				GameID:          g.RpsGame.ID,
				Status:          models.RpsGameStatusCompleted,
				Move:            models.RpsParticipantMoveScissors,
			}); err != nil {
				t.Fatalf("game2 Respond: %v", err)
			}
		}

		// Game 3: guest wins (200 pt bet) — Paper beats Rock
		{
			p := pairs[3]
			g, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   p.host.ID,
				InvitedPlayerID:      p.guest.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      3600,
				BetAmount:            betPtr(200),
				HostUserID:           p.host.UserID,
			})
			if err != nil {
				t.Fatalf("game3 RequestGame: %v", err)
			}
			if _, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
				InvitedPlayerID: p.guest.ID,
				GameID:          g.RpsGame.ID,
				Status:          models.RpsGameStatusCompleted,
				Move:            models.RpsParticipantMovePaper,
			}); err != nil {
				t.Fatalf("game3 Respond: %v", err)
			}
		}

		// Game 4: tie (150 pt bet) — Rock vs Rock
		{
			p := pairs[4]
			g, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   p.host.ID,
				InvitedPlayerID:      p.guest.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      3600,
				BetAmount:            betPtr(150),
				HostUserID:           p.host.UserID,
			})
			if err != nil {
				t.Fatalf("game4 RequestGame: %v", err)
			}
			if _, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
				InvitedPlayerID: p.guest.ID,
				GameID:          g.RpsGame.ID,
				Status:          models.RpsGameStatusCompleted,
				Move:            models.RpsParticipantMoveRock,
			}); err != nil {
				t.Fatalf("game4 Respond: %v", err)
			}
		}

		// Game 5: expires (100 pt bet)
		{
			p := pairs[5]
			_, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   p.host.ID,
				InvitedPlayerID:      p.guest.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      1,
				BetAmount:            betPtr(100),
				HostUserID:           p.host.UserID,
			})
			if err != nil {
				t.Fatalf("game5 RequestGame: %v", err)
			}
		}
		// Game 5's expiry timer started at creation (before games 0–4 ran).
		// Even if games 0–4 were slow, the game is already expired by now.
		time.Sleep(2 * time.Second)
		if _, err := rpsService.ExpireGamesAndRefundBets(ctx); err != nil {
			t.Fatalf("ExpireGamesAndRefundBets: %v", err)
		}

		// --- Invariant assertions ---

		// 1. Total system conservation: sum(all user balances) + escrow == totalIssued
		var totalUserBalance int64
		for i, p := range pairs {
			hBal, err := ledger.GetUserBalance(ctx, *p.host.UserID)
			if err != nil {
				t.Fatalf("GetUserBalance host[%d]: %v", i, err)
			}
			gBal, err := ledger.GetUserBalance(ctx, *p.guest.UserID)
			if err != nil {
				t.Fatalf("GetUserBalance guest[%d]: %v", i, err)
			}
			totalUserBalance += hBal + gBal
		}

		escrow, err = ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("re-fetch escrow: %v", err)
		}
		totalAfter := totalUserBalance + escrow.Balance()
		if totalAfter != totalIssued {
			t.Errorf("total system balance = %d, want %d (money conservation violated)", totalAfter, totalIssued)
		}

		// 2. Escrow nets to zero.
		if escrow.Balance() != escrowStart {
			t.Errorf("escrow balance = %d, want %d (must return to start)", escrow.Balance(), escrowStart)
		}

		// 3. No pending transfers remain anywhere in the ledger.
		pendingCount, err := ledger.CountTransfers(ctx, &stores.LedgerTransferFilter{
			Statuses: []models.LedgerTransferStatus{models.LedgerTransferStatusPending},
		})
		if err != nil {
			t.Fatalf("CountTransfers pending: %v", err)
		}
		if pendingCount != 0 {
			t.Errorf("pending transfers remaining = %d, want 0", pendingCount)
		}
	})
}

// TestBettingInvariant_EscrowNetsZeroAfterEachGame verifies that the escrow
// account returns to its starting balance immediately after each individual game
// settles, not just in aggregate. This catches a leak in one game being masked
// by another game's accounting.
func TestBettingInvariant_EscrowNetsZeroAfterEachGame(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}
		escrowStart := escrow.Balance()
		if escrowStart != 0 {
			t.Fatalf("escrow account has unexpected pre-existing balance %d; test assumes a clean starting state", escrowStart)
		}

		assertEscrowAtStart := func(label string) {
			t.Helper()
			e, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
			if err != nil {
				t.Fatalf("%s: re-fetch escrow: %v", label, err)
			}
			if e.Balance() != escrowStart {
				t.Errorf("%s: escrow = %d, want %d", label, e.Balance(), escrowStart)
			}
		}

		betPtr := func(v int64) *int64 { return &v }

		makePlayerPair := func(tag string) (host, guest *models.Player) {
			hUser := mustCreateUser(t, ctx, adapter, tag+"_h@example.com")
			gUser := mustCreateUser(t, ctx, adapter, tag+"_g@example.com")
			h := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
				stores.WithPlayerEmail(tag+"_h@example.com"),
				stores.WithUserID(hUser.ID),
			)
			g := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
				stores.WithPlayerEmail(tag+"_g@example.com"),
				stores.WithUserID(gUser.ID),
			)
			mustFundPlayerWallet(t, ctx, adapter, ledger, h.UserID, 500)
			mustFundPlayerWallet(t, ctx, adapter, ledger, g.UserID, 500)
			return h, g
		}

		// Game 1: complete (host wins) — escrow must return to start immediately.
		{
			h, g := makePlayerPair("ezg1")
			game, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   h.ID,
				InvitedPlayerID:      g.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      3600,
				BetAmount:            betPtr(100),
				HostUserID:           h.UserID,
			})
			if err != nil {
				t.Fatalf("game1 RequestGame: %v", err)
			}
			if _, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
				InvitedPlayerID: g.ID,
				GameID:          game.RpsGame.ID,
				Status:          models.RpsGameStatusCompleted,
				Move:            models.RpsParticipantMoveScissors,
			}); err != nil {
				t.Fatalf("game1 Respond: %v", err)
			}
			assertEscrowAtStart("after game1 (host wins)")
		}

		// Game 2: cancelled — escrow must return to start immediately.
		{
			h, g := makePlayerPair("ezg2")
			game, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   h.ID,
				InvitedPlayerID:      g.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      3600,
				BetAmount:            betPtr(100),
				HostUserID:           h.UserID,
			})
			if err != nil {
				t.Fatalf("game2 RequestGame: %v", err)
			}
			if _, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
				InvitedPlayerID: g.ID,
				GameID:          game.RpsGame.ID,
				Status:          models.RpsGameStatusCancelled,
			}); err != nil {
				t.Fatalf("game2 Respond: %v", err)
			}
			assertEscrowAtStart("after game2 (cancelled)")
		}

		// Game 3: expired — escrow must return to start after sweep.
		{
			h, g := makePlayerPair("ezg3")
			_, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   h.ID,
				InvitedPlayerID:      g.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      1,
				BetAmount:            betPtr(100),
				HostUserID:           h.UserID,
			})
			if err != nil {
				t.Fatalf("game3 RequestGame: %v", err)
			}
			time.Sleep(2 * time.Second)
			if n, err := rpsService.ExpireGamesAndRefundBets(ctx); err != nil {
				t.Fatalf("game3 ExpireGamesAndRefundBets: %v", err)
			} else if n != 1 {
				t.Errorf("game3 ExpireGamesAndRefundBets processed = %d, want 1", n)
			}
			assertEscrowAtStart("after game3 (expired)")
		}
	})
}

// TestBettingInvariant_NoOrphanPendingAfterAllPaths verifies that after running
// all three terminal game paths (complete / cancel / expire), the ledger transfer
// table contains zero pending bet_escrow rows. This is a direct audit of the
// transfer table — stronger than checking balances alone.
func TestBettingInvariant_NoOrphanPendingAfterAllPaths(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		betPtr := func(v int64) *int64 { return &v }

		makePlayerPair := func(tag string) (host, guest *models.Player) {
			hUser := mustCreateUser(t, ctx, adapter, tag+"_h@example.com")
			gUser := mustCreateUser(t, ctx, adapter, tag+"_g@example.com")
			h := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
				stores.WithPlayerEmail(tag+"_h@example.com"),
				stores.WithUserID(hUser.ID),
			)
			g := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
				stores.WithPlayerEmail(tag+"_g@example.com"),
				stores.WithUserID(gUser.ID),
			)
			mustFundPlayerWallet(t, ctx, adapter, ledger, h.UserID, 500)
			mustFundPlayerWallet(t, ctx, adapter, ledger, g.UserID, 500)
			return h, g
		}

		// Path 1: complete (host wins).
		{
			h, g := makePlayerPair("orp1")
			game, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   h.ID,
				InvitedPlayerID:      g.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      3600,
				BetAmount:            betPtr(100),
				HostUserID:           h.UserID,
			})
			if err != nil {
				t.Fatalf("path1 RequestGame: %v", err)
			}
			if _, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
				InvitedPlayerID: g.ID,
				GameID:          game.RpsGame.ID,
				Status:          models.RpsGameStatusCompleted,
				Move:            models.RpsParticipantMoveScissors,
			}); err != nil {
				t.Fatalf("path1 Respond: %v", err)
			}
		}

		// Path 2: cancelled.
		{
			h, g := makePlayerPair("orp2")
			game, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   h.ID,
				InvitedPlayerID:      g.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      3600,
				BetAmount:            betPtr(100),
				HostUserID:           h.UserID,
			})
			if err != nil {
				t.Fatalf("path2 RequestGame: %v", err)
			}
			if _, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
				InvitedPlayerID: g.ID,
				GameID:          game.RpsGame.ID,
				Status:          models.RpsGameStatusCancelled,
			}); err != nil {
				t.Fatalf("path2 Respond: %v", err)
			}
		}

		// Path 3: expired.
		{
			h, g := makePlayerPair("orp3")
			_, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   h.ID,
				InvitedPlayerID:      g.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      1,
				BetAmount:            betPtr(100),
				HostUserID:           h.UserID,
			})
			if err != nil {
				t.Fatalf("path3 RequestGame: %v", err)
			}
			time.Sleep(2 * time.Second)
			if n, err := rpsService.ExpireGamesAndRefundBets(ctx); err != nil {
				t.Fatalf("path3 ExpireGamesAndRefundBets: %v", err)
			} else if n != 1 {
				t.Errorf("path3 ExpireGamesAndRefundBets processed = %d, want 1", n)
			}
		}

		// No pending bet_escrow transfers anywhere in the ledger.
		count, err := ledger.CountTransfers(ctx, &stores.LedgerTransferFilter{
			TransferCodes: []string{models.TransferCodeBetEscrow},
			Statuses:      []models.LedgerTransferStatus{models.LedgerTransferStatusPending},
		})
		if err != nil {
			t.Fatalf("CountTransfers: %v", err)
		}
		if count != 0 {
			t.Errorf("orphaned pending bet_escrow transfers = %d, want 0", count)
		}
	})
}
